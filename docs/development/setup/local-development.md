# Local Development Guide

This guide walks you through cloning the repository, building the CLI from source, running it locally, and setting up debug configurations.

---

## Clone the Repository

```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
```

---

## Project Structure

```text
openframe-cli/
├── main.go                    # Application entry point
├── go.mod                     # Go module definition
├── go.sum                     # Dependency lock file
├── cmd/                       # Cobra CLI command definitions
│   ├── root.go
│   ├── bootstrap/
│   ├── cluster/
│   ├── chart/
│   └── dev/
├── internal/                  # Internal packages (not exported)
│   ├── bootstrap/
│   ├── cluster/
│   ├── chart/
│   ├── dev/
│   └── shared/
└── tests/                     # Integration and unit test utilities
    ├── testutil/
    ├── integration/
    └── mocks/
```

---

## Install Dependencies

```bash
go mod download
```

Verify all modules are downloaded and tidy:

```bash
go mod tidy
go mod verify
```

---

## Build the CLI

### Standard Build

```bash
go build -o openframe ./main.go
```

### Build with Version Information

```bash
go build \
  -ldflags="-X github.com/flamingo-stack/openframe-cli/cmd.version=dev \
            -X github.com/flamingo-stack/openframe-cli/cmd.commit=$(git rev-parse --short HEAD) \
            -X github.com/flamingo-stack/openframe-cli/cmd.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o openframe \
  ./main.go
```

### Verify the Build

```bash
./openframe --help
./openframe --version
```

---

## Running Locally

You can run the CLI directly using `go run` without building a binary:

```bash
# Run with go run
go run ./main.go --help
go run ./main.go cluster list
go run ./main.go bootstrap --help
```

Or use the compiled binary:

```bash
./openframe cluster list
./openframe bootstrap my-dev-cluster --verbose
```

---

## Running Commands with Verbose Output

During development, always use `--verbose` / `-v` to see detailed logs:

```bash
./openframe bootstrap my-dev-cluster --verbose
./openframe cluster create test-cluster --verbose
./openframe chart install test-cluster --deployment-mode=oss-tenant --verbose
```

---

## Hot Reload / Watch Mode

The CLI is a compiled binary — there is no native hot-reload. Use the following workflow for rapid iteration:

### Using a Watch Script

```bash
# Install air (Go live reload tool)
go install github.com/cosmtrek/air@latest

# Or use a simple rebuild loop in your terminal
while true; do
  go build -o openframe ./main.go && echo "Build OK"
  inotifywait -r -e modify ./cmd/ ./internal/ 2>/dev/null
done
```

### Manual Rebuild Pattern

The most common pattern during development:

```bash
go build -o openframe ./main.go && ./openframe <your-command>
```

---

## Debug Configuration

### VS Code — `launch.json`

Add the following to `.vscode/launch.json` for debugging with VS Code:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug: bootstrap",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/main.go",
      "args": ["bootstrap", "--verbose"],
      "env": {
        "OPENFRAME_FANCY_LOGO": "false"
      }
    },
    {
      "name": "Debug: cluster create",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/main.go",
      "args": ["cluster", "create", "debug-cluster", "--skip-wizard"],
      "env": {
        "OPENFRAME_FANCY_LOGO": "false"
      }
    },
    {
      "name": "Debug: cluster list",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/main.go",
      "args": ["cluster", "list"]
    }
  ]
}
```

### GoLand

1. Create a **Run/Debug Configuration** → **Go Build**
2. Set **Package path** to `github.com/flamingo-stack/openframe-cli`
3. Set **Program arguments** to the command you want to debug (e.g., `bootstrap --verbose`)
4. Add environment variables as needed

---

## Running Tests During Development

```bash
# Run all unit tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test -v ./internal/cluster/...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

See the [Testing Guide](../testing/README.md) for full details on test organization and writing new tests.

---

## Linting

Run the linter before submitting any changes:

```bash
golangci-lint run ./...
```

Fix formatting issues:

```bash
gofmt -w .
goimports -w .
```

---

## Working with the Mock Executor

The CLI uses a `CommandExecutor` interface to abstract all shell command execution. During development and testing, you can use the mock executor to simulate command outputs without running real tools:

```go
import "github.com/flamingo-stack/openframe-cli/internal/shared/executor"
import "github.com/flamingo-stack/openframe-cli/tests/testutil"

// Create a mock executor
mockExecutor := testutil.NewTestMockExecutor()

// Configure mock responses
mockExecutor.SetResponse("k3d cluster list", &executor.CommandResult{
    ExitCode: 0,
    Stdout:   `[{"name": "test-cluster"}]`,
})

// Inject into a service
clusterService := cluster.NewClusterService(mockExecutor, true)
```

---

## Common Development Tasks

### Adding a New Command

1. Create a new file under `cmd/<group>/<command>.go`
2. Define the Cobra command with `Use`, `Short`, `Long`, `RunE`
3. Register the command in the parent group's `cmd/<group>/<group>.go`
4. Create a corresponding service in `internal/<group>/services/`
5. Write unit tests in `internal/<group>/services/<service>_test.go`

### Adding a New Provider

1. Create the provider under `internal/<group>/providers/<tool>/`
2. Implement the provider interface defined in `internal/<group>/utils/types/interfaces.go`
3. Add the provider to the prerequisite checker if it requires external tool installation

### Modifying Helm Values Wizard

The configuration wizard is in `internal/chart/ui/configuration/wizard.go`. It coordinates:
- `BranchConfigurator` — Git branch selection
- `DockerConfigurator` — Docker registry settings
- `IngressConfigurator` — Ingress/domain configuration

---

## Useful Development Commands

```bash
# Check for vulnerabilities
govulncheck ./...

# List all available CLI commands
./openframe --help

# Dry-run a cluster create (no actual cluster created)
./openframe cluster create test --dry-run

# Dry-run a chart install
./openframe chart install my-cluster --dry-run

# Build integration test binary
go test -c ./tests/integration/... -o /tmp/openframe-integration-tests
```

---

## Next Steps

- Review the [Architecture Overview](../architecture/README.md) to understand how components fit together
- Read the [Contributing Guidelines](../contributing/guidelines.md) before opening a pull request
