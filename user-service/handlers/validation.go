package handlers

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

func formatValidationError(err error) string {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		fe := ve[0]
		switch fe.Tag() {
		case "required":
			return fmt.Sprintf("%s is required", fe.Field())
		case "email":
			return "Invalid email format"
		case "min":
			return fmt.Sprintf("%s must be at least %s characters", fe.Field(), fe.Param())
		case "max":
			return fmt.Sprintf("%s must be at most %s characters", fe.Field(), fe.Param())
		case "alphanum":
			return fmt.Sprintf("%s must contain only letters and numbers", fe.Field())
		case "numeric":
			return fmt.Sprintf("%s must contain only digits", fe.Field())
		}
		return fmt.Sprintf("%s is invalid", fe.Field())
	}
	return "Invalid request body"
}