package permission

import (
	"context"
)

type Service interface {
	Create(ctx context.Context, req PermissionCreateRequest) (*PermissionResponse, error)
	GetByID(ctx context.Context, id uint) (*PermissionResponse, error)
	List(ctx context.Context) ([]PermissionResponse, error)
	Update(ctx context.Context, id uint, req PermissionUpdateRequest) (*PermissionResponse, error)
	Delete(ctx context.Context, id uint) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo}
}

func (s *service) Create(ctx context.Context, req PermissionCreateRequest) (*PermissionResponse, error) {
	perm := &Permission{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.repo.Create(ctx, perm); err != nil {
		return nil, err
	}

	return s.toResponse(perm), nil
}

func (s *service) GetByID(ctx context.Context, id uint) (*PermissionResponse, error) {
	perm, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toResponse(perm), nil
}

func (s *service) List(ctx context.Context) ([]PermissionResponse, error) {
	perms, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	var res []PermissionResponse
	for _, perm := range perms {
		res = append(res, *s.toResponse(&perm))
	}

	return res, nil
}

func (s *service) Update(ctx context.Context, id uint, req PermissionUpdateRequest) (*PermissionResponse, error) {
	perm, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		perm.Name = req.Name
	}
	perm.Description = req.Description

	if err := s.repo.Update(ctx, perm); err != nil {
		return nil, err
	}

	return s.toResponse(perm), nil
}

func (s *service) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) toResponse(perm *Permission) *PermissionResponse {
	return &PermissionResponse{
		ID:          perm.ID,
		Name:        perm.Name,
		Description: perm.Description,
		CreatedAt:   perm.CreatedAt,
		UpdatedAt:   perm.UpdatedAt,
	}
}
