#!/bin/bash
# Pre-commit hook for golangci-lint-auto-configure
# This hook runs golangci-lint-auto-configure to ensure your config is optimized

set -e

echo "Running golangci-lint-auto-configure pre-commit hook..."

# Check if golangci-lint-auto-configure is installed
if ! command -v golangci-lint-auto-configure &>/dev/null; then
	echo "Error: golangci-lint-auto-configure is not installed"
	echo "Install it from: https://github.com/LarsArtmann/golangci-lint-auto-configure"
	exit 1
fi

# Find config file
CONFIG_FILE=""
for file in ".golangci.yml" ".golangci.yaml"; do
	if [ -f "$file" ]; then
		CONFIG_FILE="$file"
		break
	fi
done

if [ -z "$CONFIG_FILE" ]; then
	echo "No .golangci.yml config file found, skipping..."
	exit 0
fi

echo "Found config: $CONFIG_FILE"

# Run analyze to check for recommendations
echo "Analyzing configuration..."
if ! golangci-lint-auto-configure analyze --config "$CONFIG_FILE"; then
	echo "Warning: Analysis found issues with your configuration"
	echo "Run 'golangci-lint-auto-configure configure' to auto-fix"
fi

# Optional: Auto-configure (uncomment to enable)
# echo "Auto-configuring..."
# golangci-lint-auto-configure configure --config "$CONFIG_FILE"

echo "Pre-commit hook completed successfully!"
exit 0
