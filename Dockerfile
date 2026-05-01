# Multi-stage Dockerfile for golangci-lint-auto-configure
# Build: docker build -t golangci-lint-auto-configure .
# Run: docker run --rm -v $(pwd):/app golangci-lint-auto-configure analyze

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
    -o /usr/local/bin/golangci-lint-auto-configure \
    ./cmd/golangci-lint-auto-configure

# =============================================================================
# Runtime Stage
# =============================================================================
FROM golangci/golangci-lint:v2.1-alpine AS runtime

# Install git (needed for version check)
RUN apk add --no-cache git bash

# Copy the auto-configure binary from builder
COPY --from=builder /usr/local/bin/golangci-lint-auto-configure /usr/local/bin/

# Copy example configurations
COPY examples/ /examples/

# Set working directory
WORKDIR /app

# Default command
CMD ["golangci-lint-auto-configure", "--help"]

# =============================================================================
# Slim Variant (alternative, smaller but without golangci-lint)
# =============================================================================
# FROM alpine:3.20 AS slim
#
# RUN apk add --no-cache git bash
#
# COPY --from=builder /usr/local/bin/golangci-lint-auto-configure /usr/local/bin/
# COPY examples/ /examples/
#
# CMD ["golangci-lint-auto-configure", "--help"]

# =============================================================================
# Usage Examples
# =============================================================================
#
# Build the image:
#   docker build -t golangci-lint-auto-configure .
#
# Analyze current project:
#   docker run --rm -v $(pwd):/app golangci-lint-auto-configure analyze
#
# Configure with dry-run:
#   docker run --rm -v $(pwd):/app golangci-lint-auto-configure configure --dry-run
#
# Configure for high priority linters:
#   docker run --rm -v $(pwd):/app golangci-lint-auto-configure configure --priority high
#
# Run with custom config:
#   docker run --rm -v $(pwd):/app -v $(pwd)/.golangci.yml:/app/.golangci.yml \
#       golangci-lint-auto-configure analyze
#
# Use as base image in Dockerfile:
#   FROM golangci-lint-auto-configure AS linter
#
# CI/CD with GitHub Actions:
#   - name: Lint with golangci-lint-auto-configure
#     uses: docker://golangci-lint-auto-configure
#     with:
#       args: analyze