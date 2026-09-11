#!/usr/bin/env bash
#
# vendorHash guard: keeps vendorHash.nix in sync with go.mod/go.sum.
#
# When go.mod or go.sum changes, the Nix fixed-output derivation for go-modules
# fails with an opaque "hash mismatch ... got: sha256-..." error, breaking the
# build for every clone until someone hand-pastes the new hash. This guard
# turns that failure into either a self-repair (--fix) or a precise instruction.
#
# Usage:
#   scripts/vendorhash-guard.sh          # check: fail with exact fix instructions
#   scripts/vendorhash-guard.sh --fix    # fix: write the got: hash and re-build
#
# CI usage (nix job):
#   scripts/vendorhash-guard.sh --fix && git diff --exit-code vendorHash.nix
#
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

VENDORHASH_FILE="vendorHash.nix"
MODE="check"
if [ "${1:-}" = "--fix" ] || [ "${1:-}" = "fix" ]; then
	MODE="fix"
elif [ -n "${1:-}" ]; then
	echo "usage: $0 [--fix]" >&2
	exit 2
fi

if [ ! -f "$VENDORHASH_FILE" ]; then
	echo "✗ $VENDORHASH_FILE not found — is this the right repo?" >&2
	exit 1
fi

echo "Building with Nix to verify vendorHash (this may take a while)..."
LOG=$(mktemp)
trap 'rm -f "$LOG"' EXIT

if nix build >"$LOG" 2>&1; then
	echo "✓ vendorHash is in sync — nix build succeeded"
	exit 0
fi

GOT=$(grep -oP 'got:\s+\Ksha256-[A-Za-z0-9+/=]+' "$LOG" | head -1 || true)
if [ -z "$GOT" ]; then
	echo "✗ nix build failed for a reason OTHER than a vendorHash mismatch:" >&2
	tail -30 "$LOG" >&2
	exit 1
fi

CURRENT=$(grep -oP '^\s*"\Ksha256-[^"]+' "$VENDORHASH_FILE" || true)

if [ "$MODE" = "check" ]; then
	echo "✗ vendorHash mismatch — go.mod/go.sum changed without updating $VENDORHASH_FILE" >&2
	echo "" >&2
	echo "  expected: ${CURRENT:-<unreadable>}" >&2
	echo "  got:      $GOT" >&2
	echo "" >&2
	echo "Fix (data-safe, no history rewrite):" >&2
	echo "  scripts/vendorhash-guard.sh --fix    # writes the hash and re-builds" >&2
	echo "or manually:" >&2
	echo "  printf '%s\\n' '\"$GOT\"' > $VENDORHASH_FILE && nix build" >&2
	echo "" >&2
	echo "Then commit $VENDORHASH_FILE together with your go.mod/go.sum change." >&2
	exit 1
fi

# --fix mode: write the got: hash and re-build.
{
	echo "# Managed by buildflow nix-checker — update with: nix build 2>&1 | rg \"got:\""
	echo "\"$GOT\""
} >"$VENDORHASH_FILE"

echo "Wrote $GOT to $VENDORHASH_FILE — re-building..."
if nix build >"$LOG" 2>&1; then
	echo "✓ vendorHash fixed — nix build succeeded"
	echo ""
	echo "Commit $VENDORHASH_FILE together with your go.mod/go.sum change:"
	echo "  git add $VENDORHASH_FILE && git commit"
	exit 0
fi

echo "✗ nix build still fails after writing the got: hash:" >&2
tail -30 "$LOG" >&2
exit 1
