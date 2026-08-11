#!/bin/bash

# Linter Count Verification Script
# This script verifies that the correct number of linter documentation files exist

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get expected count from golangci-lint CLI
echo "Getting linter count from golangci-lint CLI..."
EXPECTED_COUNT=$(golangci-lint linters --json 2>/dev/null | jq -r '.Enabled[].name' | wc -l | tr -d ' ')
echo -e "${GREEN}✓${NC} Expected linters: $EXPECTED_COUNT"

# Get actual count from reports directory
echo "Counting documentation files..."
ACTUAL_COUNT=$(ls -1 "$(cd "$(dirname "$0")/.." && pwd)/reports/"*.md 2>/dev/null | wc -l | tr -d ' ')
echo -e "${GREEN}✓${NC} Documentation files: $ACTUAL_COUNT"

# Calculate completion percentage
if [ "$EXPECTED_COUNT" -gt 0 ]; then
	PERCENTAGE=$(echo "scale=1; $ACTUAL_COUNT * 100 / $EXPECTED_COUNT" | bc 2>/dev/null || echo "0")
	echo -e "${GREEN}✓${NC} Completion: ${PERCENTAGE}%"
else
	echo -e "${RED}✗${NC} ERROR: Invalid expected count"
	exit 1
fi

# Check for mismatch
if [ "$ACTUAL_COUNT" -eq "$EXPECTED_COUNT" ]; then
	echo -e "\n${GREEN}✓ SUCCESS${NC}: All linters documented!"
	exit 0
elif [ "$ACTUAL_COUNT" -lt "$EXPECTED_COUNT" ]; then
	MISSING=$((EXPECTED_COUNT - ACTUAL_COUNT))
	echo -e "\n${YELLOW}⚠ WARNING${NC}: $MISSING linter(s) not yet documented"
	echo "Expected: $EXPECTED_COUNT, Actual: $ACTUAL_COUNT"
	exit 0
else
	EXTRA=$((ACTUAL_COUNT - EXPECTED_COUNT))
	echo -e "\n${RED}✗ ERROR${NC}: $EXTRA extra documentation file(s) found"
	echo "Expected: $EXPECTED_COUNT, Actual: $ACTUAL_COUNT"
	exit 1
fi
