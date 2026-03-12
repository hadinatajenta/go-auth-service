package permission

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, permission *Permission) error
	List(ctx context.Context) ([]Permission, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) Create(ctx context.Context, permission *Permission) error {
	return r.db.WithContext(ctx).Create(permission).Error
}

func (r *repository) List(ctx context.Context) ([]Permission, error) {
	var permissions []Permission
	if err := r.db.WithContext(ctx).Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}
