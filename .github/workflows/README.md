# GitHub Actions CI/CD Workflow Documentation

## Overview

The `.github/workflows/test.yml` workflow provides comprehensive testing and quality checks for the Auth-Service project using GitHub Actions. The workflow is triggered on every push and pull request to the `main` branch.

## Workflow Structure

The workflow consists of 5 sequential and parallel jobs:

```
┌─────────────────────────────────────────────┐
│  Lint (go vet, gofmt)                       │
│  - Checks code quality                      │
│  - Verifies formatting                      │
└─────────────────────┬───────────────────────┘
                      │
          ┌───────────┴───────────┐
          │                       │
    ┌─────▼──────┐         ┌─────▼──────┐
    │ Unit Tests │         │ E2E Tests  │
    │ + Coverage │         │ + Docker   │
    └─────┬──────┘         └─────┬──────┘
          │                       │
          └───────────┬───────────┘
                      │
          ┌───────────┴───────────┐
          │                       │
    ┌─────▼──────┐         ┌─────▼──────┐
    │  Security  │         │Status Check│
    │   Scan     │         │ (Summary)  │
    └────────────┘         └────────────┘
```

## Jobs

### 1. Lint Job
**Name**: `lint`  
**Runs on**: `ubuntu-latest`  
**Duration**: ~2-3 minutes

**Steps**:
- Checkout code
- Set up Go 1.25
- Run `go vet ./...` - Detects suspicious constructs
- Run `gofmt` check - Ensures code formatting compliance

**Artifacts**: None

### 2. Unit Tests Job
**Name**: `test`  
**Runs on**: `ubuntu-latest`  
**Depends on**: `lint`  
**Duration**: ~5-10 minutes

**Steps**:
- Checkout code
- Set up Go 1.25
- Cache Go modules (faster builds)
- Download dependencies
- Run unit tests with coverage:
  ```bash
  go test ./internal/... -v -race -coverprofile=coverage.out
  ```
- Generate coverage report using `go tool cover`
- Upload coverage to Codecov
- Archive coverage report as artifact

**Coverage**:
- Runs with `-race` flag to detect race conditions
- Generates coverage profile in `coverage.out`
- Publishes coverage to Codecov.io
- Displays coverage percentage in job summary

**Artifacts**:
- `coverage-report` - Coverage profile (30 days retention)

### 3. E2E Tests Job
**Name**: `e2e`  
**Runs on**: `ubuntu-latest` (has Docker available)  
**Depends on**: `lint`  
**Timeout**: 15 minutes  
**Duration**: ~8-12 minutes

**Steps**:
- Checkout code
- Set up Go 1.25
- Set up Docker Buildx (for container operations)
- Cache Go modules
- Download dependencies
- Run E2E tests:
  ```bash
  go test ./tests/e2e/... -v -timeout 120s
  ```

**Environment Variables**:
- `JWT_SECRET`: 32-character test key (meets production requirements)
- `ENVIRONMENT`: Set to `development` for testing

**Requirements**:
- Docker available (ubuntu-latest includes Docker)
- 120-second timeout per test
- testcontainers-go handles container lifecycle

**Artifacts**:
- `e2e-test-results` - Test output (7 days retention)

### 4. Security Job
**Name**: `security`  
**Runs on**: `ubuntu-latest`  
**Depends on**: `lint`  
**Duration**: ~3-5 minutes

**Steps**:
- Checkout code
- Set up Go 1.25
- Run gosec security scanner
  - Scans for security vulnerabilities
  - Outputs JSON report
  - Continues on error (doesn't block pipeline)
- Upload results as artifact

**Scanner**: `securego/gosec`  
**Artifacts**:
- `gosec-results` - Security scan results in JSON (7 days retention)

### 5. Status Job
**Name**: `status`  
**Runs on**: `ubuntu-latest`  
**Depends on**: All previous jobs  
**Always runs**: `true` (runs even if previous jobs fail)  
**Duration**: ~1 minute

**Purpose**:
- Generates summary of all test results
- Creates GitHub job summary with status table
- Fails if any critical job failed (lint, test, e2e)
- Passes if only security scan fails

**Output**:
- GitHub Step Summary with results table
- Clear pass/fail indication

## Environment Variables

### Set in Workflow
```yaml
JWT_SECRET: "test_jwt_secret_must_be_at_least_32_characters_long1234"
GO_VERSION: "1.25"
```

### Used in E2E Tests
```yaml
JWT_SECRET: ${{ env.JWT_SECRET }}
ENVIRONMENT: development
```

**Note**: In production, these would come from GitHub Secrets, but for testing they're hardcoded since they're test values.

## Triggers

The workflow runs on:
- **Push to main**: Every commit to main branch
- **Pull Requests to main**: Every PR targeting main branch

### Example Trigger Events
```yaml
on:
  push:
    branches:
      - main
  pull_request:
    branches:
      - main
```

## Running Tests Locally (Equivalent)

To run the same checks locally:

```bash
cd auth-service

# 1. Lint checks (like GitHub Actions)
go vet ./...
gofmt -s -l .

# 2. Unit tests with coverage
go test ./internal/... -v -race -coverprofile=coverage.out

# 3. View coverage report
go tool cover -func=coverage.out
go tool cover -html=coverage.out  # Opens in browser

# 4. E2E tests (requires Docker)
export JWT_SECRET="test_jwt_secret_must_be_at_least_32_characters_long1234"
export ENVIRONMENT=development
go test ./tests/e2e/... -v -timeout 120s

# 5. Security scan
gosec ./...
```

## Artifacts and Reports

### Coverage Report
- **Location**: `auth-service/coverage.out`
- **Format**: Go coverage format
- **Usage**: Upload to Codecov, generate HTML report
- **Retention**: 30 days

### E2E Test Results
- **Location**: `auth-service/test-results/`
- **Format**: Go test output
- **Retention**: 7 days

### Security Results
- **Location**: `gosec-results.json`
- **Format**: JSON
- **Retention**: 7 days

## Codecov Integration

Coverage is automatically uploaded to Codecov.io:
- Shows coverage trends over time
- Provides PR coverage reports
- Integrates with GitHub PR checks
- No additional configuration needed

To view coverage:
1. Go to [codecov.io](https://codecov.io)
2. Connect your GitHub account
3. Find `pookie1811` repository
4. View coverage reports and trends

## Troubleshooting

### Issue: E2E Tests Timeout
**Cause**: Container startup too slow or insufficient resources  
**Solution**:
- Increase timeout in workflow (currently 120s per test, 15m job timeout)
- Check Docker availability on runner
- Review container logs in workflow output

### Issue: Coverage Report Not Generated
**Cause**: No tests were executed or test files not found  
**Solution**:
- Verify test files exist in `./internal/...`
- Check test file naming (`*_test.go`)
- Ensure tests are not skipped

### Issue: Security Scan Failures
**Cause**: Code has potential security issues  
**Solution**:
- Review gosec output in artifacts
- Fix issues or suppress false positives
- Re-run workflow

### Issue: Module Download Failure
**Cause**: Network issue or dependency problem  
**Solution**:
- Check go.mod and go.sum are committed
- Verify dependencies are accessible
- Check mod cache with `go mod verify`

## Performance Optimization

### Module Caching
The workflow caches Go modules between runs:
```yaml
uses: actions/cache@v3
with:
  path: ~/go/pkg/mod
  key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
```

**Benefits**:
- Saves ~1-2 minutes per run (module download time)
- Cache invalidates when go.sum changes
- Improves CI/CD speed significantly

### Parallel Jobs
Jobs 2-4 (test, e2e, security) run in parallel after lint:
- Lint must pass first (no point testing bad code)
- Other jobs can run simultaneously
- Reduces total pipeline time

**Typical Timing**:
- Lint: 2-3 minutes
- Test + E2E + Security (parallel): 8-12 minutes
- Status: 1 minute
- **Total**: ~13-16 minutes

## Matrix Testing (Optional Enhancement)

To test against multiple Go versions:

```yaml
strategy:
  matrix:
    go-version: ["1.24", "1.25"]

steps:
  - uses: actions/setup-go@v4
    with:
      go-version: ${{ matrix.go-version }}
```

This would run tests for each Go version.

## GitHub Branch Protection Rules

Recommended settings to require passing workflow:

1. Go to **Settings** → **Branches**
2. Add rule for `main`
3. Enable:
   - ✅ Require status checks to pass before merging
   - ✅ Require branches to be up to date before merging
   - Select required checks:
     - `lint`
     - `test`
     - `e2e`
     - `status`

## Security Considerations

### Secrets Management
Currently, `JWT_SECRET` is hardcoded as a test value. For production CI/CD:

```yaml
env:
  JWT_SECRET: ${{ secrets.JWT_SECRET }}
```

Add as GitHub Secret via:
1. Settings → Secrets and variables → Actions
2. Create `JWT_SECRET` secret
3. Reference in workflow

### Docker Security
- Uses official Docker images (postgres:15, golang:1.25)
- Pulls latest security patches (ubuntu-latest)
- No credentials or secrets in logs
- Container operations isolated per job

## Monitoring and Alerts

### GitHub Actions Dashboard
- View workflow runs at repository → Actions
- See each job's status and duration
- Download artifacts

### Email Notifications
Automatic notifications for:
- Workflow failures
- Can be configured in GitHub account settings

### Branch Status Badges
Add to README.md:
```markdown
![Tests](https://github.com/hadinatajenta/pookie1811/actions/workflows/test.yml/badge.svg?branch=main)
```

## Common Modifications

### Add Lint Tools
```yaml
- name: Run golangci-lint
  uses: golangci/golangci-lint-action@v3
```

### Add Dependency Check
```yaml
- name: Check for vulnerable dependencies
  run: cd auth-service && go list -json -m all | nancy sleuth
```

### Custom Coverage Threshold
```yaml
- name: Check coverage threshold
  run: |
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | cut -d'%' -f1)
    if (( $(echo "$COVERAGE < 80" | bc -l) )); then
      echo "Coverage is ${COVERAGE}%, expected at least 80%"
      exit 1
    fi
```

### Run Tests on Schedule
```yaml
on:
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM UTC
```

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Go Testing Packages](https://golang.org/pkg/testing/)
- [Codecov Documentation](https://docs.codecov.io)
- [golangci-lint](https://golangci-lint.run/)
- [gosec](https://github.com/securego/gosec)

## Contact & Support

For questions about the workflow:
1. Check GitHub Actions logs for specific failures
2. Review this documentation
3. See `.github/workflows/test.yml` for implementation details
4. Consult Go and GitHub Actions official docs

---

**Last Updated**: March 30, 2026  
**Workflow Version**: 1.0  
**Go Version**: 1.25  
**Auth-Service Version**: v1.0.0-READY
