# List all available commands
default:
    @just --list

# Format all Go code
fmt:
    go fmt ./...
    gofmt -s -w .

# Run go mod tidy to clean up dependencies
tidy:
    go mod tidy

# Verify dependencies
verify:
    go mod verify

# Run tests with coverage
test:
    go test -cover ./...

# Run tests with coverage and generate HTML report
test-coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html

# Run golangci-lint
lint:
    golangci-lint run

# Build debug versions of both commands
build:
    go build -v ./cmd/ai-stdio

# Clean build artifacts
clean:
    rm -f ai-stdio
    rm -f coverage.out
    rm -f coverage.html

# Build optimized release versions for current platform
release:
    go build -v -ldflags="-s -w" ./cmd/ai-stdio

# Build release versions for multiple platforms
release-all: clean
    #!/usr/bin/env sh
    mkdir -p dist
    GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/ai-stdio-linux-amd64 ./cmd/ai-stdio
    GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/ai-stdio-darwin-amd64 ./cmd/ai-stdio
    GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/ai-stdio-darwin-arm64 ./cmd/ai-stdio
    GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/ai-stdio-windows-amd64.exe ./cmd/ai-stdio

# Run all quality checks (format, lint, test)
check: fmt lint test

# Install the commands to $GOPATH/bin
install:
    go install ./cmd/ai-stdio

# Update all dependencies
update:
    go get -u ./...
    go mod tidy

# Run security check using govulncheck
security:
    govulncheck ./...

# Generate and show test coverage statistics
coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out

# Run tests in verbose mode
test-verbose:
    go test -v ./...

# Run tests with race detection
test-race:
    go test -race ./...

# Verify all code is properly formatted
verify-fmt:
    #!/usr/bin/env sh
    if [ -n "$(gofmt -l .)" ]; then
        echo "These files are not formatted correctly:"
        gofmt -l .
        exit 1
    fi
