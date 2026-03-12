package permission

import "time"

type PermissionCreateRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=100"`
	Description string `json:"description" binding:"max=255"`
}

type PermissionUpdateRequest struct {
	Name        string `json:"name" binding:"max=100"`
	Description string `json:"description" binding:"max=255"`
}

type PermissionResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
