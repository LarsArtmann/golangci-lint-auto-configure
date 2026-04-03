package detection_test

import (
	"os"
	"path/filepath"
	"testing"

	detectionpkg "github.com/larsartmann/golangci-lint-auto-configure/pkg/detection"
)

func setupBenchmarkProject(b *testing.B) string {
	b.Helper()

	tempDir := b.TempDir()

	createBenchmarkGoMod(b, tempDir)
	createBenchmarkMainGo(b, tempDir)

	return tempDir
}

func createBenchmarkGoMod(b *testing.B, tempDir string) {
	goMod := `module test

go 1.21

require (
	github.com/gin-gonic/gin v1.9.0
	github.com/spf13/cobra v1.8.0
)
`

	err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goMod), 0o644)
	if err != nil {
		b.Fatalf("Failed to write go.mod: %v", err)
	}
}

func createBenchmarkMainGo(b *testing.B, tempDir string) {
	mainGo := `package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.Default()
	r.Run()
}
`

	err := os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(mainGo), 0o644)
	if err != nil {
		b.Fatalf("Failed to write main.go: %v", err)
	}
}

func BenchmarkDetector_Detect(b *testing.B) {
	tempDir := setupBenchmarkProject(b)

	// b.TempDir() automatically cleans up
	detector := detectionpkg.NewDetector(tempDir)

	for b.Loop() {
		_ = detector.Detect()
	}
}

func BenchmarkGetRecommendedLinters(b *testing.B) {
	projectTypes := []detectionpkg.ProjectType{
		detectionpkg.ProjectTypeCLI,
		detectionpkg.ProjectTypeLibrary,
		detectionpkg.ProjectTypeWeb,
		detectionpkg.ProjectTypeAPI,
	}

	for b.Loop() {
		for _, projectType := range projectTypes {
			_ = detectionpkg.GetRecommendedLinters(projectType)
		}
	}
}
