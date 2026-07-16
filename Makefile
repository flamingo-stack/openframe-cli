# OpenFrame CLI Makefile

.PHONY: all build build-all clean test test-unit test-race test-integration lint fmt vet tidy help

# Variables
BINARY_NAME := openframe
# -trimpath drops absolute build paths from the binary (reproducibility), matching
# the release build in .goreleaser.yml.
GO_BUILD := CGO_ENABLED=0 go build -trimpath

# Detect current OS and architecture
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
BINARY_SUFFIX := $(if $(filter windows,$(GOOS)),.exe,)

# Unit-test package set. Includes the root package (main_test.go — the only
# exit-code fidelity tests) and tests/testutil, which `./cmd/... ./internal/...`
# silently skipped. Deliberately excludes ./tests/integration/... (real clusters).
UNIT_PKGS := . ./cmd/... ./internal/... ./tests/testutil/...

# Default target
all: build

build: ## Build binary for the current platform
	@echo "Building $(BINARY_NAME) for $(GOOS)/$(GOARCH)..."
	@$(GO_BUILD) -o $(BINARY_NAME)-$(GOOS)-$(GOARCH)$(BINARY_SUFFIX) .

build-all: ## Cross-compile every release platform (matches .goreleaser.yml)
	@echo "Building $(BINARY_NAME) for all release platforms..."
	@GOOS=linux   GOARCH=amd64 $(GO_BUILD) -o $(BINARY_NAME)-linux-amd64 .
	@GOOS=linux   GOARCH=arm64 $(GO_BUILD) -o $(BINARY_NAME)-linux-arm64 .
	@GOOS=darwin  GOARCH=amd64 $(GO_BUILD) -o $(BINARY_NAME)-darwin-amd64 .
	@GOOS=darwin  GOARCH=arm64 $(GO_BUILD) -o $(BINARY_NAME)-darwin-arm64 .
	@GOOS=windows GOARCH=amd64 $(GO_BUILD) -o $(BINARY_NAME)-windows-amd64.exe .
	@GOOS=windows GOARCH=arm64 $(GO_BUILD) -o $(BINARY_NAME)-windows-arm64.exe .

test-unit: ## Run unit tests (vet on; incl. root main_test.go + tests/testutil)
	@echo "Running unit tests..."
	@go test -count=1 $(UNIT_PKGS)

test-race: ## Run unit tests with the race detector (requires CGO)
	@echo "Running unit tests with -race..."
	@CGO_ENABLED=1 go test -race -count=1 $(UNIT_PKGS)

# Integration tests are opt-in via a build tag: they create REAL k3d clusters and
# run a full bootstrap, so `go test ./...` must never trigger them by accident.
# The harness builds its own CLI binary (see tests/integration/common/cli_runner.go),
# so `make build` is NOT required — but docker and k3d must be installed.
test-integration: ## Run integration tests (real k3d clusters; needs docker + k3d)
	@echo "Running integration tests (real clusters!)..."
	@go test -tags integration -count=1 ./tests/integration/...

test: test-unit test-integration ## Run unit + integration tests

lint: ## Run golangci-lint (govet, staticcheck, errcheck, gosec, ineffassign)
	@echo "Running golangci-lint..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed: https://golangci-lint.run/usage/install/"; exit 1; }
	@golangci-lint run ./...

fmt: ## Format Go source in place (gofmt -w)
	@gofmt -l -w .

vet: ## Run go vet on all packages
	@go vet ./...

tidy: ## Check go.mod/go.sum are tidy (fails if `go mod tidy` would change them)
	@go mod tidy -diff

# clean matches only the platform-suffixed binaries, NOT a broad `openframe-*`
# glob — that also matched tracked files like openframe-helm-values.example.yaml
# and deleted them.
clean: ## Remove build artifacts
	@rm -f $(BINARY_NAME) \
		$(BINARY_NAME)-linux-* $(BINARY_NAME)-darwin-* $(BINARY_NAME)-windows-*
	@echo "Cleaned build artifacts"

help: ## Show this help
	@echo "Available targets:"
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'
