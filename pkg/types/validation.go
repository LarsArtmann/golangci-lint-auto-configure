package types

import (
	"github.com/go-playground/validator/v10"
)

// Validator is the global validator instance (lazy-initialized).
var Validator *validator.Validate

// initValidator initializes the global validator instance.
func initValidator() *validator.Validate {
	if Validator == nil {
		Validator = validator.New()
	}

	return Validator
}

// ValidateConfig validates a Config struct using go-playground/validator.
func ValidateConfig(cfg *Config) error {
	v := initValidator()

	return v.Struct(cfg)
}

// ValidateRunConfig validates a RunConfig struct.
func ValidateRunConfig(cfg *RunConfig) error {
	v := initValidator()

	return v.Struct(cfg)
}

// ValidateLintersConfig validates a LintersConfig struct.
func ValidateLintersConfig(cfg *LintersConfig) error {
	v := initValidator()

	return v.Struct(cfg)
}

// ValidationErrors converts validator.ValidationErrors to a slice of ValidationError.
func ValidationErrors(err error) []ValidationError {
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return []ValidationError{
			{
				Field:   "unknown",
				Message: err.Error(),
			},
		}
	}

	errors := make([]ValidationError, 0, len(validationErrors))
	for _, e := range validationErrors {
		errors = append(errors, ValidationError{
			Field:   e.Field(),
			Message: e.Tag(),
		})
	}

	return errors
}

// IsValidationError checks if an error is a validation error.
func IsValidationError(err error) bool {
	if err == nil {
		return false
	}

	_, ok := err.(validator.ValidationErrors)

	return ok
}
