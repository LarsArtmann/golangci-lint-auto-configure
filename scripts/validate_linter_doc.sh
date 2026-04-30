#!/bin/bash

# Documentation Validation Script
# Validates that a linter documentation file meets quality standards

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

LINTER_NAME=$1
DOC_FILE="$(cd "$(dirname "$0")/.." && pwd)/reports/${LINTER_NAME}.md"

if [ -z "$LINTER_NAME" ]; then
    echo -e "${RED}✗ ERROR${NC}: Usage: $0 <linter_name>"
    exit 1
fi

echo "Validating documentation for: $LINTER_NAME"

# Check if file exists
if [ ! -f "$DOC_FILE" ]; then
    echo -e "${RED}✗ FAIL${NC}: File does not exist: $DOC_FILE"
    exit 1
fi
echo -e "${GREEN}✓${NC} File exists"

# Check if file is not empty
if [ ! -s "$DOC_FILE" ]; then
    echo -e "${RED}✗ FAIL${NC}: File is empty"
    exit 1
fi
echo -e "${GREEN}✓${NC} File is not empty"

# Check file size (should be > 10KB for comprehensive docs)
FILE_SIZE=$(du -k "$DOC_FILE" | cut -f1)
if [ "$FILE_SIZE" -lt 10 ]; then
    echo -e "${YELLOW}⚠ WARNING${NC}: File is less than 10KB ($FILE_SIZE KB) - may not be comprehensive"
else
    echo -e "${GREEN}✓${NC} File size adequate ($FILE_SIZE KB)"
fi

# Check for required sections (case-insensitive)
# Accept either separate "Enabled/Disabled" sections or combined "Enabled" section
REQUIRED_SECTIONS=(
    "What the Linter Does"
    "How It Should Be Configured"
    "How It Interferes or Works Together With Other Linters"
)

MISSING_SECTIONS=()
for section in "${REQUIRED_SECTIONS[@]}"; do
    if ! grep -qi "$section" "$DOC_FILE"; then
        MISSING_SECTIONS+=("$section")
    fi
done

# Check for When It Should Be Enabled (either separate or combined)
if ! grep -qi "When It Should Be Enabled" "$DOC_FILE"; then
    MISSING_SECTIONS+=("When It Should Be Enabled")
fi

if [ ${#MISSING_SECTIONS[@]} -gt 0 ]; then
    echo -e "${RED}✗ FAIL${NC}: Missing sections:"
    for section in "${MISSING_SECTIONS[@]}"; do
        echo "  - $section"
    done
    exit 1
fi
echo -e "${GREEN}✓${NC} All required sections present"

# Check for practical examples
if ! grep -qi "Example" "$DOC_FILE"; then
    echo -e "${YELLOW}⚠ WARNING${NC}: No examples found in documentation"
else
    echo -e "${GREEN}✓${NC} Examples present"
fi

# Check for configuration YAML examples
if ! grep -qi "yaml\|yml\|configuration" "$DOC_FILE"; then
    echo -e "${YELLOW}⚠ WARNING${NC}: No configuration examples found"
else
    echo -e "${GREEN}✓${NC} Configuration examples present"
fi

# Check for inter-linter relationships
if ! grep -qi "linter\|complementary\|conflict" "$DOC_FILE"; then
    echo -e "${YELLOW}⚠ WARNING${NC}: No inter-linter relationships documented"
else
    echo -e "${GREEN}✓${NC} Inter-linter relationships documented"
fi

echo -e "\n${GREEN}✓ SUCCESS${NC}: $LINTER_NAME documentation validated!"
exit 0
