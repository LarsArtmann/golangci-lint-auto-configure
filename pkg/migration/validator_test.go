package migration_test

import (
	"fmt"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/migration"
)

// FailingValidator is a mock validator that always fails.
type FailingValidator struct {
	ErrorMessage string
}

func (v FailingValidator) ValidateConfig(_ *migration.Migrator) error {
	if v.ErrorMessage == "" {
		return migration.ErrMockValidationFailed
	}

	return fmt.Errorf("%s", v.ErrorMessage)
}
