package detection

import (
	"os"
	"path/filepath"
	"testing"
)

func setupBenchmarkProject(b *testing.B) string {
	tempDir, err := os.MkdirTemp("", "detection-bench-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create go.mod
	goMod := `module test

go 1.21

require (
	github.com/gin-gonic/gin v1.9.0
	github.com/spf13/cobra v1.8.0
)
`
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goMod), 0o644); err != nil {
		b.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create main.go
	mainGo := `package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.Default()
	r.Run()
}
`
	if err := os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(mainGo), 0o644); err != nil {
		b.Fatalf("Failed to write main.go: %v", err)
	}

	return tempDir
}

func BenchmarkDetector_Detect(b *testing.B) {
	tempDir := setupBenchmarkProject(b)
	defer os.RemoveAll(tempDir)

	detector := NewDetector(tempDir)

	for b.Loop() {
		_ = detector.Detect()
	}
}

func BenchmarkDetector_hasMainPackage(b *testing.B) {
	tempDir := setupBenchmarkProject(b)
	defer os.RemoveAll(tempDir)

	detector := NewDetector(tempDir)

	for b.Loop() {
		_ = detector.hasMainPackage()
	}
}

func BenchmarkDetector_analyzeGoMod(b *testing.B) {
	tempDir := setupBenchmarkProject(b)
	defer os.RemoveAll(tempDir)

	detector := NewDetector(tempDir)

	for b.Loop() {
		_, _ = detector.analyzeGoMod()
	}
}

func BenchmarkGetRecommendedLinters(b *testing.B) {
	projectTypes := []ProjectType{
		ProjectTypeCLI,
		ProjectTypeLibrary,
		ProjectTypeWeb,
		ProjectTypeAPI,
	}

	for b.Loop() {
		for _, pt := range projectTypes {
			_ = GetRecommendedLinters(pt)
		}
	}
}
