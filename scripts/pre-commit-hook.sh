#!/bin/bash
# Pre-commit hook for golangci-linter-auto-configure
# This hook runs golangci-linter-auto-configure to ensure your config is optimized

set -e

echo "Running golangci-linter-auto-configure pre-commit hook..."

# Check if golangci-linter-auto-configure is installed
if ! command -v golangci-linter-auto-configure &> /dev/null; then
    echo "Error: golangci-linter-auto-configure is not installed"
    echo "Install it from: https://github.com/LarsArtmann/golangcli-linter-auto-configure"
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
if ! golangci-linter-auto-configure analyze --config "$CONFIG_FILE"; then
    echo "Warning: Analysis found issues with your configuration"
    echo "Run 'golangci-linter-auto-configure configure' to auto-fix"
fi

# Optional: Auto-configure (uncomment to enable)
# echo "Auto-configuring..."
# golangci-linter-auto-configure configure --config "$CONFIG_FILE"

echo "Pre-commit hook completed successfully!"
exit 0
