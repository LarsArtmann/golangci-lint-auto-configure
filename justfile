# Justfile for golangci-lint-auto-configure
# Common development tasks

default: help

help:
    @echo "Available commands:"
    @echo "  just build         - Build the CLI binary"
    @echo "  just test          - Run all tests with coverage"
    @echo "  just test-coverage  - Show coverage report"
    @echo "  just coverage-html  - Generate HTML coverage report"
    @echo "  just lint          - Run linters"
    @echo "  just run           - Run the CLI (default command)"
    @echo "  just clean         - Clean build artifacts"
    @echo "  just install       - Install the CLI to GOPATH/bin"
    @echo "  just install-local  - Install locally with version ldflags"
    @echo "  just dogfood       - Run tool on itself (analyze our own config)"
    @echo ""
    @echo "Nix commands:"
    @echo "  just nix-build     - Build with Nix (reproducible)"
    @echo "  just nix-check     - Run all Nix checks"
    @echo "  just nix-update    - Update Nix flake inputs"
    @echo "  just nix-vendor    - Update vendorHash after go.mod changes"

build:
    @echo "Building CLI..."
    @GOWORK=off GOTOOLCHAIN=local go build -o bin/golangci-lint-auto-configure ./cmd/golangci-lint-auto-configure

test:
    @echo "Running tests..."
    @GOWORK=off GOTOOLCHAIN=local ginkgo -r --cover

test-coverage:
    @echo "Test coverage summary:"
    @GOWORK=off GOTOOLCHAIN=local go test ./... -coverprofile=coverage.out -covermode=atomic 2>&1 | grep coverage:
    @echo ""
    @echo "Total coverage:"
    @go tool cover -func=coverage.out | grep total | awk '{print "  " $$3 " of statements"}'

coverage-html:
    @echo "Generating HTML coverage report..."
    @GOWORK=off GOTOOLCHAIN=local go test ./... -coverprofile=coverage.out -covermode=atomic > /dev/null 2>&1
    @GOWORK=off GOTOOLCHAIN=local go tool cover -html=coverage.out -o coverage.html
    @echo "HTML coverage report: coverage.html"
    @open coverage.html 2>/dev/null || echo "Open coverage.html in your browser"

lint:
    @echo "Running linters..."
    @GOWORK=off GOTOOLCHAIN=local golangci-lint run --config .golangci.yml

run build *args:
    @echo "Running CLI..."
    @./bin/golangci-lint-auto-configure {{args}}

analyze build *args:
    @./bin/golangci-lint-auto-configure analyze {{args}}

configure build *args:
    @./bin/golangci-lint-auto-configure configure {{args}}

validate build *args:
    @./bin/golangci-lint-auto-configure validate {{args}}

dogfood: build
    @echo "🐕 Dogfooding: Running golangci-lint-auto-configure on itself..."
    @./bin/golangci-lint-auto-configure analyze
    @echo ""
    @echo "✅ Dogfooding complete!"

report build *args:
    @./bin/golangci-lint-auto-configure report {{args}}

migrate build *args:
    @./bin/golangci-lint-auto-configure migrate {{args}}

clean:
    @echo "Cleaning build artifacts..."
    @rm -rf bin/
    @echo "Clean!"

install: build
    @echo "Installing CLI..."
    @GOWORK=off GOTOOLCHAIN=local go install ./cmd/golangci-lint-auto-configure

install-local:
    #!/usr/bin/env bash
    set -e
    echo "Installing locally with version..."
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
    GOPATH=$(go env GOPATH)
    GOWORK=off GOTOOLCHAIN=local go build -ldflags "-X main.version=$VERSION" -o "$GOPATH/bin/golangci-lint-auto-configure" ./cmd/golangci-lint-auto-configure
    echo "Installed golangci-lint-auto-configure v$VERSION to $GOPATH/bin/"

fmt:
    @echo "Formatting code..."
    @go fmt ./...

fmt-check:
    @echo "Checking formatting..."
    @test -z "$(gofmt -l .)"

tidy:
    @echo "Tidying go.mod..."
    @GOWORK=off GOTOOLCHAIN=local go mod tidy

deps:
    @echo "Installing dependencies..."
    @GOWORK=off GOTOOLCHAIN=local go mod download

# Nix commands

nix-build:
    @echo "Building with Nix..."
    @nix build
    @echo "Binary: ./result/bin/golangci-lint-auto-configure"

nix-check:
    @echo "Running Nix checks..."
    @nix flake check

nix-update:
    @echo "Updating Nix flake inputs..."
    @nix flake update

nix-vendor:
    #!/usr/bin/env bash
    set -e
    echo "Building to calculate vendorHash..."
    nix build 2>&1 | tail -5
