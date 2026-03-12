package user

import (
	"auth-service/internal/utils"
	"context"
	"errors"
)

type Service interface {
	GetProfile(ctx context.Context, id uint) (*UserProfileResponse, error)
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

	return &UserProfileResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
	}, nil
}
