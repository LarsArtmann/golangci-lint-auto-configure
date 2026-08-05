package detection_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	detectionpkg "github.com/larsartmann/golangci-lint-auto-configure/pkg/detection"
)

func writeGoMod(dir string, requires ...string) error {
	content := "module test\n\ngo 1.21\n"

	var b strings.Builder
	for _, req := range requires {
		b.WriteString("\nrequire " + req)
	}

	return os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content+b.String()+"\n"), 0o644)
}

func writeGoFile(dir, name, content string) error {
	return os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)
}

type detectTestCase struct {
	name  string
	setup func(string) error
	want  detectionpkg.ProjectType
}

var detectTests = []detectTestCase{
	{name: "CLI", setup: setupCLI, want: detectionpkg.ProjectTypeCLI},
	{name: "Library", setup: setupLibrary, want: detectionpkg.ProjectTypeLibrary},
	{name: "Web", setup: setupWeb, want: detectionpkg.ProjectTypeWeb},
	{name: "Monorepo", setup: setupMonorepo, want: detectionpkg.ProjectTypeMonorepo},
}

func setupCLI(dir string) error {
	return setupProjectWithMain(dir, "github.com/spf13/cobra v1.8.0", "package main\n\nfunc main() {}\n")
}

func setupLibrary(dir string) error {
	err := writeGoMod(dir)
	if err != nil {
		return err
	}

	return writeGoFile(dir, "lib.go", "package test\n\nfunc Hello() string { return \"hello\" }\n")
}

func setupWeb(dir string) error {
	return setupProjectWithMain(dir, "github.com/gin-gonic/gin v1.9.0",
		"package main\n\nimport \"github.com/gin-gonic/gin\"\n\nfunc main() {\n\tr := gin.Default()\n\tr.Run()\n}\n")
}

func setupMonorepo(dir string) error {
	err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0o644)
	if err != nil {
		return err
	}

	sub := filepath.Join(dir, "subproject")
	err = os.MkdirAll(sub, 0o755)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(sub, "go.mod"), []byte("module test/sub\n\ngo 1.21\n"), 0o644)
}

func setupProjectWithMain(dir, require, mainContent string) error {
	err := writeGoMod(dir, require)
	if err != nil {
		return err
	}

	return writeGoFile(dir, "main.go", mainContent)
}

func TestDetector_Detect(t *testing.T) {
	t.Parallel()

	for _, testCase := range detectTests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			err := testCase.setup(dir)
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			if got := detectionpkg.NewDetector(dir).Detect(); got != testCase.want {
				t.Errorf("Detect() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestProjectType_StringAndPreset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		pt           detectionpkg.ProjectType
		stringResult string
		presetResult string
	}{
		{detectionpkg.ProjectTypeCLI, "CLI", "standard"},
		{detectionpkg.ProjectTypeLibrary, "Library", "minimal"},
		{detectionpkg.ProjectTypeWeb, "Web", "strict"},
		{detectionpkg.ProjectTypeAPI, "API", "strict"},
		{detectionpkg.ProjectTypeMonorepo, "Monorepo", "strict"},
		{detectionpkg.ProjectTypeUnknown, "Unknown", "standard"},
	}

	for _, tc := range tests {
		t.Run(tc.stringResult, func(t *testing.T) {
			t.Parallel()

			if got := tc.pt.String(); got != tc.stringResult {
				t.Errorf("String() = %v, want %v", got, tc.stringResult)
			}

			if got := tc.pt.Preset(); got != tc.presetResult {
				t.Errorf("Preset() = %v, want %v", got, tc.presetResult)
			}
		})
	}
}

func TestGetRecommendedLinters(t *testing.T) {
	for _, projType := range allProjectTypes() {
		t.Run(projType.String(), func(t *testing.T) {
			linters := detectionpkg.GetRecommendedLinters(projType)
			if len(linters) == 0 {
				t.Errorf("GetRecommendedLinters(%v) returned empty", projType)
			}

			checkEssentialLinters(t, projType, linters)
		})
	}
}

func TestDetector_HasClickHouse(t *testing.T) {
	t.Parallel()

	t.Run("detects ClickHouse driver in go.mod", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		err := writeGoMod(dir, "github.com/ClickHouse/clickhouse-go/v2 v2.20.0")
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		if !detectionpkg.NewDetector(dir).HasClickHouse() {
			t.Error("HasClickHouse() = false, want true")
		}
	})

	t.Run("returns false for non-ClickHouse project", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		err := writeGoMod(dir, "github.com/gin-gonic/gin v1.9.0")
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		if detectionpkg.NewDetector(dir).HasClickHouse() {
			t.Error("HasClickHouse() = true, want false")
		}
	})

	t.Run("returns false when go.mod is missing", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()

		if detectionpkg.NewDetector(dir).HasClickHouse() {
			t.Error("HasClickHouse() = true, want false")
		}
	})
}

func TestDetector_HasArangoDB(t *testing.T) {
	t.Parallel()

	t.Run("detects ArangoDB driver in go.mod", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		err := writeGoMod(dir, "github.com/arangodb/go-driver v1.6.0")
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		if !detectionpkg.NewDetector(dir).HasArangoDB() {
			t.Error("HasArangoDB() = false, want true")
		}
	})

	t.Run("returns false for non-ArangoDB project", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		err := writeGoMod(dir, "github.com/gin-gonic/gin v1.9.0")
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		if detectionpkg.NewDetector(dir).HasArangoDB() {
			t.Error("HasArangoDB() = true, want false")
		}
	})
}

func TestDetector_HasGoHumanize(t *testing.T) {
	t.Parallel()

	t.Run("detects go-humanize dependency in go.mod", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		err := writeGoMod(dir, "github.com/dustin/go-humanize v1.0.1")
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		if !detectionpkg.NewDetector(dir).HasGoHumanize() {
			t.Error("HasGoHumanize() = false, want true")
		}
	})

	t.Run("returns false for project without go-humanize", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		err := writeGoMod(dir, "github.com/gin-gonic/gin v1.9.0")
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		if detectionpkg.NewDetector(dir).HasGoHumanize() {
			t.Error("HasGoHumanize() = true, want false")
		}
	})
}

func TestDetector_HasSwaggo_PropagatesScannerError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	err := writeGoMod(dir) // no swaggo dependency
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// A line exceeding bufio.MaxScanTokenSize (64 KiB) triggers
	// bufio.ErrTooLong. scanFileForSwaggo must propagate this error
	// through hasSwaggoInCode back to the caller.
	longLine := "// " + strings.Repeat("x", 100_000)
	err = writeGoFile(dir, "overflow.go", "package main\n"+longLine+"\n")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	_, err = detectionpkg.NewDetector(dir).HasSwaggo()
	if err == nil {
		t.Fatal("HasSwaggo() expected error for unscannable .go file, got nil")
	}
}

func TestDetector_DetectResilientToScannerErrorsInSiblingFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	err := writeGoMod(dir, "github.com/spf13/cobra v1.8.0")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// aaa_overflow.go sorts before main.go, so filepath.Walk visits it first.
	// Its >64 KiB line triggers bufio.ErrTooLong. hasMainPackage must swallow
	// the per-file scanner error, continue the walk, and still find "package
	// main" in main.go — without resilience the walk would abort at the bad file.
	longLine := "// " + strings.Repeat("x", 100_000)
	if err := writeGoFile(dir, "aaa_overflow.go", longLine+"\n"); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	if err := writeGoFile(dir, "main.go", "package main\n\nfunc main() {}\n"); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	got := detectionpkg.NewDetector(dir).Detect()
	if got != detectionpkg.ProjectTypeCLI {
		t.Errorf("Detect() = %v, want %v (scanner error in sibling file must not abort detection)",
			got, detectionpkg.ProjectTypeCLI)
	}
}

func allProjectTypes() []detectionpkg.ProjectType {
	return []detectionpkg.ProjectType{
		detectionpkg.ProjectTypeCLI,
		detectionpkg.ProjectTypeLibrary,
		detectionpkg.ProjectTypeWeb,
		detectionpkg.ProjectTypeAPI,
		detectionpkg.ProjectTypeMonorepo,
		detectionpkg.ProjectTypeUnknown,
	}
}

func checkEssentialLinters(t *testing.T, projType detectionpkg.ProjectType, linters []string) {
	t.Helper()

	hasGosec := false
	hasErrcheck := false

	for _, l := range linters {
		if l == "gosec" {
			hasGosec = true
		}

		if l == "errcheck" {
			hasErrcheck = true
		}
	}

	if !hasGosec {
		t.Errorf("GetRecommendedLinters(%v) missing gosec", projType)
	}

	if !hasErrcheck {
		t.Errorf("GetRecommendedLinters(%v) missing errcheck", projType)
	}
}
