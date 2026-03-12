package menu

import (
	"context"
)

type Service interface {
	Create(ctx context.Context, req MenuCreateRequest) (*MenuResponse, error)
	GetByID(ctx context.Context, id uint) (*MenuResponse, error)
	List(ctx context.Context) ([]MenuResponse, error)
	Update(ctx context.Context, id uint, req MenuUpdateRequest) (*MenuResponse, error)
	Delete(ctx context.Context, id uint) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo}
}

func (s *service) Create(ctx context.Context, req MenuCreateRequest) (*MenuResponse, error) {
	menu := &Menu{
		Name:        req.Name,
		Description: req.Description,
		Path:        req.Path,
		Icon:        req.Icon,
		ParentID:    req.ParentID,
		SortOrder:   req.SortOrder,
	}

	if err := s.repo.Create(ctx, menu); err != nil {
		return nil, err
	}

	return s.toResponse(menu), nil
}

func (s *service) GetByID(ctx context.Context, id uint) (*MenuResponse, error) {
	menu, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toResponse(menu), nil
}

func (s *service) List(ctx context.Context) ([]MenuResponse, error) {
	menus, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	var res []MenuResponse
	for _, m := range menus {
		res = append(res, *s.toResponse(&m))
	}

	return res, nil
}

func (s *service) Update(ctx context.Context, id uint, req MenuUpdateRequest) (*MenuResponse, error) {
	menu, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		menu.Name = req.Name
	}
	menu.Description = req.Description
	if req.Path != "" {
		menu.Path = req.Path
	}
	menu.Icon = req.Icon
	menu.ParentID = req.ParentID
	menu.SortOrder = req.SortOrder

	if err := s.repo.Update(ctx, menu); err != nil {
		return nil, err
	}

	return s.toResponse(menu), nil
}

func (s *service) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) toResponse(menu *Menu) *MenuResponse {
	return &MenuResponse{
		ID:          menu.ID,
		Name:        menu.Name,
		Description: menu.Description,
		Path:        menu.Path,
		Icon:        menu.Icon,
		ParentID:    menu.ParentID,
		SortOrder:   menu.SortOrder,
		CreatedAt:   menu.CreatedAt,
		UpdatedAt:   menu.UpdatedAt,
	}
}
