package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator wraps the go-playground/validator instance with application-level
// convenience methods.
type Validator struct {
	validate *validator.Validate
}

// New creates a new Validator with default struct-level validation rules.
func New() *Validator {
	v := validator.New()

	// Register custom validations here as needed.
	// v.RegisterValidation("custom_tag", customValidationFunc)

	return &Validator{validate: v}
}

// Validate checks the given struct against its validation tags.
// It returns a map of field names to error messages, or nil if validation passes.
func (v *Validator) Validate(s any) map[string]string {
	err := v.validate.Struct(s)
	if err == nil {
		return nil
	}

	errors := make(map[string]string)
	for _, e := range err.(validator.ValidationErrors) {
		field := e.Field()
		errors[field] = formatValidationError(e)
	}

	return errors
}

// ValidateVar validates a single variable against a tag.
func (v *Validator) ValidateVar(field any, tag string) error {
	return v.validate.Var(field, tag)
}

// formatValidationError converts a validator field error into a human-readable message.
func formatValidationError(e validator.FieldError) string {
	field := e.Field()

	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, e.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, e.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, e.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, e.Param())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	default:
		return fmt.Sprintf("%s failed %s validation", field, e.Tag())
	}
}

// FormatErrors converts a map of field errors into a single string suitable
// for use as an error message (e.g., in AppError).
func FormatErrors(errors map[string]string) string {
	parts := make([]string, 0, len(errors))
	for _, msg := range errors {
		parts = append(parts, msg)
	}
	return strings.Join(parts, "; ")
}
