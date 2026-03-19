# Auth-Service

## Overview

Auth-Service is a centralized authentication and authorization microservice built with Go and the Gin framework. It provides secure identity management, access control, and session tracking for internal applications.

## Features

- **JWT Authentication**: Secure stateless authentication using JSON Web Tokens.
- **Refresh Token Sessions**: Long-lived sessions with persistent refresh tokens in PostgreSQL.
- **RBAC (Role-Based Access Control)**: Granular permission management with role hierarchy support.
- **Permission Caching**: In-memory caching for high-performance authorization checks.
- **API Key Authentication**: Machine-to-machine authentication for internal services.
- **Audit Logging**: Comprehensive tracking of security-sensitive operations.
- **Prometheus Metrics**: Built-in monitoring for login attempts, refresh operations, and request durations.
- **Token Introspection**: RFC 7662-inspired endpoint for internal services to verify token validity.
- **Session Metadata Tracking**: Capture client IP addresses and User-Agent strings for security auditing.

## Architecture

The service follows a clean architecture pattern:
`Handler -> Service -> Repository -> Database`

- **Framework**: [Gin Gonic](https://gin-gonic.com/)
- **Database**: [PostgreSQL](https://www.postgresql.org/)
- **ORM**: [GORM](https://gorm.io/)

## Project Structure

- `cmd/server`: Entry point for the application.
- `internal/module`: Domain-specific logic (auth, user, role, permission, etc.).
- `internal/middleware`: Custom Gin middlewares (auth, metrics, audit, rate limiting).
- `internal/utils`: Shared utilities and helpers (cache, security, responses).
- `internal/router`: Global route registration and dependency injection.
- `migrations`: SQL migration files for database schema management.
- `pkg/logger`: Unified logging wrapper.

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

| Endpoint | Method | Description |
| :--- | :--- | :--- |
| `/api/v1/auth/login` | POST | Authenticate and obtain tokens |
| `/api/v1/auth/refresh` | POST | Refresh an expired access token |
| `/api/v1/auth/introspect` | POST | Verify token validity from internal services |
| `/api/v1/me` | GET | Current user profile |
| `/api/v1/users` | GET | User management (admin) |
| `/api/v1/roles` | GET | RBAC Role management (admin) |
| `/api/v1/service-accounts` | POST | API Key management (admin) |

## Metrics

Prometheus metrics are exposed at:
`GET /metrics`

Available metrics include login successes/failures, token refreshes, and API request latencies.

## Health Check

Check service health at:
`GET /health`

The response includes database connectivity status.

## Deployment Notes

1. **Migrations**: Ensure all migrations in the `migrations/` directory are applied before starting the service.
2. **Security**: Ensure `JWT_SECRET` is set to a strong, random value in production environments.
3. **Caching**: The current implementation uses an in-memory cache; consider extending to Redis for distributed deployments.
