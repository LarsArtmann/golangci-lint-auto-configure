package types

import (
	"errors"
	"fmt"
	"sync"

	"github.com/go-playground/validator/v10"
)

// validatorOnce ensures the global validator is initialized exactly once.
var (
	validatorOnce     sync.Once
	validatorInstance *validator.Validate
)

// initValidator initializes the global validator instance.
func initValidator() *validator.Validate {
	validatorOnce.Do(func() {
		validatorInstance = validator.New()
	})

	return validatorInstance
}

// ValidateStruct validates any struct using go-playground/validator.
func ValidateStruct[T any](cfg *T, name string) error {
	v := initValidator()

	err := v.Struct(cfg)
	if err != nil {
		return fmt.Errorf("%s validation failed: %w", name, err)
	}

	return nil
}

// ValidateConfig validates a Config struct using go-playground/validator.
func ValidateConfig(cfg *Config) error {
	return ValidateStruct(cfg, "config")
}

// ValidateRunConfig validates a RunConfig struct.
func ValidateRunConfig(cfg *RunConfig) error {
	return ValidateStruct(cfg, "run config")
}

// ValidateLintersConfig validates a LintersConfig struct.
func ValidateLintersConfig(cfg *LintersConfig) error {
	return ValidateStruct(cfg, "linters config")
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

	validationErrs := make([]ValidationError, 0, len(validationErrors))
	for _, e := range validationErrors {
		validationErrs = append(validationErrs, ValidationError{
			Field:   e.Field(),
			Message: e.Tag(),
		})
	}

	return validationErrs
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
