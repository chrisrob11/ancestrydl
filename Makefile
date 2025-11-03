.PHONY: build test lint clean install help

# Binary name
BINARY_NAME=ancestrydl

# Version information
VERSION ?= dev
BUILD_DATE := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Build flags
# Use -linkmode=external to ensure LC_UUID is generated on macOS
LDFLAGS := -ldflags "-linkmode=external -X main.Version=$(VERSION) -X main.BuildDate=$(BUILD_DATE)"

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@rm -f $(BINARY_NAME)
	go build $(LDFLAGS) -o $(BINARY_NAME) .

# Run tests
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...

# Run tests with coverage report
test-coverage: test
	@echo "Generating coverage report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@rm -rf dist/

# Install the binary
install: build
	@echo "Installing $(BINARY_NAME)..."
	@go install .

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Run the application
run: build
	@./$(BINARY_NAME)

# Display help
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint          - Run linter"
	@echo "  clean         - Clean build artifacts"
	@echo "  install       - Install the binary"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  run           - Build and run the application"
	@echo "  help          - Display this help message"
