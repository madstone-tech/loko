.PHONY: build test lint fmt clean install dev help

# Variables
BINARY_NAME=loko
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# Default target
help:
	@echo "loko - C4 Architecture Documentation Tool"
	@echo ""
	@echo "Usage:"
	@echo "  make build      Build the binary"
	@echo "  make test       Run all tests"
	@echo "  make test-v     Run tests with verbose output"
	@echo "  make coverage   Run tests with coverage report"
	@echo "  make lint       Run linter"
	@echo "  make fmt        Format code"
	@echo "  make clean      Remove build artifacts"
	@echo "  make install    Install to GOPATH/bin"
	@echo "  make dev        Build and run in development mode"
	@echo "  make deps       Download dependencies"
	@echo "  make tools      Install development tools"

# Build
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

# Test
test:
	go test ./...

test-v:
	go test -v ./...

coverage:
	go test -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Integration tests
test-integration:
	go test -tags=integration -v ./tests/integration/...

# Lint
lint:
	golangci-lint run

# Format
fmt:
	go fmt ./...
	goimports -w .

# Clean
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	rm -rf dist/

# Install
install: build
	cp $(BINARY_NAME) $(GOPATH)/bin/

# Development
dev: build
	./$(BINARY_NAME) --help

# Dependencies
deps:
	go mod download
	go mod tidy

# Install development tools
tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/goreleaser/goreleaser@latest

# Run (for development)
run:
	go run . $(ARGS)

# Watch mode (requires entr or similar)
watch:
	find . -name "*.go" | entr -r make run ARGS="serve"
