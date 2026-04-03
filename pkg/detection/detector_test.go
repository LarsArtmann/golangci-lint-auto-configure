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
	if err := writeGoMod(dir); err != nil {
		return err
	}

	return writeGoFile(dir, "lib.go", "package test\n\nfunc Hello() string { return \"hello\" }\n")
}

func setupWeb(dir string) error {
	return setupProjectWithMain(dir, "github.com/gin-gonic/gin v1.9.0",
		"package main\n\nimport \"github.com/gin-gonic/gin\"\n\nfunc main() {\n\tr := gin.Default()\n\tr.Run()\n}\n")
}

func setupMonorepo(dir string) error {
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0o644); err != nil {
		return err
	}

	sub := filepath.Join(dir, "subproject")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(sub, "go.mod"), []byte("module test/sub\n\ngo 1.21\n"), 0o644)
}

func setupProjectWithMain(dir, require, mainContent string) error {
	if err := writeGoMod(dir, require); err != nil {
		return err
	}

	return writeGoFile(dir, "main.go", mainContent)
}

func TestDetector_Detect(t *testing.T) {
	t.Parallel()

	for _, tc := range detectTests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			if err := tc.setup(dir); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			if got := detectionpkg.NewDetector(dir).Detect(); got != tc.want {
				t.Errorf("Detect() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestProjectType_String(t *testing.T) {
	tests := []struct {
		pt   detectionpkg.ProjectType
		want string
	}{
		{detectionpkg.ProjectTypeCLI, "CLI"},
		{detectionpkg.ProjectTypeLibrary, "Library"},
		{detectionpkg.ProjectTypeWeb, "Web"},
		{detectionpkg.ProjectTypeAPI, "API"},
		{detectionpkg.ProjectTypeMonorepo, "Monorepo"},
		{detectionpkg.ProjectTypeUnknown, "Unknown"},
	}

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			if got := tc.pt.String(); got != tc.want {
				t.Errorf("String() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestGetRecommendedLinters(t *testing.T) {
	for _, pt := range allProjectTypes() {
		t.Run(pt.String(), func(t *testing.T) {
			linters := detectionpkg.GetRecommendedLinters(pt)
			if len(linters) == 0 {
				t.Errorf("GetRecommendedLinters(%v) returned empty", pt)
			}

			checkEssentialLinters(t, pt, linters)
		})
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

func checkEssentialLinters(t *testing.T, pt detectionpkg.ProjectType, linters []string) {
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
		t.Errorf("GetRecommendedLinters(%v) missing gosec", pt)
	}

	if !hasErrcheck {
		t.Errorf("GetRecommendedLinters(%v) missing errcheck", pt)
	}
}
