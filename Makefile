.PHONY: build test clean dev install lint fmt vet

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet
GOLINT=golangci-lint

# Binary name
BINARY_NAME=pricetrek
BINARY_UNIX=$(BINARY_NAME)_unix
BINARY_WINDOWS=$(BINARY_NAME).exe
BINARY_DARWIN=$(BINARY_NAME)_darwin

# Build directory
BUILD_DIR=build

# Version
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Default target
all: clean fmt vet test build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for multiple platforms
build-all: clean fmt vet test
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_UNIX) .
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_WINDOWS) .
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_DARWIN) .
	@echo "Build complete for all platforms"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "Clean complete"

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .

# Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

# Run linter
lint:
	@echo "Running linter..."
	@if command -v $(GOLINT) >/dev/null 2>&1; then \
		$(GOLINT) run; \
	else \
		echo "golangci-lint not found, installing..."; \
		$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint; \
		$(GOLINT) run; \
	fi

# Development workflow
dev: fmt vet test build
	@echo "Development build complete"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Install the binary
install: build
	@echo "Installing $(BINARY_NAME)..."
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installation complete"

# Uninstall the binary
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "Uninstallation complete"

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Run with specific command
run-init: build
	@echo "Running pricetrek init..."
	./$(BUILD_DIR)/$(BINARY_NAME) init

# Generate mocks (if using mockgen)
mocks:
	@echo "Generating mocks..."
	@if command -v mockgen >/dev/null 2>&1; then \
		mockgen -source=internal/storage/storage.go -destination=internal/storage/mocks/storage_mock.go; \
		mockgen -source=internal/providers/provider.go -destination=internal/providers/mocks/provider_mock.go; \
	else \
		echo "mockgen not found, installing..."; \
		$(GOGET) -u github.com/golang/mock/mockgen; \
		mockgen -source=internal/storage/storage.go -destination=internal/storage/mocks/storage_mock.go; \
		mockgen -source=internal/providers/provider.go -destination=internal/providers/mocks/provider_mock.go; \
	fi

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t pricetrek:$(VERSION) .
	docker tag pricetrek:$(VERSION) pricetrek:latest

# Docker run
docker-run: docker-build
	@echo "Running Docker container..."
	docker run --rm -it pricetrek:latest

# Release preparation
release: clean fmt vet test build-all
	@echo "Preparing release..."
	@mkdir -p release
	cp $(BUILD_DIR)/$(BINARY_UNIX) release/$(BINARY_NAME)-linux-amd64
	cp $(BUILD_DIR)/$(BINARY_WINDOWS) release/$(BINARY_NAME)-windows-amd64.exe
	cp $(BUILD_DIR)/$(BINARY_DARWIN) release/$(BINARY_NAME)-darwin-amd64
	@echo "Release files prepared in release/ directory"

# Help
help:
	@echo "Available targets:"
	@echo "  build        - Build the application"
	@echo "  build-all    - Build for all platforms"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage report"
	@echo "  clean        - Clean build artifacts"
	@echo "  fmt          - Format code"
	@echo "  vet          - Run go vet"
	@echo "  lint         - Run linter"
	@echo "  dev          - Development workflow (fmt, vet, test, build)"
	@echo "  deps         - Install dependencies"
	@echo "  install      - Install the binary"
	@echo "  uninstall    - Uninstall the binary"
	@echo "  run          - Run the application"
	@echo "  run-init     - Run pricetrek init"
	@echo "  mocks        - Generate mocks"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  release      - Prepare release files"
	@echo "  help         - Show this help"