package middleware

import (
	"auth-service/internal/module/user"
	"auth-service/internal/utils/cache"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type mockUserRepo struct {
	permissions []string
	err         error
}

func (m *mockUserRepo) GetUserPermissions(ctx context.Context, userID uint) ([]string, error) {
	return m.permissions, m.err
}

func (m *mockUserRepo) Create(ctx context.Context, u *user.User) error { return nil }
func (m *mockUserRepo) GetByID(ctx context.Context, id uint) (*user.User, error) { return nil, nil }
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*user.User, error) { return nil, nil }
func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*user.User, error) { return nil, nil }
func (m *mockUserRepo) GetByResetToken(ctx context.Context, token string) (*user.User, error) { return nil, nil }
func (m *mockUserRepo) Update(ctx context.Context, u *user.User) error { return nil }
func (m *mockUserRepo) List(ctx context.Context) ([]user.User, error) { return nil, nil }
func (m *mockUserRepo) Delete(ctx context.Context, id uint) error { return nil }
func (m *mockUserRepo) AddRole(ctx context.Context, userID uint, roleID uint) error { return nil }
func (m *mockUserRepo) RemoveRole(ctx context.Context, userID uint, roleID uint) error { return nil }
func (m *mockUserRepo) ListRoles(ctx context.Context, userID uint) ([]user.Role, error) { return nil, nil }

type mockCache struct {
	cache.Cache
	store map[string]interface{}
}

func (m *mockCache) SIsMember(ctx context.Context, key string, member string) (bool, error) {
	if m.store == nil { return false, errors.New("not found") }
	val, ok := m.store[key]
	if !ok {
		return false, errors.New("not found")
	}
	set := val.(map[string]struct{})
	_, exists := set[member]
	return exists, nil
}

func (m *mockCache) SAdd(ctx context.Context, key string, members ...string) error {
	if m.store == nil {
		m.store = make(map[string]interface{})
	}
	val, ok := m.store[key]
	var set map[string]struct{}
	if !ok {
		set = make(map[string]struct{})
	} else {
		set = val.(map[string]struct{})
	}
	for _, mm := range members {
		set[mm] = struct{}{}
	}
	m.store[key] = set
	return nil
}

func TestPermissionMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		userID            uint
		requiredPerm       string
		userPerms         []string
		cachedPerms       []string
		wantStatus         int
	}{
		{
			name:         "Access Granted - Direct Match",
			userID:       1,
			requiredPerm: "read",
			userPerms:    []string{"read", "write"},
			wantStatus:   http.StatusOK,
		},
		{
			name:         "Access Denied - Missing Permission",
			userID:       2,
			requiredPerm: "admin",
			userPerms:    []string{"read"},
			wantStatus:   http.StatusForbidden,
		},
		{
			name:         "Access Granted - From Cache",
			userID:       3,
			requiredPerm: "write",
			cachedPerms:  []string{"write"},
			wantStatus:   http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)

			repo := &mockUserRepo{permissions: tt.userPerms}
			csh := &mockCache{store: make(map[string]interface{})}
			
			if len(tt.cachedPerms) > 0 {
				_ = csh.SAdd(context.Background(), fmt.Sprintf("user_perms:%d", tt.userID), tt.cachedPerms...)
			}

			r.Use(func(ctx *gin.Context) {
				ctx.Set("user_id", tt.userID)
				ctx.Next()
			})
			
			r.GET("/test", PermissionMiddleware(repo, csh, tt.requiredPerm), func(ctx *gin.Context) {
				ctx.String(http.StatusOK, "OK")
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("%s: got status %d, want %d", tt.name, w.Code, tt.wantStatus)
			}
		})
	}
}
