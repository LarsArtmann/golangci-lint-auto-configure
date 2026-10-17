package linter

import (
	"testing"
)

func TestMergeNormalizerChanges(t *testing.T) {
	t.Parallel()

	shapeRepair := Change{Path: "run.go", Kind: ChangeUpdate, From: "1.28", To: "1.27"}
	sharedBadKey := Change{Path: PathLintersSettings + ".goconst", Element: "min-length", Kind: ChangeUpdate}
	redundancyRemoval := Change{Path: PathLintersEnable, Element: "lll", Kind: ChangeRemove}

	tests := []struct {
		name      string
		persisted []Change
		derived   []Change
		want      []string
	}{
		{
			name:    "converge-only repairs survive the union",
			derived: []Change{redundancyRemoval},
			want:    []string{redundancyRemoval.Key()},
		},
		{
			name:      "shared repairs are counted once",
			persisted: []Change{sharedBadKey},
			derived:   []Change{sharedBadKey, redundancyRemoval},
			want:      []string{sharedBadKey.Key(), redundancyRemoval.Key()},
		},
		{
			name:      "phase-one-only repairs are kept",
			persisted: []Change{shapeRepair},
			derived:   nil,
			want:      []string{shapeRepair.Key()},
		},
		{
			name:      "order keeps persisted first, then converge-exclusive",
			persisted: []Change{shapeRepair, sharedBadKey},
			derived:   []Change{sharedBadKey, redundancyRemoval},
			want:      []string{shapeRepair.Key(), sharedBadKey.Key(), redundancyRemoval.Key()},
		},
		{
			name: "empty sides merge to empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := mergeNormalizerChanges(tt.persisted, tt.derived)

			keys := make([]string, 0, len(got))
			for _, change := range got {
				keys = append(keys, change.Key())
			}

			if len(keys) != len(tt.want) {
				t.Fatalf("mergeNormalizerChanges() = %v, want keys %v", keys, tt.want)
			}

			for i := range tt.want {
				if keys[i] != tt.want[i] {
					t.Fatalf("mergeNormalizerChanges()[%d] = %q, want %q (all: %v)", i, keys[i], tt.want[i], tt.want)
				}
			}
		})
	}
}
