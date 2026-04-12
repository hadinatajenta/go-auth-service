package user

import (
	"auth-service/internal/utils"
	"errors"
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

func (h *Handler) Create(c *gin.Context) {
	var req UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, utils.MsgInvalidInput, utils.FormatValidationError(err))
		return
	}

	res, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		code := http.StatusInternalServerError
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			code = appErr.Code
		}
		utils.ErrorResponse(c, code, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, utils.MsgCreateSuccess, res)
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

	_, err := h.service.ForgotPassword(c.Request.Context(), req)
	if err != nil {
		// Always return success to prevent user enumeration
		utils.SuccessResponse(c, "If that email exists, a reset link has been sent.", nil)
		return
	}

	// SECURITY: Never return the reset token in the API response.
	// Deliver it out-of-band (email/SMS) in production.
	utils.SuccessResponse(c, "If that email exists, a reset link has been sent.", nil)
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

// AddRole assigns a role to a user
// @Summary Assign role to user
// @Description Assign a specific role to a user
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param body body UserRoleRequest true "Role assignment request"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.APIResponse
// @Router /users/{id}/roles [post]
func (h *Handler) AddRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationErrorResponse(c, utils.MsgInvalidInput, nil)
		return
	}

	var req UserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, utils.MsgInvalidInput, utils.FormatValidationError(err))
		return
	}

	if err := h.service.AddRole(c.Request.Context(), uint(id), req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "Role assigned successfully", nil)
}

// RemoveRole removes a role from a user
// @Summary Remove role from user
// @Description Remove a specific role from a user
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param roleId path int true "Role ID"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.APIResponse
// @Router /users/{id}/roles/{roleId} [delete]
func (h *Handler) RemoveRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	roleID, err2 := strconv.ParseUint(c.Param("roleId"), 10, 32)
	if err != nil || err2 != nil {
		utils.ValidationErrorResponse(c, utils.MsgInvalidInput, nil)
		return
	}

	if err := h.service.RemoveRole(c.Request.Context(), uint(id), uint(roleID)); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "Role removed successfully", nil)
}

// ListRoles lists roles assigned to a user
// @Summary List user roles
// @Description Get all roles assigned to a specific user
// @Tags users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} utils.APIResponse{data=[]RoleResponse}
// @Router /users/{id}/roles [get]
func (h *Handler) ListRoles(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationErrorResponse(c, utils.MsgInvalidInput, nil)
		return
	}

	res, err := h.service.ListRoles(c.Request.Context(), uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, utils.MsgFetchSuccess, res)
}

// GetPermissions returns effective permissions for a user
// @Summary Get user effective permissions
// @Description Get all permissions assigned to a user including those inherited via role hierarchy
// @Tags users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} utils.APIResponse{data=[]string}
// @Router /users/{id}/permissions [get]
func (h *Handler) GetPermissions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationErrorResponse(c, utils.MsgInvalidInput, nil)
		return
	}

	res, err := h.service.GetPermissions(c.Request.Context(), uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, utils.MsgFetchSuccess, res)
}
// GetMyPermissions returns the effective permissions for the authenticated user.
func (h *Handler) GetMyPermissions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User context not found", nil)
		return
	}

	perms, err := h.service.GetPermissions(c.Request.Context(), userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	if perms == nil {
		perms = []string{}
	}
	utils.SuccessResponse(c, utils.MsgFetchSuccess, gin.H{"permissions": perms})
}

// GetMyRoles returns the roles assigned to the authenticated user.
func (h *Handler) GetMyRoles(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User context not found", nil)
		return
	}

	roles, err := h.service.ListRoles(c.Request.Context(), userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	if roles == nil {
		roles = []RoleResponse{}
	}
	utils.SuccessResponse(c, utils.MsgFetchSuccess, gin.H{"roles": roles})
}
