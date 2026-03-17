package user

import (
	"auth-service/internal/utils"
	"context"
	"errors"
	"time"
)

type Service interface {
	GetProfile(ctx context.Context, id uint) (*UserProfileResponse, error)
	Update(ctx context.Context, id uint, req UserUpdateRequest) (*UserResponse, error)
	List(ctx context.Context) ([]UserResponse, error)
	Delete(ctx context.Context, id uint) error
	ChangePassword(ctx context.Context, id uint, req ChangePasswordRequest) error
	ForgotPassword(ctx context.Context, req ForgotPasswordRequest) (string, error)
	ResetPassword(ctx context.Context, req ResetPasswordRequest) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo}
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

	if err := s.repo.Update(ctx, u); err != nil {
		return nil, err
	}

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
	return s.repo.Delete(ctx, id)
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
	return s.repo.Update(ctx, u)
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
	u.LoginAttempts = 0 // Reset lockout on password reset
	u.LockedUntil = nil

	return s.repo.Update(ctx, u)
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
