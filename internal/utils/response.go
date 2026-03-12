package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// GeneralResponse is the base function for all responses
func GeneralResponse(c *gin.Context, code int, success bool, message string, data interface{}, meta interface{}, err interface{}) {
	c.JSON(code, APIResponse{
		Success: success,
		Message: message,
		Data:    data,
		Meta:    meta,
		Error:   err,
	})
}

// SuccessResponse sends a 200 OK response
func SuccessResponse(c *gin.Context, message string, data interface{}) {
	GeneralResponse(c, http.StatusOK, true, message, data, nil, nil)
}

// CreatedResponse sends a 201 Created response
func CreatedResponse(c *gin.Context, message string, data interface{}) {
	GeneralResponse(c, http.StatusCreated, true, message, data, nil, nil)
}

// ErrorResponse sends a generic error response
func ErrorResponse(c *gin.Context, code int, message string, err interface{}) {
	GeneralResponse(c, code, false, message, nil, nil, err)
}

// ValidationErrorResponse specifically handles validation issues
func ValidationErrorResponse(c *gin.Context, message string, errors interface{}) {
	ErrorResponse(c, http.StatusBadRequest, message, errors)
}

// AbortWithError is a helper to abort the request with an ErrorResponse
func AbortWithError(c *gin.Context, code int, message string, err interface{}) {
	ErrorResponse(c, code, message, err)
	c.Abort()
}
