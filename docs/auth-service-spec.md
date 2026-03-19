# Auth Service Specification

> **Version**: 1.0  
> **Status**: Production  
> **Stack**: Go, Gin, PostgreSQL, JWT  

---

## 1. System Overview

Auth-Service is an internal **centralized identity and access management microservice**. It acts as the single source of truth for authentication and authorization across all internal company applications.

Rather than each application managing its own user accounts and permission logic, they delegate identity verification and access control to Auth-Service. This eliminates duplicated security logic and gives the security team a single system to audit and harden.

---

## 2. Goals

| Goal | Description |
| :--- | :--- |
| **Centralized Authentication** | Handle user login, token issuance, and session lifecycle for all internal apps. |
| **Role-Based Authorization** | Manage roles and permissions that consuming apps can enforce. |
| **Service Authentication** | Allow internal services to authenticate via API keys without human users. |
| **Auditability** | Record security-relevant events (logins, permission changes, etc.). |
| **Observability** | Expose operational metrics and health endpoints. |

---

## 3. Non-Goals

The following capabilities are **intentionally out of scope for v1**:

| Out of Scope | Reason |
| :--- | :--- |
| **OAuth2 Provider / OIDC** | Standards-based federation is complexity not needed for internal-only apps. |
| **SSO / SAML** | No cross-organization federation requirement exists at this stage. |
| **Social Login** | Internal services do not require third-party identity providers. |
| **LDAP / Active Directory Integration** | Company does not manage user identities in an LDAP directory. |
| **Multi-Tenant Organizations** | All applications share a single tenant. |
| **Self-Service User Registration** | User accounts are provisioned by administrators. |

---

## 4. Architecture

### Layers

Each request flows through a strict layered pipeline:

```
HTTP Request
     │
     ▼
[Middleware] (auth, permission, audit, rate limit, metrics)
     │
     ▼
[Handler] (request binding, validation, response)
     │
     ▼
[Service] (business logic, cache, token operations)
     │
     ▼
[Repository] (SQL via GORM)
     │
     ▼
[PostgreSQL]
```

### Middleware Chain (in order)

1. **Recovery** — panic recovery
2. **Logger** — Gin access logging
3. **CORS** — cross-origin header management
4. **Audit** — records mutations for audit log
5. **MetricsMiddleware** — tracks request count and latency
6. **RateLimit** — token bucket limiting on auth endpoints
7. **AuthMiddleware** — validates Bearer JWT or `ApiKey` header
8. **PermissionMiddleware** — checks RBAC permission on protected sub-routes

---

## 5. Authentication Model

### Login Flow

1. Client sends `POST /api/v1/auth/login` with `username` and `password`.
2. Service authenticates credentials using bcrypt comparison.
3. On success, two tokens are issued:
   - **Access Token** (JWT, short-lived): carries `user_id` and expiry in Claims.
   - **Refresh Token** (opaque, long-lived): stored in `user_sessions`.
4. Session record stores IP address and User-Agent for tracking.

### Token Validation

Consuming applications validate tokens either:
- **Locally**: by verifying the JWT signature with the shared `JWT_SECRET`.
- **Via introspection**: `POST /api/v1/auth/introspect` — the service checks both signature validity **and** session existence (prevents post-logout token reuse).

### Refresh Flow

`POST /api/v1/auth/refresh` accepts a refresh token and returns a new access token. The old session remains valid until explicit logout.

### Logout

- `POST /api/v1/auth/logout` — deletes the specific session by refresh token.
- `POST /api/v1/auth/logout-all` — deletes **all** sessions for the authenticated user.

---

## 6. Authorization Model (RBAC)

### Concepts

- **User** — an identity that can be assigned one or more Roles.
- **Role** — a named group that holds a collection of Permissions.
- **Permission** — a string identifier (e.g., `manage_users`, `view_audit_logs`).

### Data Model

```
users ──< user_roles >── roles ──< role_permissions >── permissions
```

### Enforcement

The `PermissionMiddleware` is applied per route group. It:
1. Reads `user_id` from the request context (set by `AuthMiddleware`).
2. Checks the in-memory cache for a `user_perms:<user_id>` key.
3. On cache miss, queries the DB and writes the result to cache.
4. Checks whether the required permission string is in the resolved set.

### Cache Invalidation

The permission cache is invalidated automatically when:
- A user's roles change (`AddRole`, `RemoveRole`)
- A role's permissions change (`AddPermission`, `RemovePermission`)
- A role or permission is updated or deleted

---

## 7. Service-to-Service Authentication

Internal services that need to call the API without a human user authenticate using **API Keys** (Service Accounts).

### Key Format

```
sk_live_<8-char-prefix>_<32-char-secret>
```

The raw key is returned **only once** at creation time. Only the prefix and a bcrypt hash of the secret are stored.

### Authentication Header

```
Authorization: ApiKey sk_live_9f2ab4cd_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

`AuthMiddleware` detects the `ApiKey` scheme, looks up the service account by prefix, verifies the secret with bcrypt, and sets `actor_type=service` in the request context.

### API Key Lifecycle

| Endpoint | Action |
| :--- | :--- |
| `POST /api/v1/service-accounts` | Create a new service account (returns raw key once) |
| `GET /api/v1/service-accounts` | List service accounts |
| `POST /api/v1/service-accounts/:id/revoke` | Permanently revoke a key |

All endpoints require the `manage_service_accounts` permission.

---

## 8. Session Model

Each successful login creates a `user_sessions` record:

| Field | Description |
| :--- | :--- |
| `user_id` | Owner of the session |
| `access_token` | Issued JWT (for introspection lookup) |
| `refresh_token` | Opaque refresh token |
| `ip_address` | Client IP at login time (IPv4/IPv6) |
| `user_agent` | Client User-Agent header |
| `expired_at` | Expiry timestamp |

Sessions are cleaned up on logout. The `last_used_at` field on Service Accounts is also updated on each API key authentication.

---

## 9. API Surface

| Group | Base Path | Purpose |
| :--- | :--- | :--- |
| **Auth** | `/api/v1/auth` | Login, logout, refresh, introspect, password reset |
| **Self** | `/api/v1/me` | Current user profile, own roles, own permissions |
| **Users** | `/api/v1/users` | Admin user management with RBAC role assignment |
| **Roles** | `/api/v1/roles` | Role CRUD and permission assignment |
| **Permissions** | `/api/v1/permissions` | Permission CRUD |
| **Service Accounts** | `/api/v1/service-accounts` | API key lifecycle management |
| **Menus** | `/api/v1/menus` | Frontend navigation menu management |
| **Audit Logs** | `/api/v1/audit-logs` | Queryable security event log |

All routes under `/api/v1/` require a valid `Authorization` header except `/auth/login`, `/auth/register`, `/auth/refresh`, `/auth/forgot-password`, `/auth/reset-password`, and `/auth/introspect`.

---

## 10. Security Model

| Mechanism | Implementation |
| :--- | :--- |
| **Password Hashing** | bcrypt with default cost |
| **JWT Signing** | HMAC-SHA256 using `JWT_SECRET` env var |
| **JWT Validation** | Signature and expiry enforced on every request |
| **Refresh Token** | Stored in DB; deleted on logout |
| **API Key Hashing** | bcrypt; raw key returned only at creation |
| **Rate Limiting** | Token bucket on auth endpoints (5 req/min, burst 10) |
| **Password Reset Token** | Never returned in API response; must be delivered out-of-band |
| **Audit Logging** | Mutations recorded with user, IP, method, and path |

---

## 11. Observability

### Prometheus Metrics

Exposed at `GET /metrics` (standard Prometheus scrape endpoint).

| Metric | Type | Description |
| :--- | :--- | :--- |
| `login_attempt_total` | Counter | All login attempts |
| `login_success_total` | Counter | Successful logins |
| `login_failure_total` | Counter | Failed logins (also increments on 401/403) |
| `token_refresh_total` | Counter | Successful token refreshes |
| `auth_request_total` | CounterVec | All requests, labelled by method, path, status |
| `auth_request_duration_seconds` | HistogramVec | Request latencies, labelled by method and path |

### Health Check

`GET /health` returns database connectivity status. Used as a liveness probe.

### Audit Log

All security-sensitive mutations are stored in the `audit_logs` table and queryable via `GET /api/v1/audit-logs` (requires appropriate permission).

---

## 12. Deployment Model

The service is containerized using Docker:

- **Dockerfile**: Multi-stage build. Final image is Alpine-based with a non-root user.
- **Docker Compose**: Orchestrates the service alongside PostgreSQL. An `auth-network` bridge network isolates traffic.
- **Migrations**: SQL files in `migrations/` are applied automatically via the PostgreSQL `docker-entrypoint-initdb.d` mount on first run.
- **Configuration**: All runtime configuration is injected via environment variables using a `.env` file.

---

## 13. Scaling Considerations

The current architecture is designed for low-to-medium traffic internal loads:

- **Stateless application layer**: The service itself holds no in-process user state beyond the permission cache. Multiple instances can be run behind a load balancer.
- **In-memory cache limitation**: The permission cache is per-process. If multiple instances run simultaneously, cache invalidation events only apply to the instance that received the write request, potentially causing brief inconsistency. A Redis-backed distributed cache would resolve this.
- **Database**: PostgreSQL is a single-node setup. Connection pooling is handled by GORM's built-in pool.

---

## 14. Future Considerations (Non-V1)

These are potential improvements for future iterations, with no current commitment:

- **Redis distributed cache**: Shared permission cache across multiple instances.
- **OAuth2 / OIDC Provider**: If the company requires integration with third-party tools.
- **Token rotation**: Implement refresh token rotation for stronger session security.
- **Webhook events**: Emit events (user created, role changed) for downstream consumers.
- **Admin UI**: A dedicated management dashboard for RBAC administration.
- **LDAP sync**: Import users from an Active Directory or LDAP source.
