# Auth-Service: Internal Identity & Access Management (IAM) System

## Overview

**Auth-Service** is a comprehensive **centralized Identity & Access Management (IAM) microservice** built with Go, Gin, and PostgreSQL. It serves as the single source of truth for:

- **User Authentication**: JWT-based stateless authentication with persistent sessions
- **Access Control (RBAC)**: Role-Based Access Control with granular permission management
- **Service Authentication**: Machine-to-machine (M2M) authentication via API Keys (Service Accounts)
- **Audit & Compliance**: Complete audit trail of all security-sensitive operations
- **Multi-tenant Support**: Support for multi-user, multi-role scenarios with hierarchical permission management

Instead of each application managing its own user database and permission logic, all internal services delegate these responsibilities to Auth-Service, eliminating duplicated security logic and providing a unified security audit log.

## Core Features

### Authentication & Sessions
- **JWT Authentication**: Secure stateless authentication using JSON Web Tokens (15-minute expiry)
- **Refresh Token Sessions**: Long-lived sessions with JWT refresh tokens (7-day expiry) persisted in PostgreSQL
- **Session Metadata Tracking**: Capture client IP addresses and User-Agent strings for security auditing
- **Account Lockout Protection**: 5 consecutive failed login attempts trigger a 15-minute account lock
- **Password Management**: Bcrypt-based password hashing, password reset flow, and change password operations

### Authorization (RBAC)
- **Role-Based Access Control**: Granular permission management with role hierarchy support
- **In-Memory Permission Cache**: O(1) permission lookups with automatic cache invalidation
- **Dynamic Permission Resolution**: Permissions fetched on-demand with configurable TTL (default 15 minutes)
- **Permission Grouping**: Organize permissions by functional category for easier management

### Service-to-Service Authentication
- **API Key (Service Account) Management**: Cryptographically secure M2M authentication
- **API Key Format**: `sk_live_<8-char-prefix>_<32-char-secret>` with bcrypt hashing
- **Key Lifecycle Management**: Create, list, and revoke API keys with audit tracking
- **Last Used Tracking**: Asynchronous tracking of API key usage for security monitoring

### Audit & Compliance
- **Comprehensive Audit Logging**: All security-sensitive mutations recorded with user, IP, method, and path
- **Audit Trail Queryable**: Built-in endpoint for querying historical audit logs
- **Prometheus Metrics**: Production-grade monitoring with metrics for logins, token operations, and latencies
- **Health Check Endpoint**: Database connectivity status for liveness probes

### Additional Features
- **Token Introspection**: RFC 7662-inspired endpoint for internal services to verify token validity
- **Menu Management**: Hierarchical (parent-child) frontend navigation menu system with role-based access
- **CORS Support**: Cross-origin request handling with configurable origins
- **Rate Limiting**: Token bucket rate limiting on authentication endpoints (5 req/min, burst 10)

## Architecture

The service follows a **Clean Architecture** pattern with strict layering:

```
HTTP Request
    ↓
[Middleware Stack]
    • Recovery (panic recovery)
    • Logger (Gin access logging)
    • CORS (cross-origin handling)
    • Audit (mutation recording)
    • Metrics (Prometheus instrumentation)
    • Rate Limit (auth endpoints)
    • Auth (JWT/API Key validation)
    • Permission (RBAC enforcement)
    ↓
[Handler Layer] (request binding, validation, response formatting)
    ↓
[Service Layer] (business logic, token operations, cache management)
    ↓
[Repository Layer] (GORM-based data access)
    ↓
[PostgreSQL Database]
```

**Technology Stack**:
- **Framework**: [Gin Gonic](https://gin-gonic.com/) v1.12.0 — High-performance HTTP routing
- **Database**: [PostgreSQL](https://www.postgresql.org/) — Persistent session and audit storage
- **ORM**: [GORM](https://gorm.io/) v1.31.1 — Type-safe database queries
- **JWT**: [golang-jwt/jwt](https://github.com/golang-jwt/jwt) v5.3.1 — Token generation and validation
- **Security**: [golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto) — Bcrypt password hashing
- **Monitoring**: [Prometheus client_golang](https://github.com/prometheus/client_golang) — Metrics export
- **Documentation**: [Swaggo](https://github.com/swaggo/swag) — OpenAPI/Swagger auto-generation
- **Validation**: [go-playground/validator](https://github.com/go-playground/validator) — Struct validation

## Project Structure

```
auth-service/
├── cmd/
│   ├── server/           # Application entry point
│   └── seed/             # CLI for initial data seeding
├── internal/
│   ├── config/           # Environment variables and configuration
│   ├── database/         # PostgreSQL connection and setup
│   ├── middleware/       # Gin middlewares (auth, metrics, audit, rate limit, CORS)
│   ├── module/           # Domain-specific logic (modular by feature)
│   │   ├── auth/         # Authentication (login, logout, refresh, introspect)
│   │   ├── user/         # User management (profile, password, profile update)
│   │   ├── role/         # Role CRUD and permission assignment
│   │   ├── permission/   # Permission management
│   │   ├── menu/         # Menu hierarchy management
│   │   ├── service_account/  # API key lifecycle
│   │   └── audit/        # Audit log querying
│   ├── router/           # Route registration and dependency injection
│   └── utils/            # Shared utilities (response formatting, validation, cache)
│       ├── cache/        # In-memory set-based cache for permissions
│       ├── metrics/      # Prometheus metric definitions
│       ├── crypto.go     # Password hashing utilities
│       ├── jwt.go        # Token generation and validation
│       ├── response.go   # Standardized JSON response formatting
│       └── errors.go     # Error handling
├── migrations/           # SQL migration files (database schema)
├── pkg/
│   └── logger/           # log/slog wrapper for structured logging
├── docs/                 # OpenAPI/Swagger documentation
├── Dockerfile            # Multi-stage Docker build
├── docker-compose.yml    # Local development orchestration
├── go.mod               # Go module definition
└── README.md            # This file
```

## Modular Design

Each feature is organized as an independent module under `internal/module/<feature>` with:

- **Handler** (`handler.go`): Gin HTTP handlers, request binding, validation
- **Service** (`service.go`): Core business logic, external API calls, cache coordination
- **Repository** (`repository.go`): GORM queries, database operations
- **Model** (`model.go`): Domain entities and database structs
- **DTO** (`dto.go` or inline): Request/response data transfer objects

This modular structure makes it easy to:
- Add new features without affecting existing modules
- Test each layer independently
- Maintain clear separation of concerns
- Scale specific modules as needed

## Requirements

- Go 1.25+
- PostgreSQL

## Installation

1. Clone the repository.
2. Install dependencies:
   ```bash
   go mod tidy
   ```
3. Run the application:
   ```bash
   go run cmd/server/main.go
   ```

## Environment Variables

Copy `.env.example` to `.env` and configure the following:

- `APP_PORT`: Port for the server to listen on.
- `DB_HOST`: PostgreSQL host.
- `DB_PORT`: PostgreSQL port.
- `DB_USER`: PostgreSQL username.
- `DB_PASSWORD`: PostgreSQL password.
- `DB_NAME`: PostgreSQL database name.
- `DB_SSLMODE`: Database SSL mode (e.g., disable).
- `JWT_SECRET`: Secret key for signing JWTs.

## Running with Docker

### Using Docker Compose
```bash
docker compose up -d --build
```

### Manual Build
```bash
docker build -t auth-service .
docker run -p 8080:8080 auth-service
```

## API Overview

### Public Endpoints (No Authentication Required)
| Endpoint | Method | Description |
| :--- | :--- | :--- |
| `POST /api/v1/auth/login` | POST | Authenticate with username/password, obtain JWT and refresh token |
| `POST /api/v1/auth/register` | POST | Register new user account |
| `POST /api/v1/auth/refresh` | POST | Refresh expired access token using refresh token |
| `POST /api/v1/auth/forgot-password` | POST | Request password reset token |
| `POST /api/v1/auth/reset-password` | POST | Reset password with reset token |
| `POST /api/v1/auth/introspect` | POST | Verify token validity (internal service consumption) |
| `GET /health` | GET | Health check / liveness probe |

### Authenticated User Endpoints
| Endpoint | Method | Description |
| :--- | :--- | :--- |
| `GET /api/v1/me` | GET | Current authenticated user profile |
| `GET /api/v1/me/permissions` | GET | List permissions for current user |
| `GET /api/v1/me/roles` | GET | List roles assigned to current user |
| `POST /api/v1/me/change-password` | POST | Change own password |

### User Management Endpoints (Admin)
| Endpoint | Method | Description |
| :--- | :--- | :--- |
| `GET /api/v1/users` | GET | List all users (paginated) |
| `POST /api/v1/users` | POST | Create new user account |
| `GET /api/v1/users/:id` | GET | Get user profile by ID |
| `PUT /api/v1/users/:id` | PUT | Update user profile (name, email) |
| `DELETE /api/v1/users/:id` | DELETE | Delete user and all associated sessions |
| `GET /api/v1/users/:id/roles` | GET | List roles assigned to user |
| `POST /api/v1/users/:id/roles` | POST | Assign role to user (cache invalidation) |
| `DELETE /api/v1/users/:id/roles/:role_id` | DELETE | Remove role from user (cache invalidation) |

### Role Management Endpoints (Admin)
| Endpoint | Method | Description |
| :--- | :--- | :--- |
| `GET /api/v1/roles` | GET | List all roles (paginated) |
| `POST /api/v1/roles` | POST | Create new role |
| `GET /api/v1/roles/:id` | GET | Get role details |
| `PUT /api/v1/roles/:id` | PUT | Update role name/description |
| `DELETE /api/v1/roles/:id` | DELETE | Delete role (cache invalidation) |
| `GET /api/v1/roles/:id/permissions` | GET | List permissions in role |
| `POST /api/v1/roles/:id/permissions` | POST | Add permission to role (cache invalidation) |
| `DELETE /api/v1/roles/:id/permissions/:perm_id` | DELETE | Remove permission from role (cache invalidation) |
| `GET /api/v1/roles/:id/users` | GET | List users with this role |

### Permission Management Endpoints (Admin)
| Endpoint | Method | Description |
| :--- | :--- | :--- |
| `GET /api/v1/permissions` | GET | List all permissions (paginated) |
| `POST /api/v1/permissions` | POST | Create new permission |
| `GET /api/v1/permissions/:id` | GET | Get permission details |
| `PUT /api/v1/permissions/:id` | PUT | Update permission name/description |
| `DELETE /api/v1/permissions/:id` | DELETE | Delete permission (cache invalidation) |
| `GET /api/v1/permissions/grouped` | GET | List permissions grouped by category |

### Service Account Endpoints (Admin)
| Endpoint | Method | Description |
| :--- | :--- | :--- |
| `POST /api/v1/service-accounts` | POST | Create new API key (raw key returned only once) |
| `GET /api/v1/service-accounts` | GET | List all service accounts |
| `POST /api/v1/service-accounts/:id/revoke` | POST | Revoke API key (immediate deactivation) |

### Menu Management Endpoints (Admin & User)
| Endpoint | Method | Description |
| :--- | :--- | :--- |
| `GET /api/v1/menus/allowed` | GET | List accessible menus for current user |
| `GET /api/v1/menus/tree` | GET | Get menu hierarchy tree (current user) |
| `POST /api/v1/menus` | POST | Create new menu item |
| `GET /api/v1/menus/:id` | GET | Get menu item details |
| `PUT /api/v1/menus/:id` | PUT | Update menu item |
| `DELETE /api/v1/menus/:id` | DELETE | Delete menu item |

### Audit & Debug Endpoints (Admin)
| Endpoint | Method | Description |
| :--- | :--- | :--- |
| `GET /api/v1/audit-logs` | GET | Query security audit log (paginated, filterable) |
| `GET /api/v1/rbac/debug/user/:id` | GET | Debug user permissions and roles (development) |

## Prometheus Metrics

Prometheus metrics are exposed at: `GET /metrics`

### Available Metrics

| Metric | Type | Description |
| :--- | :--- | :--- |
| `login_attempt_total` | Counter | Total login attempts |
| `login_success_total` | Counter | Successful logins |
| `login_failure_total` | Counter | Failed login attempts |
| `token_refresh_total` | Counter | Successful token refreshes |
| `auth_request_total` | CounterVec | All requests (labels: method, path, status) |
| `auth_request_duration_seconds` | HistogramVec | Request latencies (labels: method, path) |

Example Prometheus scrape config:
```yaml
scrape_configs:
  - job_name: 'auth-service'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
```

## Health Check

`GET /health` returns database connectivity status. Used as a liveness probe in Kubernetes deployments.

**Response** (healthy):
```json
{
  "message": "Server is healthy",
  "data": {
    "status": "up"
  }
}
```

**Status Codes**:
- `200 OK` — Service is healthy
- `503 Service Unavailable` — Database connection failed

Check service health at:
`GET /health`

The response includes database connectivity status.

## Deployment Notes

1. **Migrations**: Ensure all migrations in the `migrations/` directory are applied before starting the service.
2. **Security**: Ensure `JWT_SECRET` is set to a strong, random value in production environments.
3. **Caching**: The current implementation uses an in-memory cache; consider extending to Redis for distributed deployments.
