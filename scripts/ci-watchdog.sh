#!/usr/bin/env bash
#
# CI-health watchdog: detects the two silent-failure modes that left this repo
# with a disabled CI for two months (July 2026) — a workflow manually disabled,
# and a red master nobody looked at.
#
# Checks:
#   1. The ci.yml workflow state is "active" (not disabled_inactivity /
#      manually_disabled / disabled_data_retention).
#   2. The latest master run of ci.yml concluded "success".
#
# On drift (or with --force-drift), opens or updates an idempotent GitHub issue
# labeled ci-health. Exit codes: 0 healthy, 1 drift detected, 2 script error.
#
# Usage:
#   scripts/ci-watchdog.sh                     # real check
#   scripts/ci-watchdog.sh --force-drift       # dry mode: report drift without
#                                              # touching the issue tracker
#   scripts/ci-watchdog.sh --check-only        # real check, never post issues
#
set -euo pipefail

REPO="${GITHUB_REPOSITORY:-LarsArtmann/golangci-lint-auto-configure}"
WORKFLOW="ci.yml"
LABEL="ci-health"
ISSUE_TITLE="ci-health: master CI is unhealthy"

FORCE_DRIFT=false
CHECK_ONLY=false
case "${1:-}" in
	"--force-drift") FORCE_DRIFT=true ;;
	"--check-only") CHECK_ONLY=true ;;
	"") ;;
	*) echo "usage: $0 [--force-drift|--check-only]" >&2; exit 2 ;;
esac

problems=()

workflow_state=$(
	gh api "repos/${REPO}/actions/workflows/${WORKFLOW}" \
		--jq '.state' 2>/dev/null || echo "api_error"
)
if [ "$workflow_state" != "active" ]; then
	problems+=("workflow ci.yml state is '${workflow_state}' (expected 'active' — the workflow may be disabled again, as it silently was from 2026-07-16 to 2026-09-11)")
else
	echo "✓ ci.yml workflow state: active"
fi

last_run=$(
	gh api "repos/${REPO}/actions/workflows/${WORKFLOW}/runs?branch=master&status=completed&per_page=1" \
		--jq '.workflow_runs[0] | "\(.conclusion) \(.html_url)"' 2>/dev/null || echo "api_error -"
)
conclusion="${last_run%% *}"
url="${last_run#* }"
if [ "$conclusion" != "success" ]; then
	problems+=("latest master run of ci.yml concluded '${conclusion}' (expected 'success') — ${url}")
else
	echo "✓ latest master run: success (${url})"
fi

if [ "$FORCE_DRIFT" = true ]; then
	problems=("FORCED-DRIFT dry run: no real problem exists; this validates the watchdog wiring end-to-end")
fi

if [ ${#problems[@]} -eq 0 ]; then
	echo "✓ CI healthy"
	exit 0
fi

body=$(
	printf '%s\n\n' "The CI-health watchdog detected drift on \`${REPO}\` master. Silence here has history: the main workflow was manually disabled from 2026-07-16 to 2026-09-11 without anyone noticing."
)
for p in "${problems[@]}"; do
	body+=$(printf -- '- %s\n' "$p")
done
body+=$(
	printf '\nFix: re-enable the workflow (Actions tab → enable, or `gh workflow enable ci.yml`), repair the red run, then close this issue.\n_Run by `.github/workflows/ci-watchdog.yml` — do not delete the label `%s`._\n' "$LABEL"
)

echo ""
echo "✗ CI drift detected:"
for p in "${problems[@]}"; do
	echo "  - $p"
done

if [ "$FORCE_DRIFT" = true ] || [ "$CHECK_ONLY" = true ]; then
	echo "(dry mode: issue not posted)"
	exit 1
fi

existing=$(
	gh api "search/issues?q=repo:${REPO}+is:issue+is:open+label:${LABEL}+in:title" \
		--jq '.items[0].number' 2>/dev/null || echo ""
)
if [ -n "$existing" ] && [ "$existing" != "null" ]; then
	gh issue edit "$existing" --repo "$REPO" --body "$body" > /dev/null
	echo "Updated existing ci-health issue #${existing}"
else
	gh issue create --repo "$REPO" --title "$ISSUE_TITLE" --label "$LABEL" --body "$body" > /dev/null
	echo "Created ci-health issue"
fi

exit 1
