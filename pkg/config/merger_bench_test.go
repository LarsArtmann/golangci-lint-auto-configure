package config

import (
	"os"
	"path/filepath"
	"testing"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

func BenchmarkMergeConfigInto(b *testing.B) {
	cm := &Merger{
		logger: log.NewWithOptions(nil, log.Options{Level: log.FatalLevel}),
		fs:     osFS{},
	}

	b.ResetTimer()

	for range b.N {
		b.StopTimer()

		primary := &types.Config{
			Version: "2",
			Run: types.RunConfig{
				Timeout: "5m",
			},
			Linters: types.LintersConfig{
				Enable:  []string{"linterE", "linterA", "linterD"},
				Disable: []string{"typecheck"},
			},
		}

		secondary := &types.Config{
			Version: "2",
			Linters: types.LintersConfig{
				Enable: []string{"linterA", "linterB", "linterC"},
				Exclusions: types.LintersExclusionsConfig{
					Paths: []string{"zz_generated"},
				},
			},
			Issues: types.IssuesConfig{
				MaxIssuesPerLinter: 50,
				MaxSameIssues:      10,
			},
		}

		b.StartTimer()

		cm.mergeConfigInto(primary, secondary)
	}
}

func BenchmarkMergeConfigsFiles(b *testing.B) {
	dir := b.TempDir()

	primaryPath := filepath.Join(dir, ".golangci.yml")
	secondaryPath := filepath.Join(dir, ".golangci.yaml")

	err := os.WriteFile(primaryPath, []byte(`version: "2"
run:
  timeout: 5m
linters:
  enable:
    - linterE
    - gosec
`), 0o644)
	if err != nil {
		b.Fatal(err)
	}

	err = os.WriteFile(secondaryPath, []byte(`version: "2"
linters:
  enable:
    - linterB
    - linterC
`), 0o644)
	if err != nil {
		b.Fatal(err)
	}

	cm := NewMerger(log.NewWithOptions(nil, log.Options{Level: log.FatalLevel}))
	paths := []string{primaryPath, secondaryPath}

	b.ResetTimer()

	for range b.N {
		_, _, err := cm.MergeConfigs(paths)
		if err != nil {
			b.Fatal(err)
		}
	}
}
