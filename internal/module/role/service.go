package role

import (
	"context"
)

type Service interface {
	Create(ctx context.Context, req RoleCreateRequest) (*RoleResponse, error)
	GetByID(ctx context.Context, id uint) (*RoleResponse, error)
	List(ctx context.Context) ([]RoleResponse, error)
	Update(ctx context.Context, id uint, req RoleUpdateRequest) (*RoleResponse, error)
	Delete(ctx context.Context, id uint) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo}
}

func (s *service) Create(ctx context.Context, req RoleCreateRequest) (*RoleResponse, error) {
	role := &Role{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.repo.Create(ctx, role); err != nil {
		return nil, err
	}

	return s.toResponse(role), nil
}

func (s *service) GetByID(ctx context.Context, id uint) (*RoleResponse, error) {
	role, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toResponse(role), nil
}

func (s *service) List(ctx context.Context) ([]RoleResponse, error) {
	roles, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	var res []RoleResponse
	for _, role := range roles {
		res = append(res, *s.toResponse(&role))
	}

	return res, nil
}

func (s *service) Update(ctx context.Context, id uint, req RoleUpdateRequest) (*RoleResponse, error) {
	role, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		role.Name = req.Name
	}
	role.Description = req.Description

	if err := s.repo.Update(ctx, role); err != nil {
		return nil, err
	}

	return s.toResponse(role), nil
}

func (s *service) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) toResponse(role *Role) *RoleResponse {
	return &RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}
