package user

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uint) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByResetToken(ctx context.Context, token string) (*User, error)
	Update(ctx context.Context, user *User) error
	List(ctx context.Context) ([]User, error)
	Delete(ctx context.Context, id uint) error
	GetUserPermissions(ctx context.Context, userID uint) ([]string, error)
	AddRole(ctx context.Context, userID uint, roleID uint) error
	RemoveRole(ctx context.Context, userID uint, roleID uint) error
	ListRoles(ctx context.Context, userID uint) ([]Role, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) Create(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *repository) GetByID(ctx context.Context, id uint) (*User, error) {
	var u User
	if err := r.db.WithContext(ctx).Preload("Roles").First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repository) GetByUsername(ctx context.Context, username string) (*User, error) {
	var u User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repository) GetByResetToken(ctx context.Context, token string) (*User, error) {
	var u User
	if err := r.db.WithContext(ctx).Where("reset_token = ?", token).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repository) Update(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *repository) List(ctx context.Context) ([]User, error) {
	var users []User
	if err := r.db.WithContext(ctx).Debug().
		Preload("Roles").
		Select("users.*").
		Joins("LEFT JOIN user_roles ON user_roles.user_id = users.id").
		Joins("LEFT JOIN roles ON roles.id = user_roles.role_id").
		Group("users.id").
		Order("MIN(roles.id) ASC").
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&User{}, id).Error
}

func (r *repository) GetUserPermissions(ctx context.Context, userID uint) ([]string, error) {
	var permissions []string

	query := `
		WITH RECURSIVE role_hierarchy AS (
			-- Base case: roles directly assigned to the user
			SELECT r.id, r.parent_id, 1 as depth
			FROM roles r
			INNER JOIN user_roles ur ON r.id = ur.role_id
			WHERE ur.user_id = ? AND r.deleted_at IS NULL
			
			UNION ALL
			
			-- Recursive step: find parents of assigned roles
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

	if err := r.db.WithContext(ctx).Raw(query, userID).Scan(&permissions).Error; err != nil {
		return nil, err
	}

	return permissions, nil
}

func (r *repository) AddRole(ctx context.Context, userID uint, roleID uint) error {
	return r.db.WithContext(ctx).Table("user_roles").Create(map[string]interface{}{
		"user_id": userID,
		"role_id": roleID,
	}).Error
}

func (r *repository) RemoveRole(ctx context.Context, userID uint, roleID uint) error {
	return r.db.WithContext(ctx).Table("user_roles").
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(nil).Error
}

func (r *repository) ListRoles(ctx context.Context, userID uint) ([]Role, error) {
	var roles []Role
	err := r.db.WithContext(ctx).
		Table("roles").
		Joins("INNER JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND roles.deleted_at IS NULL", userID).
		Find(&roles).Error
	return roles, err
}
