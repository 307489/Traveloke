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
		case "min":
			return fmt.Sprintf("%s must be at least %s", fe.Field(), fe.Param())
		case "max":
			return fmt.Sprintf("%s must be at most %s", fe.Field(), fe.Param())
		case "oneof":
			return fmt.Sprintf("%s must be one of: %s", fe.Field(), fe.Param())
		}
		return fmt.Sprintf("%s is invalid", fe.Field())
	}
	return "Invalid request body"
}