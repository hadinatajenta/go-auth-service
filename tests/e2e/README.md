# End-to-End Tests for Auth-Service

This directory contains comprehensive end-to-end (E2E) tests for the Auth-Service microservice using `testcontainers-go` and the Go standard library.

## Overview

The E2E test suite validates the complete authentication flow including:
- User registration
- User login with JWT tokens
- Token validation and refresh
- Account lockout after failed attempts
- API key creation and usage
- Service account revocation

## Architecture

### Files

- **setup_test.go**: Initializes test environment
  - Spins up a PostgreSQL container with testcontainers
  - Runs all SQL migrations
  - Starts the Gin HTTP server
  - Provides test helpers for making HTTP requests

- **auth_flow_test.go**: Core test cases covering all authentication flows
  - 8 comprehensive test functions
  - Uses only `net/http` and `encoding/json` (no test helper libraries)
  - Tests raw HTTP requests and responses

## Prerequisites

- Go 1.25.0+
- Docker (required for testcontainers)
- Docker daemon must be running

## Running Tests

### Run all E2E tests
```bash
cd auth-service
go test -v ./tests/e2e/...
```

### Run specific test
```bash
go test -v -run TestLoginUser ./tests/e2e/...
```

### Run with timeout
```bash
go test -v -timeout 5m ./tests/e2e/...
```

### Run with verbose output and race detection
```bash
go test -v -race ./tests/e2e/...
```

## Test Cases

### 1. TestRegisterUser (201)
Validates user registration with email, password, and username.

**Endpoint**: `POST /api/v1/auth/register`
**Expected**: HTTP 201 or 200

### 2. TestLoginUser (200 + tokens)
Validates login returns both access and refresh tokens.

**Endpoint**: `POST /api/v1/auth/login`
**Expected**: HTTP 200 with `access_token` and `refresh_token`

### 3. TestGetMeWithAccessToken (200 + email validation)
Validates that authenticated requests work with Bearer token.

**Endpoint**: `GET /api/v1/auth/me`
**Expected**: HTTP 200 with email matching login user

### 4. TestGetMeWithExpiredToken (401)
Validates that expired tokens are rejected.

**Endpoint**: `GET /api/v1/auth/me`
**Expected**: HTTP 401 Unauthorized

### 5. TestRefreshToken (200 + new token)
Validates refresh token exchange for new access token.

**Endpoint**: `POST /api/v1/auth/refresh`
**Expected**: HTTP 200 with new `access_token`

### 6. TestAccountLockout (429 or 401 after 5 failed attempts)
Validates account lockout protection after 5 failed login attempts.

**Endpoint**: `POST /api/v1/auth/login`
**Expected**: 
- Attempts 1-5: HTTP 401
- Attempt 6: HTTP 429 (Too Many Requests) or HTTP 401 with locked message

### 7. TestAPIKeyUsage (200 with API key)
Validates that API keys can be created and used to access protected endpoints.

**Endpoints**:
- `POST /api/v1/service-accounts` - Create API key
- `GET /api/v1/auth/me` - Access with API key via `X-API-Key` header

**Expected**: HTTP 200

### 8. TestRevokeAPIKey (401 after revocation)
Validates that revoked API keys no longer work.

**Endpoints**:
- `POST /api/v1/service-accounts/:id/revoke` - Revoke key
- `GET /api/v1/auth/me` - Try to access with revoked key

**Expected**: HTTP 401

## Test Environment

### Automatic Setup
- PostgreSQL 15 container starts automatically
- All migrations run before tests
- Gin server starts with test configuration
- Database is cleaned up after tests complete

### Configuration
Tests use the following environment:
- **JWT Secret**: 64-character test key (meets 32-char minimum requirement)
- **CORS Origins**: http://localhost:3000
- **Environment**: development (debug endpoints enabled)
- **Database**: PostgreSQL 15 Alpine
- **Port**: Random (testcontainers)

## Key Features

### No External Dependencies for HTTP Testing
- Uses only `net/http` and `encoding/json`
- No test helper libraries (e.g., no testify, httpexpect)
- Raw HTTP requests and response parsing
- Easy to understand and maintain

### Database Isolation
- Each test run gets fresh PostgreSQL container
- Full schema initialization from migrations
- Automatic cleanup
- No test data persistence between runs

### Concurrent Test Safety
- All tests can run in parallel with `-parallel` flag
- PostgreSQL connection pooling configured
- Tests use unique email addresses for isolation

### Comprehensive Error Checking
- HTTP status code validation
- Response body parsing and validation
- Email/token verification
- Account lockout scenario testing

## Response Structures

### AuthResponse (Login/Register)
```json
{
  "success": true,
  "data": {
    "access_token": "eyJ...",
    "refresh_token": "eyJ...",
    "user": {
      "id": 1,
      "email": "user@example.com"
    }
  },
  "message": "Login successful"
}
```

### UserResponse (Get Me)
```json
{
  "success": true,
  "data": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe"
  },
  "message": "Success"
}
```

### RefreshResponse (Token Refresh)
```json
{
  "success": true,
  "data": {
    "access_token": "eyJ..."
  },
  "message": "Token refreshed successfully"
}
```

## Troubleshooting

### Tests timeout
- Increase timeout: `go test -timeout 10m ./tests/e2e/...`
- Check Docker daemon is running
- Check system resources (disk, memory)

### PostgreSQL container fails to start
```bash
# Check Docker is running
docker ps

# Check for port conflicts
docker ps -a | grep postgres

# Clean up old containers
docker container prune
```

### Tests cannot connect to database
- Verify migrations directory path is correct (relative to test file)
- Check migrations are valid SQL
- Check database user/password in setup_test.go

### Tests fail with "connection refused"
- Ensure `testcontainers` can reach Docker daemon
- On Mac/Windows, Docker Desktop must be running
- On Linux, Docker socket must be accessible

## Development Notes

### Adding New Tests
1. Add test function in `auth_flow_test.go`
2. Use `testCtx.DoRequest()` to make HTTP calls
3. Parse JSON responses with `json.Unmarshal()`
4. Use standard assertions with `if` statements and `t.Fatal()`

### Test Helpers
- `marshalJSON(v)` - Marshal struct to JSON bytes
- `unmarshalJSON(data, v)` - Unmarshal JSON bytes to struct
- `testCtx.DoRequest(req)` - Make HTTP request
- `readResponseBody(resp)` - Read response body

### Running Specific Test Functions
```bash
# Run only TestLoginUser
go test -v -run TestLoginUser ./tests/e2e/...

# Run all tests matching pattern
go test -v -run "^TestAccount" ./tests/e2e/...
```

## CI/CD Integration

### GitHub Actions Example
```yaml
- name: Run E2E Tests
  run: |
    cd auth-service
    go test -v -timeout 10m ./tests/e2e/...
```

### GitLab CI Example
```yaml
e2e-tests:
  image: golang:1.25
  services:
    - docker:dind
  script:
    - cd auth-service
    - go test -v -timeout 10m ./tests/e2e/...
```

## Performance

- **Total runtime**: ~30-60 seconds (varies by system)
- **Container startup**: ~5-10 seconds
- **Migrations**: ~1-2 seconds
- **Test execution**: ~20-40 seconds
- **Cleanup**: ~5 seconds

## Dependencies

```go
require (
    github.com/testcontainers/testcontainers-go v0.27.0
    gorm.io/driver/postgres v1.6.0
    gorm.io/gorm v1.31.1
    github.com/gin-gonic/gin v1.12.0
)
```

## License

Same as Auth-Service (Apache 2.0)
