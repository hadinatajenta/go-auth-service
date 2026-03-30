package e2e_test

import (
	"auth-service/internal/config"
	"auth-service/internal/router"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestContext holds shared test state
type TestContext struct {
	DB               *gorm.DB
	Router           *gin.Engine
	BaseURL          string
	HTTPClient       *http.Client
	pgContainer      testcontainers.Container
	TestAccessToken  string
	TestRefreshToken string
	TestUserEmail    string
}

var testCtx *TestContext

// TestMain initializes the test environment
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Create PostgreSQL container
	pgContainer, dsn, err := setupPostgresContainer(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup Postgres container: %v\n", err)
		os.Exit(1)
	}

	// Connect to database
	db, err := connectDatabase(dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		pgContainer.Terminate(ctx) //nolint:errcheck
		os.Exit(1)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run migrations: %v\n", err)
		pgContainer.Terminate(ctx) //nolint:errcheck
		os.Exit(1)
	}

	// Create test config
	testConfig := &config.Config{
		DBHost:             "localhost",
		DBUser:             "test",
		DBPassword:         "test",
		DBName:             "test_db",
		DBPort:             "5432", // Will be overridden by DSN
		DBSSLMode:          "disable",
		JWTSecret:          "test_jwt_secret_must_be_at_least_32_characters_long",
		AppPort:            "8080",
		CORSAllowedOrigins: "http://localhost:3000",
		Environment:        "development",
	}

	// Setup Gin router
	ginRouter := router.SetupRouter(db, testConfig)

	// Start test server
	ts := httptest.NewServer(ginRouter)
	defer ts.Close()

	// Initialize test context
	testCtx = &TestContext{
		DB:         db,
		Router:     ginRouter,
		BaseURL:    ts.URL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		pgContainer: pgContainer,
	}

	// Run tests
	code := m.Run()

	// Cleanup
	if err := pgContainer.Terminate(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to terminate container: %v\n", err)
	}

	os.Exit(code)
}

// setupPostgresContainer creates and starts a PostgreSQL container
func setupPostgresContainer(ctx context.Context) (testcontainers.Container, string, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "test_db",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to create container: %w", err)
	}

	// Get container host and port
	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx) //nolint:errcheck
		return nil, "", fmt.Errorf("failed to get host: %w", err)
	}

	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		container.Terminate(ctx) //nolint:errcheck
		return nil, "", fmt.Errorf("failed to get mapped port: %w", err)
	}

	// Build DSN
	dsn := fmt.Sprintf("user=test password=test dbname=test_db host=%s port=%s sslmode=disable",
		host, port.Port())

	// Wait for database to be ready
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port.Port()), 1*time.Second)
		if err == nil {
			conn.Close()
			break
		}
		if i == maxRetries-1 {
			container.Terminate(ctx) //nolint:errcheck
			return nil, "", fmt.Errorf("database not ready after retries")
		}
		time.Sleep(100 * time.Millisecond)
	}

	return container, dsn, nil
}

// connectDatabase establishes a connection to the test database
func connectDatabase(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

// runMigrations executes all SQL migration files
func runMigrations(db *gorm.DB) error {
	// Get the project root directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Navigate to migrations directory relative to test file location
	migrationsDir := filepath.Join(cwd, "../../migrations")

	// If that doesn't work, try from auth-service root
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		migrationsDir = filepath.Join(cwd, "migrations")
	}

	// List and read migration files
	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Sort migration files (should be named with numeric prefix)
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".up.sql") {
			filePath := filepath.Join(migrationsDir, file.Name())

			// Read migration file
			sqlBytes, err := ioutil.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read migration file %s: %w", file.Name(), err)
			}

			// Execute migration
			if err := db.Exec(string(sqlBytes)).Error; err != nil {
				return fmt.Errorf("failed to execute migration %s: %w", file.Name(), err)
			}
		}
	}

	return nil
}

// HTTPRequest is a helper for making HTTP requests in tests
type HTTPRequest struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    []byte
}

// DoRequest executes an HTTP request and returns the response
func (tc *TestContext) DoRequest(req HTTPRequest) (*http.Response, error) {
	httpReq, err := http.NewRequest(req.Method, tc.BaseURL+req.Path, strings.NewReader(string(req.Body)))
	if err != nil {
		return nil, err
	}

	// Set default headers
	httpReq.Header.Set("Content-Type", "application/json")

	// Apply custom headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	return tc.HTTPClient.Do(httpReq)
}

// Helper to read response body
func readResponseBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
