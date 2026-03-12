package auth

import (
	"auth-service/internal/config"
	"auth-service/internal/module/user"
	UserRepository "auth-service/internal/module/user"
	"auth-service/internal/utils"
	"context"
	"errors"
	"time"
)

type Service interface {
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	Register(ctx context.Context, req RegisterRequest) error
}

type service struct {
	userRepo UserRepository.Repository
	authRepo Repository
	cfg      *config.Config
}

func NewService(userRepo UserRepository.Repository, authRepo Repository, cfg *config.Config) Service {
	return &service{userRepo, authRepo, cfg}
}

func (s *service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	u, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, errors.New(utils.MsgInvalidCredentials)
	}

	if !utils.CheckPasswordHash(req.Password, u.Password) {
		return nil, errors.New(utils.MsgInvalidCredentials)
	}

	token, err := utils.GenerateToken(u.ID, s.cfg.JWTSecret)
	if err != nil {
		return nil, err
	}

	// Create session
	sess := &UserSession{
		UserID:    u.ID,
		Token:     token,
		ExpiredAt: time.Now().Add(time.Hour * 72),
	}
	if err := s.authRepo.CreateSession(ctx, sess); err != nil {
		return nil, err
	}

	// Update last login
	now := time.Now()
	u.LastLogin = &now
	s.userRepo.Update(ctx, u)

	return &LoginResponse{Token: token}, nil
}

func (s *service) Register(ctx context.Context, req RegisterRequest) error {
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return err
	}

	u := &user.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Username:  req.Username,
		Email:     req.Email,
		Password:  hashedPassword,
	}

	return s.userRepo.Create(ctx, u)
}
