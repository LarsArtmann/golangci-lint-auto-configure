package diff

import (
	"testing"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

func BenchmarkCompare(b *testing.B) {
	differ := NewDiffer()

	old := &types.Config{
		Version: "2",
		Linters: types.LintersConfig{
			Enable: []string{"gosec", "errcheck", "staticcheck", "govet"},
		},
	}

	newConfig := &types.Config{
		Version: "2",
		Linters: types.LintersConfig{
			Enable: []string{"gosec", "errcheck", "staticcheck", "govet", "gocritic", "unused", "ineffassign"},
		},
		Formatters: types.FormattersConfig{
			Enable: []string{"gci", "gofumpt"},
		},
	}

	b.ResetTimer()

	for b.Loop() {
		differ.Compare(old, newConfig)
	}
}

func BenchmarkFormatChanges(b *testing.B) {
	differ := NewDiffer()

	changes := make([]Change, 50)
	for i := range changes {
		changes[i] = Change{
			Type:        ChangeTypeAdded,
			Path:        "linters.enable.linter-" + string(rune('a'+i%26)),
			NewValue:    "linter-" + string(rune('a'+i%26)),
			Description: "Enabled linter: linter-" + string(rune('a'+i%26)),
		}
	}

	b.ResetTimer()

	for b.Loop() {
		differ.FormatChanges(changes)
	}
}
