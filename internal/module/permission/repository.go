package permission

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, permission *Permission) error
	GetByID(ctx context.Context, id uint) (*Permission, error)
	List(ctx context.Context) ([]Permission, error)
	Update(ctx context.Context, permission *Permission) error
	Delete(ctx context.Context, id uint) error
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

func (r *repository) GetByID(ctx context.Context, id uint) (*Permission, error) {
	var permission Permission
	if err := r.db.WithContext(ctx).First(&permission, id).Error; err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *repository) List(ctx context.Context) ([]Permission, error) {
	var permissions []Permission
	if err := r.db.WithContext(ctx).Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *repository) Update(ctx context.Context, permission *Permission) error {
	return r.db.WithContext(ctx).Save(permission).Error
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Permission{}, id).Error
}
