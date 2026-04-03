package detection

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type ProjectType int

const (
	ProjectTypeUnknown ProjectType = iota
	ProjectTypeCLI
	ProjectTypeLibrary
	ProjectTypeWeb
	ProjectTypeAPI
	ProjectTypeMonorepo
)

func closeFile(c io.Closer) {
	_ = c.Close()
}

func (p ProjectType) String() string {
	switch p {
	case ProjectTypeUnknown:
		return "Unknown"
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
	}

	return "Unknown"
}

type Detector struct {
	rootDir string
	cache   ProjectType
	cached  bool
	mu      sync.Mutex
}

func NewDetector(rootDir string) *Detector {
	return &Detector{
		rootDir: rootDir,
		cache:   ProjectTypeUnknown,
		cached:  false,
		mu:      sync.Mutex{},
	}
}

func (d *Detector) Detect() ProjectType {
	d.mu.Lock()
	if d.cached {
		d.mu.Unlock()

		return d.cache
	}
	d.mu.Unlock()

	projectType := d.detect()

	d.mu.Lock()
	d.cache = projectType
	d.cached = true
	d.mu.Unlock()

	return projectType
}

func (d *Detector) HasSwaggo() (bool, error) {
	if d.hasSwaggoInGoMod() {
		return true, nil
	}

	return d.hasSwaggoInCode()
}

func (d *Detector) detect() ProjectType {
	if d.isMonorepo() {
		return ProjectTypeMonorepo
	}

	modulePath, imports := d.analyzeGoMod()
	hasMain := d.hasMainPackage()
	hasHTTP := d.hasHTTPFramework(imports)
	hasCLI := d.hasCLIFramework(imports)

	return classifyProject(modulePath, hasMain, hasHTTP, hasCLI, d.hasAPICodePatterns())
}

func classifyProject(modulePath string, hasMain, hasHTTP, hasCLI, hasAPI bool) ProjectType {
	switch {
	case hasHTTP && !hasMain:
		return ProjectTypeLibrary
	case hasHTTP && hasMain:
		return ProjectTypeWeb
	case hasCLI && hasMain:
		return ProjectTypeCLI
	case hasMain && hasAPI:
		return ProjectTypeAPI
	case hasMain:
		return ProjectTypeCLI
	case modulePath != "":
		return ProjectTypeLibrary
	default:
		return ProjectTypeUnknown
	}
}

func (d *Detector) isMonorepo() bool {
	count := 0

	_ = filepath.Walk(d.rootDir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return filepath.SkipDir
		}

		if info.Name() == "go.mod" {
			count++
		}

		return nil
	})

	return count > 1
}

type goModInfo struct {
	modulePath string
	imports    []string
}

func scanGoMod(scanner *bufio.Scanner) goModInfo {
	var info goModInfo

	inRequire := false

	for scanner.Scan() {
		inRequire = processGoModLine(scanner.Text(), &info, inRequire)
	}

	return info
}

func processGoModLine(rawLine string, info *goModInfo, inRequire bool) bool {
	line := strings.TrimSpace(rawLine)

	if strings.HasPrefix(line, "module ") {
		info.modulePath = strings.TrimSpace(strings.TrimPrefix(line, "module"))

		return inRequire
	}

	if line == "require (" {
		return true
	}

	if line == ")" {
		return false
	}

	if imp := extractImportFromLine(line, inRequire); imp != "" {
		info.imports = append(info.imports, imp)
	}

	return inRequire
}

func extractImportFromLine(line string, inRequire bool) string {
	if !inRequire && !strings.HasPrefix(line, "require ") {
		return ""
	}

	fields := strings.Fields(line)
	for i, field := range fields {
		if field == "require" && i == 0 {
			continue
		}

		if strings.Contains(field, "/") {
			return field
		}
	}

	return ""
}

func (d *Detector) analyzeGoMod() (string, []string) {
	goModPath := filepath.Join(d.rootDir, "go.mod")

	file, err := os.Open(goModPath)
	if err != nil {
		return "", nil
	}

	defer closeFile(file)

	info := scanGoMod(bufio.NewScanner(file))

	return info.modulePath, info.imports
}

func (d *Detector) analyzeGoModWithError() (string, []string, error) {
	goModPath := filepath.Join(d.rootDir, "go.mod")

	file, err := os.Open(goModPath)
	if err != nil {
		return "", nil, fmt.Errorf("open go.mod: %w", err)
	}

	defer closeFile(file)

	scanner := bufio.NewScanner(file)
	info := scanGoMod(scanner)

	err = scanner.Err()
	if err != nil {
		return "", nil, fmt.Errorf("scan go.mod: %w", err)
	}

	return info.modulePath, info.imports, nil
}

func (d *Detector) hasMainPackage() bool {
	found := false

	_ = filepath.Walk(d.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer closeFile(file)

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

func (d *Detector) hasHTTPFramework(imports []string) bool {
	for _, imp := range imports {
		for _, framework := range HTTPFrameworks {
			if strings.Contains(imp, framework) {
				return true
			}
		}
	}

	return false
}

func (d *Detector) hasCLIFramework(imports []string) bool {
	for _, imp := range imports {
		for _, framework := range CLIFrameworks {
			if strings.Contains(imp, framework) {
				return true
			}
		}
	}

	return false
}

func (d *Detector) hasAPICodePatterns() bool {
	found := false

	_ = filepath.Walk(d.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer closeFile(file)

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			for _, pattern := range APIPatterns {
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

func (d *Detector) hasSwaggoInGoMod() bool {
	_, imports, err := d.analyzeGoModWithError()
	if err != nil {
		return false
	}

	for _, imp := range imports {
		for _, swaggoImport := range SwaggoImports {
			if strings.Contains(imp, swaggoImport) {
				return true
			}
		}
	}

	return false
}

func (d *Detector) hasSwaggoInCode() (bool, error) {
	found := false

	walkErr := filepath.Walk(d.rootDir, func(path string, info os.FileInfo, err error) error {
		return d.checkFileForSwaggo(path, info, err, &found)
	})
	if walkErr != nil {
	return false, fmt.Errorf("walk directory: %w", walkErr)
	}

	return found, nil
}

func (d *Detector) checkFileForSwaggo(path string, info os.FileInfo, err error, found *bool) error {
	if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer closeFile(file)

	return d.scanFileForSwaggo(file, found)
}

func (d *Detector) scanFileForSwaggo(file *os.File, found *bool) error {
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if d.containsSwaggoPattern(line) {
			*found = true

			return filepath.SkipAll
		}
	}

	return nil
}

func (d *Detector) containsSwaggoPattern(line string) bool {
	for _, pattern := range SwaggoPatterns {
		if strings.Contains(line, pattern) {
			return true
		}
	}

	return false
}
