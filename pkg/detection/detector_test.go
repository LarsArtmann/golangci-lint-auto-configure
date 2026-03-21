package detection

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeGoMod creates a go.mod file with optional require dependencies.
func writeGoMod(dir string, requires ...string) error {
	content := `module test

go 1.21
`

	var contentSb15 strings.Builder
	for _, req := range requires {
		contentSb15.WriteString("\nrequire " + req)
	}

	content += contentSb15.String()

	content += "\n"

	return os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0o644)
}

// writeGoFile creates a Go source file with the given name and content.
func writeGoFile(dir, name, content string) error {
	return os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)
}

func TestDetector_Detect(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(dir string) error
		want        ProjectType
		description string
	}{
		{
			name: "CLI project with cobra",
			setupFunc: func(dir string) error {
				err := writeGoMod(dir, "github.com/spf13/cobra v1.8.0")
				if err != nil {
					return err
				}

				return writeGoFile(dir, "main.go", `package main

func main() {}
`)
			},
			want:        ProjectTypeCLI,
			description: "Should detect CLI project with cobra and main package",
		},
		{
			name: "Library project",
			setupFunc: func(dir string) error {
				err := writeGoMod(dir)
				if err != nil {
					return err
				}

				return writeGoFile(dir, "lib.go", `package test

func Hello() string { return "hello" }
`)
			},
			want:        ProjectTypeLibrary,
			description: "Should detect library project without main",
		},
		{
			name: "Web project with gin",
			setupFunc: func(dir string) error {
				err := writeGoMod(dir, "github.com/gin-gonic/gin v1.9.0")
				if err != nil {
					return err
				}

				return writeGoFile(dir, "main.go", `package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.Default()
	r.Run()
}
`)
			},
			want:        ProjectTypeWeb,
			description: "Should detect web project with gin and main",
		},
		{
			name: "Monorepo project",
			setupFunc: func(dir string) error {
				// Create root go.mod
				err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(`module test

go 1.21
`), 0o644)
				if err != nil {
					return err
				}
				// Create subdirectory with another go.mod
				subDir := filepath.Join(dir, "subproject")

				err = os.MkdirAll(subDir, 0o755)
				if err != nil {
					return err
				}

				return os.WriteFile(filepath.Join(subDir, "go.mod"), []byte(`module test/sub

go 1.21
`), 0o644)
			},
			want:        ProjectTypeMonorepo,
			description: "Should detect monorepo with multiple go.mod files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temp directory (t.TempDir() automatically cleans up)
			tempDir := t.TempDir()

			// Setup test files
			if err := tt.setupFunc(tempDir); err != nil {
				t.Fatalf("Failed to setup test: %v", err)
			}

			// Run detection
			detector := NewDetector(tempDir)
			got := detector.Detect()

			if got != tt.want {
				t.Errorf("Detect() = %v, want %v - %s", got, tt.want, tt.description)
			}
		})
	}
}

func TestProjectType_String(t *testing.T) {
	tests := []struct {
		projectType ProjectType
		want        string
	}{
		{ProjectTypeCLI, "CLI"},
		{ProjectTypeLibrary, "Library"},
		{ProjectTypeWeb, "Web"},
		{ProjectTypeAPI, "API"},
		{ProjectTypeMonorepo, "Monorepo"},
		{ProjectTypeUnknown, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.projectType.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRecommendedLinters(t *testing.T) {
	// Test that each project type returns linters
	projectTypes := []ProjectType{
		ProjectTypeCLI,
		ProjectTypeLibrary,
		ProjectTypeWeb,
		ProjectTypeAPI,
		ProjectTypeMonorepo,
		ProjectTypeUnknown,
	}

	for _, pt := range projectTypes {
		t.Run(pt.String(), func(t *testing.T) {
			linters := GetRecommendedLinters(pt)
			if len(linters) == 0 {
				t.Errorf("GetRecommendedLinters(%v) returned empty list", pt)
			}
			// Check that essential linters are present
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
		})
	}
}
