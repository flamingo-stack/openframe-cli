# Local Development Guide

This guide covers cloning the repository, building the CLI from source, running it locally, and configuring your debug environment.

---

## Clone the Repository

```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
```

---

## Project Layout

```text
openframe-cli/
├── main.go                         # Binary entry point
├── cmd/                            # CLI command definitions (Cobra)
│   ├── root.go                     # Root command, global flags, version
│   ├── bootstrap/bootstrap.go      # bootstrap command
│   ├── cluster/                    # cluster subcommands
│   ├── chart/                      # chart subcommands
│   └── dev/                        # dev subcommands
├── internal/                       # Private packages
│   ├── bootstrap/service.go        # Bootstrap orchestration
│   ├── cluster/                    # Cluster lifecycle (K3D)
│   ├── chart/                      # ArgoCD + Helm chart installation
│   ├── dev/                        # Telepresence + Skaffold workflows
│   └── shared/                     # Executor, UI, errors, config
├── tests/
│   ├── integration/                # Integration tests (real k3d required)
│   ├── mocks/                      # Mock implementations
│   └── testutil/                   # Test helpers and fixtures
└── go.mod                          # Go module definition
```

---

## Install Dependencies

```bash
go mod download
```

This downloads all Go module dependencies declared in `go.mod` and `go.sum`.

---

## Build the CLI

### Quick Build

```bash
go build -o openframe main.go
```

This produces an `openframe` binary in the current directory.

### Build with Version Info

```bash
go build \
  -ldflags "-X main.Version=dev-local -X main.Commit=$(git rev-parse --short HEAD) -X main.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o openframe \
  main.go
```

### Verify the Build

```bash
./openframe --version
./openframe --help
```

---

## Run Locally

You can run the CLI directly without installing it system-wide:

```bash
# Run from the project root
./openframe --help
./openframe cluster list
./openframe bootstrap --help
```

Or add the project directory to your `$PATH` temporarily:

```bash
export PATH="$PWD:$PATH"
openframe --help
```

---

## Running Tests

### Unit Tests

Unit tests use mock executors and do not require Docker or K3D:

```bash
go test ./...
```

Run with verbose output:

```bash
go test -v ./...
```

Run with race detection:

```bash
go test -race ./...
```

Run a specific package:

```bash
go test -v ./internal/cluster/...
go test -v ./internal/chart/...
```

### Integration Tests

Integration tests execute real K3D operations and require Docker + K3D installed and running:

```bash
# Ensure K3D and Docker are running first
docker info
k3d version

# Run integration tests
go test -v -tags integration ./tests/integration/...
```

> **Warning:** Integration tests create and delete real K3D clusters. They take several minutes to run and should not be run in environments without sufficient resources (see [Prerequisites](../../../getting-started/prerequisites.md)).

---

## Watch Mode (Auto-Rebuild)

For a fast feedback loop during development, use `go run` with automatic reloads via [air](https://github.com/air-verse/air):

```bash
# Install air
go install github.com/air-verse/air@latest

# Run with hot reload
air
```

Or simply use `go run` for quick iteration:

```bash
go run main.go --help
go run main.go cluster list
```

---

## Debug Configuration

### VS Code Debug Launch Config

Create `.vscode/launch.json` in the project root:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug: openframe bootstrap",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/main.go",
      "args": ["bootstrap", "--verbose"],
      "env": {
        "KUBECONFIG": "${env:HOME}/.kube/config"
      }
    },
    {
      "name": "Debug: cluster list",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/main.go",
      "args": ["cluster", "list"]
    },
    {
      "name": "Debug: chart install",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/main.go",
      "args": ["chart", "install", "--verbose"]
    }
  ]
}
```

### GoLand / IntelliJ Run Configuration

1. Go to **Run → Edit Configurations**
2. Click **+** → **Go Build**
3. Set **Run kind** to `File`
4. Set **Files** to `main.go`
5. Set **Program arguments** to e.g. `bootstrap --verbose`
6. Click **OK** and press **Debug**

---

## Linting

Run the linter before committing:

```bash
golangci-lint run ./...
```

Auto-fix simple issues:

```bash
golangci-lint run --fix ./...
```

---

## Code Formatting

```bash
# Format all Go files
gofmt -w .

# Format and fix imports
goimports -w .
```

---

## Build the CLI Binary for Integration Tests

The integration test suite uses a pre-built binary. To build it:

```bash
go build -o tests/integration/openframe main.go
```

The test suite (`tests/integration/common/cli_runner.go`) will automatically detect and cache this binary based on source file timestamps.

---

## Common Development Commands Reference

| Task | Command |
|---|---|
| Build binary | `go build -o openframe main.go` |
| Run all tests | `go test ./...` |
| Run with race detection | `go test -race ./...` |
| Lint | `golangci-lint run ./...` |
| Format code | `goimports -w .` |
| Run specific test | `go test -v -run TestClusterCreate ./internal/cluster/...` |
| Download deps | `go mod download` |
| Tidy deps | `go mod tidy` |
| Check vulnerabilities | `govulncheck ./...` |
