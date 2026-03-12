package menu

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, menu *Menu) error
	List(ctx context.Context) ([]Menu, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) Create(ctx context.Context, menu *Menu) error {
	return r.db.WithContext(ctx).Create(menu).Error
}

func (r *repository) List(ctx context.Context) ([]Menu, error) {
	var menus []Menu
	if err := r.db.WithContext(ctx).Find(&menus).Error; err != nil {
		return nil, err
	}
	return menus, nil
}
