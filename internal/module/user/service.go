package user

import (
	"auth-service/internal/utils"
	"context"
	"errors"
)

type Service interface {
	GetProfile(ctx context.Context, id uint) (*UserProfileResponse, error)
	Update(ctx context.Context, id uint, req UserUpdateRequest) (*UserResponse, error)
	List(ctx context.Context) ([]UserResponse, error)
	Delete(ctx context.Context, id uint) error
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
