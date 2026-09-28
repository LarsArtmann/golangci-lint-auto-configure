# Contributing

Thanks for your interest in contributing!

## Prerequisites

- **Go 1.27+** — the codebase uses `encoding/json/v2` (non-experimental since Go 1.27)
- **Nix** (recommended) — provides a reproducible dev shell with all tools
- **golangci-lint v2.x**

## Development Setup

### Option A: Nix dev shell (recommended)

```bash
nix develop          # enters dev shell with Go, golangci-lint, templ, gopls, etc.
```

The dev shell still sets `GOEXPERIMENT=jsonv2` (inert on Go 1.27, where
`encoding/json/v2` ships enabled by default; kept for consistency).

### Option B: Manual setup

Any Go 1.27+ toolchain works out of the box — no environment variable needed.
(On Go 1.26 you would have to `export GOEXPERIMENT=jsonv2`; the flag no longer
exists as a gate on 1.27.)

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
