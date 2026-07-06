#!/usr/bin/env bash
# Coverage threshold gate — fails CI if total coverage drops below minimum.
# Usage: scripts/coverage-check.sh [min-percentage] (default: 60)
set -euo pipefail

MIN="${1:-60}"

if [[ ! -f coverage.out ]]; then
	echo "coverage.out not found — run 'go test -coverprofile=coverage.out ./...'" >&2
	exit 1
fi

TOTAL=$(go tool cover -func=coverage.out | grep '^total:' | awk '{print $NF}' | tr -d '%')

echo "Coverage: ${TOTAL}% (minimum: ${MIN}%)"

if (( $(echo "${TOTAL} < ${MIN}" | bc -l) )); then
	echo "❌ Coverage ${TOTAL}% is below minimum ${MIN}%"
	exit 1
fi

echo "✅ Coverage ${TOTAL}% meets minimum ${MIN}%"
