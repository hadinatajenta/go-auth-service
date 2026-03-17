package audit

import (
	"auth-service/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service}
}

func (h *Handler) List(c *gin.Context) {
	userIDStr := c.Query("user_id")
	action := c.Query("action")
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)
	var userID uint
	if userIDStr != "" {
		id, _ := strconv.ParseUint(userIDStr, 10, 32)
		userID = uint(id)
	}

	logs, total, err := h.service.List(c.Request.Context(), userID, action, limit, offset)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch audit logs", err.Error())
		return
	}

	utils.SuccessResponse(c, "Audit logs fetched successfully", gin.H{
		"items": logs,
		"total": total,
		"limit": limit,
		"offset": offset,
	})
}
