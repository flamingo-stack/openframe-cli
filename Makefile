# OpenFrame CLI Makefile

.PHONY: build build-all clean test test-unit test-race test-integration lint help

# Variables
BINARY_NAME := openframe
GO_BUILD := CGO_ENABLED=0 go build

# Detect current OS and architecture
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
BINARY_SUFFIX := $(if $(filter windows,$(GOOS)),.exe,)

# Default target
all: build

## Build binary for current platform
build:
	@echo "Building $(BINARY_NAME) for $(GOOS)/$(GOARCH)..."
	@$(GO_BUILD) -o $(BINARY_NAME)-$(GOOS)-$(GOARCH)$(BINARY_SUFFIX) .

## Build binaries for all platforms
build-all:
	@echo "Building $(BINARY_NAME) for all platforms..."
	@GOOS=linux GOARCH=amd64 $(GO_BUILD) -o $(BINARY_NAME)-linux-amd64 .
	@GOOS=darwin GOARCH=arm64 $(GO_BUILD) -o $(BINARY_NAME)-darwin-arm64 .
	@GOOS=windows GOARCH=amd64 $(GO_BUILD) -o $(BINARY_NAME)-windows-amd64.exe .

## Run unit tests (vet enabled; -vet=off removed per audit remediation §0).
## Includes the root package (main_test.go — the only exit-code fidelity tests)
## and tests/testutil, which `./cmd/... ./internal/...` silently skipped.
## Deliberately NOT ./tests/integration/... — those create real k3d clusters.
test-unit:
	@echo "Running unit tests..."
	@go test -count=1 . ./cmd/... ./internal/... ./tests/testutil/...

## Run unit tests with the race detector (CGO required)
test-race:
	@echo "Running unit tests with -race..."
	@CGO_ENABLED=1 go test -race -count=1 . ./cmd/... ./internal/... ./tests/testutil/...

## Run golangci-lint (static-analysis gate: govet, staticcheck, errcheck, gosec, ineffassign)
lint:
	@echo "Running golangci-lint..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed: https://golangci-lint.run/usage/install/"; exit 1; }
	@golangci-lint run ./...

## Run integration tests
test-integration:
	@echo "Running integration tests..."
	@go test ./tests/integration/...

## Run all tests
test: test-unit test-integration

## Clean build artifacts
clean:
	@rm -f $(BINARY_NAME) $(BINARY_NAME)-*
	@echo "Cleaned build artifacts"

## Show help
help:
	@echo "Available targets:"
	@echo "  build            - Build binary for current platform (default)"
	@echo "  build-all        - Build binaries for all platforms"
	@echo "  test             - Run all tests"
	@echo "  test-unit        - Run unit tests (vet enabled)"
	@echo "  test-race        - Run unit tests with the race detector"
	@echo "  lint             - Run golangci-lint static-analysis gate"
	@echo "  test-integration - Run integration tests"
	@echo "  clean            - Clean build artifacts"
