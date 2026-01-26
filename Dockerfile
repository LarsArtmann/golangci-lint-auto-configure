# Multi-stage Dockerfile for golangci-linter-auto-configure
# Build: docker build -t golangci-linter-auto-configure .
# Run: docker run --rm -v $(pwd):/app golangci-linter-auto-configure analyze

# =============================================================================
# Build Stage
# =============================================================================
FROM golang:1.26-alpine AS builder

# Install git for go install
RUN apk add --no-cache git

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the binary
# -s: strip symbols (smaller binary)
# -w: omit DWARF symbols (smaller binary)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /usr/local/bin/golangci-linter-auto-configure \
    ./cmd/golangci-linter-auto-configure

# =============================================================================
# Runtime Stage
# =============================================================================
FROM golangci/golangci-lint:2.1.5-alpine AS runtime

# Install git (needed for version check)
RUN apk add --no-cache git bash

# Copy the auto-configure binary from builder
COPY --from=builder /usr/local/bin/golangci-linter-auto-configure /usr/local/bin/

# Copy example configurations
COPY examples/ /examples/

# Set working directory
WORKDIR /app

# Default command
CMD ["golangci-linter-auto-configure", "--help"]

# =============================================================================
# Slim Variant (alternative, smaller but without golangci-lint)
# =============================================================================
# FROM alpine:3.20 AS slim
#
# RUN apk add --no-cache git bash
#
# COPY --from=builder /usr/local/bin/golangci-linter-auto-configure /usr/local/bin/
# COPY examples/ /examples/
#
# CMD ["golangci-linter-auto-configure", "--help"]

# =============================================================================
# Usage Examples
# =============================================================================
#
# Build the image:
#   docker build -t golangci-linter-auto-configure .
#
# Analyze current project:
#   docker run --rm -v $(pwd):/app golangci-linter-auto-configure analyze
#
# Configure with dry-run:
#   docker run --rm -v $(pwd):/app golangci-linter-auto-configure configure --dry-run
#
# Configure for high priority linters:
#   docker run --rm -v $(pwd):/app golangci-linter-auto-configure configure --priority high
#
# Run with custom config:
#   docker run --rm -v $(pwd):/app -v $(pwd)/.golangci.yml:/app/.golangci.yml \
#       golangci-linter-auto-configure analyze
#
# Use as base image in Dockerfile:
#   FROM golangci-linter-auto-configure AS linter
#
# CI/CD with GitHub Actions:
#   - name: Lint with golangci-linter-auto-configure
#     uses: docker://golangci-linter-auto-configure
#     with:
#       args: analyze