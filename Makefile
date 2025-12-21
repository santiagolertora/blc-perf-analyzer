.PHONY: build test clean install bench coverage lint help

# Binary name
BINARY_NAME=blc-perf-analyzer
OUTPUT_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Build parameters
BUILD_FLAGS=-v
LDFLAGS=-s -w

all: test build

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(OUTPUT_DIR)
	$(GOBUILD) $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME) cmd/blc-perf-analyzer/main.go
	@echo "✓ Build complete: $(OUTPUT_DIR)/$(BINARY_NAME)"

## build-linux: Build for Linux (useful for cross-compilation)
build-linux:
	@echo "Building $(BINARY_NAME) for Linux..."
	@mkdir -p $(OUTPUT_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME)-linux cmd/blc-perf-analyzer/main.go
	@echo "✓ Linux build complete: $(OUTPUT_DIR)/$(BINARY_NAME)-linux"

## test: Run all tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./internal/...
	@echo "✓ All tests passed"

## test-coverage: Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -cover -coverprofile=coverage.txt ./internal/...
	@echo "✓ Coverage report saved to coverage.txt"
	@echo "\nCoverage summary:"
	@$(GOCMD) tool cover -func=coverage.txt | tail -1

## test-coverage-html: Generate HTML coverage report
coverage-html: coverage
	@echo "Generating HTML coverage report..."
	$(GOCMD) tool cover -html=coverage.txt -o coverage.html
	@echo "✓ HTML coverage report saved to coverage.html"

## bench: Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./internal/parser ./internal/heatmap
	@echo "✓ Benchmarks complete"

## lint: Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
		echo "✓ Linting complete"; \
	else \
		echo "⚠ golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...
	@echo "✓ Code formatted"

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(OUTPUT_DIR)/$(BINARY_NAME)*
	rm -f coverage.txt coverage.html
	rm -rf blc-perf-analyzer-2*
	@echo "✓ Clean complete"

## install: Install to /usr/local/bin (requires sudo)
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo cp $(OUTPUT_DIR)/$(BINARY_NAME) /usr/local/bin/
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "✓ Installation complete"

## uninstall: Remove from /usr/local/bin (requires sudo)
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "✓ Uninstall complete"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOGET) -v ./...
	$(GOMOD) tidy
	@echo "✓ Dependencies downloaded"

## help: Show this help message
help:
	@echo "BLC Perf Analyzer - Makefile Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
	@echo ""

