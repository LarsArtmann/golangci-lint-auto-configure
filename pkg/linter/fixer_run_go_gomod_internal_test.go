package linter

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
)

func TestReadGoModGoDirective(t *testing.T) {
	tests := []struct {
		name    string
		goMod   string
		want    string
		noGoMod bool
	}{
		{
			name:  "classic directive",
			goMod: "module example.com/foo\n\ngo 1.27\n",
			want:  "1.27",
		},
		{
			name:  "patch-form directive normalizes",
			goMod: "module example.com/foo\n\ngo 1.27.0\n",
			want:  "1.27",
		},
		{
			name:  "go-prefixed directive",
			goMod: "go go1.27.1\n",
			want:  "1.27",
		},
		{
			name:  "toolchain directive is not the go directive",
			goMod: "module example.com/foo\n\ngo 1.26\n\ntoolchain go1.28.1\n",
			want:  "1.26",
		},
		{
			name:  "directive inside require block is ignored",
			goMod: "module example.com/foo\n\ngo 1.26\n\nrequire (\n\tgo 1.99\n)\n",
			want:  "1.26",
		},
		{
			name:  "unparsable directive is silent",
			goMod: "module example.com/foo\n\ngo devel\n",
			want:  "",
		},
		{
			name:  "no directive",
			goMod: "module example.com/foo\n",
			want:  "",
		},
		{
			name:    "missing go.mod",
			noGoMod: true,
			want:    "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()

			if !test.noGoMod {
				if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(test.goMod), 0o600); err != nil {
					t.Fatalf("writing go.mod: %v", err)
				}
			}

			configPath := filepath.Join(dir, ".golangci.yml")

			if got := readGoModGoDirective(configPath); got != test.want {
				t.Fatalf("readGoModGoDirective(%q) = %q, want %q", configPath, got, test.want)
			}
		})
	}
}

func TestGoModExceedsBinary(t *testing.T) {
	tests := []struct {
		directive string
		binary    string
		want      bool
	}{
		{directive: "1.28", binary: "1.27", want: true},
		{directive: "1.27", binary: "1.27", want: false},
		{directive: "1.27.0", binary: "1.27", want: false},
		{directive: "1.26", binary: "1.27", want: false},
		{directive: "2.0", binary: "1.27", want: true},
	}

	for _, test := range tests {
		t.Run(test.directive+" vs "+test.binary, func(t *testing.T) {
			if got := goModExceedsBinary(test.directive, test.binary); got != test.want {
				t.Fatalf("goModExceedsBinary(%q, %q) = %v, want %v", test.directive, test.binary, got, test.want)
			}
		})
	}
}

// TestRescueWarnsOnOverspecifiedGoMod pins the full contract: a go.mod directive
// newer than the binary produces a warning and nothing else — the config is
// untouched, no error is raised, and no ledger entry is recorded.
func TestRescueWarnsOnOverspecifiedGoMod(t *testing.T) {
	dir := t.TempDir()

	goMod := "module example.com/fleet\n\ngo 1.99\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o600); err != nil {
		t.Fatalf("writing go.mod: %v", err)
	}

	configPath := filepath.Join(dir, ".golangci.yml")

	fineConfig := `version: "2"
run:
  timeout: 5m
  go: "1.26"
linters:
  default: standard
`
	if err := os.WriteFile(configPath, []byte(fineConfig), 0o600); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	var logBuf bytes.Buffer

	logger := log.NewWithOptions(&logBuf, log.Options{Level: log.WarnLevel})

	ledger := &rescueRecorder{}
	loader := config.NewLoader(logger)

	fixer := &Fixer{
		configLoader: loader,
		analyzer:     stubRescueAnalyzer{golangciLintGoVersion: "1.26"},
		logger:       logger,
		ledger:       ledger,
	}

	cfg, err := loader.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if err := fixer.rescueOverspecifiedRunGo(cfg, configPath, false); err != nil {
		t.Fatalf("rescueOverspecifiedRunGo returned error for fine run.go: %v", err)
	}

	if cfg.Run.Go != "1.26" {
		t.Fatalf("run.go mutated to %q; go.mod warning must never rewrite the config", cfg.Run.Go)
	}

	if ledger.hasRescuedRunGo() {
		t.Fatal("go.mod warning must not record a rescue ledger entry")
	}

	if !strings.Contains(logBuf.String(), "go.mod declares go 1.99") {
		t.Fatalf("expected go.mod warning in log output, got: %s", logBuf.String())
	}
}
