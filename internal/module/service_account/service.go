package service_account

import (
	"auth-service/internal/utils"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Service interface {
	Create(ctx context.Context, req CreateServiceAccountRequest) (*CreateServiceAccountResponse, error)
	List(ctx context.Context) ([]ServiceAccountResponse, error)
	Revoke(ctx context.Context, id uint) error
	AuthenticateByKey(ctx context.Context, rawKey string) (*ServiceAccount, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo}
}

func (s *service) Create(ctx context.Context, req CreateServiceAccountRequest) (*CreateServiceAccountResponse, error) {
	// Generate key components
	prefix := utils.GenerateRandomString(8)
	secret := utils.GenerateRandomString(32)
	rawKey := fmt.Sprintf("sk_live_%s_%s", prefix, secret)

	// Hash the secret portion only
	hashedSecret, err := utils.HashPassword(secret)
	if err != nil {
		return nil, errors.New("failed to generate API key")
	}

	sa := &ServiceAccount{
		Name:         req.Name,
		Description:  req.Description,
		APIKeyPrefix: prefix,
		APIKeyHash:   hashedSecret,
	}

	if err := s.repo.Create(ctx, sa); err != nil {
		return nil, errors.New("failed to create service account")
	}

	return &CreateServiceAccountResponse{
		ServiceAccountResponse: toResponse(sa),
		APIKey:                 rawKey,
	}, nil
}

func (s *service) List(ctx context.Context) ([]ServiceAccountResponse, error) {
	accounts, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	var responses []ServiceAccountResponse
	for _, a := range accounts {
		responses = append(responses, toResponse(&a))
	}
	return responses, nil
}

func (s *service) Revoke(ctx context.Context, id uint) error {
	return s.repo.Revoke(ctx, id)
}

// AuthenticateByKey validates an API key and returns the service account.
// It expects a key in the format: sk_live_<prefix>_<secret>
func (s *service) AuthenticateByKey(ctx context.Context, rawKey string) (*ServiceAccount, error) {
	// Key format: sk_live_<8-char prefix>_<secret>
	// e.g. sk_live_9f2ab4cd_xxxxx...
	if !strings.HasPrefix(rawKey, "sk_live_") {
		return nil, errors.New("invalid API key format")
	}

	// Strip "sk_live_" then split on first underscore to get prefix and secret
	trimmed := strings.TrimPrefix(rawKey, "sk_live_")
	parts := strings.SplitN(trimmed, "_", 2)
	if len(parts) != 2 {
		return nil, errors.New("invalid API key format")
	}

	prefix := parts[0]
	secret := parts[1]

	sa, err := s.repo.GetByPrefix(ctx, prefix)
	if err != nil {
		return nil, errors.New("invalid API key")
	}

	if sa.Revoked {
		return nil, errors.New("API key has been revoked")
	}

	if !utils.CheckPasswordHash(secret, sa.APIKeyHash) {
		return nil, errors.New("invalid API key")
	}

	// Update last used timestamp asynchronously — do not fail auth on error
	go s.repo.UpdateLastUsed(context.Background(), sa.ID, time.Now())

	return sa, nil
}

func toResponse(sa *ServiceAccount) ServiceAccountResponse {
	return ServiceAccountResponse{
		ID:          sa.ID,
		Name:        sa.Name,
		Description: sa.Description,
		Revoked:     sa.Revoked,
		LastUsedAt:  sa.LastUsedAt,
		CreatedAt:   sa.CreatedAt,
	}
}
