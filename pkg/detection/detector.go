package detection

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// ProjectType represents the type of Go project.
type ProjectType int

const (
	ProjectTypeUnknown ProjectType = iota
	ProjectTypeCLI
	ProjectTypeLibrary
	ProjectTypeWeb
	ProjectTypeAPI
	ProjectTypeMonorepo
)

func (p ProjectType) String() string {
	switch p {
	case ProjectTypeCLI:
		return "CLI"
	case ProjectTypeLibrary:
		return "Library"
	case ProjectTypeWeb:
		return "Web"
	case ProjectTypeAPI:
		return "API"
	case ProjectTypeMonorepo:
		return "Monorepo"
	default:
		return "Unknown"
	}
}

// Detector analyzes project structure to determine project type.
type Detector struct {
	rootDir string
}

// NewDetector creates a new project type detector.
func NewDetector(rootDir string) *Detector {
	return &Detector{rootDir: rootDir}
}

// Detect analyzes the project and returns the detected type.
func (d *Detector) Detect() ProjectType {
	// Check for monorepo first (multiple go.mod files)
	if d.isMonorepo() {
		return ProjectTypeMonorepo
	}

	// Analyze go.mod
	modulePath, imports := d.analyzeGoMod()

	// Check for main package
	hasMain := d.hasMainPackage()

	// Check for HTTP frameworks
	hasHTTPFramework := d.hasHTTPFramework(imports)

	// Check for CLI frameworks
	hasCLIFramework := d.hasCLIFramework(imports)

	// Decision logic
	switch {
	case hasHTTPFramework && !hasMain:
		return ProjectTypeLibrary
	case hasHTTPFramework && hasMain:
		return ProjectTypeWeb
	case hasCLIFramework && hasMain:
		return ProjectTypeCLI
	case hasMain && d.hasAPICodePatterns():
		return ProjectTypeAPI
	case hasMain:
		return ProjectTypeCLI
	case modulePath != "" && !hasMain:
		return ProjectTypeLibrary
	default:
		return ProjectTypeUnknown
	}
}

// isMonorepo checks if there are multiple go.mod files.
func (d *Detector) isMonorepo() bool {
	count := 0

	filepath.Walk(d.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.Name() == "go.mod" {
			count++
		}

		return nil
	})

	return count > 1
}

// analyzeGoMod extracts module path and imports from go.mod.
func (d *Detector) analyzeGoMod() (string, []string) {
	goModPath := filepath.Join(d.rootDir, "go.mod")

	file, err := os.Open(goModPath)
	if err != nil {
		return "", nil
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	inRequire := false

	var modulePath string

	var imports []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Extract module path
		if strings.HasPrefix(line, "module ") {
			modulePath = strings.TrimSpace(strings.TrimPrefix(line, "module"))

			continue
		}

		// Track require block
		if line == "require (" {
			inRequire = true

			continue
		}

		if line == ")" {
			inRequire = false

			continue
		}

		// Extract imports
		if inRequire || strings.HasPrefix(line, "require ") {
			// Parse "require package version" or "package version" (in block)
			fields := strings.Fields(line)
			for i, field := range fields {
				if field == "require" && i == 0 {
					continue
				}
				// First non-require field that looks like a package path
				if strings.Contains(field, "/") {
					imports = append(imports, field)

					break
				}
			}
		}
	}

	return modulePath, imports
}

// hasMainPackage checks if there's a main package in the project.
func (d *Detector) hasMainPackage() bool {
	found := false

	filepath.Walk(d.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "package main") {
				found = true

				return filepath.SkipAll
			}
		}

		return nil
	})

	return found
}

// hasHTTPFramework checks if common HTTP frameworks are imported.
func (d *Detector) hasHTTPFramework(imports []string) bool {
	httpFrameworks := []string{
		"github.com/gin-gonic/gin",
		"github.com/labstack/echo",
		"github.com/gofiber/fiber",
		"github.com/gorilla/mux",
		"github.com/go-chi/chi",
		"net/http",
		"github.com/valyala/fasthttp",
		"goa.design/goa",
	}

	for _, imp := range imports {
		for _, framework := range httpFrameworks {
			if strings.Contains(imp, framework) {
				return true
			}
		}
	}

	return false
}

// hasCLIFramework checks if common CLI frameworks are imported.
func (d *Detector) hasCLIFramework(imports []string) bool {
	cliFrameworks := []string{
		"github.com/spf13/cobra",
		"github.com/urfave/cli",
		"github.com/alecthomas/kingpin",
		"github.com/charmbracelet/bubbletea",
		"github.com/charmbracelet/lipgloss",
		"github.com/manifoldco/promptui",
	}

	for _, imp := range imports {
		for _, framework := range cliFrameworks {
			if strings.Contains(imp, framework) {
				return true
			}
		}
	}

	return false
}

// hasAPICodePatterns checks for common API patterns in code.
func (d *Detector) hasAPICodePatterns() bool {
	apiPatterns := []string{
		"json.Marshal",
		"json.Unmarshal",
		"http.Handler",
		"grpc.",
		"proto.",
		"REST",
		"API",
	}

	found := false

	filepath.Walk(d.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			for _, pattern := range apiPatterns {
				if strings.Contains(line, pattern) {
					found = true

					return filepath.SkipAll
				}
			}
		}

		return nil
	})

	return found
}

// GetRecommendedLinters returns recommended linters for a project type.
func GetRecommendedLinters(projectType ProjectType) []string {
	switch projectType {
	case ProjectTypeCLI:
		return []string{
			"gosec", "errcheck", "staticcheck", "govet", "ineffassign",
			"wrapcheck", "errorlint", "gocritic", "nolintlint",
		}
	case ProjectTypeLibrary:
		return []string{
			"gosec", "errcheck", "staticcheck", "govet", "ineffassign",
			"wrapcheck", "errorlint", "gocritic", "musttag",
		}
	case ProjectTypeWeb:
		return []string{
			"gosec", "errcheck", "staticcheck", "govet", "ineffassign",
			"noctx", "bodyclose", "wrapcheck", "errorlint",
		}
	case ProjectTypeAPI:
		return []string{
			"gosec", "errcheck", "staticcheck", "govet", "ineffassign",
			"noctx", "bodyclose", "wrapcheck", "errorlint", "musttag",
		}
	case ProjectTypeMonorepo:
		return []string{
			"gosec", "errcheck", "staticcheck", "govet", "ineffassign",
			"gocritic", "errorlint",
		}
	default:
		return []string{
			"gosec", "errcheck", "staticcheck", "govet", "ineffassign",
		}
	}
}
