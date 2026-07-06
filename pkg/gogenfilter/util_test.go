package gogenfilter

import (
	"testing"
	"testing/fstest"
)

func TestMergeExclusionPaths(t *testing.T) {
	t.Run("merges without duplicates", func(t *testing.T) {
		existing := []string{`\.pb\.go$`, `_templ\.go$`}
		newPaths := []string{`\.pb\.go$`, `wire_gen\.go$`}

		result := MergeExclusionPaths(existing, newPaths)

		if len(result) != 3 {
			t.Fatalf("expected 3 paths, got %d: %v", len(result), result)
		}

		if result[0] != `\.pb\.go$` {
			t.Errorf("expected first sorted path to be pb (backslash < underscore), got %q", result[0])
		}
	})

	t.Run("empty existing returns new sorted", func(t *testing.T) {
		result := MergeExclusionPaths(nil, []string{`wire_gen\.go$`, `\.pb\.go$`})

		if len(result) != 2 {
			t.Fatalf("expected 2, got %d", len(result))
		}

		if result[0] != `\.pb\.go$` {
			t.Errorf("expected pb first, got %q", result[0])
		}
	})

	t.Run("both empty returns empty", func(t *testing.T) {
		result := MergeExclusionPaths(nil, nil)
		if len(result) != 0 {
			t.Errorf("expected empty, got %v", result)
		}
	})
}

func TestExclusionPaths(t *testing.T) {
	exclusions := []GeneratedExclusion{
		{Path: `\.pb\.go$`, Reason: "protobuf"},
		{Path: `_templ\.go$`, Reason: "templ"},
	}

	paths := ExclusionPaths(exclusions)

	if len(paths) != 2 {
		t.Fatalf("expected 2, got %d", len(paths))
	}

	if paths[0] != `\.pb\.go$` {
		t.Errorf("expected pb, got %q", paths[0])
	}
}

func TestGeneratedExclusionString(t *testing.T) {
	e := GeneratedExclusion{Path: `\.pb\.go$`, Reason: "protobuf generated code"}
	s := e.String()

	if s != `\.pb\.go$ (protobuf generated code)` {
		t.Errorf("unexpected string: %q", s)
	}
}

func TestShouldSkipDir(t *testing.T) {
	tests := []struct {
		name string
		dir  string
		want bool
	}{
		{"vendor", "vendor", true},
		{"node_modules", "node_modules", true},
		{"hidden dir", ".cache", true},
		{"git dir", ".git", true},
		{"normal dir", "src", false},
		{"pkg dir", "pkg", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldSkipDir(tt.dir); got != tt.want {
				t.Errorf("shouldSkipDir(%q) = %v, want %v", tt.dir, got, tt.want)
			}
		})
	}
}

func TestScanProjectEmptyDir(t *testing.T) {
	fsys := fstest.MapFS{}
	result, err := ScanProject(fsys, "/nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ScannedFiles != 0 {
		t.Errorf("expected 0 files, got %d", result.ScannedFiles)
	}

	if result.GeneratedFiles != 0 {
		t.Errorf("expected 0 generated, got %d", result.GeneratedFiles)
	}
}
