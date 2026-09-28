package linter

import (
	"context"
	"os"
	"testing"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

func TestNormalizeGoMajorMinor(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
		ok   bool
	}{
		{name: "go-prefixed patch version", in: "go1.27.1", want: "1.27", ok: true},
		{name: "bare patch version", in: "1.27.1", want: "1.27", ok: true},
		{name: "major.minor only", in: "1.27", want: "1.27", ok: true},
		{name: "older version", in: "1.26.7", want: "1.26", ok: true},
		{name: "empty", in: "", want: "", ok: false},
		{name: "devel keyword", in: "devel", want: "", ok: false},
		{name: "devel pseudo version", in: "go1.28-0f9a9bc", want: "", ok: false},
		{name: "non-numeric minor", in: "1.x", want: "", ok: false},
		{name: "major only", in: "1", want: "", ok: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := normalizeGoMajorMinor(test.in)
			if got != test.want || ok != test.ok {
				t.Fatalf("normalizeGoMajorMinor(%q) = (%q, %t), want (%q, %t)",
					test.in, got, ok, test.want, test.ok)
			}
		})
	}
}

func TestCompareGoMajorMinor(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want int
	}{
		{name: "newer minor", a: "1.27", b: "1.26", want: 1},
		{name: "older minor", a: "1.26", b: "1.27", want: -1},
		{name: "equal", a: "1.27", b: "1.27", want: 0},
		{name: "equal ignoring patch", a: "1.27.0", b: "go1.27.9", want: 0},
		{name: "numeric minor not lexicographic", a: "1.9", b: "1.10", want: -1},
		{name: "newer major", a: "2.0", b: "1.99", want: 1},
		{name: "unparsable compares equal", a: "devel", b: "1.27", want: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := compareGoMajorMinor(test.a, test.b); got != test.want {
				t.Fatalf("compareGoMajorMinor(%q, %q) = %d, want %d", test.a, test.b, got, test.want)
			}
		})
	}
}

func TestParseBuiltWithGoVersion(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "real version output",
			in:   "golangci-lint has version 2.13.2 built with go1.27.1 from v2.13.2 on 1970-01-01T00:00:00Z\n",
			want: "go1.27.1",
		},
		{name: "no go version", in: "golangci-lint has version 2.13.2", want: ""},
		{name: "empty", in: "", want: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := parseBuiltWithGoVersion(test.in); got != test.want {
				t.Fatalf("parseBuiltWithGoVersion(%q) = %q, want %q", test.in, got, test.want)
			}
		})
	}
}

func TestUpdateGoVersion(t *testing.T) {
	logger := log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel})

	tests := []struct {
		name            string
		localGoVersion  string
		golangciLintGo  string
		currentRunGo    string
		wantRunGo       string
		wantChangeCount int
	}{
		{
			name:            "strips patch from local version",
			localGoVersion:  "1.27.1",
			golangciLintGo:  "1.27",
			currentRunGo:    "1.26",
			wantRunGo:       "1.27",
			wantChangeCount: 1,
		},
		{
			name:            "caps at golangci-lint build version",
			localGoVersion:  "1.27.1",
			golangciLintGo:  "1.26",
			currentRunGo:    "1.26",
			wantRunGo:       "1.26",
			wantChangeCount: 0,
		},
		{
			name:            "cap rewrites newer config value",
			localGoVersion:  "1.27.1",
			golangciLintGo:  "1.26",
			currentRunGo:    "1.27",
			wantRunGo:       "1.26",
			wantChangeCount: 1,
		},
		{
			name:            "idempotent when already normalized",
			localGoVersion:  "1.27.1",
			golangciLintGo:  "1.27",
			currentRunGo:    "1.27",
			wantRunGo:       "1.27",
			wantChangeCount: 0,
		},
		{
			name:            "normalizes patch-form config value",
			localGoVersion:  "1.27",
			golangciLintGo:  "1.27",
			currentRunGo:    "1.27.1",
			wantRunGo:       "1.27",
			wantChangeCount: 1,
		},
		{
			name:            "no cap when golangci-lint go version unknown",
			localGoVersion:  "1.27.1",
			golangciLintGo:  "",
			currentRunGo:    "1.26",
			wantRunGo:       "1.27",
			wantChangeCount: 1,
		},
		{
			name:            "unparsable local version leaves config unchanged",
			localGoVersion:  "devel go1.28-0f9a9bc",
			golangciLintGo:  "1.27",
			currentRunGo:    "1.26",
			wantRunGo:       "1.26",
			wantChangeCount: 0,
		},
		{
			name:            "empty local version skips update",
			localGoVersion:  "",
			golangciLintGo:  "1.27",
			currentRunGo:    "1.26",
			wantRunGo:       "1.26",
			wantChangeCount: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			updater := newConfigUpdater(logger, func(context.Context) string {
				return test.localGoVersion
			}, test.golangciLintGo)

			cfg := &types.Config{}
			cfg.Run.Go = test.currentRunGo

			changed := updater.updateGoVersion(context.Background(), cfg)
			if changed != test.wantChangeCount {
				t.Fatalf("change count = %d, want %d", changed, test.wantChangeCount)
			}

			if cfg.Run.Go != test.wantRunGo {
				t.Fatalf("run.go = %q, want %q", cfg.Run.Go, test.wantRunGo)
			}
		})
	}
}
