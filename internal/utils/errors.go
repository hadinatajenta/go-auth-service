package utils

import (
	"fmt"
	"net/http"
)

type AppError struct {
	Code       int         `json:"-"`
	Message    string      `json:"message"`
	Err        error       `json:"-"`
	Validation interface{} `json:"validation,omitempty"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func NewError(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func NewValidationError(message string, validation interface{}) *AppError {
	return &AppError{
		Code:       http.StatusBadRequest,
		Message:    message,
		Validation: validation,
	}
}

// Common Errors using global constants
var (
	ErrUnauthorized = NewError(http.StatusUnauthorized, MsgUnauthorized, nil)
	ErrForbidden    = NewError(http.StatusForbidden, MsgForbidden, nil)
	ErrNotFound     = NewError(http.StatusNotFound, MsgNotFound, nil)
	ErrInternal     = NewError(http.StatusInternalServerError, MsgInternalError, nil)
	ErrBadRequest   = NewError(http.StatusBadRequest, MsgInvalidInput, nil)
)
