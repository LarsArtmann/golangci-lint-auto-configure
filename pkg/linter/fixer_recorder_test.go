package linter

import (
	"testing"
)

func TestConfigChangeRecorder_Normalize(t *testing.T) {
	rec := configChangeRecorder{}

	rec.normalize(func() int { return 3 })
	rec.normalize(func() int { return 5 })
	rec.normalize(func() int { return 0 })

	if rec.counts.normalization != 8 {
		t.Errorf("normalization = %d, want 8", rec.counts.normalization)
	}
}

func TestConfigChangeRecorder_Generated(t *testing.T) {
	rec := configChangeRecorder{}

	rec.generated(func() int { return 2 })
	rec.generated(func() int { return 4 })

	if rec.counts.generated != 6 {
		t.Errorf("generated = %d, want 6", rec.counts.generated)
	}
}

func TestConfigChangeRecorder_PreservesInitialCounts(t *testing.T) {
	initial := fixCounts{
		deprecation: 2,
		enable:      5,
		formatter:   3,
	}

	rec := configChangeRecorder{counts: initial}
	rec.normalize(func() int { return 4 })
	rec.generated(func() int { return 1 })

	if rec.counts.deprecation != 2 {
		t.Errorf("deprecation = %d, want 2 (should be preserved)", rec.counts.deprecation)
	}

	if rec.counts.enable != 5 {
		t.Errorf("enable = %d, want 5 (should be preserved)", rec.counts.enable)
	}

	if rec.counts.formatter != 3 {
		t.Errorf("formatter = %d, want 3 (should be preserved)", rec.counts.formatter)
	}

	if rec.counts.normalization != 4 {
		t.Errorf("normalization = %d, want 4", rec.counts.normalization)
	}

	if rec.counts.generated != 1 {
		t.Errorf("generated = %d, want 1", rec.counts.generated)
	}

	expectedTotal := 2 + 5 + 3 + 4 + 1
	if rec.counts.total() != expectedTotal {
		t.Errorf("total() = %d, want %d", rec.counts.total(), expectedTotal)
	}
}

func TestConfigChangeRecorder_ExecutesMutation(t *testing.T) {
	rec := configChangeRecorder{}

	executed := false

	rec.normalize(func() int {
		executed = true

		return 1
	})

	if !executed {
		t.Error("normalize did not execute the mutation closure")
	}
}
