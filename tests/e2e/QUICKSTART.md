# E2E Tests Quick Start & Best Practices

## Quick Start (60 seconds)

### Prerequisites
- Go 1.25+
- Docker running (`docker ps` should work)

### Run Tests
```bash
cd auth-service

# Run all tests
go test -v ./tests/e2e/...

# Expected output:
# === RUN   TestRegisterUser
# --- PASS: TestRegisterUser (2.34s)
# === RUN   TestLoginUser
# --- PASS: TestLoginUser (2.45s)
# ... (more tests)
# ok      auth-service/tests/e2e  45.23s
```

That's it! Tests automatically:
- Spin up PostgreSQL container
- Run migrations
- Start the app
- Test all flows
- Clean up

## Anatomy of a Test

Here's TestLoginUser simplified to show the pattern:

```go
func TestLoginUser(t *testing.T) {
    // 1. ARRANGE: Prepare test data
    email := "test@example.com"
    password := "Password123!"
    
    // Register first
    registerBody := marshalJSON(map[string]string{
        "email":    email,
        "password": password,
    })
    
    // 2. ACT: Make HTTP request
    resp, err := testCtx.DoRequest(HTTPRequest{
        Method: http.MethodPost,
        Path:   "/api/v1/auth/login",
        Body:   registerBody,
    })
    if err != nil {
        t.Fatalf("Request failed: %v", err)
    }
    defer resp.Body.Close()
    
    // 3. ASSERT: Verify results
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("Expected 200, got %d", resp.StatusCode)
    }
    
    // Parse and verify response
    respBody, _ := ioutil.ReadAll(resp.Body)
    var authResp AuthResponse
    unmarshalJSON(respBody, &authResp)
    
    if authResp.Data.AccessToken == "" {
        t.Fatal("Access token missing")
    }
}
```

**Key Pattern: Arrange → Act → Assert**

## Best Practices

### 1. Use Unique Emails Per Test
```go
// ❌ BAD: Same email reused
func TestExample(t *testing.T) {
    email := "test@example.com" // Used by other tests too!
}

// ✅ GOOD: Unique email per test
func TestExample(t *testing.T) {
    email := "testexample@example.com"
}
```

### 2. Always Defer Response Body Close
```go
// ❌ BAD: Can leak goroutines
resp, _ := testCtx.DoRequest(req)
body, _ := ioutil.ReadAll(resp.Body)

// ✅ GOOD: Always close
resp, _ := testCtx.DoRequest(req)
defer resp.Body.Close()
body, _ := ioutil.ReadAll(resp.Body)
```

### 3. Check Status Code First
```go
// ❌ BAD: Try to parse error response as success
if err := unmarshalJSON(respBody, &SuccessResp); err != nil {
    // Error might be because response is error format
}

// ✅ GOOD: Check status code first
if resp.StatusCode != http.StatusOK {
    t.Fatalf("Unexpected status: %d", resp.StatusCode)
}
var resp SuccessResp
unmarshalJSON(respBody, &resp)
```

### 4. Store Tokens in TestContext for Reuse
```go
// ✅ GOOD: Store tokens once, reuse in tests
testCtx.TestAccessToken = authResp.Data.AccessToken

// Later test can use it:
resp, _ := testCtx.DoRequest(HTTPRequest{
    Headers: map[string]string{
        "Authorization": fmt.Sprintf("Bearer %s", testCtx.TestAccessToken),
    },
})
```

### 5. Handle Variable Shadowing
```go
// ❌ BAD: Redefining resp variable
resp, _ := firstRequest()
resp.Body.Close()

resp, _ := secondRequest() // Shadows previous resp!

// ✅ GOOD: Use different names
r1, _ := firstRequest()
r1.Body.Close()

r2, _ := secondRequest()
```

### 6. Use Helper Functions
```go
// Define once, use everywhere
func marshalJSON(v interface{}) []byte {
    data, _ := json.Marshal(v)
    return data
}

func unmarshalJSON(data []byte, v interface{}) error {
    return json.Unmarshal(data, v)
}

// Now tests are cleaner:
body := marshalJSON(map[string]string{
    "email": "test@example.com",
})
```

## Common Patterns

### Pattern 1: Testing with Authentication
```go
func TestAuthenticatedRequest(t *testing.T) {
    // Ensure we have a token (stored in testCtx)
    if testCtx.TestAccessToken == "" {
        t.Skip("No access token available")
    }
    
    // Make authenticated request
    resp, _ := testCtx.DoRequest(HTTPRequest{
        Method: http.MethodGet,
        Path:   "/api/v1/auth/me",
        Headers: map[string]string{
            "Authorization": fmt.Sprintf("Bearer %s", testCtx.TestAccessToken),
        },
    })
    // ... assertions
}
```

### Pattern 2: Testing Error Cases
```go
func TestUnauthorizedAccess(t *testing.T) {
    resp, _ := testCtx.DoRequest(HTTPRequest{
        Method: http.MethodGet,
        Path:   "/api/v1/auth/me",
        Headers: map[string]string{
            "Authorization": "Bearer invalid_token",
        },
    })
    
    if resp.StatusCode != http.StatusUnauthorized {
        t.Fatalf("Expected 401, got %d", resp.StatusCode)
    }
    
    respBody, _ := ioutil.ReadAll(resp.Body)
    var errResp ErrorResponse
    unmarshalJSON(respBody, &errResp)
    
    if errResp.Message == "" {
        t.Fatal("Error message missing")
    }
}
```

### Pattern 3: Testing Sequential Flows
```go
func TestMultipleSteps(t *testing.T) {
    // Step 1: Register
    registerResp := /* ... register logic ... */
    
    // Step 2: Login
    loginResp := /* ... login with registered user ... */
    token := loginResp.Data.AccessToken
    
    // Step 3: Use token
    resp := testCtx.DoRequest(HTTPRequest{
        Headers: map[string]string{
            "Authorization": fmt.Sprintf("Bearer %s", token),
        },
    })
    
    // Step 4: Verify
    if resp.StatusCode != http.StatusOK {
        t.Fatal("Token doesn't work")
    }
}
```

## Debugging Tests

### 1. Print Response Body on Failure
```go
if resp.StatusCode != http.StatusOK {
    respBody, _ := ioutil.ReadAll(resp.Body)
    t.Logf("Response: %s", string(respBody))
    t.Fatalf("Expected 200, got %d", resp.StatusCode)
}
```

### 2. Run Single Test with Verbose Output
```bash
go test -v -run TestLoginUser ./tests/e2e/...
```

### 3. Run with More Logging
```bash
# Disable test caching
go test -v -count=1 ./tests/e2e/...

# With environment variable
DEBUG=1 go test -v ./tests/e2e/...
```

### 4. Check Container Logs if Test Fails
```bash
# See what containers are running
docker ps

# View logs of postgres container
docker logs <container_id>
```

## Extending Tests

### Adding a New Test
1. Create function in auth_flow_test.go:
```go
func TestNewFeature(t *testing.T) {
    // Your test here
}
```

2. Follow the Arrange → Act → Assert pattern

3. Use existing helpers (marshalJSON, etc.)

4. Reuse tokens from testCtx if needed

5. Run: `go test -v -run TestNewFeature ./tests/e2e/...`

### Adding New Response Type
If you need to parse a new response:

1. Add struct to auth_flow_test.go:
```go
type MyResponse struct {
    Success bool `json:"success"`
    Data    struct {
        Field1 string `json:"field1"`
    } `json:"data"`
    Message string `json:"message"`
}
```

2. Use in test:
```go
var resp MyResponse
unmarshalJSON(respBody, &resp)
if resp.Data.Field1 == "" {
    t.Fatal("Field missing")
}
```

## Troubleshooting Guide

### Problem: "database not ready after retries"
**Cause**: PostgreSQL container isn't starting  
**Solution**: 
```bash
# Check Docker is running
docker ps

# Check for disk space
df -h

# Try with longer timeout in setup_test.go
maxRetries := 60  // Increase from 30
```

### Problem: "Address already in use"
**Cause**: Port 5432 already in use  
**Solution**:
```bash
# Find what's using port 5432
lsof -i :5432

# Stop conflicting service or use different port
```

### Problem: "no such file or directory" for migrations
**Cause**: Migration path is wrong  
**Solution**: Check migrations directory path in runMigrations():
```go
// Print actual path being used
fmt.Println("Looking for migrations at:", migrationsDir)
files, err := ioutil.ReadDir(migrationsDir)
```

### Problem: Tests timeout
**Cause**: Container too slow to start  
**Solution**:
```bash
# Run with longer timeout
go test -timeout 10m ./tests/e2e/...

# Or increase timeout in code
WaitingFor: wait.ForListeningPort("5432/tcp").
    WithStartupTimeout(30 * time.Second)
```

## Performance Tips

### 1. Run Tests in Parallel
```bash
go test -v -parallel 4 ./tests/e2e/...
```

### 2. Skip Container Logs if Not Needed
```bash
# Don't capture container logs
container, _ := testcontainers.GenericContainer(ctx, 
    testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started: true,
    })
```

### 3. Reuse testCtx Tokens
Tests that run after TestLoginUser can use testCtx.TestAccessToken instead of logging in again.

### 4. Run Only What You Need
```bash
# Run one test
go test -run TestLoginUser ./tests/e2e/...

# Run tests matching pattern
go test -run "^Test.*Token" ./tests/e2e/...
```

## CI/CD Integration Examples

### GitHub Actions
```yaml
name: E2E Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: 1.25
      - name: Run E2E Tests
        run: cd auth-service && go test -v ./tests/e2e/...
```

### GitLab CI
```yaml
e2e-tests:
  image: golang:1.25
  services:
    - docker:dind
  script:
    - cd auth-service
    - go test -v ./tests/e2e/...
```

## Maintenance

### Regular Checks
- [ ] Test all major flows monthly
- [ ] Update response structs when API changes
- [ ] Keep testcontainers-go updated
- [ ] Review and fix flaky tests

### When Adding API Endpoints
1. Add to auth_flow_test.go
2. Create test function
3. Test happy path + error cases
4. Run: `go test -v ./tests/e2e/...`
5. Commit with test

### When Changing Response Format
1. Update response struct
2. Run tests to see failures
3. Update assertions
4. Ensure backwards compatibility

## Summary

**Remember:**
- ✅ Always close response bodies
- ✅ Check status code before parsing
- ✅ Use unique test data (emails)
- ✅ Follow Arrange → Act → Assert
- ✅ Store reusable tokens in testCtx
- ✅ Handle variable shadowing
- ✅ Keep tests independent

**Run tests regularly to catch regressions early!**

```bash
go test -v ./tests/e2e/...
```

This is your safety net for production deployments. Keep it green! ✅
