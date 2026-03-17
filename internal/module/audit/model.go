package audit

import (
	"time"
)

type AuditLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	Action    string    `gorm:"size:50" json:"action"`
	Entity    string    `gorm:"size:50;index" json:"entity"`
	EntityID  uint      `gorm:"index" json:"entity_id"`
	OldData   string    `gorm:"type:jsonb" json:"old_data"`
	NewData   string    `gorm:"type:jsonb" json:"new_data"`
	RequestID string    `gorm:"size:100;index" json:"request_id"`
	Method    string    `gorm:"size:10" json:"method"`
	Path      string    `gorm:"size:255" json:"path"`
	IPAddress string    `gorm:"size:45" json:"ip_address"`
	UserAgent string    `gorm:"size:255" json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}
