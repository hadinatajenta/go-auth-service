package user

import (
	"auth-service/internal/module/audit"
	"auth-service/internal/utils"
	"auth-service/internal/utils/cache"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type Service interface {
	Create(ctx context.Context, req UserCreateRequest) (*UserResponse, error)
	GetProfile(ctx context.Context, id uint) (*UserProfileResponse, error)
	Update(ctx context.Context, id uint, req UserUpdateRequest) (*UserResponse, error)
	List(ctx context.Context) ([]UserResponse, error)
	Delete(ctx context.Context, id uint) error
	ChangePassword(ctx context.Context, id uint, req ChangePasswordRequest) error
	ForgotPassword(ctx context.Context, req ForgotPasswordRequest) (string, error)
	ResetPassword(ctx context.Context, req ResetPasswordRequest) error
	AddRole(ctx context.Context, userID uint, req UserRoleRequest) error
	RemoveRole(ctx context.Context, userID uint, roleID uint) error
	ListRoles(ctx context.Context, userID uint) ([]RoleResponse, error)
	GetPermissions(ctx context.Context, userID uint) ([]string, error)
}

type service struct {
	repo     Repository
	auditSvc audit.Service
	cache    cache.Cache
}

func NewService(repo Repository, auditSvc audit.Service, cache cache.Cache) Service {
	return &service{repo, auditSvc, cache}
}

func (s *service) Create(ctx context.Context, req UserCreateRequest) (*UserResponse, error) {
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	u := &User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  hashedPassword,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}

	s.logActivity(ctx, "CREATE", "user", u.ID, nil, u)

	return s.toResponse(u), nil
}

func (s *service) GetProfile(ctx context.Context, id uint) (*UserProfileResponse, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New(utils.MsgNotFound)
	}

	var roleNames []string
	for _, r := range u.Roles {
		roleNames = append(roleNames, r.Name)
	}

	return &UserProfileResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Roles:     roleNames,
	}, nil
}

func (s *service) Update(ctx context.Context, id uint, req UserUpdateRequest) (*UserResponse, error) {
	if id == 1 {
		return nil, errors.New("cannot modify system bootstrap administrator")
	}
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New(utils.MsgNotFound)
	}

	if req.FirstName != "" {
		u.FirstName = req.FirstName
	}
	if req.LastName != "" {
		u.LastName = req.LastName
	}

	oldUser := *u

	if err := s.repo.Update(ctx, u); err != nil {
		return nil, err
	}

	// Invalidate permission cache for this user
	_ = s.cache.Delete(ctx, fmt.Sprintf("user_perms:%d", u.ID))

	s.logActivity(ctx, "UPDATE", "user", u.ID, &oldUser, u)

	return s.toResponse(u), nil
}

func (s *service) List(ctx context.Context) ([]UserResponse, error) {
	users, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	var res []UserResponse
	for _, u := range users {
		res = append(res, *s.toResponse(&u))
	}

	return res, nil
}

func (s *service) Delete(ctx context.Context, id uint) error {
	auditCtx := audit.FromContext(ctx)
	if auditCtx != nil {
		if id == auditCtx.UserID {
			return errors.New("cannot delete your own account")
		}
	}

	if id == 1 {
		return errors.New("cannot delete system bootstrap administrator")
	}

	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate permission cache for this user
	_ = s.cache.Delete(ctx, fmt.Sprintf("user_perms:%d", id))

	s.logActivity(ctx, "DELETE", "user", id, u, nil)
	return nil
}

func (s *service) ChangePassword(ctx context.Context, id uint, req ChangePasswordRequest) error {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.New(utils.MsgNotFound)
	}

	if !utils.CheckPasswordHash(req.OldPassword, u.Password) {
		return errors.New("invalid old password")
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	u.Password = hashedPassword
	if err := s.repo.Update(ctx, u); err != nil {
		return err
	}

	// Invalidate permission cache for this user
	_ = s.cache.Delete(ctx, fmt.Sprintf("user_perms:%d", u.ID))

	s.logActivity(ctx, "CHANGE_PASSWORD", "user", u.ID, nil, nil)
	return nil
}

func (s *service) ForgotPassword(ctx context.Context, req ForgotPasswordRequest) (string, error) {
	u, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return "", errors.New(utils.MsgNotFound)
	}

	token := utils.GenerateRandomString(32)
	expires := time.Now().Add(time.Hour * 1)

	u.ResetToken = token
	u.ResetTokenExpires = &expires

	if err := s.repo.Update(ctx, u); err != nil {
		return "", err
	}

	return token, nil
}

func (s *service) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
	u, err := s.repo.GetByResetToken(ctx, req.Token)
	if err != nil {
		return errors.New("invalid or expired reset token")
	}

	if u.ResetTokenExpires != nil && u.ResetTokenExpires.Before(time.Now()) {
		return errors.New("invalid or expired reset token")
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	u.Password = hashedPassword
	u.ResetToken = ""
	u.ResetTokenExpires = nil
	u.LoginAttempts = 0
	u.LockedUntil = nil

	if err := s.repo.Update(ctx, u); err != nil {
		return err
	}

	s.logActivity(ctx, "RESET_PASSWORD", "user", u.ID, nil, nil)
	return nil
}

func (s *service) AddRole(ctx context.Context, userID uint, req UserRoleRequest) error {
	auditCtx := audit.FromContext(ctx)
	if auditCtx != nil && userID == auditCtx.UserID {
		return errors.New("cannot manage roles for your own account")
	}
	if err := s.repo.AddRole(ctx, userID, req.RoleID); err != nil {
		return err
	}

	// Invalidate permission cache for this user
	_ = s.cache.Delete(ctx, fmt.Sprintf("user_perms:%d", userID))

	s.logActivity(ctx, "ADD_ROLE", "user", userID, nil, req)
	return nil
}

func (s *service) RemoveRole(ctx context.Context, userID uint, roleID uint) error {
	auditCtx := audit.FromContext(ctx)
	if auditCtx != nil && userID == auditCtx.UserID {
		return errors.New("cannot manage roles for your own account")
	}
	if userID == 1 && roleID == 1 {
		return errors.New("cannot remove system admin role from bootstrap user")
	}

	if err := s.repo.RemoveRole(ctx, userID, roleID); err != nil {
		return err
	}

	// Invalidate permission cache for this user
	_ = s.cache.Delete(ctx, fmt.Sprintf("user_perms:%d", userID))

	s.logActivity(ctx, "REMOVE_ROLE", "user", userID, nil, map[string]interface{}{"role_id": roleID})
	return nil
}

func (s *service) ListRoles(ctx context.Context, userID uint) ([]RoleResponse, error) {
	roles, err := s.repo.ListRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	var res []RoleResponse
	for _, r := range roles {
		res = append(res, RoleResponse{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
		})
	}
	return res, nil
}

func (s *service) GetPermissions(ctx context.Context, userID uint) ([]string, error) {
	return s.repo.GetUserPermissions(ctx, userID)
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

func (s *service) toResponse(u *User) *UserResponse {
	var roleNames []string
	for _, r := range u.Roles {
		roleNames = append(roleNames, r.Description)
	}

	return &UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Roles:     roleNames,
	}
}
