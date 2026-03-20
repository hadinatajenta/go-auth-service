# Auth-Service Integration Guide

## 1. Overview

**Auth-Service** is the centralized identity and access management (IAM) provider for internal company applications. 

Rather than each service managing its own user database and permission logic, they delegate these responsibilities to Auth-Service. This ensures a single source of truth for user identities and a unified audit log for security events.

---

## 2. Authentication

Auth-Service uses **stateless JWT tokens** for user authentication.

### How to Verify a User

When your service receives a request with an `Authorization: Bearer <token>` header, follow these steps:

1.  **Validate JWT Signature**: Use the shared `JWT_SECRET` environment variable to verify the token's HMAC-SHA256 signature.
2.  **Check Expiry**: Ensure the `exp` claim is in the future.
3.  **Extract Identity**: Extract the `user_id` claim. This is the unique identifier for the user in the system.
4.  **Optional Introspection**: For high-value operations (e.g., financial transactions), call the **Introspection Endpoint** to ensure the session hasn't been revoked:
    ```http
    POST /api/v1/auth/introspect
    Content-Type: application/json

    { "token": "<access_token>" }
    ```
    If `active` is `false`, reject the request.

---

## 3. Authorization (RBAC)

**Important**: Access tokens do NOT contain roles or permissions. This prevents tokens from becoming stale when a user's permissions change.

### How to Resolve Permissions

To verify if a user is authorized to perform an action, follow this workflow:

1.  **Extract user_id** from the validated JWT.
2.  **Fetch Permissions**: Call the "Self Permissions" endpoint using the user's own token:
    ```http
    GET /api/v1/me/permissions
    Authorization: Bearer <user_access_token>
    ```
3.  **Local Check**: Verify if the required permission string (e.g., `view_reports`) exists in the returned list.
4.  **Local Caching**: Store this list in your service's memory or a local Redis.

---

## 4. Service-to-Service Authentication

When your service needs to call another internal service (or Auth-Service itself) without a human user, use a **Service Account API Key**.

### Authorization Header
```http
Authorization: ApiKey sk_live_<prefix>_<secret>
```

### API Key Lifecycle
1.  **Create**: Use `POST /api/v1/service-accounts` to generate a key. The raw key is returned **only once**.
2.  **List**: Use `GET /api/v1/service-accounts` to see active accounts.
3.  **Revoke**: Use `POST /api/v1/service-accounts/:id/revoke` to immediately invalidate a key.

---

## 5. Recommended Integration Pattern

The standard request flow for a microservice integrating with Auth-Service is as follows:

1.  **User** sends request with JWT to **Service A**.
2.  **Service A** validates the JWT locally (Decentralized Auth).
3.  **Service A** checks its local cache for the user's permissions.
4.  If cache miss: **Service A** calls **Auth-Service** `/me/permissions` (Centralized AuthZ) and updates its cache.
5.  **Service A** approves/denies the request.
6.  If approved, **Service A** performs the action.

---

## 6. Caching Strategy

To maintain high performance and reduce load on Auth-Service, you **must** cache permissions locally.

*   **TTL**: Recommended Time-to-Live (TTL) is **15 minutes**, matching the default access token lifetime.
*   **Invalidation**: Since permissions are cached, there may be a 15-minute lag if a user's roles are revoked. If your use case requires real-time enforcement, reduce the TTL or call `/me/permissions` for every request.
*   **Scope**: Cache permissions per `user_id`.

---

## 7. Security Considerations

*   **Never Trust Frontend Authorization**: The frontend may hide buttons based on permissions, but the backend **must** always perform a hard check.
*   **Always Validate Signature**: Never use `jwt.ParseUnverified`. Always verify the signature using the shared `JWT_SECRET`.
*   **Secure API Keys**: Treat Service Account keys like passwords. Never log them or check them into source control.
*   **Enforce RBAC on the Backend**: Every sensitive API endpoint in your service should map to a specific permission string that is checked against the user's permission set.
