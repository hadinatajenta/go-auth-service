package service_account

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
// @Summary Create a service account
// @Description Create a new service account and return its API key (returned only once)
// @Tags service-accounts
// @Accept json
// @Produce json
// @Param body body CreateServiceAccountRequest true "Service account details"
// @Success 201 {object} utils.APIResponse{data=CreateServiceAccountResponse}
// @Failure 400 {object} utils.APIResponse
// @Security Bearer
// @Router /service-accounts [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateServiceAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, utils.MsgInvalidInput, utils.FormatValidationError(err))
		return
	}

	res, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "Service account created. Store the API key — it will not be shown again.", res)
}

// List godoc
// @Summary List all service accounts
// @Description Retrieve all service accounts (api_key is never returned)
// @Tags service-accounts
// @Produce json
// @Success 200 {object} utils.APIResponse{data=[]ServiceAccountResponse}
// @Security Bearer
// @Router /service-accounts [get]
func (h *Handler) List(c *gin.Context) {
	res, err := h.service.List(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MsgInternalError, nil)
		return
	}

	utils.SuccessResponse(c, utils.MsgFetchSuccess, res)
}

// Revoke godoc
// @Summary Revoke a service account
// @Description Mark a service account as revoked. All future authentication attempts will be rejected.
// @Tags service-accounts
// @Produce json
// @Param id path int true "Service Account ID"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.APIResponse
// @Security Bearer
// @Router /service-accounts/{id}/revoke [post]
func (h *Handler) Revoke(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationErrorResponse(c, utils.MsgInvalidInput, nil)
		return
	}

	if err := h.service.Revoke(c.Request.Context(), uint(id)); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MsgInternalError, nil)
		return
	}

	utils.SuccessResponse(c, "Service account revoked successfully", nil)
}
