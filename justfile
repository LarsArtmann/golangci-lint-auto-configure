# Justfile for golangci-linter-auto-configure
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

build:
    @echo "Building CLI..."
    @go build -o bin/golangci-linter-auto-configure ./cmd/golangci-linter-auto-configure

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
    @./bin/golangci-linter-auto-configure {{args}}

analyze build *args:
    @./bin/golangci-linter-auto-configure analyze {{args}}

configure build *args:
    @./bin/golangci-linter-auto-configure configure {{args}}

validate build *args:
    @./bin/golangci-linter-auto-configure validate {{args}}

report build *args:
    @./bin/golangci-linter-auto-configure report {{args}}

migrate build *args:
    @./bin/golangci-linter-auto-configure migrate {{args}}

clean:
    @echo "Cleaning build artifacts..."
    @rm -rf bin/
    @echo "Clean!"

install: build
    @echo "Installing CLI..."
    @go install ./cmd/golangci-linter-auto-configure

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