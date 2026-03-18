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
