package audit

import (
	"auth-service/internal/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service}
}

// List godoc
// @Summary List audit logs
// @Description Retrieve a paginated list of audit logs with optional filtering
// @Tags audit
// @Accept  json
// @Produce  json
// @Param user_id query int false "User ID filter"
// @Param action query string false "Action filter (CREATE, UPDATE, DELETE)"
// @Param limit query int false "Pagination limit"
// @Param offset query int false "Pagination offset"
// @Success 200 {object} utils.APIResponse{data=object}
// @Failure 500 {object} utils.APIResponse
// @Security Bearer
// @Router /audit-logs [get]
func (h *Handler) List(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Query("userId"), 10, 32)
	action := c.Query("action")
	entity := c.Query("entity")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	var from, to *time.Time
	if fromStr := c.Query("from"); fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			from = &t
		}
	}
	if toStr := c.Query("to"); toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			to = &t
		}
	}

	logs, total, err := h.service.List(c.Request.Context(), uint(userID), action, entity, from, to, limit, offset)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, utils.MsgFetchSuccess, gin.H{
		"logs":  logs,
		"total": total,
	})
}
