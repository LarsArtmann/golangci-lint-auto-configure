# Justfile for golangci-lint-auto-configure
# Common development tasks

# Version metadata (shared across build targets)
VERSION := `git describe --tags --always --dirty 2>/dev/null || echo dev`
COMMIT := `git rev-parse --short HEAD 2>/dev/null || echo none`
DATE := `date -u +%Y-%m-%dT%H:%M:%SZ`
TREE_STATE := `if git diff --quiet 2>/dev/null; then echo clean; else echo dirty; fi`

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
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Generating templ files..."
    templ generate
    echo "Building CLI..."
    GOTOOLCHAIN=local go build -ldflags "-X github.com/larsartmann/golangci-lint-auto-configure/pkg/version.version={{VERSION}} -X github.com/larsartmann/golangci-lint-auto-configure/pkg/version.commit={{COMMIT}} -X github.com/larsartmann/golangci-lint-auto-configure/pkg/version.date={{DATE}} -X github.com/larsartmann/golangci-lint-auto-configure/pkg/version.treeState={{TREE_STATE}}" -o bin/golangci-lint-auto-configure ./cmd/golangci-lint-auto-configure
    echo "Built v{{VERSION}}"

test:
    @echo "Running tests..."
    @GOTOOLCHAIN=local go run github.com/onsi/ginkgo/v2/ginkgo -r --cover

test-coverage:
    @echo "Test coverage summary:"
    @GOTOOLCHAIN=local go test ./... -coverprofile=coverage.out -covermode=atomic 2>&1 | grep coverage:
    @echo ""
    @echo "Total coverage:"
    @go tool cover -func=coverage.out | grep total | awk '{print "  " $$3 " of statements"}'

coverage-html:
    @echo "Generating HTML coverage report..."
    @GOTOOLCHAIN=local go test ./... -coverprofile=coverage.out -covermode=atomic > /dev/null 2>&1
    @GOTOOLCHAIN=local go tool cover -html=coverage.out -o coverage.html
    @echo "HTML coverage report: coverage.html"
    @open coverage.html 2>/dev/null || echo "Open coverage.html in your browser"

lint:
    @echo "Running linters..."
    @GOTOOLCHAIN=local golangci-lint run --config .golangci.yml

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
    @cp bin/golangci-lint-auto-configure "$(go env GOPATH)/bin/golangci-lint-auto-configure"
    @echo "Installed to $(go env GOPATH)/bin/"

install-local: build
    @echo "Installing locally..."
    @GOPATH=$(go env GOPATH) && cp bin/golangci-lint-auto-configure "$GOPATH/bin/golangci-lint-auto-configure"
    @echo "Installed v{{VERSION}} to $(go env GOPATH)/bin/"

fmt:
    @echo "Formatting code..."
    @go fmt ./...

fmt-check:
    @echo "Checking formatting..."
    @test -z "$(gofmt -l .)"

tidy:
    @echo "Tidying go.mod..."
    @GOTOOLCHAIN=local go mod tidy

deps:
    @echo "Installing dependencies..."
    @GOTOOLCHAIN=local go mod download

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

# Bump version, commit, push (GitHub Actions auto-tags on push to master)
# Usage: just release 0.2.0
release VERSION:
    #!/usr/bin/env bash
    set -euo pipefail
    new_version="{{ VERSION }}"
    
    # Validate semver format
    if ! echo "$new_version" | grep -qP '^\d+\.\d+(\.\d+)?$'; then
        echo "ERROR: Version must be semver (X.Y.Z), got: $new_version"
        exit 1
    fi
    
    # Find the file containing the version
    version_file=""
    for f in flake.nix nix/packages/default.nix package.nix; do
        if [ -f "$f" ] && grep -qP 'version\s*=\s*"[^"]' "$f"; then
            version_file="$f"
            break
        fi
    done
    
    if [ -z "$version_file" ]; then
        echo "ERROR: No version found in flake.nix, nix/packages/default.nix, or package.nix"
        exit 1
    fi
    
    old_version=$(grep -oP 'version\s*=\s*"\K[^"]+' "$version_file" | head -1)
    echo "Bumping $old_version -> $new_version in $version_file"
    
    sed -i "s|version = \"$old_version\"|version = \"$new_version\"|" "$version_file"
    
    git add "$version_file"
    git commit -m "release: v$new_version"
    git push
    echo "Pushed. GitHub Actions will auto-tag v$new_version."
