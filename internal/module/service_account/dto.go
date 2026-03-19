package service_account

import "time"

type CreateServiceAccountRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// ServiceAccountResponse is safe to return — never exposes hash or raw key.
type ServiceAccountResponse struct {
	ID          uint       `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Revoked     bool       `json:"revoked"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// CreateServiceAccountResponse is returned ONCE on creation, containing the raw key.
type CreateServiceAccountResponse struct {
	ServiceAccountResponse
	APIKey string `json:"api_key"`
}
