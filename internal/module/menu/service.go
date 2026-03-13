package menu

import (
	"context"
	"sort"
)

type Service interface {
	Create(ctx context.Context, req MenuCreateRequest) (*MenuResponse, error)
	GetByID(ctx context.Context, id uint) (*MenuResponse, error)
	List(ctx context.Context) ([]MenuResponse, error)
	GetUserMenusTree(ctx context.Context, userID uint) ([]MenuTreeResponse, error)
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

func (s *service) GetUserMenusTree(ctx context.Context, userID uint) ([]MenuTreeResponse, error) {
	menus, err := s.repo.GetAccessibleMenus(ctx, userID)
	if err != nil {
		return nil, err
	}

	menuMap := make(map[uint]*MenuTreeResponse)
	for _, m := range menus {
		menuMap[m.ID] = &MenuTreeResponse{
			ID:          m.ID,
			Name:        m.Name,
			Description: m.Description,
			Path:        m.Path,
			Icon:        m.Icon,
			ParentID:    m.ParentID,
			SortOrder:   m.SortOrder,
			Children:    []MenuTreeResponse{},
		}
	}

	var tree []MenuTreeResponse

	for _, m := range menus {
		node := menuMap[m.ID]
		if m.ParentID == 0 {
			tree = append(tree, *node)
		} else {
			if parent, exists := menuMap[m.ParentID]; exists {
				parent.Children = append(parent.Children, *node)
			} else {
				tree = append(tree, *node)
			}
		}
	}

	s.sortMenuTree(tree)

	return tree, nil
}

func (s *service) sortMenuTree(tree []MenuTreeResponse) {
	for i := range tree {
		if len(tree[i].Children) > 0 {
			s.sortMenuTree(tree[i].Children)
		}
	}

	sort.Slice(tree, func(i, j int) bool {
		return tree[i].SortOrder < tree[j].SortOrder
	})
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
