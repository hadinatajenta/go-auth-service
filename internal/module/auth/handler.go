package auth

import (
	"auth-service/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service}
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, utils.MsgInvalidInput, utils.FormatValidationError(err))
		return
	}

	res, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, utils.MsgLoginSuccess, res)
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, utils.MsgInvalidInput, utils.FormatValidationError(err))
		return
	}

	if err := h.service.Register(c.Request.Context(), req); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MsgInternalError, err.Error())
		return
	}

	utils.SuccessResponse(c, utils.MsgRegisterSuccess, nil)
}
