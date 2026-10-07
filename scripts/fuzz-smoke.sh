#!/usr/bin/env bash
# Short fuzz smoke: runs every fuzz target for a few seconds each.
# Exit non-zero on any crasher. CI runs this as the `fuzz` job; locally it is
# the pre-push confidence check for fuzz-seeded invariants.
#
# Usage: scripts/fuzz-smoke.sh [seconds-per-target]
set -euo pipefail

cd "$(dirname "$0")/.."

SECONDS_PER_TARGET="${1:-10}"

if ! command -v go >/dev/null 2>&1; then
	echo "error: go not found in PATH" >&2
	exit 1
fi

# jsonv2 is required on Go 1.26 toolchains; a no-op on Go 1.27+.
export CGO_ENABLED=1
export GOEXPERIMENT="${GOEXPERIMENT:-jsonv2}"

TARGETS=(
	"pkg/types:FuzzNormalizeGoMajorMinor"
	"pkg/types:FuzzCompareGoMajorMinor"
	"pkg/config:FuzzDetectYAMLIndent"
	"pkg/linter:FuzzMergeExclusionLinters"
)

failed=0

for target in "${TARGETS[@]}"; do
	pkg="${target%%:*}"
	fn="${target#*:}"

	echo "=== fuzzing ${fn} (${pkg}) for ${SECONDS_PER_TARGET}s ==="

	if ! go test "./${pkg}" -run "^${fn}\$" -fuzz "^${fn}\$" -fuzztime="${SECONDS_PER_TARGET}s"; then
		echo "FAIL: ${fn} found a crasher" >&2
		failed=1
	fi
done

if [ "${failed}" -ne 0 ]; then
	echo "fuzz smoke FAILED — crashers written to testdata/fuzz/, triage before pushing" >&2
	exit 1
fi

echo "fuzz smoke: all ${#TARGETS[@]} targets passed"
