package user

import (
	"auth-service/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service}
}

func (h *Handler) GetProfile(c *gin.Context) {
	// Example: get ID from param or context
	idStr := c.Param("id")
	if idStr == "" {
		// Fallback for /me route if user ID is in context
		val, exists := c.Get("user_id")
		if exists {
			idStr = strconv.FormatUint(uint64(val.(uint)), 10)
		}
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ValidationErrorResponse(c, utils.MsgInvalidInput, nil)
		return
	}

	res, err := h.service.GetProfile(c.Request.Context(), uint(id))
	if err != nil {
		utils.ErrorResponse(c, 404, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, utils.MsgFetchSuccess, res)
}
