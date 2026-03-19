package auth

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	CreateSession(ctx context.Context, session *UserSession) error
	GetSessionByAccessToken(ctx context.Context, token string) (*UserSession, error)
	GetSessionByRefreshToken(ctx context.Context, token string) (*UserSession, error)
	DeleteSession(ctx context.Context, token string) error
	DeleteSessionByRefreshToken(ctx context.Context, token string) error
	DeleteAllSessionsByUserID(ctx context.Context, userID uint) error
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

func (r *repository) GetSessionByAccessToken(ctx context.Context, token string) (*UserSession, error) {
	var s UserSession
	if err := r.db.WithContext(ctx).Where("access_token = ?", token).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *repository) GetSessionByRefreshToken(ctx context.Context, token string) (*UserSession, error) {
	var s UserSession
	if err := r.db.WithContext(ctx).Where("refresh_token = ?", token).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *repository) DeleteSession(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Where("access_token = ?", token).Or("refresh_token = ?", token).Delete(&UserSession{}).Error
}

func (r *repository) DeleteSessionByRefreshToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Where("refresh_token = ?", token).Delete(&UserSession{}).Error
}

func (r *repository) DeleteAllSessionsByUserID(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&UserSession{}).Error
}
