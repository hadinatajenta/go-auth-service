package menu

import (
	"time"
)

type Menu struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"size:255" json:"description"`
	Path        string    `gorm:"size:255" json:"path"`
	Icon        string    `gorm:"size:50" json:"icon"`
	ParentID    uint      `json:"parent_id"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type MenuPermission struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	MenuID       uint      `gorm:"uniqueIndex:idx_menu_permission" json:"menu_id"`
	PermissionID uint      `gorm:"uniqueIndex:idx_menu_permission" json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
