# Code-to-Spec Verification Audit

**Document under review**: `docs/auth-service-spec.md` v1.0  
**Source code path**: `auth-service/`  
**Audit date**: 2026-03-19

---

## 1. Accurate Sections

The following sections accurately reflect the implementation:

- **§1 System Overview** — correct framing of centralized IAM.
- **§2 Goals** — all five goals map to real implemented features.
- **§3 Non-Goals** — all items are correctly absent from the codebase.
- **§4 Architecture / Layers** — the `Handler → Service → Repository → Database` flow is accurate. The global middleware chain order (Recovery → Logger → CORS → Audit → Metrics) is correct per `router.go` lines 60–64.
- **§6 Cache Invalidation** — the four invalidation triggers (`AddRole`, `RemoveRole`, `AddPermission`, `RemovePermission`, role update/delete, permission update/delete) are all confirmed in the service layer.
- **§7 Service Account lifecycle** — the three endpoints (Create, List, Revoke), the required permission (`manage_service_accounts`), and the key format `sk_live_<prefix>_<secret>` are all accurate.
- **§8 Session Model table** — all six fields (`user_id`, `access_token`, `refresh_token`, `ip_address`, `user_agent`, `expired_at`) exist in the `UserSession` model.
- **§9 API Surface** — all seven groups are registered in `router.go`.
- **§10 Security Model** — bcrypt, HS256, rate limiting (5/min, burst 10), and audit logging rows are all confirmed.
- **§11 Observability** — all six metric names match `internal/utils/metrics/metrics.go` exactly.
- **§12 Deployment** — Docker multi-stage build, non-root user, Alpine image, and `auth-network` bridge are all confirmed in `Dockerfile` and `docker-compose.yml`.
- **§13 Scaling Considerations** — per-process cache limitation and GORM connection pooling description are accurate.

---

## 2. Minor Inaccuracies

### 2a — Refresh Token is not opaque

**Location**: §5 Authentication Model, Login Flow

> **Original**: "Refresh Token (opaque, long-lived): stored in `user_sessions`."

**Finding**: `GenerateRefreshToken` in `internal/utils/jwt.go` produces a **signed JWT** using HS256 and `JWT_SECRET`, with a `user_id` claim and a 7-day expiry. It is structurally identical to the access token — not an opaque random string.

> **Corrected**: "Refresh Token (JWT, 7-day expiry): signed with the same HS256 algorithm as the access token and stored in `user_sessions`."

---

### 2b — Access token lifetime not stated

**Location**: §5 Authentication Model, Login Flow

> **Original**: "Access Token (JWT, short-lived): carries `user_id` and expiry in Claims."

**Finding**: `GenerateToken` sets `exp` to `time.Now().Add(time.Minute * 15)`. The spec is correct but imprecise.

> **Corrected**: "Access Token (JWT, 15-minute expiry): carries `user_id` and `exp` in Claims, signed with HS256."

---

### 2c — Account lockout specifics missing from Security Model

**Location**: §10 Security Model

> **Original**: The row for account lockout is absent from the table.

**Finding**: `auth/service.go` implements a lockout after **5 consecutive failed login attempts**, locking the account for **15 minutes**.

> **Corrected**: Add this row to the Security Model table:

| Mechanism | Implementation |
| :--- | :--- |
| **Account Lockout** | 5 consecutive failed logins trigger a 15-minute account lock enforced at the service layer. |

---

### 2d — last_used_at update is asynchronous

**Location**: §8 Session Model

> **Original**: "The `last_used_at` field on Service Accounts is also updated on each API key authentication."

**Finding**: `service_account/service.go` line 106: `go s.repo.UpdateLastUsed(...)` — the update is dispatched in a **goroutine** and deliberately decoupled from the authentication response. It may lag or fail silently.

> **Corrected**: "The `last_used_at` field on Service Accounts is updated asynchronously after each successful API key authentication. Failures are silently ignored to avoid blocking the auth response."

---

### 2e — Permission cache uses Set semantics, not simple key lookup

**Location**: §6 Authorization Model, Enforcement

> **Original**: "Checks the in-memory cache for a `user_perms:<user_id>` key."

**Finding**: `permission_middleware.go` calls `csh.SIsMember(ctx, cacheKey, requiredPermission)` and populates via `csh.SAdd(...)`. The cache stores a **set** per user, not a serialized list under a single key. The middleware does an O(1) set-membership test, not a key fetch-and-scan.

> **Corrected**: "Checks the in-memory cache for a set stored at key `user_perms:<user_id>` using `SIsMember` for O(1) membership testing. On cache miss, queries the DB and repopulates the set with `SAdd`."

---

## 3. Missing Implementation Details

### 3a — RBAC Debug Endpoint not documented

The following route exists in `router.go` lines 175–180 but is absent from the §9 API Surface:

```
GET /api/v1/rbac/debug/user/:id  (requires manage_rbac_debug permission)
```

This endpoint (handled by `role.Handler.DebugUser`) resolves the full effective permission set for a given user and is useful for RBAC troubleshooting. It should appear in §9 API Surface.

---

### 3b — PermissionMiddleware requires user_id — API key actors are excluded

**Location**: §6 Enforcement / §7 Service-to-Service Authentication

`permission_middleware.go` line 15: `userID, exists := c.Get("user_id")`. The `user_id` context key is **only set for JWT Bearer auth** (see `auth_middleware.go` line 72). For API key actors, only `actor_type`, `actor_id`, and `service_account_id` are set — **`user_id` is never set**.

This means service accounts using `Authorization: ApiKey` will fail `PermissionMiddleware` with a 401, even for routes nominally protected by that middleware.

This is a real behavioral constraint that should be documented explicitly in §7.

---

### 3c — Swagger / OpenAPI endpoint not documented

`router.go` line 89 registers `GET /swagger/*any`. This is a live interactive API documentation endpoint and represents part of the service's public surface.

---

## 4. Incorrect Claims

### 4a — logout-all is not in the `auth` route group

**Location**: §5 Logout section

> **Original**: "`POST /api/v1/auth/logout-all` — deletes all sessions for the authenticated user."

**Finding**: In `router.go` line 112, `POST /auth/logout-all` is inside the `protected` group (requires `AuthMiddleware`), **not** inside the `authorizedAuth` group. The URL is correct, but the spec's implicit grouping with unauthenticated auth endpoints is misleading. This endpoint always requires a valid Bearer token.

> **Corrected**: Move this endpoint to a separate note: "`POST /api/v1/auth/logout-all` — requires authentication (Bearer token). Deletes all active sessions for the current user."

---

### 4b — Rate limiting applies to all 7 auth group routes, not just login/register

**Location**: §10 Security Model

> **Original**: "Token bucket on auth endpoints (5 req/min, burst 10)"

**Finding**: `router.go` line 95-103: `authorizedAuth.Use(middleware.RateLimitMiddleware(5.0/60.0, 10))` is applied to the entire `authorizedAuth` group, which includes: login, register, refresh, logout, forgot-password, reset-password, and **introspect**. The description is correct but the impression that only login/register is rate-limited is misleading.

> **Corrected**: "Token bucket (5 req/min, burst 10) applied to the entire `/api/v1/auth` group: login, register, refresh, logout, forgot-password, reset-password, and introspect."

---

## 5. Recommended Spec Fixes (Summary)

| Section | Fix Required | Priority |
| :--- | :--- | :--- |
| §5 – Refresh Token | Change "opaque" to "JWT, 7-day expiry, HS256" | High |
| §5 – Access Token | Add "15-minute expiry" | Low |
| §5 – logout-all | Clarify it requires AuthMiddleware (Bearer) | Medium |
| §6 – Enforcement | Update cache description to Set semantics (`SIsMember`/`SAdd`) | Medium |
| §7 – API Key actors | Document that API key actors cannot use PermissionMiddleware-protected routes | High |
| §8 – last_used_at | Note async update via goroutine | Low |
| §9 – API Surface | Add `GET /api/v1/rbac/debug/user/:id` group entry | Medium |
| §10 – Security | Add account lockout row (5 attempts, 15 min) | Medium |
| §10 – Rate Limiting | Clarify all 7 auth routes are rate-limited, not just login/register | Low |
