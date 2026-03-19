package role

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, role *Role) error
	GetByID(ctx context.Context, id uint) (*Role, error)
	List(ctx context.Context) ([]Role, error)
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id uint) error
	GetEffectivePermissions(ctx context.Context, roleID uint) ([]string, error)
	AddPermission(ctx context.Context, roleID uint, permissionID uint) error
	RemovePermission(ctx context.Context, roleID uint, permissionID uint) error
	ListPermissions(ctx context.Context, roleID uint) ([]map[string]interface{}, error)
	ListUsers(ctx context.Context, roleID uint) ([]map[string]interface{}, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) Create(ctx context.Context, role *Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *repository) GetByID(ctx context.Context, id uint) (*Role, error) {
	var role Role
	if err := r.db.WithContext(ctx).First(&role, id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *repository) List(ctx context.Context) ([]Role, error) {
	var roles []Role
	if err := r.db.WithContext(ctx).Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *repository) Update(ctx context.Context, role *Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Role{}, id).Error
}

func (r *repository) GetEffectivePermissions(ctx context.Context, roleID uint) ([]string, error) {
	var permissions []string

	query := `
		WITH RECURSIVE role_hierarchy AS (
			-- Base case: the initial role
			SELECT id, parent_id, 1 as depth FROM roles WHERE id = ? AND deleted_at IS NULL
			UNION ALL
			-- Recursive step: find parents
			SELECT r.id, r.parent_id, rh.depth + 1
			FROM roles r
			INNER JOIN role_hierarchy rh ON r.id = rh.parent_id
			WHERE r.deleted_at IS NULL AND rh.depth < 10
		)
		SELECT DISTINCT p.name
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		INNER JOIN role_hierarchy rh ON rp.role_id = rh.id
	`

	if err := r.db.WithContext(ctx).Raw(query, roleID).Scan(&permissions).Error; err != nil {
		return nil, err
	}

	return permissions, nil
}

func (r *repository) AddPermission(ctx context.Context, roleID uint, permissionID uint) error {
	return r.db.WithContext(ctx).Table("role_permissions").Create(map[string]interface{}{
		"role_id":       roleID,
		"permission_id": permissionID,
	}).Error
}

func (r *repository) RemovePermission(ctx context.Context, roleID uint, permissionID uint) error {
	return r.db.WithContext(ctx).Table("role_permissions").
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(nil).Error
}

func (r *repository) ListPermissions(ctx context.Context, roleID uint) ([]map[string]interface{}, error) {
	var perms []map[string]interface{}
	err := r.db.WithContext(ctx).
		Table("permissions").
		Select("permissions.*").
		Joins("INNER JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&perms).Error
	return perms, err
}

func (r *repository) ListUsers(ctx context.Context, roleID uint) ([]map[string]interface{}, error) {
	var users []map[string]interface{}
	err := r.db.WithContext(ctx).
		Table("users").
		Select("users.id, users.username, users.email, users.first_name, users.last_name").
		Joins("INNER JOIN user_roles ON user_roles.user_id = users.id").
		Where("user_roles.role_id = ? AND users.deleted_at IS NULL", roleID).
		Find(&users).Error
	return users, err
}
