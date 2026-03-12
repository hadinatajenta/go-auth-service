package menu

import "time"

type MenuCreateRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=100"`
	Description string `json:"description" binding:"max=255"`
	Path        string `json:"path" binding:"required,max=255"`
	Icon        string `json:"icon" binding:"max=50"`
	ParentID    uint   `json:"parent_id"`
	SortOrder   int    `json:"sort_order"`
}

type MenuUpdateRequest struct {
	Name        string `json:"name" binding:"max=100"`
	Description string `json:"description" binding:"max=255"`
	Path        string `json:"path" binding:"max=255"`
	Icon        string `json:"icon" binding:"max=50"`
	ParentID    uint   `json:"parent_id"`
	SortOrder   int    `json:"sort_order"`
}

type MenuResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Path        string    `json:"path"`
	Icon        string    `json:"icon"`
	ParentID    uint      `json:"parent_id"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
