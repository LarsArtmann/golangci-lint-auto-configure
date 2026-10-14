package constants_test

import (
	"testing"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// The pragmatic-noise tier DERIVES from friction evidence
// (types.DeriveTier at table load): the hand-maintained noise map is gone.
// These specs pin the derivation against the historically measured set, so a
// new Evidence row silently reclassifying a linter fails here first.
func TestPragmaticNoiseDerivation(t *testing.T) {
	derived := constants.PragmaticNoiseLinters

	expected := map[string]string{
		"funlen":          derived["funlen"],
		"gochecknoglobals": derived["gochecknoglobals"],
		"ireturn":          derived["ireturn"],
		"wrapcheck":        derived["wrapcheck"],
	}

	if len(derived) != len(expected) {
		t.Fatalf("derived noise set = %v, want exactly the 4 historically measured linters", derived)
	}

	for name := range expected {
		if _, known := derived[name]; !known {
			t.Errorf("expected %q in the derived noise set, got %v", name, derived)
		}

		record, known := constants.PolicyRecord(types.LinterName(name))
		if !known {
			t.Fatalf("noise linter %q missing from the policy table", name)
		}

		if record.Friction == nil {
			t.Errorf("%q derives into noise without friction evidence", name)

			continue
		}

		if record.Friction.Ratio < types.PragmaticNoiseFrictionThreshold {
			t.Errorf("%q ratio %.2f is below the %.1f threshold yet derived into noise",
				name, record.Friction.Ratio, types.PragmaticNoiseFrictionThreshold)
		}
	}
}

func TestDeriveTierThresholds(t *testing.T) {
	hot := &types.Evidence{Ratio: 0.83}
	cold := &types.Evidence{Ratio: 0.79}

	cases := []struct {
		name  string
		stored types.Tier
		fric  *types.Evidence
		want  types.TierKind
	}{
		{"unmanaged below threshold stays recommendable", types.Tier{}, cold, types.TierRecommendable},
		{"unmanaged at threshold derives to noise", types.Tier{}, hot, types.TierPragmaticNoise},
		{"unmanaged without evidence stays recommendable", types.Tier{}, nil, types.TierRecommendable},
		{"explicit disabled wins over measurement", types.Tier{Kind: types.TierDisabled, Rationale: "x"}, hot, types.TierDisabled},
		{"explicit never-auto-enable wins over measurement",
			types.Tier{Kind: types.TierNeverAutoEnable, Rationale: "x"}, hot, types.TierNeverAutoEnable},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := types.DeriveTier(tc.stored, tc.fric)
			if got.Kind != tc.want {
				t.Errorf("DeriveTier(kind=%d, friction=%v) = %s, want %s",
					tc.stored.Kind, tc.fric, got.Kind, tc.want)
			}
		})
	}
}

func TestDerivedNoiseTierCarriesMeasuredRationale(t *testing.T) {
	record, known := constants.PolicyRecord("wrapcheck")
	if !known {
		t.Fatal("wrapcheck missing from the policy table")
	}

	if record.Tier.Kind != types.TierPragmaticNoise {
		t.Fatalf("wrapcheck tier = %s, want pragmatic-noise", record.Tier.Kind)
	}

	if record.Tier.Rationale == "" {
		t.Fatal("derived tier must cite the measurement in its rationale")
	}
}
