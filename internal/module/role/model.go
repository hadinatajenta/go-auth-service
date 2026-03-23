package role

import (
	"time"

	"gorm.io/gorm"
)

type Role struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:50;unique;not null" json:"name"`
	Description string         `gorm:"size:255" json:"description"`
	ParentID    *uint          `json:"parent_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type UserRole struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"uniqueIndex:idx_user_role;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user_id"`
	RoleID    uint      `gorm:"uniqueIndex:idx_user_role;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
