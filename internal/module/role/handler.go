package role

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

// Create godoc
// @Summary Create a new role
// @Description Create a new role with optional hierarchy
// @Tags roles
// @Accept  json
// @Produce  json
// @Param role body RoleCreateRequest true "Role details"
// @Success 200 {object} utils.APIResponse{data=RoleResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Security Bearer
// @Router /roles [post]
func (h *Handler) Create(c *gin.Context) {
	var req RoleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, utils.MsgInvalidInput, err.Error())
		return
	}

	res, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, utils.MsgCreateSuccess, res)
}

// GetByID godoc
// @Summary Get role by ID
// @Description Retrieve a specific role by its unique ID
// @Tags roles
// @Accept  json
// @Produce  json
// @Param id path int true "Role ID"
// @Success 200 {object} utils.APIResponse{data=RoleResponse}
// @Failure 404 {object} utils.APIResponse
// @Security Bearer
// @Router /roles/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationErrorResponse(c, utils.MsgInvalidInput, nil)
		return
	}

	res, err := h.service.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, utils.MsgFetchSuccess, res)
}

// List godoc
// @Summary List roles
// @Description Retrieve a list of all roles
// @Tags roles
// @Accept  json
// @Produce  json
// @Success 200 {object} utils.APIResponse{data=[]RoleResponse}
// @Failure 500 {object} utils.APIResponse
// @Security Bearer
// @Router /roles [get]
func (h *Handler) List(c *gin.Context) {
	res, err := h.service.List(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, utils.MsgFetchSuccess, res)
}

// Update godoc
// @Summary Update role
// @Description Update role details, including parent/hierarchy
// @Tags roles
// @Accept  json
// @Produce  json
// @Param id path int true "Role ID"
// @Param role body RoleUpdateRequest true "Updated role details"
// @Success 200 {object} utils.APIResponse{data=RoleResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Security Bearer
// @Router /roles/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationErrorResponse(c, utils.MsgInvalidInput, nil)
		return
	}

	var req RoleUpdateRequest
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

// Delete godoc
// @Summary Delete role
// @Description Soft-delete a role by ID
// @Tags roles
// @Accept  json
// @Produce  json
// @Param id path int true "Role ID"
// @Success 200 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Security Bearer
// @Router /roles/{id} [delete]
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

// AddPermission assigns a permission to a role
// @Summary Assign permission to role
// @Description Assign a specific permission to a role
// @Tags roles
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Param body body RolePermissionRequest true "Permission assignment request"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.APIResponse
// @Router /roles/{id}/permissions [post]
func (h *Handler) AddPermission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationErrorResponse(c, utils.MsgInvalidInput, nil)
		return
	}

	var req RolePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, utils.MsgInvalidInput, utils.FormatValidationError(err))
		return
	}

	if err := h.service.AddPermission(c.Request.Context(), uint(id), req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "Permission assigned successfully", nil)
}

// RemovePermission removes a permission from a role
// @Summary Remove permission from role
// @Description Remove a specific permission from a role
// @Tags roles
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Param permissionId path int true "Permission ID"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.APIResponse
// @Router /roles/{id}/permissions/{permissionId} [delete]
func (h *Handler) RemovePermission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	permID, err2 := strconv.ParseUint(c.Param("permissionId"), 10, 32)
	if err != nil || err2 != nil {
		utils.ValidationErrorResponse(c, utils.MsgInvalidInput, nil)
		return
	}

	if err := h.service.RemovePermission(c.Request.Context(), uint(id), uint(permID)); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "Permission removed successfully", nil)
}

// ListPermissions lists permissions assigned to a role
// @Summary List role permissions
// @Description Get all permissions assigned directly to a specific role
// @Tags roles
// @Produce json
// @Param id path int true "Role ID"
// @Success 200 {object} utils.APIResponse{data=[]map[string]interface{}}
// @Router /roles/{id}/permissions [get]
func (h *Handler) ListPermissions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationErrorResponse(c, utils.MsgInvalidInput, nil)
		return
	}

	res, err := h.service.ListPermissions(c.Request.Context(), uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, utils.MsgFetchSuccess, res)
}

// ListUsers lists users who have this role
// @Summary List role users
// @Description Get all users assigned to a specific role
// @Tags roles
// @Produce json
// @Param id path int true "Role ID"
// @Success 200 {object} utils.APIResponse{data=[]map[string]interface{}}
// @Router /roles/{id}/users [get]
func (h *Handler) ListUsers(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationErrorResponse(c, utils.MsgInvalidInput, nil)
		return
	}

	res, err := h.service.ListUsers(c.Request.Context(), uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, utils.MsgFetchSuccess, res)
}

// DebugUser returns detailed RBAC information for a user
// @Summary Debug user RBAC
// @Description Get direct roles, inherited roles, and all effective permissions for a user
// @Tags rbac
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} utils.APIResponse
// @Router /rbac/debug/user/{id} [get]
func (h *Handler) DebugUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationErrorResponse(c, utils.MsgInvalidInput, nil)
		return
	}

	res, err := h.service.DebugUser(c.Request.Context(), uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, utils.MsgFetchSuccess, res)
}
