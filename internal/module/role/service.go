package role

import (
	"context"
	"encoding/json"
	"errors"
	"auth-service/internal/module/audit"
)

type Service interface {
	Create(ctx context.Context, req RoleCreateRequest) (*RoleResponse, error)
	GetByID(ctx context.Context, id uint) (*RoleResponse, error)
	List(ctx context.Context) ([]RoleResponse, error)
	Update(ctx context.Context, id uint, req RoleUpdateRequest) (*RoleResponse, error)
	Delete(ctx context.Context, id uint) error
}

type service struct {
	repo      Repository
	auditSvc  audit.Service
}

func NewService(repo Repository, auditSvc audit.Service) Service {
	return &service{repo, auditSvc}
}

func (s *service) Create(ctx context.Context, req RoleCreateRequest) (*RoleResponse, error) {
	role := &Role{
		Name:        req.Name,
		Description: req.Description,
		ParentID:    req.ParentID,
	}

	if err := s.repo.Create(ctx, role); err != nil {
		return nil, err
	}

	// Audit Log
	s.logActivity(ctx, "CREATE", "role", role.ID, nil, role)

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

	oldRole := *role

	if req.ParentID != nil {
		// Cycle detection
		if *req.ParentID == role.ID {
			return nil, errors.New("a role cannot be its own parent")
		}
		if s.hasCycle(ctx, *req.ParentID, role.ID) {
			return nil, errors.New("circular dependency detected in role hierarchy")
		}
		role.ParentID = req.ParentID
	} else if req.ParentID == nil {
		role.ParentID = nil
	}

	if req.Name != "" {
		role.Name = req.Name
	}
	role.Description = req.Description

	if err := s.repo.Update(ctx, role); err != nil {
		return nil, err
	}

	// Audit Log
	s.logActivity(ctx, "UPDATE", "role", role.ID, &oldRole, role)

	return s.toResponse(role), nil
}

func (s *service) Delete(ctx context.Context, id uint) error {
	role, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Audit Log
	s.logActivity(ctx, "DELETE", "role", id, role, nil)

	return nil
}

func (s *service) hasCycle(ctx context.Context, parentID, childID uint) bool {
	if parentID == 0 {
		return false
	}

	parent, err := s.repo.GetByID(ctx, parentID)
	if err != nil || parent.ParentID == nil {
		return false
	}

	if *parent.ParentID == childID {
		return true
	}

	return s.hasCycle(ctx, *parent.ParentID, childID)
}

func (s *service) logActivity(ctx context.Context, action, entity string, entityID uint, oldData, newData interface{}) {
	auditCtx := audit.FromContext(ctx)
	if auditCtx == nil {
		return
	}

	var oldJSON, newJSON string
	if oldData != nil {
		b, _ := json.Marshal(oldData)
		oldJSON = string(b)
	}
	if newData != nil {
		b, _ := json.Marshal(newData)
		newJSON = string(b)
	}

	log := &audit.AuditLog{
		UserID:    auditCtx.UserID,
		Action:    action,
		Entity:    entity,
		EntityID:  entityID,
		OldData:   oldJSON,
		NewData:   newJSON,
		RequestID: auditCtx.RequestID,
		Method:    auditCtx.Method,
		Path:      auditCtx.Path,
		IPAddress: auditCtx.IPAddress,
		UserAgent: auditCtx.UserAgent,
	}

	_ = s.auditSvc.Log(ctx, log)
}

func (s *service) toResponse(role *Role) *RoleResponse {
	return &RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		ParentID:    role.ParentID,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}
