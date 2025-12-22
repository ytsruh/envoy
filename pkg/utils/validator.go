package utils

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	validate *validator.Validate
)

func init() {
	validate = validator.New()

	// Register custom validation functions
	validate.RegisterValidation("project_name", validateProjectName)
}

// Validate validates a struct using the validator package
func Validate(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var errorMessages []string
			for _, e := range validationErrors {
				errorMessages = append(errorMessages, formatValidationError(e))
			}
			return fmt.Errorf("%s", strings.Join(errorMessages, "; "))
		}
		return fmt.Errorf("validation failed: %w", err)
	}
	return nil
}

// validateProjectName custom validation for project names
func validateProjectName(fl validator.FieldLevel) bool {
	name := fl.Field().String()
	if len(name) < 1 || len(name) > 100 {
		return false
	}
	// Allow letters, numbers, spaces, hyphens, and underscores
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == ' ' || char == '-' || char == '_') {
			return false
		}
	}
	return true
}

// formatValidationError converts validation errors to user-friendly messages
func formatValidationError(fe validator.FieldError) string {
	field := fe.Field()
	tag := fe.Tag()
	param := fe.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", field, param)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "project_name":
		return fmt.Sprintf("%s must be 1-100 characters and contain only letters, numbers, spaces, hyphens, and underscores", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// GetFieldTagName returns the JSON field name for validation error formatting
func GetFieldTagName(fld reflect.StructField) string {
	tag := fld.Tag.Get("json")
	if tag == "" {
		return fld.Name
	}
	// Remove JSON tags like "omitempty"
	if strings.Contains(tag, ",") {
		parts := strings.Split(tag, ",")
		return parts[0]
	}
	return tag
}
