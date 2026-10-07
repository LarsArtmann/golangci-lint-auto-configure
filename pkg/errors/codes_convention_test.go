package apperrors_test

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
)

// codeUsageRegexp matches the code literal (second argument) of every
// errorfamily constructor form used in this codebase: Wrap, Wrapf,
// WrapRejection(f), WrapTransient(f), WrapClassified(f), etc.
var codeUsageRegexp = regexp.MustCompile(`Wrap[A-Za-z]*f?\(\s*[^,]+,\s*"([a-z][a-z0-9_.]+)"`)

// codeShapeRegexp pins the "<domain>[.<subdomain>].<action>" naming convention.
var codeShapeRegexp = regexp.MustCompile(`^[a-z]+(_[a-z]+)?(\.[a-z_]+)+$`)

var scanSkipSuffixes = []string{"_test.go", "_templ.go", "codes.go"}

// scanUsedCodes collects every error-family code literal in non-test source.
func scanUsedCodes(t *testing.T, root string) map[string]string {
	t.Helper()

	used := map[string]string{}

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			name := d.Name()
			if name == "vendor" || name == ".git" || name == "testdata" || name == ".direnv" || name == "bin" {
				return filepath.SkipDir
			}

			return nil
		}

		name := d.Name()
		if !strings.HasSuffix(name, ".go") || slices.Contains(scanSkipSuffixes, name) {
			return nil // skip non-source files by design
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		for _, match := range codeUsageRegexp.FindAllStringSubmatch(string(data), -1) {
			if _, known := used[match[1]]; !known {
				used[match[1]] = path
			}
		}

		return nil
	})
	if err != nil {
		t.Fatalf("scanning for used codes: %v", err)
	}

	return used
}

// TestEveryUsedCodeIsRegistered forbids unregistered codes: a typo in a code
// literal escapes every template lookup and degrades the user-facing error to
// a raw log line. Add new codes to pkg/errors/codes.go in the same change.
func TestEveryUsedCodeIsRegistered(t *testing.T) {
	repoRoot := filepath.Join("..", "..")

	used := scanUsedCodes(t, repoRoot)

	var unregistered []string

	for code, path := range used {
		if !apperrors.RegisteredCodes[code] {
			unregistered = append(unregistered, code+" (first use: "+path+")")
		}
	}

	if len(unregistered) > 0 {
		slices.Sort(unregistered)
		t.Fatalf("unregistered error codes found — add them to RegisteredCodes:\n%s",
			strings.Join(unregistered, "\n"))
	}
}

// TestRegisteredCodesAreUsedAndShaped catches the opposite drift: dead
// registry entries (nothing constructs the code anymore) and codes that
// violate the "<domain>[.<subdomain>].<action>" naming convention.
func TestRegisteredCodesAreUsedAndShaped(t *testing.T) {
	repoRoot := filepath.Join("..", "..")

	used := scanUsedCodes(t, repoRoot)

	var dead, malformed []string

	for code := range apperrors.RegisteredCodes {
		if _, isUsed := used[code]; !isUsed {
			dead = append(dead, code)
		}

		if !codeShapeRegexp.MatchString(code) {
			malformed = append(malformed, code)
		}
	}

	if len(dead) > 0 {
		slices.Sort(dead)
		t.Fatalf("registry entries never used in source — remove or start using them:\n%s",
			strings.Join(dead, "\n"))
	}

	if len(malformed) > 0 {
		slices.Sort(malformed)
		t.Fatalf("registry entries violating the code naming convention:\n%s",
			strings.Join(malformed, "\n"))
	}
}
