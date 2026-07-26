package detection

import (
	"bufio"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	errorfamily "github.com/larsartmann/go-error-family"
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

// walkGoFiles walks all .go files in the directory and calls processFile for each.
// Returns early if processFile returns filepath.SkipAll.
func (d *Detector) walkGoFiles(processFile func(*os.File) error) error {
	walkErr := filepath.Walk(d.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			//nolint:nilerr // intentionally skip inaccessible paths during walk
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return errorfamily.WrapTransientf(err, "detector.open_file",
				"opening file %s", path)
		}

		processErr := processFile(file)
		closeFile(file)

		return processErr
	})
	if walkErr != nil {
		return errorfamily.WrapTransientf(walkErr, "detector.walk_dir",
			"walking directory %s", d.rootDir)
	}

	return nil
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

// Preset returns the recommended linter preset for this project type.
// "standard" is the default; "strict" for web/API/monorepo;
// "minimal" for library projects.
func (p ProjectType) Preset() string {
	switch p {
	case ProjectTypeUnknown, ProjectTypeCLI:
		return "standard"
	case ProjectTypeWeb, ProjectTypeAPI, ProjectTypeMonorepo:
		return "strict"
	case ProjectTypeLibrary:
		return "minimal"
	default:
		return "standard"
	}
}

// RecommendPresets analyzes the project and returns a prioritized list of
// recommended presets that can be combined with --preset flags. The first
// element is always the base preset from project type detection. Additional
// presets are added based on detected technologies.
func (d *Detector) RecommendPresets() []string {
	base := d.Detect().Preset()
	recommendations := []string{base}

	if hasSwaggo, err := d.HasSwaggo(); err == nil && hasSwaggo {
		recommendations = append(recommendations, "format")
	}

	projectType := d.Detect()
	if projectType == ProjectTypeWeb || projectType == ProjectTypeAPI {
		recommendations = append(recommendations, "security")
	}

	return recommendations
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

	if d.hasSwaggoConfigFile() {
		return true, nil
	}

	return d.hasSwaggoInCode()
}

// HasClickHouse returns true if the project imports a ClickHouse driver.
func (d *Detector) HasClickHouse() bool {
	_, imports, err := d.analyzeGoModWithError()
	if err != nil {
		return false
	}

	return hasAnyImport(imports, ClickHouseImports)
}

// HasArangoDB returns true if the project imports an ArangoDB driver.
func (d *Detector) HasArangoDB() bool {
	_, imports, err := d.analyzeGoModWithError()
	if err != nil {
		return false
	}

	return hasAnyImport(imports, ArangoDBImports)
}

func hasAnyImport(imports, targets []string) bool {
	for _, imp := range imports {
		for _, target := range targets {
			if strings.Contains(imp, target) {
				return true
			}
		}
	}

	return false
}

func (d *Detector) hasSwaggoConfigFile() bool {
	swaggoConfigFiles := []string{
		"swag.yml",
		"swag.yaml",
		"docs/swagger.yaml",
		"docs/swagger.json",
	}

	for _, f := range swaggoConfigFiles {
		_, err := os.Stat(filepath.Join(d.rootDir, f))
		if err == nil {
			return true
		}
	}

	return false
}

func (d *Detector) detect() ProjectType {
	if d.isMonorepo() {
		return ProjectTypeMonorepo
	}

	modulePath, imports, _ := d.analyzeGoModWithError()
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

func (d *Detector) analyzeGoModWithError() (string, []string, error) {
	goModPath := filepath.Join(d.rootDir, "go.mod")

	file, err := os.Open(goModPath)
	if err != nil {
		return "", nil, errorfamily.WrapTransient(err, "detector.open_gomod", "open go.mod")
	}

	defer closeFile(file)

	scanner := bufio.NewScanner(file)
	info := scanGoMod(scanner)

	err = scanner.Err()
	if err != nil {
		return "", nil, errorfamily.WrapTransient(err, "detector.scan_gomod", "scan go.mod")
	}

	return info.modulePath, info.imports, nil
}

func (d *Detector) hasMainPackage() bool {
	found := false

	if walkErr := d.walkGoFiles(func(file *os.File) error {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "package main") {
				found = true

				return filepath.SkipAll
			}
		}

		if err := scanner.Err(); err != nil {
			// Best-effort by design: swallow per-file scan errors so one
			// pathologically long line does not abort detection of the project.
			slog.Debug("detector: hasMainPackage scan error, skipping file", "error", err)
		}

		return nil
	}); walkErr != nil {
		slog.Debug("detector: hasMainPackage walk failed", "error", walkErr)
	}

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

	if walkErr := d.walkGoFiles(func(file *os.File) error {
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

		if err := scanner.Err(); err != nil {
			// Best-effort by design: swallow per-file scan errors so one
			// pathologically long line does not abort detection of the project.
			slog.Debug("detector: hasAPICodePatterns scan error, skipping file", "error", err)
		}

		return nil
	}); walkErr != nil {
		slog.Debug("detector: hasAPICodePatterns walk failed", "error", walkErr)
	}

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

	walkErr := d.walkGoFiles(func(file *os.File) error {
		return d.scanFileForSwaggo(file, &found)
	})
	if walkErr != nil {
		return false, errorfamily.WrapTransient(walkErr, "detector.walk_failed", "walk directory")
	}

	return found, nil
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

	if err := scanner.Err(); err != nil {
		return errorfamily.WrapTransient(err, "detector.scan_swaggo", "scan swaggo patterns")
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
