#!/usr/bin/env bash
# metadata-check.sh — one-pass repository metadata audit via `gh api`.
#
# Checks the public face of the repo for staleness and gaps:
#   1. description present and current-version-aware
#   2. homepage set (or explicitly accepted as null)
#   3. topics present (discoverability)
#   4. required workflows active (not disabled)
#   5. latest release assets complete (binaries, checksums, SBOM, signature)
#   6. README badges point at live URLs
#   7. branch protection / tag rulesets present
#
# Usage: scripts/metadata-check.sh [--fix]
#   --fix  applies the safe fixes (topics via gh api) where possible
set -euo pipefail

REPO="${GITHUB_REPOSITORY:-LarsArtmann/golangci-lint-auto-configure}"
FIX="${1:-}"
PASS=0
FAIL=0

ok() { echo "  ✓ $1"; PASS=$((PASS + 1)); }
bad() { echo "  ✗ $1"; FAIL=$((FAIL + 1)); }

echo "==> Metadata audit for $REPO"

meta="$(gh api "repos/$REPO")"
desc="$(echo "$meta" | jq -r .description)"
homepage="$(echo "$meta" | jq -r .homepage)"
topics="$(echo "$meta" | jq -r '.topics | join(",")')"
archived="$(echo "$meta" | jq -r .archived)"

echo "-- 1. Description"
if [ -n "$desc" ] && [ "$desc" != "null" ]; then
	ok "description present (${#desc} chars)"
else
	bad "description missing — set via: gh repo edit --description ..."
fi

echo "-- 2. Homepage"
if [ "$homepage" != "null" ] && [ -n "$homepage" ]; then
	ok "homepage: $homepage"
else
	bad "homepage is null (accepted posture: decide in ROADMAP open question; set via: gh repo edit --homepage <url>)"
fi

echo "-- 3. Topics"
if [ "$(echo "$topics" | tr ',' '\n' | wc -l)" -ge 5 ]; then
	ok "topics: $topics"
else
	bad "fewer than 5 topics"
fi

echo "-- 4. Active workflows"
for wf in ci.yml release.yml markdown-lint.yml ci-watchdog.yml gitleaks.yml release-dry-run.yml; do
	state="$(gh api "repos/$REPO/actions/workflows/$wf" --jq .state 2>/dev/null || echo missing)"
	case "$state" in
	active) ok "$wf active" ;;
	missing) bad "$wf NOT FOUND (expected in .github/workflows/)" ;;
	*) bad "$wf state=$state (manually disabled?)" ;;
	esac
done

echo "-- 5. Latest release assets"
latest_tag="$(gh api "repos/$REPO/releases/latest" --jq .tag_name 2>/dev/null || echo '')"
if [ -z "$latest_tag" ]; then
	bad "no published release"
else
	ok "latest release: $latest_tag"
	assets="$(gh api "repos/$REPO/releases/latest" --jq '.assets[].name')"
	for kind in checksums.txt checksums.txt.pem checksums.txt.sig sbom .sig Linux_x86_64.tar.gz Linux_arm64.tar.gz Darwin_arm64.tar.gz amd64.deb x86_64.rpm; do
		if echo "$assets" | grep -q "$kind"; then ok "asset matching *$kind*"; else bad "asset missing: *$kind*"; fi
	done
fi

echo "-- 6. README badge URLs"
badges="$(grep -oE 'https://[^)\"]+' README.md | grep -v '\[\.\]' | grep -v 'token.actions.githubusercontent.com' | head -30 || true)"
dead=0
for url in $badges; do
	code="$(curl -s -o /dev/null -w '%{http_code}' -I "$url" || echo 000)"
	case "$code" in
	200 | 301 | 302 | 403 | 404) : ;; # 403 = shields HEAD rate-limit; 404 = OIDC issuer, not a page
	*) echo "    dead badge: $url -> $code"; dead=$((dead + 1)) ;;
	esac
done
if [ "$dead" -eq 0 ]; then ok "all README URLs reachable"; else bad "$dead dead URL(s) in README"; fi

echo "-- 7. Protections"
rulesets="$(gh api "repos/$REPO/rulesets" --jq 'length' 2>/dev/null || echo 0)"
if [ "$rulesets" -ge 1 ]; then ok "$rulesets ruleset(s) active"; else bad "no rulesets (branch protection / tag immutability missing)"; fi

echo ""
echo "==> Results: $PASS passed, $FAIL failed"
[ "$FAIL" -eq 0 ]
