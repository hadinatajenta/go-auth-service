# Auth Service API Documentation

Comprehensive reference for all Auth-Service API endpoints.

## Base URL

`http://localhost:8080/api/v1`

## Authentication Methods

### 1. JWT Bearer Token (User Authentication)

For user-initiated requests, include a Bearer token in the Authorization header:

```http
Authorization: Bearer <access_token>
```

**Token Format**: JWT (JSON Web Token)
- **Expiry**: 15 minutes
- **Algorithm**: HS256 (HMAC-SHA256)
- **Claims**: `user_id`, `exp`, `iat`

**Obtaining a Token**: Use `/auth/login` endpoint

### 2. API Key (Service-to-Service Authentication)

For machine-to-machine authentication, include an API Key:

```http
Authorization: ApiKey sk_live_<prefix>_<secret>
```

**Key Format**: `sk_live_<8-char-prefix>_<32-char-secret>`
**Obtaining a Key**: Use `/service-accounts` endpoint (Admin only)

---

## Standard Response Format

All API responses follow a consistent format:

### Success Response (2xx)
```json
{
  "status": "success",
  "message": "Operation completed successfully",
  "data": {
    // Response payload varies by endpoint
  }
}
```

### Error Response (4xx, 5xx)
```json
{
  "status": "error",
  "message": "Human-readable error message",
  "error": "machine_readable_error_code",
  "timestamp": "2026-03-30T10:30:00Z"
}
```

### Pagination Response
```json
{
  "status": "success",
  "message": "Data retrieved",
  "data": {
    "items": [...],
    "total": 42,
    "page": 1,
    "per_page": 10,
    "total_pages": 5
  }
}
```

---

## Public Endpoints (No Authentication)

### Health Check

- **URL**: `/health`
- **Method**: `GET`
- **Auth**: None
- **Response**: 
  ```json
  {
    "message": "Server is healthy",
    "data": {
      "status": "up"
    }
  }
  ```

### Login

- **URL**: `/auth/login`
- **Method**: `POST`
- **Auth**: None
- **Request Body**:
  ```json
  {
    "username": "john_doe",
    "password": "secure_password"
  }
  ```
- **Response** (200):
  ```json
  {
    "status": "success",
    "message": "Login successful",
    "data": {
      "access_token": "eyJhbGciOiJIUzI1NiIs...",
      "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
      "user": {
        "id": "uuid",
        "username": "john_doe",
        "email": "john@example.com",
        "first_name": "John",
        "last_name": "Doe"
      }
    }
  }
  ```
- **Error Cases**:
  - `401 Unauthorized`: Invalid credentials
  - `423 Locked`: Account locked due to failed login attempts (15-minute lockout)

### Register

- **URL**: `/auth/register`
- **Method**: `POST`
- **Auth**: None
- **Request Body**:
  ```json
  {
    "username": "new_user",
    "email": "user@example.com",
    "password": "strong_password",
    "first_name": "Jane",
    "last_name": "Doe"
  }
  ```
- **Response** (201): User created with tokens (same as login response)
- **Error Cases**:
  - `400 Bad Request`: Validation failed (username/email already exists)
  - `422 Unprocessable Entity`: Password too weak

### Refresh Token

- **URL**: `/auth/refresh`
- **Method**: `POST`
- **Auth**: None
- **Request Body**:
  ```json
  {
    "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
  }
  ```
- **Response** (200):
  ```json
  {
    "status": "success",
    "message": "Token refreshed",
    "data": {
      "access_token": "eyJhbGciOiJIUzI1NiIs...",
      "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
    }
  }
  ```
- **Error Cases**:
  - `401 Unauthorized`: Invalid or expired refresh token
  - `410 Gone`: Session revoked (logout was called)

### Forgot Password

- **URL**: `/auth/forgot-password`
- **Method**: `POST`
- **Auth**: None
- **Request Body**:
  ```json
  {
    "email": "user@example.com"
  }
  ```
- **Response** (200): Reset token sent via email
- **Note**: Reset token is never returned in API response; user receives it via email

### Reset Password

- **URL**: `/auth/reset-password`
- **Method**: `POST`
- **Auth**: None
- **Request Body**:
  ```json
  {
    "reset_token": "token_from_email",
    "new_password": "new_secure_password"
  }
  ```
- **Response** (200): Password changed successfully

### Token Introspection (RFC 7662)

- **URL**: `/auth/introspect`
- **Method**: `POST`
- **Auth**: None (intended for internal service use with API Key)
- **Request Body**:
  ```json
  {
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
  ```
- **Response** (200):
  ```json
  {
    "status": "success",
    "data": {
      "active": true,
      "user_id": "uuid",
      "exp": 1700000000,
      "iat": 1699990000,
      "username": "john_doe"
    }
  }
  ```
- **Inactive Token Response** (200):
  ```json
  {
    "status": "success",
    "data": {
      "active": false
    }
  }
  ```
- **Use Case**: Internal services verify token validity without storing JWT_SECRET

---

## Authenticated Endpoints - Current User

These endpoints return information about the currently authenticated user.

### Get Current User Profile

- **URL**: `/me`
- **Method**: `GET`
- **Auth**: Bearer Token required
- **Response** (200):
  ```json
  {
    "status": "success",
    "data": {
      "id": "uuid",
      "username": "john_doe",
      "email": "john@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "created_at": "2026-01-15T10:30:00Z"
    }
  }
  ```

### Get Current User Permissions

- **URL**: `/me/permissions`
- **Method**: `GET`
- **Auth**: Bearer Token required
- **Query Parameters**:
  - `page` (optional): Page number (default: 1)
  - `per_page` (optional): Items per page (default: 10)
- **Response** (200): Paginated list of permission strings
  ```json
  {
    "status": "success",
    "data": {
      "items": ["manage_users", "view_reports", "edit_roles"],
      "total": 12,
      "page": 1
    }
  }
  ```

### Get Current User Roles

- **URL**: `/me/roles`
- **Method**: `GET`
- **Auth**: Bearer Token required
- **Response** (200):
  ```json
  {
    "status": "success",
    "data": [
      {
        "id": "uuid",
        "name": "Administrator",
        "description": "Full system access"
      },
      {
        "id": "uuid",
        "name": "User Manager",
        "description": "Manage user accounts"
      }
    ]
  }
  ```

### Change Password

- **URL**: `/me/change-password`
- **Method**: `POST`
- **Auth**: Bearer Token required
- **Request Body**:
  ```json
  {
    "current_password": "old_password",
    "new_password": "new_secure_password"
  }
  ```
- **Response** (200): Password changed
- **Error Cases**:
  - `400 Bad Request`: Current password incorrect
  - `422 Unprocessable Entity`: New password too weak

### Logout (Current Session)

- **URL**: `/auth/logout`
- **Method**: `POST`
- **Auth**: Bearer Token required
- **Request Body**: `{}`
- **Response** (200): Session terminated
- **Note**: Only current session is terminated; other sessions remain active

### Logout All Sessions

- **URL**: `/auth/logout-all`
- **Method**: `POST`
- **Auth**: Bearer Token required
- **Request Body**: `{}`
- **Response** (200): All sessions terminated
- **Effect**: All refresh tokens are invalidated

---

## Admin Endpoints - User Management

Requires `manage_users` permission.

### List All Users

- **URL**: `/users`
- **Method**: `GET`
- **Auth**: Bearer Token + `manage_users` permission
- **Query Parameters**:
  - `page` (optional): Page number (default: 1)
  - `per_page` (optional): Items per page (default: 10)
  - `search` (optional): Search by username, email, or name
- **Response** (200): Paginated user list

### Create User

- **URL**: `/users`
- **Method**: `POST`
- **Auth**: Bearer Token + `manage_users` permission
- **Request Body**:
  ```json
  {
    "username": "newuser",
    "email": "newuser@example.com",
    "password": "initial_password",
    "first_name": "New",
    "last_name": "User"
  }
  ```
- **Response** (201): User created

### Get User by ID

- **URL**: `/users/:id`
- **Method**: `GET`
- **Auth**: Bearer Token + `manage_users` permission
- **Response** (200): User details

### Update User

- **URL**: `/users/:id`
- **Method**: `PUT`
- **Auth**: Bearer Token + `manage_users` permission
- **Request Body**: (all fields optional)
  ```json
  {
    "first_name": "Updated",
    "last_name": "Name",
    "email": "newemail@example.com"
  }
  ```
- **Response** (200): User updated

### Delete User

- **URL**: `/users/:id`
- **Method**: `DELETE`
- **Auth**: Bearer Token + `manage_users` permission
- **Response** (204): User deleted and all sessions revoked

### List User Roles

- **URL**: `/users/:id/roles`
- **Method**: `GET`
- **Auth**: Bearer Token + `manage_users` permission
- **Response** (200): Roles assigned to user

### Assign Role to User

- **URL**: `/users/:id/roles`
- **Method**: `POST`
- **Auth**: Bearer Token + `manage_users` permission
- **Request Body**:
  ```json
  {
    "role_id": "uuid"
  }
  ```
- **Response** (201): Role assigned (triggers permission cache invalidation)

### Remove Role from User

- **URL**: `/users/:id/roles/:role_id`
- **Method**: `DELETE`
- **Auth**: Bearer Token + `manage_users` permission
- **Response** (204): Role removed (triggers permission cache invalidation)

---

## Admin Endpoints - Role Management

Requires `manage_roles` permission.

### List All Roles

- **URL**: `/roles`
- **Method**: `GET`
- **Auth**: Bearer Token + `manage_roles` permission
- **Query Parameters**:
  - `page` (optional): Page number (default: 1)
  - `per_page` (optional): Items per page (default: 10)
- **Response** (200): Paginated role list

### Create Role

- **URL**: `/roles`
- **Method**: `POST`
- **Auth**: Bearer Token + `manage_roles` permission
- **Request Body**:
  ```json
  {
    "name": "Editor",
    "description": "Can edit content"
  }
  ```
- **Response** (201): Role created

### Get Role by ID

- **URL**: `/roles/:id`
- **Method**: `GET`
- **Auth**: Bearer Token + `manage_roles` permission
- **Response** (200): Role details with permission count

### Update Role

- **URL**: `/roles/:id`
- **Method**: `PUT`
- **Auth**: Bearer Token + `manage_roles` permission
- **Request Body**:
  ```json
  {
    "name": "Updated Role",
    "description": "Updated description"
  }
  ```
- **Response** (200): Role updated (triggers cache invalidation)

### Delete Role

- **URL**: `/roles/:id`
- **Method**: `DELETE`
- **Auth**: Bearer Token + `manage_roles` permission
- **Response** (204): Role deleted (triggers cache invalidation for all users)

### List Role Permissions

- **URL**: `/roles/:id/permissions`
- **Method**: `GET`
- **Auth**: Bearer Token + `manage_roles` permission
- **Response** (200): Permissions in role

### Add Permission to Role

- **URL**: `/roles/:id/permissions`
- **Method**: `POST`
- **Auth**: Bearer Token + `manage_roles` permission
- **Request Body**:
  ```json
  {
    "permission_id": "uuid"
  }
  ```
- **Response** (201): Permission added (triggers cache invalidation)

### Remove Permission from Role

- **URL**: `/roles/:id/permissions/:permission_id`
- **Method**: `DELETE`
- **Auth**: Bearer Token + `manage_roles` permission
- **Response** (204): Permission removed (triggers cache invalidation)

### List Users in Role

- **URL**: `/roles/:id/users`
- **Method**: `GET`
- **Auth**: Bearer Token + `manage_roles` permission
- **Query Parameters**:
  - `page` (optional): Page number (default: 1)
  - `per_page` (optional): Items per page (default: 10)
- **Response** (200): Paginated list of users with this role

---

## Admin Endpoints - Permission Management

Requires `manage_permissions` permission.

### List All Permissions

- **URL**: `/permissions`
- **Method**: `GET`
- **Auth**: Bearer Token + `manage_permissions` permission
- **Query Parameters**:
  - `page` (optional): Page number (default: 1)
  - `per_page` (optional): Items per page (default: 10)
- **Response** (200): Paginated permission list

### Create Permission

- **URL**: `/permissions`
- **Method**: `POST`
- **Auth**: Bearer Token + `manage_permissions` permission
- **Request Body**:
  ```json
  {
    "name": "view_reports",
    "description": "Can view analytics reports",
    "category": "reports"
  }
  ```
- **Response** (201): Permission created

### Get Permission by ID

- **URL**: `/permissions/:id`
- **Method**: `GET`
- **Auth**: Bearer Token + `manage_permissions` permission
- **Response** (200): Permission details

### Update Permission

- **URL**: `/permissions/:id`
- **Method**: `PUT`
- **Auth**: Bearer Token + `manage_permissions` permission
- **Request Body**:
  ```json
  {
    "name": "updated_name",
    "description": "Updated description",
    "category": "updated_category"
  }
  ```
- **Response** (200): Permission updated (triggers cache invalidation)

### Delete Permission

- **URL**: `/permissions/:id`
- **Method**: `DELETE`
- **Auth**: Bearer Token + `manage_permissions` permission
- **Response** (204): Permission deleted (triggers cache invalidation)

### List Permissions Grouped by Category

- **URL**: `/permissions/grouped`
- **Method**: `GET`
- **Auth**: Bearer Token + `manage_permissions` permission
- **Response** (200):
  ```json
  {
    "status": "success",
    "data": {
      "users": ["manage_users", "create_users", "delete_users"],
      "reports": ["view_reports", "export_reports"],
      "audit": ["view_audit_logs", "export_audit_logs"]
    }
  }
  ```

---

## Admin Endpoints - Service Account Management

Requires `manage_service_accounts` permission.

### Create Service Account (Generate API Key)

- **URL**: `/service-accounts`
- **Method**: `POST`
- **Auth**: Bearer Token + `manage_service_accounts` permission
- **Request Body**:
  ```json
  {
    "name": "payment-service",
    "description": "API key for payment service"
  }
  ```
- **Response** (201):
  ```json
  {
    "status": "success",
    "message": "Service account created",
    "data": {
      "id": "uuid",
      "name": "payment-service",
      "description": "API key for payment service",
      "api_key": "sk_live_abc12345_...",
      "created_at": "2026-03-30T10:30:00Z"
    }
  }
  ```
- **Important**: Raw API key is returned **only once**. Store it securely.

### List Service Accounts

- **URL**: `/service-accounts`
- **Method**: `GET`
- **Auth**: Bearer Token + `manage_service_accounts` permission
- **Query Parameters**:
  - `page` (optional): Page number (default: 1)
  - `per_page` (optional): Items per page (default: 10)
- **Response** (200): List of service accounts (raw key **not** included)

### Revoke Service Account

- **URL**: `/service-accounts/:id/revoke`
- **Method**: `POST`
- **Auth**: Bearer Token + `manage_service_accounts` permission
- **Request Body**: `{}`
- **Response** (200): API key immediately invalidated
- **Effect**: All requests using this key will be rejected

---

## Admin Endpoints - Menu Management

Requires `manage_menus` permission (for create/update/delete).

### List Available Menus (User)

- **URL**: `/menus/allowed`
- **Method**: `GET`
- **Auth**: Bearer Token
- **Response** (200): List of menus accessible to current user

### Get Menu Tree (User)

- **URL**: `/menus/tree`
- **Method**: `GET`
- **Auth**: Bearer Token
- **Response** (200): Hierarchical menu structure for current user

### Create Menu

- **URL**: `/menus`
- **Method**: `POST`
- **Auth**: Bearer Token + `manage_menus` permission
- **Request Body**:
  ```json
  {
    "name": "Dashboard",
    "label": "Dashboard",
    "path": "/dashboard",
    "icon": "dashboard",
    "parent_id": null,
    "order": 1,
    "required_permission": "view_dashboard"
  }
  ```
- **Response** (201): Menu created

### Get Menu by ID

- **URL**: `/menus/:id`
- **Method**: `GET`
- **Auth**: Bearer Token + `manage_menus` permission
- **Response** (200): Menu details

### Update Menu

- **URL**: `/menus/:id`
- **Method**: `PUT`
- **Auth**: Bearer Token + `manage_menus` permission
- **Request Body**: (all fields optional)
  ```json
  {
    "name": "Updated Name",
    "label": "Updated Label",
    "path": "/updated-path"
  }
  ```
- **Response** (200): Menu updated

### Delete Menu

- **URL**: `/menus/:id`
- **Method**: `DELETE`
- **Auth**: Bearer Token + `manage_menus` permission
- **Response** (204): Menu deleted

---

## Admin Endpoints - Audit Logs

Requires `view_audit_logs` permission.

### Query Audit Logs

- **URL**: `/audit-logs`
- **Method**: `GET`
- **Auth**: Bearer Token + `view_audit_logs` permission
- **Query Parameters**:
  - `page` (optional): Page number (default: 1)
  - `per_page` (optional): Items per page (default: 10)
  - `user_id` (optional): Filter by user
  - `action` (optional): Filter by action (CREATE, UPDATE, DELETE, LOGIN, LOGOUT)
  - `resource` (optional): Filter by resource type
  - `start_date` (optional): ISO 8601 date filter
  - `end_date` (optional): ISO 8601 date filter
- **Response** (200):
  ```json
  {
    "status": "success",
    "data": {
      "items": [
        {
          "id": "uuid",
          "user_id": "uuid",
          "action": "CREATE",
          "resource_type": "user",
          "resource_id": "uuid",
          "ip_address": "192.168.1.100",
          "user_agent": "Mozilla/5.0...",
          "changes": {"username": "newuser"},
          "timestamp": "2026-03-30T10:30:00Z"
        }
      ],
      "total": 1524,
      "page": 1,
      "total_pages": 153
    }
  }
  ```

---

## Debug Endpoints (Development Only)

### RBAC Debug - User Permissions

- **URL**: `/rbac/debug/user/:id`
- **Method**: `GET`
- **Auth**: Bearer Token + `manage_users` permission
- **Response** (200):
  ```json
  {
    "status": "success",
    "data": {
      "user_id": "uuid",
      "username": "john_doe",
      "roles": [
        {
          "id": "uuid",
          "name": "Administrator"
        }
      ],
      "permissions": ["manage_users", "manage_roles", "manage_permissions"],
      "effective_permissions": 12
    }
  }
  ```
- **Use Case**: Troubleshoot RBAC configuration
- **Recommendation**: Disable in production
