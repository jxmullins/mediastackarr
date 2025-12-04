# MediaStack CLI Makefile

# Variables
BINARY_NAME=mediastack
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X github.com/jxmullins/mediastack/internal/cli.Version=$(VERSION) -X github.com/jxmullins/mediastack/internal/cli.BuildTime=$(BUILD_TIME)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Directories
BUILD_DIR=build
CMD_DIR=cmd/mediastack

.PHONY: all build clean test deps install lint help

# Default target
all: deps build

# Build the binary for current platform
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for all platforms
build-all: build-linux build-darwin build-windows

# Build for Linux (amd64 and arm64)
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./$(CMD_DIR)
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./$(CMD_DIR)
	@echo "Built Linux binaries"

# Build for macOS (amd64 and arm64)
build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./$(CMD_DIR)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./$(CMD_DIR)
	@echo "Built macOS binaries"

# Build for Windows
build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./$(CMD_DIR)
	@echo "Built Windows binary"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Install to GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME)..."
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)
	@echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"

# Install locally to /usr/local/bin
install-local: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "Installed to /usr/local/bin/$(BINARY_NAME)"

# Lint the code
lint:
	@echo "Linting..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping..."; \
	fi

# Format code
fmt:
	@echo "Formatting..."
	$(GOCMD) fmt ./...

# Vet code
vet:
	@echo "Vetting..."
	$(GOCMD) vet ./...

# Run the binary
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

# Show help
help:
	@echo "MediaStack CLI Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make              Build for current platform"
	@echo "  make build        Build for current platform"
	@echo "  make build-all    Build for all platforms (linux, darwin, windows)"
	@echo "  make build-linux  Build for Linux (amd64 + arm64)"
	@echo "  make build-darwin Build for macOS (amd64 + arm64)"
	@echo "  make build-windows Build for Windows (amd64)"
	@echo "  make clean        Remove build artifacts"
	@echo "  make test         Run tests"
	@echo "  make deps         Download dependencies"
	@echo "  make install      Install to GOPATH/bin"
	@echo "  make install-local Install to /usr/local/bin (requires sudo)"
	@echo "  make lint         Run linter"
	@echo "  make fmt          Format code"
	@echo "  make vet          Vet code"
	@echo "  make help         Show this help"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION           Set version (default: git describe or 'dev')"
	@echo ""
	@echo "Examples:"
	@echo "  make VERSION=1.0.0 build"
	@echo "  make build-all"
