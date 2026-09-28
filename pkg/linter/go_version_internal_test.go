package linter

import (
	"context"
	"os"
	"testing"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

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
