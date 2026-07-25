// Package main implements a portable coverage threshold gate.
// It reads a Go coverage profile and fails if total coverage is below a minimum.
//
// Usage:
//
//	go run ./cmd/coverage-check [min-percentage] [profile-path]
//
// Defaults: min-percentage=60, profile-path=coverage.out
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	minFlag := flag.Float64("min", 60.0, "minimum coverage percentage")
	profileFlag := flag.String("profile", "coverage.out", "path to coverage profile")
	flag.Parse()

	if err := run(*minFlag, *profileFlag); err != nil {
		fmt.Fprintf(os.Stderr, "❌ %s\n", err)
		os.Exit(1)
	}
}

func run(min float64, profilePath string) error {
	if _, err := os.Stat(profilePath); err != nil {
		return fmt.Errorf(
			"%s not found — run 'go test -coverprofile=%s ./...' first",
			profilePath,
			profilePath,
		)
	}

	total, err := parseTotalCoverage(profilePath)
	if err != nil {
		return fmt.Errorf("failed to parse coverage: %w", err)
	}

	fmt.Printf("Coverage: %.1f%% (minimum: %.0f%%)\n", total, min)

	if total < min {
		return fmt.Errorf("coverage %.1f%% is below minimum %.0f%%", total, min)
	}

	fmt.Printf("✅ Coverage %.1f%% meets minimum %.0f%%\n", total, min)

	return nil
}

func parseTotalCoverage(profilePath string) (float64, error) {
	cmd := exec.Command("go", "tool", "cover", "-func="+profilePath)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("go tool cover failed: %w", err)
	}

	for _, line := range strings.Split(string(output), "\n") {
		if !strings.HasPrefix(line, "total:") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			return 0, fmt.Errorf("unexpected total line format: %s", line)
		}

		percentStr := strings.TrimSuffix(fields[len(fields)-1], "%")
		percent, err := strconv.ParseFloat(percentStr, 64)
		if err != nil {
			return 0, fmt.Errorf("failed to parse percentage %q: %w", percentStr, err)
		}

		return percent, nil
	}

	return 0, fmt.Errorf("no total line found in coverage output")
}
