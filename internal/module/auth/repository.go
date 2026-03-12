package auth

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	CreateSession(ctx context.Context, session *UserSession) error
	GetSessionByToken(ctx context.Context, token string) (*UserSession, error)
	DeleteSession(ctx context.Context, token string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) CreateSession(ctx context.Context, session *UserSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *repository) GetSessionByToken(ctx context.Context, token string) (*UserSession, error) {
	var s UserSession
	if err := r.db.WithContext(ctx).Where("token = ?", token).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *repository) DeleteSession(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Where("token = ?", token).Delete(&UserSession{}).Error
}
