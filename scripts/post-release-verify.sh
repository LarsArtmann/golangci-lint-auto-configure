#!/usr/bin/env bash
#
# Post-release verification for golangci-lint-auto-configure
#
# Downloads a published binary and verifies it works. Also checks release metadata.
#
# Usage: ./scripts/post-release-verify.sh <version>
#   version: required, e.g. "0.6.0"
#
set -euo pipefail

VERSION="${1:-}"
if [ -z "$VERSION" ]; then
	echo "Usage: $0 <version> (e.g. 0.6.0)"
	exit 1
fi

TAG="v${VERSION}"
REPO="LarsArtmann/golangci-lint-auto-configure"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

pass=0
fail=0

check() {
	local desc="$1"
	shift
	if "$@"; then
		echo -e "${GREEN}✓${NC} $desc"
		pass=$((pass + 1))
	else
		echo -e "${RED}✗${NC} $desc"
		fail=$((fail + 1))
	fi
}

echo "=== Post-Release Verification: ${TAG} ==="
echo ""

# 1. Tag exists locally and remotely
check "Tag ${TAG} exists locally" git rev-parse "${TAG}"

REMOTE_TAG=$(git ls-remote --tags origin "refs/tags/${TAG}" 2>/dev/null || true)
if [ -n "$REMOTE_TAG" ]; then
	echo -e "${GREEN}✓${NC} Tag ${TAG} exists on remote"
	pass=$((pass + 1))
else
	echo -e "${RED}✗${NC} Tag ${TAG} not found on remote"
	fail=$((fail + 1))
fi

# 2. GitHub release exists
check "GitHub release ${TAG} exists" gh release view "${TAG}" --repo "${REPO}"

# 3. Release has assets
ASSET_COUNT=$(gh release view "${TAG}" --repo "${REPO}" --json assets --jq '.assets | length' 2>/dev/null || echo "0")
if [ "$ASSET_COUNT" -ge 3 ]; then
	echo -e "${GREEN}✓${NC} Release has ${ASSET_COUNT} assets"
	pass=$((pass + 1))
else
	echo -e "${RED}✗${NC} Release has only ${ASSET_COUNT} assets (expected >= 3)"
	fail=$((fail + 1))
fi

# 4. Release notes are not the commit dump (should be < 15KB of curated content)
NOTES_LENGTH=$(gh release view "${TAG}" --repo "${REPO}" --json body --jq '.body | length' 2>/dev/null || echo "0")
if [ "$NOTES_LENGTH" -gt 0 ] && [ "$NOTES_LENGTH" -lt 20000 ]; then
	echo -e "${GREEN}✓${NC} Release notes are curated (${NOTES_LENGTH} bytes)"
	pass=$((pass + 1))
elif [ "$NOTES_LENGTH" -ge 20000 ]; then
	echo -e "${RED}✗${NC} Release notes look like raw commit dump (${NOTES_LENGTH} bytes — rewrite with curated content)"
	fail=$((fail + 1))
else
	echo -e "${RED}✗${NC} Release notes are empty"
	fail=$((fail + 1))
fi

# 5. Checksums file exists
ASSET_NAMES=$(gh release view "${TAG}" --repo "${REPO}" --json assets --jq '.assets[].name' 2>/dev/null || true)
if echo "$ASSET_NAMES" | grep -q checksums; then
	echo -e "${GREEN}✓${NC} Checksums asset exists"
	pass=$((pass + 1))
else
	echo -e "${RED}✗${NC} Checksums asset not found"
	fail=$((fail + 1))
fi

# 6. Download and smoke-test a binary
echo ""
echo "  Downloading Linux x86_64 binary for smoke test..."
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

if gh release download "${TAG}" --repo "${REPO}" --pattern '*Linux_x86_64*' --dir "$TMPDIR" --clobber 2>/dev/null; then
	tar xzf "$TMPDIR"/golangci-lint-auto-configure_*_Linux_x86_64.tar.gz -C "$TMPDIR"
	BINARY="$TMPDIR/golangci-lint-auto-configure"

	if "$BINARY" --version >/dev/null 2>&1; then
		echo -e "${GREEN}✓${NC} Binary --version runs"
		pass=$((pass + 1))
	else
		echo -e "${RED}✗${NC} Binary --version runs"
		fail=$((fail + 1))
	fi

	if "$BINARY" --help >/dev/null 2>&1; then
		echo -e "${GREEN}✓${NC} Binary --help runs"
		pass=$((pass + 1))
	else
		echo -e "${RED}✗${NC} Binary --help runs"
		fail=$((fail + 1))
	fi

	# Verify version string contains the version number (literal match, not regex)
	VERSION_OUTPUT=$("$BINARY" --version 2>&1)
	if echo "$VERSION_OUTPUT" | grep -qF "$VERSION"; then
		echo -e "${GREEN}✓${NC} Binary --version shows correct version (${VERSION})"
		pass=$((pass + 1))
	else
		echo -e "${RED}✗${NC} Binary --version does not show version ${VERSION}:"
		echo "  $VERSION_OUTPUT"
		fail=$((fail + 1))
	fi
else
	echo -e "${RED}✗${NC} Could not download Linux x86_64 binary"
	fail=$((fail + 1))
fi

echo ""
echo "=== Results ==="
echo -e "  ${GREEN}Passed:${NC}  $pass"
echo -e "  ${RED}Failed:${NC}  $fail"

if [ "$fail" -gt 0 ]; then
	echo ""
	echo -e "${RED}✗ ${fail} check(s) failed. Review the release.${NC}"
	exit 1
else
	echo ""
	echo -e "${GREEN}✓ Release ${TAG} verified successfully.${NC}"
	exit 0
fi
