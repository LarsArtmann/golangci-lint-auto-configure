# Contributing

Thanks for your interest in contributing!

## Prerequisites

- **Go 1.26+** — the codebase uses `encoding/json/v2` (experimental in Go 1.26)
- **Nix** (recommended) — provides a reproducible dev shell with all tools
- **golangci-lint v2.x**

## Development Setup

### Option A: Nix dev shell (recommended)

```bash
nix develop          # enters dev shell with Go, golangci-lint, templ, gopls, etc.
```

The dev shell automatically sets `GOEXPERIMENT=jsonv2` (required — the codebase
imports `encoding/json/v2` which is behind this experiment flag in Go 1.26).

### Option B: Manual setup

If you're not using Nix, you **must** export this environment variable before
any `go` command:

```bash
export GOEXPERIMENT=jsonv2   # required for encoding/json/v2
```

Without it, `go build` and `go test` will fail with
"build constraints exclude all Go files in encoding/json/v2".

## Common Commands

```bash
# Build
go build ./...

# Test (Ginkgo BDD specs, run via go test)
GOEXPERIMENT=jsonv2 go test -race ./pkg/... ./internal/...

# Lint
golangci-lint run --config=.golangci.yml --timeout=5m

# Format Go code
gofumpt -w .
goimports -w .

# Generate templ output (after editing .templ files)
templ generate
```

## Reporting Issues

Please use GitHub Issues to report bugs or request features.
