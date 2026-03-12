package utils

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func FormatValidationError(err error) []ValidationError {
	if ve, ok := err.(validator.ValidationErrors); ok {
		out := make([]ValidationError, len(ve))
		for i, fe := range ve {
			out[i] = ValidationError{
				Field:   fe.Field(),
				Message: getErrorMsg(fe),
			}
		}
		return out
	}
	return nil
}

func getErrorMsg(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return fmt.Sprintf("Minimum length is %s", fe.Param())
	}
	return "Invalid value"
}
