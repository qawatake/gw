.PHONY: build install clean test fmt lint help

# Binary name
BINARY_NAME=gw
# Build directory
BUILD_DIR=bin
# Installation directory
INSTALL_DIR=$(HOME)/bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOFMT=$(GOCMD) fmt
GOMOD=$(GOCMD) mod

# Build the project
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/gw
	@echo "✓ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Install to ~/bin
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	@mkdir -p $(INSTALL_DIR)
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "✓ Installed to $(INSTALL_DIR)/$(BINARY_NAME)"
	@echo ""
	@echo "Add the following to your shell config (~/.bashrc or ~/.zshrc):"
	@echo "  eval \"\$$(gw init)\""

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@echo "✓ Clean complete"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...
	@echo "✓ Format complete"

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy
	@echo "✓ Tidy complete"

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
		echo "✓ Lint complete"; \
	else \
		echo "golangci-lint not found. Install it with:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Build for multiple platforms
build-all: build-linux build-darwin build-windows

build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/gw
	@echo "✓ Linux build complete"

build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/gw
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/gw
	@echo "✓ macOS build complete"

build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/gw
	@echo "✓ Windows build complete"

# Uninstall from ~/bin
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "✓ Uninstalled"
	@echo ""
	@echo "Don't forget to remove the following from your shell config:"
	@echo "  eval \"\$$(gw init)\""

# Show help
help:
	@echo "Available targets:"
	@echo "  make build        - Build the binary"
	@echo "  make install      - Build and install to ~/bin"
	@echo "  make uninstall    - Remove from ~/bin"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make test         - Run tests"
	@echo "  make fmt          - Format code"
	@echo "  make tidy         - Tidy dependencies"
	@echo "  make lint         - Run linter (requires golangci-lint)"
	@echo "  make build-all    - Build for all platforms"
	@echo "  make help         - Show this help message"

# Default target
.DEFAULT_GOAL := help
