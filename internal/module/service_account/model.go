package service_account

import "time"

type ServiceAccount struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Name         string     `gorm:"size:100;not null" json:"name"`
	Description  string     `gorm:"size:255" json:"description"`
	APIKeyPrefix string     `gorm:"column:api_key_prefix;size:16;uniqueIndex;not null" json:"-"`
	APIKeyHash   string     `gorm:"column:api_key_hash;size:255;not null" json:"-"`
	Revoked      bool       `gorm:"not null;default:false" json:"revoked"`
	LastUsedAt   *time.Time `json:"last_used_at"`
	CreatedAt    time.Time  `json:"created_at"`
}
