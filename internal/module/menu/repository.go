package menu

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, menu *Menu) error
	GetByID(ctx context.Context, id uint) (*Menu, error)
	List(ctx context.Context) ([]Menu, error)
	Update(ctx context.Context, menu *Menu) error
	Delete(ctx context.Context, id uint) error
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

func (r *repository) GetByID(ctx context.Context, id uint) (*Menu, error) {
	var menu Menu
	if err := r.db.WithContext(ctx).First(&menu, id).Error; err != nil {
		return nil, err
	}
	return &menu, nil
}

func (r *repository) List(ctx context.Context) ([]Menu, error) {
	var menus []Menu
	if err := r.db.WithContext(ctx).Find(&menus).Error; err != nil {
		return nil, err
	}
	return menus, nil
}

func (r *repository) Update(ctx context.Context, menu *Menu) error {
	return r.db.WithContext(ctx).Save(menu).Error
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Menu{}, id).Error
}
