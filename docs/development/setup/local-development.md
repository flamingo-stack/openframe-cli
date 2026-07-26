# Local Development Guide

This guide walks through cloning the OpenFrame CLI repository, building it locally, running it, and setting up a productive development workflow.

---

## Clone and Setup

```bash
# Clone the repository
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli

# Download Go module dependencies
go mod download

# Verify dependencies
go mod verify
```

---

## Building the Binary

### Standard build

```bash
go build -o openframe .
```

This produces an `openframe` binary in the current directory.

### Build with version information

In production, the binary is built with version metadata injected via linker flags. Replicate that locally:

```bash
go build \
  -ldflags "-X github.com/flamingo-stack/openframe-cli/cmd.version=dev-local \
            -X github.com/flamingo-stack/openframe-cli/cmd.commit=$(git rev-parse --short HEAD) \
            -X github.com/flamingo-stack/openframe-cli/cmd.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o openframe .
```

Verify the version is embedded:

```bash
./openframe --version
```

Expected output:

```text
openframe version dev-local (abc1234) built on 2024-xx-xxTxx:xx:xxZ
```

### Build for a different platform (cross-compilation)

```bash
# Linux amd64 (from macOS or Windows)
GOOS=linux GOARCH=amd64 go build -o openframe-linux .

# macOS arm64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o openframe-darwin-arm64 .

# Windows amd64
GOOS=windows GOARCH=amd64 go build -o openframe.exe .
```

---

## Running Locally

Use the locally built binary directly:

```bash
./openframe --help
./openframe prerequisites check
./openframe cluster list
```

Or install it to your local `$PATH` for convenience:

```bash
sudo cp openframe /usr/local/bin/openframe-dev
openframe-dev --help
```

---

## Running in Dry-Run Mode

The executor supports a dry-run mode that logs all commands it would execute without actually running them. This is useful for development and testing:

```bash
# Via code: use NewRealCommandExecutor(true, true) in tests
# Via standard usage: the --verbose flag increases logging detail
./openframe bootstrap --verbose
```

---

## Debug Configuration

### Enable verbose logging

```bash
./openframe --verbose bootstrap
./openframe --verbose cluster create
```

The `--verbose` flag surfaces pterm debug-level output, including:
- All external command invocations (k3d, helm)
- ArgoCD sync progress events
- Kubernetes API calls

### Inspect log files

The CLI writes deployment logs to the system temp directory:

```bash
# Linux / macOS
ls /tmp/openframe-deployment-logs/

# Follow logs in real time
tail -f /tmp/openframe-deployment-logs/*.log
```

### Debug with Delve (Go debugger)

Install Delve:

```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

Run the CLI with a debugger attached:

```bash
dlv debug . -- bootstrap --verbose
```

Or debug a specific test:

```bash
dlv test ./cmd/cluster/... -- -run TestCreateCommand -v
```

### VS Code Launch Configuration

Add this to `.vscode/launch.json` for one-click debugging:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug openframe bootstrap",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}",
      "args": ["bootstrap", "--verbose"],
      "env": {
        "OPENFRAME_GITHUB_TOKEN": "${env:OPENFRAME_GITHUB_TOKEN}"
      }
    },
    {
      "name": "Debug openframe cluster list",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}",
      "args": ["cluster", "list", "--output", "json"]
    }
  ]
}
```

---

## Hot Reload / Watch Mode

Go doesn't have built-in hot reload, but you can use `air` for automatic rebuilds on file changes:

```bash
# Install air
go install github.com/air-verse/air@latest

# Run with watch mode (rebuilds on .go file changes)
air -- bootstrap --verbose
```

Or use a simple shell loop for manual iteration:

```bash
# Rebuild and run on every change
while true; do
  go build -o openframe . && ./openframe cluster list
  inotifywait -e modify $(find . -name '*.go' | head -20) 2>/dev/null
done
```

---

## Working with the Integration Tests

The integration test suite builds and exercises the real CLI binary against a live cluster:

```bash
# Build the test binary (cached if already up to date)
# Handled automatically by TestMain in integration tests

# Run all integration tests
go test ./tests/integration/... -v -timeout 30m

# Run a specific integration test
go test ./tests/integration/... -run TestClusterCreate -v
```

> **Note:** Integration tests require Docker running, k3d installed, and sufficient system resources (24GB+ RAM recommended).

---

## Iterating on a Feature

A typical development loop:

```bash
# 1. Create a feature branch
git checkout -b feature/my-new-command

# 2. Make changes to source files

# 3. Build
go build -o openframe .

# 4. Test locally
./openframe my-new-command --help

# 5. Run unit tests
go test ./cmd/... ./internal/...

# 6. Run vet and format
go vet ./...
goimports -w .

# 7. Commit and push
git add -A
git commit -m "feat: add my-new-command"
git push origin feature/my-new-command
```

---

## Useful Development Commands

```bash
# List all commands and flags
./openframe --help

# Run all unit tests with race detector
go test -race ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Tidy dependencies
go mod tidy

# Check for security vulnerabilities
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

---

## Working with the External Platform Repository

The OpenFrame platform chart (`openframe-oss-tenant`) is a separate repository. When developing features that interact with chart deployment, you may need to reference it:

- **Repository:** [https://github.com/flamingo-stack/openframe-oss-tenant](https://github.com/flamingo-stack/openframe-oss-tenant)
- **Documentation:** [https://github.com/flamingo-stack/openframe-oss-tenant/tree/main/docs](https://github.com/flamingo-stack/openframe-oss-tenant/tree/main/docs)

The CLI clones this repository at runtime using `go-git` (no `git` binary required). During development, you can point the CLI at a local fork by modifying the chart configuration.

---

## Next Steps

- Review the [Architecture Overview](../architecture/README.md) to understand component relationships
- Read the [Testing Guide](../testing/README.md) for test writing conventions
- Check the [Contributing Guidelines](../contributing/guidelines.md) before opening a PR
