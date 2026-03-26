# Justfile for golangci-lint-auto-configure
# Common development tasks

default: help

help:
    @echo "Available commands:"
    @echo "  just build        - Build the CLI binary"
    @echo "  just test         - Run all tests with coverage"
    @echo "  just test-coverage - Show coverage report"
    @echo "  just coverage-html - Generate HTML coverage report"
    @echo "  just lint         - Run linters"
    @echo "  just run          - Run the CLI (default command)"
    @echo "  just clean        - Clean build artifacts"
    @echo "  just install      - Install the CLI to GOPATH/bin"
    @echo "  just install-local - Install locally with version ldflags"
    @echo "  just dogfood      - Run tool on itself (analyze our own config)"

build:
    @echo "Building CLI..."
    @go build -o bin/golangci-lint-auto-configure ./cmd/golangci-lint-auto-configure

test:
    @echo "Running tests..."
    @ginkgo -r --cover

# Show test coverage summary
test-coverage:
    @echo "Test coverage summary:"
    @go test ./... -coverprofile=coverage.out -covermode=atomic 2>&1 | grep coverage:
    @echo ""
    @echo "Total coverage:"
    @go tool cover -func=coverage.out | grep total | awk '{print "  " $$3 " of statements"}'

# Generate and open HTML coverage report
coverage-html:
    @echo "Generating HTML coverage report..."
    @go test ./... -coverprofile=coverage.out -covermode=atomic > /dev/null 2>&1
    @go tool cover -html=coverage.out -o coverage.html
    @echo "HTML coverage report: coverage.html"
    @open coverage.html 2>/dev/null || echo "Open coverage.html in your browser"

lint:
    @echo "Running linters..."
    @golangci-lint run --config .golangci.yml

run build *args:
    @echo "Running CLI..."
    @./bin/golangci-lint-auto-configure {{args}}

analyze build *args:
    @./bin/golangci-lint-auto-configure analyze {{args}}

configure build *args:
    @./bin/golangci-lint-auto-configure configure {{args}}

validate build *args:
    @./bin/golangci-lint-auto-configure validate {{args}}

# Dogfood: Run the tool on itself (follows Dogfooding First principle)
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
    @go install ./cmd/golangci-lint-auto-configure

# Install locally with version ldflags
install-local:
    #!/bin/bash
    echo "Installing locally with version..."
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
    GOPATH=$(go env GOPATH)
    go build -ldflags "-X main.version=$VERSION" -o "$GOPATH/bin/golangci-lint-auto-configure" ./cmd/golangci-lint-auto-configure
    echo "Installed golangci-lint-auto-configure v$VERSION to $GOPATH/bin/"

fmt:
    @echo "Formatting code..."
    @go fmt ./...

fmt-check:
    @echo "Checking formatting..."
    @test -z "$(gofmt -l .)"

tidy:
    @echo "Tidying go.mod..."
    @go mod tidy

deps:
    @echo "Installing dependencies..."
    @go mod download