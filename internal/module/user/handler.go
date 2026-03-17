package user

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

func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationErrorResponse(c, utils.MsgInvalidInput, nil)
		return
	}

	var req UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, utils.MsgInvalidInput, err.Error())
		return
	}

	res, err := h.service.Update(c.Request.Context(), uint(id), req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, utils.MsgUpdateSuccess, res)
}

func (h *Handler) List(c *gin.Context) {
	res, err := h.service.List(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, utils.MsgFetchSuccess, res)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationErrorResponse(c, utils.MsgInvalidInput, nil)
		return
	}

	if err := h.service.Delete(c.Request.Context(), uint(id)); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, utils.MsgDeleteSuccess, nil)
}

func (h *Handler) ChangePassword(c *gin.Context) {
	val, _ := c.Get("user_id")
	userID := val.(uint)

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, utils.MsgInvalidInput, utils.FormatValidationError(err))
		return
	}

	if err := h.service.ChangePassword(c.Request.Context(), userID, req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "Password changed successfully", nil)
}

func (h *Handler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, utils.MsgInvalidInput, utils.FormatValidationError(err))
		return
	}

	token, err := h.service.ForgotPassword(c.Request.Context(), req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// In production, send this via email. For now, we return it in the response for demo purposes.
	utils.SuccessResponse(c, "Reset token generated", gin.H{"reset_token": token})
}

func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, utils.MsgInvalidInput, utils.FormatValidationError(err))
		return
	}

	if err := h.service.ResetPassword(c.Request.Context(), req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "Password reset successfully", nil)
}
