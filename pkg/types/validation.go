package types

import (
	"errors"
	"fmt"

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

	if err := v.Struct(cfg); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}
	return nil
}

// ValidateRunConfig validates a RunConfig struct.
func ValidateRunConfig(cfg *RunConfig) error {
	v := initValidator()

	if err := v.Struct(cfg); err != nil {
		return fmt.Errorf("run config validation failed: %w", err)
	}
	return nil
}

// ValidateLintersConfig validates a LintersConfig struct.
func ValidateLintersConfig(cfg *LintersConfig) error {
	v := initValidator()

	if err := v.Struct(cfg); err != nil {
		return fmt.Errorf("linters config validation failed: %w", err)
	}
	return nil
}

// ValidationErrors converts validator.ValidationErrors to a slice of ValidationError.
func ValidationErrors(err error) []ValidationError {
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	ok := errors.As(err, &validationErrors)
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

	var valErrs validator.ValidationErrors
	ok := errors.As(err, &valErrs)

	return ok
}
