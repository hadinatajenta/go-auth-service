package permission

import (
	"auth-service/internal/utils/cache"
	"context"
	"strings"
)

type Service interface {
	Create(ctx context.Context, req PermissionCreateRequest) (*PermissionResponse, error)
	GetByID(ctx context.Context, id uint) (*PermissionResponse, error)
	List(ctx context.Context) ([]PermissionResponse, error)
	Update(ctx context.Context, id uint, req PermissionUpdateRequest) (*PermissionResponse, error)
	Delete(ctx context.Context, id uint) error
	GetGrouped(ctx context.Context) (map[string][]string, error)
}

type service struct {
	repo  Repository
	cache cache.Cache
}

func NewService(repo Repository, cache cache.Cache) Service {
	return &service{repo, cache}
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

	// Invalidate all user caches when a permission is updated (affects inheritance)
	_ = s.cache.DeleteByPrefix(ctx, "user_perms:")

	return s.toResponse(perm), nil
}

func (s *service) Delete(ctx context.Context, id uint) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	// Invalidate all user caches when a permission is deleted
	_ = s.cache.DeleteByPrefix(ctx, "user_perms:")
	return nil
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

func (s *service) GetGrouped(ctx context.Context) (map[string][]string, error) {
	perms, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	grouped := make(map[string][]string)
	for _, p := range perms {
		parts := strings.Split(p.Name, ":")
		if len(parts) > 1 {
			module := parts[0]
			grouped[module] = append(grouped[module], p.Name)
		} else {
			grouped["other"] = append(grouped["other"], p.Name)
		}
	}
	return grouped, nil
}
