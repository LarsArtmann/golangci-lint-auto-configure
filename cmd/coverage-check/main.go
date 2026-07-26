// Package main implements a portable coverage threshold gate.
// It reads a Go coverage profile and fails if total coverage is below a minimum.
//
// Usage:
//
//	go run ./cmd/coverage-check -min=60 -profile=coverage.out
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	defaultMinCoverage = 60.0
	minFieldsExpected  = 2
)

var (
	errProfileNotFound  = errors.New("coverage profile not found")
	errCoverageBelowMin = errors.New("coverage below minimum threshold")
	errTotalLineFormat  = errors.New("unexpected total line format")
	errNoTotalLine      = errors.New("no total line found in coverage output")
)

func main() {
	threshold := flag.Float64("min", defaultMinCoverage, "minimum coverage percentage")

	profilePath := flag.String("profile", "coverage.out", "path to coverage profile")

	flag.Parse()

	if err := run(*threshold, *profilePath); err != nil {
		fmt.Fprintf(os.Stderr, "❌ %s\n", err)
		os.Exit(1)
	}
}

func run(threshold float64, profilePath string) error { //nolint:erraudit // advisory: errors classified at main exit via errorfamily.ExitCode, not per-function types
	if _, err := os.Stat(profilePath); err != nil {
		return fmt.Errorf(
			"%w: %s — run 'go test -coverprofile=%s ./...' first",
			errProfileNotFound,
			profilePath,
			profilePath,
		)
	}

	total, err := parseTotalCoverage(profilePath)
	if err != nil {
		return fmt.Errorf("failed to parse coverage: %w", err)
	}

	//nolint:forbidigo // CLI tool output to stdout is intentional
	fmt.Printf("Coverage: %.1f%% (minimum: %.0f%%)\n", total, threshold)

	if total < threshold {
		return fmt.Errorf("%w: %.1f%% < %.0f%%", errCoverageBelowMin, total, threshold)
	}

	//nolint:forbidigo // CLI tool output to stdout is intentional
	fmt.Printf("✅ Coverage %.1f%% meets minimum %.0f%%\n", total, threshold)

	return nil
}

func parseTotalCoverage(profilePath string) (float64, error) {
	//nolint:gosec // profilePath comes from trusted flag input
	cmd := exec.CommandContext(context.Background(), "go", "tool", "cover", "-func="+profilePath)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("go tool cover failed: %w", err)
	}

	return parseTotalPercentage(string(output))
}

// parseTotalPercentage scans the output of `go tool cover -func` for the
// `total:` line and returns the coverage percentage. It returns an error if
// the total line is missing, malformed, or contains an unparseable percentage.
func parseTotalPercentage(output string) (float64, error) {
	for line := range strings.SplitSeq(output, "\n") {
		if !strings.HasPrefix(line, "total:") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < minFieldsExpected {
			return 0, fmt.Errorf("%w: %s", errTotalLineFormat, line)
		}

		percentStr := strings.TrimSuffix(fields[len(fields)-1], "%")

		percent, parseErr := strconv.ParseFloat(percentStr, 64)
		if parseErr != nil {
			return 0, fmt.Errorf("failed to parse percentage %q: %w", percentStr, parseErr)
		}

		return percent, nil
	}

	return 0, errNoTotalLine
}
