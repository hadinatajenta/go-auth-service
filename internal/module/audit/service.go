package audit

import (
	"context"
)

type Service interface {
	Log(ctx context.Context, log *AuditLog) error
	List(ctx context.Context, userID uint, action string, limit, offset int) ([]AuditLog, int64, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo}
}

func (s *service) Log(ctx context.Context, log *AuditLog) error {
	return s.repo.Create(ctx, log)
}

func (s *service) List(ctx context.Context, userID uint, action string, limit, offset int) ([]AuditLog, int64, error) {
	return s.repo.List(ctx, userID, action, limit, offset)
}
