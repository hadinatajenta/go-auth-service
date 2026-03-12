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
