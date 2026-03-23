package auth

import (
	"time"
)

type UserSession struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user_id"`
	AccessToken  string    `gorm:"size:512" json:"access_token"`
	RefreshToken string    `gorm:"size:512" json:"refresh_token"`
	IPAddress    string    `gorm:"size:45" json:"ip_address"`
	UserAgent    string    `gorm:"size:255" json:"user_agent"`
	ExpiredAt    time.Time `json:"expired_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
