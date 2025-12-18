# OpenFrame CLI Makefile

.PHONY: build build-all build-current clean test test-unit test-integration test-all help

# Variables
BINARY_NAME=openframe
BUILD_DIR=build

# Detect current OS and architecture
GOOS_CURRENT=$(shell go env GOOS)
GOARCH_CURRENT=$(shell go env GOARCH)

# Binary suffix for current platform
ifeq ($(GOOS_CURRENT),windows)
    BINARY_SUFFIX=.exe
else
    BINARY_SUFFIX=
endif

# Default target
all: build

## Build binaries for all platforms
build-all:
	@echo "Building $(BINARY_NAME) for all platforms..."
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(BINARY_NAME)-linux-amd64 .
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o $(BINARY_NAME)-darwin-amd64 .
	@GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o $(BINARY_NAME)-darwin-arm64 .
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o $(BINARY_NAME)-windows-amd64.exe .
	@echo "Built all platform binaries"

## Build binary for current platform only
build-current:
	@echo "Building $(BINARY_NAME) for $(GOOS_CURRENT)/$(GOARCH_CURRENT)..."
	@CGO_ENABLED=0 go build -o $(BINARY_NAME)-$(GOOS_CURRENT)-$(GOARCH_CURRENT)$(BINARY_SUFFIX) .
	@echo "Built ./$(BINARY_NAME)-$(GOOS_CURRENT)-$(GOARCH_CURRENT)$(BINARY_SUFFIX)"

## Build binary for current platform (default)
build: build-current

## Run unit tests
test-unit:
	@echo "Running unit tests..."
	@go test -vet=off ./cmd/... ./internal/...

## Run integration tests  
test-integration:
	@echo "Running integration tests..."
	@go test ./tests/integration/...

## Run all tests
test-all: test-unit test-integration

## Run tests (default)
test: test-all


## Clean build artifacts
clean:
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME) $(BINARY_NAME)-*
	@echo "Cleaned build artifacts"

## Show help
help:
	@echo "Available targets:"
	@echo "  build           - Build binary for current platform (default)"
	@echo "  build-current   - Build binary for current platform"
	@echo "  build-all       - Build binaries for all platforms"
	@echo "  test-unit       - Run unit tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-all        - Run all tests"
	@echo "  test            - Run all tests (default)"
	@echo "  clean           - Clean build artifacts"