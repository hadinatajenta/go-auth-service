package service_account

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, sa *ServiceAccount) error
	List(ctx context.Context) ([]ServiceAccount, error)
	GetByPrefix(ctx context.Context, prefix string) (*ServiceAccount, error)
	Revoke(ctx context.Context, id uint) error
	UpdateLastUsed(ctx context.Context, id uint, t time.Time) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) Create(ctx context.Context, sa *ServiceAccount) error {
	return r.db.WithContext(ctx).Create(sa).Error
}

func (r *repository) List(ctx context.Context) ([]ServiceAccount, error) {
	var accounts []ServiceAccount
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *repository) GetByPrefix(ctx context.Context, prefix string) (*ServiceAccount, error) {
	var sa ServiceAccount
	if err := r.db.WithContext(ctx).Where("api_key_prefix = ?", prefix).First(&sa).Error; err != nil {
		return nil, err
	}
	return &sa, nil
}

func (r *repository) Revoke(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&ServiceAccount{}).
		Where("id = ?", id).
		Update("revoked", true).Error
}

func (r *repository) UpdateLastUsed(ctx context.Context, id uint, t time.Time) error {
	return r.db.WithContext(ctx).
		Model(&ServiceAccount{}).
		Where("id = ?", id).
		Update("last_used_at", t).Error
}
