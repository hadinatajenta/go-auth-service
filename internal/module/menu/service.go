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
	GetFullTree(ctx context.Context) ([]MenuTreeResponse, error)
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

	if req.PermissionID > 0 {
		if err := s.repo.SetPermission(ctx, menu.ID, req.PermissionID); err != nil {
			return nil, err
		}
	}

	return s.toResponse(ctx, menu), nil
}

func (s *service) GetByID(ctx context.Context, id uint) (*MenuResponse, error) {
	menu, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toResponse(ctx, menu), nil
}

func (s *service) List(ctx context.Context) ([]MenuResponse, error) {
	menus, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	var res []MenuResponse
	for _, m := range menus {
		res = append(res, *s.toResponse(ctx, &m))
	}

	return res, nil
}

func (s *service) GetUserMenusTree(ctx context.Context, userID uint) ([]MenuTreeResponse, error) {
	menus, err := s.repo.GetAccessibleMenus(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.buildTree(ctx, menus), nil
}

func (s *service) GetFullTree(ctx context.Context) ([]MenuTreeResponse, error) {
	menus, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	return s.buildTree(ctx, menus), nil
}

func (s *service) buildTree(ctx context.Context, menus []Menu) []MenuTreeResponse {
	menuMap := make(map[uint]*MenuTreeResponse)
	for _, m := range menus {
		permID, _ := s.repo.GetPermissionID(ctx, m.ID)
		menuMap[m.ID] = &MenuTreeResponse{
			ID:           m.ID,
			Name:         m.Name,
			Description:  m.Description,
			Path:         m.Path,
			Icon:         m.Icon,
			ParentID:     m.ParentID,
			PermissionID: permID,
			SortOrder:    m.SortOrder,
			Children:     []MenuTreeResponse{},
		}
	}

	childrenMap := make(map[uint][]uint)
	var rootIDs []uint

	for _, m := range menus {
		if m.ParentID == 0 {
			rootIDs = append(rootIDs, m.ID)
		} else {
			if _, exists := menuMap[m.ParentID]; exists {
				childrenMap[m.ParentID] = append(childrenMap[m.ParentID], m.ID)
			} else {
				rootIDs = append(rootIDs, m.ID)
			}
		}
	}

	var buildNode func(id uint) MenuTreeResponse
	buildNode = func(id uint) MenuTreeResponse {
		node := menuMap[id]
		for _, childID := range childrenMap[id] {
			node.Children = append(node.Children, buildNode(childID))
		}
		return *node
	}

	var tree []MenuTreeResponse
	for _, id := range rootIDs {
		tree = append(tree, buildNode(id))
	}

	s.sortMenuTree(tree)
	return tree
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

	if err := s.repo.SetPermission(ctx, menu.ID, req.PermissionID); err != nil {
		return nil, err
	}

	return s.toResponse(ctx, menu), nil
}

func (s *service) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) toResponse(ctx context.Context, menu *Menu) *MenuResponse {
	permID, _ := s.repo.GetPermissionID(ctx, menu.ID)
	return &MenuResponse{
		ID:           menu.ID,
		Name:         menu.Name,
		Description:  menu.Description,
		Path:         menu.Path,
		Icon:         menu.Icon,
		ParentID:     menu.ParentID,
		PermissionID: permID,
		SortOrder:    menu.SortOrder,
		CreatedAt:    menu.CreatedAt,
		UpdatedAt:    menu.UpdatedAt,
	}
}

