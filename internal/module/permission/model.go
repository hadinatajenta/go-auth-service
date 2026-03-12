package permission

import (
	"time"
)

type Permission struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;unique;not null" json:"name"`
	Description string    `gorm:"size:255" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RolePermission struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	RoleID       uint      `gorm:"uniqueIndex:idx_role_permission" json:"role_id"`
	PermissionID uint      `gorm:"uniqueIndex:idx_role_permission" json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
