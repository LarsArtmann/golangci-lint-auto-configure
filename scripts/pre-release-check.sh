#!/usr/bin/env bash
#
# Pre-release checklist for golangci-lint-auto-configure
#
# Verifies that the repo is ready to cut a release. Exits non-zero on any failure.
#
# Usage: ./scripts/pre-release-check.sh [version]
#   version: optional, e.g. "0.7.0". If omitted, reads from CHANGELOG.md latest section.
#
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

export GOEXPERIMENT=jsonv2
export CGO_ENABLED=1

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

pass=0
fail=0
warn=0

check() {
	local desc="$1"
	shift
	if "$@" >/dev/null 2>&1; then
		echo -e "${GREEN}✓${NC} $desc"
		pass=$((pass + 1))
	else
		echo -e "${RED}✗${NC} $desc"
		fail=$((fail + 1))
	fi
}

warn_check() {
	local desc="$1"
	shift
	if "$@" >/dev/null 2>&1; then
		echo -e "${GREEN}✓${NC} $desc"
		pass=$((pass + 1))
	else
		echo -e "${YELLOW}⚠${NC}  $desc"
		warn=$((warn + 1))
	fi
}

info() {
	echo -e "  $1"
}

echo "=== Pre-Release Checklist ==="
echo ""

# 1. Working tree clean
if [ -n "$(git status --porcelain)" ]; then
	echo -e "${RED}✗${NC} Working tree has uncommitted changes"
	git status --short
	echo ""
	echo "Commit or stash changes before releasing."
	exit 1
else
	echo -e "${GREEN}✓${NC} Working tree clean"
	pass=$((pass + 1))
fi

# 2. Determine version
VERSION="${1:-}"
if [ -z "$VERSION" ]; then
	VERSION=$(grep -oP '(?<=## \[)[0-9]+\.[0-9]+\.[0-9]+' CHANGELOG.md | head -1)
fi
info "Target version: v${VERSION}"
echo ""

# 3. Build
check "Go build succeeds" go build ./cmd/golangci-lint-auto-configure

# 4. Tests
echo -e "  ${YELLOW}Running tests (this may take a moment)...${NC}"
check "Tests pass (-race)" go test -race ./pkg/... ./internal/...

# 5. Lint
check "golangci-lint passes" golangci-lint run --config=.golangci.yml --timeout=5m

# 6. Coverage (same gate as CI: cmd/coverage-check parses the profile total;
#    grepping per-package "coverage:" lines matches every line and mis-parses)
go test -coverprofile=/tmp/coverage-check.out -covermode=atomic ./pkg/... ./internal/... > /dev/null
if go run ./cmd/coverage-check -min=60 -profile=/tmp/coverage-check.out; then
	echo -e "${GREEN}✓${NC} Coverage >= 60% threshold"
	pass=$((pass + 1))
else
	echo -e "${RED}✗${NC} Coverage below 60% threshold"
	fail=$((fail + 1))
fi

echo ""

# 7. GoReleaser config valid
check "goreleaser check passes" goreleaser check

# 8. CHANGELOG has the version
check "CHANGELOG.md has [${VERSION}] entry" grep -q "\[${VERSION}\]" CHANGELOG.md

# 9. FEATURES.md version stamp
FEATURES_VERSION=$(grep -oP '(?<=\*\*Version:\*\* v)[0-9]+\.[0-9]+\.[0-9]+' FEATURES.md || echo "unknown")
if [ "$FEATURES_VERSION" = "$VERSION" ]; then
	echo -e "${GREEN}✓${NC} FEATURES.md version matches ($FEATURES_VERSION)"
	pass=$((pass + 1))
else
	echo -e "${RED}✗${NC} FEATURES.md version ($FEATURES_VERSION) != target ($VERSION)"
	fail=$((fail + 1))
fi

# 10. Tag doesn't already exist
if git rev-parse "v${VERSION}" >/dev/null 2>&1; then
	echo -e "${RED}✗${NC} Tag v${VERSION} already exists"
	fail=$((fail + 1))
else
	echo -e "${GREEN}✓${NC} Tag v${VERSION} does not exist yet"
	pass=$((pass + 1))
fi

# 11. Local branch is up to date with remote
LOCAL=$(git rev-parse HEAD)
REMOTE=$(git rev-parse origin/master 2>/dev/null || echo "")
if [ -n "$REMOTE" ] && [ "$LOCAL" = "$REMOTE" ]; then
	echo -e "${GREEN}✓${NC} Local master is up to date with origin"
	pass=$((pass + 1))
elif [ -z "$REMOTE" ]; then
	echo -e "${YELLOW}⚠${NC}  Could not check remote (no origin/master)"
	warn=$((warn + 1))
else
	echo -e "${YELLOW}⚠${NC}  Local master is ahead/behind origin — push before tagging"
	warn=$((warn + 1))
fi

# 12. Snapshot build (warning if it fails — non-blocking since syft/cosign may be missing)
echo ""
echo -e "  ${YELLOW}Running goreleaser snapshot...${NC}"
if goreleaser release --snapshot --clean --skip=publish 2>/tmp/goreleaser-snapshot.log; then
	echo -e "${GREEN}✓${NC} Goreleaser snapshot build succeeds"
	pass=$((pass + 1))
else
	SNAPSHOT_ERR=$(tail -5 /tmp/goreleaser-snapshot.log)
	if echo "$SNAPSHOT_ERR" | grep -q "syft\|cosign\|docker"; then
		echo -e "${YELLOW}⚠${NC}  Goreleaser snapshot fails only on missing tool (syft/cosign/docker) — acceptable for local release"
		warn=$((warn + 1))
	else
		echo -e "${RED}✗${NC} Goreleaser snapshot build fails:"
		echo "$SNAPSHOT_ERR"
		fail=$((fail + 1))
	fi
fi

rm -rf dist /tmp/coverage-check.out /tmp/goreleaser-snapshot.log 2>/dev/null || true

echo ""
echo "=== Results ==="
echo -e "  ${GREEN}Passed:${NC}  $pass"
echo -e "  ${RED}Failed:${NC}  $fail"
echo -e "  ${YELLOW}Warnings:${NC} $warn"
echo ""

if [ "$fail" -gt 0 ]; then
	echo -e "${RED}✗ ${fail} check(s) failed. Fix before releasing.${NC}"
	exit 1
else
	echo -e "${GREEN}✓ All critical checks passed. Ready to tag v${VERSION}.${NC}"
	if [ "$warn" -gt 0 ]; then
		echo -e "  ${YELLOW}($warn warnings — review above)${NC}"
	fi
	exit 0
fi
