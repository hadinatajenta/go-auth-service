package role

import "time"

type RoleCreateRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=50"`
	Description string `json:"description" binding:"max=255"`
	ParentID    *uint  `json:"parent_id"`
}

type RoleUpdateRequest struct {
	Name        string `json:"name" binding:"max=50"`
	Description string `json:"description" binding:"max=255"`
	ParentID    *uint  `json:"parent_id"`
}

type RoleResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ParentID    *uint     `json:"parent_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RolePermissionRequest struct {
	PermissionID uint `json:"permission_id" binding:"required"`
}
