package auth

import (
	"auth-service/internal/config"
	"auth-service/internal/module/user"
	UserRepository "auth-service/internal/module/user"
	"auth-service/internal/utils"
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Service interface {
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	Register(ctx context.Context, req RegisterRequest) error
	RefreshToken(ctx context.Context, req RefreshRequest) (*LoginResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	LogoutAll(ctx context.Context, userID uint) error
	Introspect(ctx context.Context, req IntrospectRequest) IntrospectResponse
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

	// Check if account is locked
	if u.LockedUntil != nil && u.LockedUntil.After(time.Now()) {
		return nil, errors.New("account is temporarily locked due to too many failed attempts")
	}

	if !utils.CheckPasswordHash(req.Password, u.Password) {
		// Increment login attempts
		u.LoginAttempts++
		if u.LoginAttempts >= 5 {
			lockout := time.Now().Add(time.Minute * 15)
			u.LockedUntil = &lockout
		}
		s.userRepo.Update(ctx, u)
		return nil, errors.New(utils.MsgInvalidCredentials)
	}

	// Reset attempts on successful login
	u.LoginAttempts = 0
	u.LockedUntil = nil

	accessToken, err := utils.GenerateToken(u.ID, s.cfg.JWTSecret)
	if err != nil {
		return nil, err
	}

	refreshToken, err := utils.GenerateRefreshToken(u.ID, s.cfg.JWTSecret)
	if err != nil {
		return nil, err
	}

	// Create session with client metadata
	sess := &UserSession{
		UserID:       u.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IPAddress:    req.IPAddress,
		UserAgent:    req.UserAgent,
		ExpiredAt:    time.Now().Add(time.Hour * 24 * 7),
	}
	if err := s.authRepo.CreateSession(ctx, sess); err != nil {
		return nil, err
	}

	// Update last login
	now := time.Now()
	u.LastLogin = &now
	if err := s.userRepo.Update(ctx, u); err != nil {
		return nil, err
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
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
func (s *service) RefreshToken(ctx context.Context, req RefreshRequest) (*LoginResponse, error) {
	// 1. Find session by refresh token
	sess, err := s.authRepo.GetSessionByRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// 2. Check expiration
	if sess.ExpiredAt.Before(time.Now()) {
		s.authRepo.DeleteSession(ctx, req.RefreshToken)
		return nil, errors.New("refresh token expired")
	}

	// 3. Generate new tokens
	accessToken, err := utils.GenerateToken(sess.UserID, s.cfg.JWTSecret)
	if err != nil {
		return nil, err
	}

	refreshToken, err := utils.GenerateRefreshToken(sess.UserID, s.cfg.JWTSecret)
	if err != nil {
		return nil, err
	}

	// 4. Update session
	s.authRepo.DeleteSession(ctx, req.RefreshToken) // Delete old session
	
	newSess := &UserSession{
		UserID:       sess.UserID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiredAt:    time.Now().Add(time.Hour * 24 * 7),
	}
	
	if err := s.authRepo.CreateSession(ctx, newSess); err != nil {
		return nil, err
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *service) Logout(ctx context.Context, refreshToken string) error {
	// check if session exists
	_, err := s.authRepo.GetSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		return errors.New("session not found")
	}

	return s.authRepo.DeleteSessionByRefreshToken(ctx, refreshToken)
}

func (s *service) LogoutAll(ctx context.Context, userID uint) error {
	return s.authRepo.DeleteAllSessionsByUserID(ctx, userID)
}

func (s *service) Introspect(ctx context.Context, req IntrospectRequest) IntrospectResponse {
	inactive := IntrospectResponse{Active: false}

	// 1. Validate JWT signature + expiry
	token, err := utils.ValidateToken(req.Token, s.cfg.JWTSecret)
	if err != nil || !token.Valid {
		return inactive
	}

	// 2. Extract claims using the correct jwt.MapClaims type
	mc, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return inactive
	}
	userIDRaw, hasUID := mc["user_id"]
	expRaw, hasExp := mc["exp"]
	if !hasUID || !hasExp {
		return inactive
	}
	userIDFloat, ok1 := userIDRaw.(float64)
	expFloat, ok2 := expRaw.(float64)
	if !ok1 || !ok2 {
		return inactive
	}

	// 3. Verify session still exists — guards against post-logout reuse
	if _, err := s.authRepo.GetSessionByAccessToken(ctx, req.Token); err != nil {
		return inactive
	}

	uid := uint(userIDFloat)
	exp := int64(expFloat)
	return IntrospectResponse{Active: true, UserID: &uid, Exp: &exp}
}
