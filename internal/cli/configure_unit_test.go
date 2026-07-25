package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

func silentLogger() *log.Logger {
	return log.NewWithOptions(nil, log.Options{Level: log.FatalLevel})
}

func TestEffectiveDryRunForCheckDiff(t *testing.T) {
	cfg := &types.Config{Version: "2"}

	tests := []struct {
		name        string
		isDryRun    bool
		check       bool
		showDiff    bool
		originalCfg *types.Config
		want        bool
	}{
		{
			name:        "check+diff+cfg forces non-dry-run",
			isDryRun:    true,
			check:       true,
			showDiff:    true,
			originalCfg: cfg,
			want:        false,
		},
		{
			name:        "check without diff stays dry-run",
			isDryRun:    true,
			check:       true,
			showDiff:    false,
			originalCfg: cfg,
			want:        true,
		},
		{
			name:        "check+diff but nil cfg stays dry-run",
			isDryRun:    true,
			check:       true,
			showDiff:    true,
			originalCfg: nil,
			want:        true,
		},
		{
			name:        "no check, no diff, isDryRun true",
			isDryRun:    true,
			check:       false,
			showDiff:    false,
			originalCfg: nil,
			want:        true,
		},
		{
			name:        "no check, no diff, isDryRun false",
			isDryRun:    false,
			check:       false,
			showDiff:    false,
			originalCfg: nil,
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := effectiveDryRunForCheckDiff(tt.isDryRun, tt.check, tt.showDiff, tt.originalCfg)
			if got != tt.want {
				t.Errorf("effectiveDryRunForCheckDiff() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolvePresetNoDetect(t *testing.T) {
	logger := silentLogger()

	t.Run("returns preset as-is when detect is false", func(t *testing.T) {
		got := resolvePresets([]string{"strict"}, false, logger)
		if len(got) != 1 || got[0] != "strict" {
			t.Errorf("expected ['strict'], got %v", got)
		}
	})

	t.Run("returns empty when preset empty and no detect", func(t *testing.T) {
		got := resolvePresets(nil, false, logger)
		if len(got) != 0 {
			t.Errorf("expected empty, got %v", got)
		}
	})
}

func TestApplyCheckDiffNoOp(t *testing.T) {
	logger := silentLogger()

	t.Run("shouldDiff false is no-op", func(t *testing.T) {
		applyCheckDiff(false, false, nil, nil, "", logger)
	})

	t.Run("nil originalCfg is no-op", func(t *testing.T) {
		applyCheckDiff(true, false, nil, nil, "", logger)
	})
}

func TestRunFmtUnlessDryDryRun(t *testing.T) {
	t.Run("dry-run does nothing", func(t *testing.T) {
		runFmtUnlessDry(context.Background(), silentLogger(), nil, "", true)
	})
}

func TestCloneConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".golangci.yml")

	err := os.WriteFile(configPath, []byte(`version: "2"
run:
  timeout: 5m
linters:
  enable:
    - errcheck
`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	loader := config.NewLoader(silentLogger())

	clone := cloneConfig(loader, configPath, silentLogger())
	if clone == nil {
		t.Fatal("expected non-nil config")
	}

	if clone.Version != "2" {
		t.Errorf("expected version '2', got %q", clone.Version)
	}

	clone.Version = "modified"

	loadedAgain := cloneConfig(loader, configPath, silentLogger())
	if loadedAgain.Version != "2" {
		t.Error("cloneConfig does not return independent copies")
	}
}

func TestLogNextStepsEmptyNoOp(t *testing.T) {
	logNextSteps(silentLogger(), nil)
	logNextSteps(silentLogger(), []string{})
}

func TestLogNextStepsWithSteps(t *testing.T) {
	logNextSteps(silentLogger(), []string{"step 1", "step 2"})
}

func TestRestoreOriginalConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".golangci.yml")

	err := os.WriteFile(configPath, []byte(`version: "2"`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	loader := config.NewLoader(silentLogger())

	original := cloneConfig(loader, configPath, silentLogger())
	original.Version = "modified-v2"

	restoreOriginalConfig(loader, original, configPath, silentLogger())

	restored := cloneConfig(loader, configPath, silentLogger())
	if restored.Version != "modified-v2" {
		t.Errorf("expected restored version 'modified-v2', got %q", restored.Version)
	}
}

func TestBackupConfigFile_NotExist_NoOp(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".golangci.yml")

	err := backupConfigFile(silentLogger(), configPath)
	if err != nil {
		t.Errorf("backupConfigFile() on nonexistent file should return nil, got %v", err)
	}

	backupPath := configPath + ".bak"

	_, statErr := os.Stat(backupPath)
	if !os.IsNotExist(statErr) {
		t.Error("backupConfigFile() should not create backup for nonexistent config")
	}
}

func TestBackupConfigFile_CreatesBackup(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".golangci.yml")
	content := []byte(`version: "2"`)

	err := os.WriteFile(configPath, content, 0o644)
	if err != nil {
		t.Fatal(err)
	}

	err = backupConfigFile(silentLogger(), configPath)
	if err != nil {
		t.Fatalf("backupConfigFile() error = %v", err)
	}

	backupPath := configPath + ".bak"

	backupData, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("backup file not created or unreadable: %v", err)
	}

	if string(backupData) != string(content) {
		t.Errorf(
			"backup content mismatch: got %q, want %q",
			string(backupData),
			string(content),
		)
	}
}

func TestBackupConfigFile_OverwritesExistingBackup(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".golangci.yml")
	backupPath := configPath + ".bak"

	err := os.WriteFile(configPath, []byte(`version: "2"`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(backupPath, []byte(`old stale content`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	err = backupConfigFile(silentLogger(), configPath)
	if err != nil {
		t.Fatalf("backupConfigFile() error = %v", err)
	}

	backupData, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("backup file unreadable: %v", err)
	}

	if string(backupData) == "old stale content" {
		t.Error("backupConfigFile() did not overwrite stale backup")
	}

	if string(backupData) != `version: "2"` {
		t.Errorf("backup content mismatch after overwrite: got %q", string(backupData))
	}
}

func TestBackupConfigFile_ReadError(t *testing.T) {
	dir := t.TempDir()

	configPath := filepath.Join(dir, "subdir")

	err := os.Mkdir(configPath, 0o755)
	if err != nil {
		t.Fatal(err)
	}

	err = backupConfigFile(silentLogger(), configPath)
	if err == nil {
		t.Error("backupConfigFile() on directory should return read error, got nil")
	}
}

func TestBackupConfigFile_WriteError(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".golangci.yml")
	backupPath := configPath + ".bak"

	err := os.WriteFile(configPath, []byte(`version: "2"`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Mkdir(backupPath, 0o755)
	if err != nil {
		t.Fatal(err)
	}

	err = backupConfigFile(silentLogger(), configPath)
	if err == nil {
		t.Error("backupConfigFile() should fail when .bak is a directory, got nil")
	}
}
