#!/usr/bin/env bash
# e2e-pin-matrix.sh — exercise `configure` against one pinned golangci-lint version.
#
# The CI job (e2e-pin-matrix in ci.yml) runs this script once per matrix cell
# (v2.10.1 pre-minimum, v2.12.0 minimum, v2.13.2 previous, v2.14.0 current).
# Each cell proves one of two contracts:
#
# Below the tool minimum (e.g. v2.10.1):
#   A0  configure refuses with the classified version.too_old error and a
#       non-zero exit — a broken, unloadable config must never be written.
#
# At or above the minimum:
#   A1  configure exits 0 on a fixture whose run.go is patch-form and
#       overspecified (run.go: "1.27.1")
#   A2  the configured run.go is capped at min(local Go, binary build Go),
#       stripped to major.minor (no patch churn)
#   A3  the PINNED golangci-lint loads the configured config (config verify
#       exits 0) — injected defaults are valid for that version, not just
#       for the latest. This is the invariant that caught the goconst
#       ignore-tests / v2.10.1 incompatibility (2026-10-07).
#   A4  a second `configure --check` on the configured fixture exits 0
#       (configure is idempotent across the pin)
#   A5  when the binary's build Go is older than the fixture's run.go,
#       `configure --check --json-errors` fails with the classified
#       config.run_go.newer_than_binary error (rescue refused in check mode)
#
# Usage: e2e-pin-matrix.sh <expected-golangci-lint-version> [min-tool-version]
#   GOLANGCI_LINT_BIN  optional path to the pinned golangci-lint binary
#                      (default: `golangci-lint` from PATH)
set -euo pipefail

EXPECTED_VERSION="${1:?usage: e2e-pin-matrix.sh <expected-golangci-lint-version> [min-tool-version]}"
MIN_TOOL_VERSION="${2:-v2.12.0}"
GOLANGCI_LINT_BIN="${GOLANGCI_LINT_BIN:-$(command -v golangci-lint)}"

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TOOL_BIN="$REPO_ROOT/bin/golangci-lint-auto-configure"

fail() { echo "✗ FAIL: $*" >&2; exit 1; }

echo "==> e2e pin matrix cell: $EXPECTED_VERSION (min supported: $MIN_TOOL_VERSION)"

# The tool's analyzer shells out to `golangci-lint`; make sure it finds the
# pinned binary, not some other install.
export PATH="$(dirname "$GOLANGCI_LINT_BIN"):$PATH"
export GOLANGCI_LINT_AUTO_CONFIGURE_NO_AUDIT=1
export GOEXPERIMENT="${GOEXPERIMENT:-jsonv2}"

binary_version="$("$GOLANGCI_LINT_BIN" --version 2>/dev/null | grep -o 'version [0-9.]*' | grep -o '[0-9.]*')"
[ "$binary_version" = "${EXPECTED_VERSION#v}" ] || fail "binary reports version $binary_version, expected ${EXPECTED_VERSION#v}"

below_minimum=false
lowest="$(printf '%s\n%s\n' "${EXPECTED_VERSION#v}" "${MIN_TOOL_VERSION#v}" | sort -V | head -1)"
if [ "$lowest" = "${EXPECTED_VERSION#v}" ] && [ "${EXPECTED_VERSION#v}" != "${MIN_TOOL_VERSION#v}" ]; then
	below_minimum=true
fi

binary_go="$("$GOLANGCI_LINT_BIN" --version 2>/dev/null | grep -o 'built with go[0-9.]*' | grep -o '[0-9.]*' | cut -d. -f1,2)"
[ -n "$binary_go" ] || fail "cannot parse build Go version from --version output"
echo "    binary Go (major.minor): $binary_go"

local_go="$(go version | grep -o 'go[0-9][0-9.]*' | head -1 | cut -d. -f1,2 | tr -d go)"
[ -n "$local_go" ] || fail "cannot parse local Go version from 'go version'"
echo "    local Go  (major.minor): $local_go"

echo "==> Building tool binary"
(cd "$REPO_ROOT" && go build -o "$TOOL_BIN" ./cmd/golangci-lint-auto-configure)

WORKDIR="$(mktemp -d)"
trap '[ -n "${WORKDIR:-}" ] && [ -d "$WORKDIR" ] && rm -rf "$WORKDIR"' EXIT

make_fixture() {
	local dir="$1"
	mkdir -p "$dir"
	cat >"$dir/go.mod" <<EOF
module example.com/pinfixture

go 1.27
EOF
	cat >"$dir/main.go" <<'EOF'
package main

func main() {}
EOF
	cat >"$dir/.golangci.yml" <<EOF
version: "2"

run:
  go: "1.27.1"

linters:
  default: none
  enable:
    - govet
EOF
	(cd "$dir" && git init -q && git -c user.email=e2e@fixture -c user.name=fixture add -A && git -c user.email=e2e@fixture -c user.name=fixture commit -qm init)
}

if [ "$below_minimum" = true ]; then
	# --- A0: below-minimum installs get a classified refusal, never a broken config ---
	echo "==> A0: configure must refuse (pinned $EXPECTED_VERSION < minimum $MIN_TOOL_VERSION)"
	make_fixture "$WORKDIR/belowmin"
	set +e
	refusal_err="$(cd "$WORKDIR/belowmin" && "$TOOL_BIN" configure --config .golangci.yml --json-errors 2>&1 >/dev/null)"
	refusal_status=$?
	set -e
	[ "$refusal_status" -ne 0 ] || fail "A0: configure exited 0 on a below-minimum golangci-lint ($EXPECTED_VERSION < $MIN_TOOL_VERSION)"
	echo "$refusal_err" | grep -q "version.too_old" \
		|| fail "A0: expected classified code version.too_old in --json-errors output, got: $refusal_err"
	echo "    ✓ classified version.too_old refusal, exit $refusal_status"
	echo "✓ ALL CHECKS PASSED for $EXPECTED_VERSION (below-minimum contract)"
	exit 0
fi

# run.go after configure must be min(local Go, binary Go).
expected_run_go="$binary_go"
if [ "$(printf '%s\n%s\n' "$binary_go" "$local_go" | sort -V | head -1)" = "$local_go" ]; then
	expected_run_go="$local_go"
fi
echo "    expected run.go:         $expected_run_go"

# --- A1+A2: configure repairs the overspecified patch-form run.go ---
echo "==> A1/A2: configure on overspecified fixture"
make_fixture "$WORKDIR/configured"
(cd "$WORKDIR/configured" && "$TOOL_BIN" configure --config .golangci.yml) >/dev/null 2>&1 \
	|| fail "A1: configure exited non-zero"

actual_run_go="$(grep -A2 '^run:' "$WORKDIR/configured/.golangci.yml" | grep 'go:' | head -1 | grep -o '[0-9.]*')"
[ "$actual_run_go" = "$expected_run_go" ] \
	|| fail "A2: run.go is $actual_run_go, expected $expected_run_go (min of local $local_go and binary $binary_go, no patch)"
echo "    ✓ run.go: 1.27.1 -> $actual_run_go"

# --- A3: the pinned binary loads the configured config ---
echo "==> A3: pinned golangci-lint loads configured config"
(cd "$WORKDIR/configured" && "$GOLANGCI_LINT_BIN" config verify) \
	|| fail "A3: pinned $EXPECTED_VERSION refuses the configured config (injected defaults incompatible?)"
echo "    ✓ config verify green on $EXPECTED_VERSION"

# --- A4: configure is idempotent under the pin ---
echo "==> A4: configure --check on configured fixture"
(cd "$WORKDIR/configured" && "$TOOL_BIN" configure --config .golangci.yml --check) >/dev/null 2>&1 \
	|| fail "A4: --check on the configured fixture wants more changes (configure not idempotent under $EXPECTED_VERSION)"
echo "    ✓ --check exits 0 (config optimal)"

# --- A5: check mode refuses to rescue, with the classified error ---
if [ "$(printf '%s\n1.27\n' "$binary_go" | sort -V | head -1)" = "$binary_go" ] && [ "$binary_go" != "1.27" ]; then
	echo "==> A5: --check must refuse rescue (binary Go $binary_go < 1.27)"
	make_fixture "$WORKDIR/broken"
	set +e
	check_err="$(cd "$WORKDIR/broken" && "$TOOL_BIN" configure --config .golangci.yml --check --json-errors 2>&1 >/dev/null)"
	check_status=$?
	set -e
	[ "$check_status" -ne 0 ] || fail "A5: --check exited 0 but run.go 1.27.1 is unloadable for $EXPECTED_VERSION"
	echo "$check_err" | grep -q "config.run_go.newer_than_binary" \
		|| fail "A5: expected classified code config.run_go.newer_than_binary in --json-errors output, got: $check_err"
	echo "    ✓ classified config.run_go.newer_than_binary, exit $check_status"
else
	echo "==> A5: skipped (binary Go $binary_go can load 1.27 — nothing to rescue)"
fi

echo "✓ ALL CHECKS PASSED for $EXPECTED_VERSION"
