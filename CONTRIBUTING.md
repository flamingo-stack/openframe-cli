# Contributing to OpenFrame CLI

Thank you for your interest in contributing to OpenFrame CLI! This guide covers how to set up your development environment, the coding standards we follow, and the process for submitting contributions.

> **Community First:** All development discussions, bug reports, and feature requests are managed in the **[OpenMSP Slack community](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)** — not GitHub Issues or GitHub Discussions.

---

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Environment Setup](#development-environment-setup)
- [Building the Project](#building-the-project)
- [Running Tests](#running-tests)
- [Code Style & Quality](#code-style--quality)
- [Architecture Guidelines](#architecture-guidelines)
- [Submitting a Pull Request](#submitting-a-pull-request)
- [Security Issues](#security-issues)

---

## Code of Conduct

Be respectful, collaborative, and constructive. The OpenMSP community is a welcoming place for contributors of all experience levels.

---

## Getting Started

### System Requirements

| Tier | RAM | CPU Cores | Disk Space |
|------|-----|-----------|------------|
| **Minimum** | 24 GB | 6 cores | 50 GB |
| **Recommended** | 32 GB | 12 cores | 100 GB |

### Required Tools

| Tool | Minimum Version | Purpose |
|------|----------------|---------|
| **Go** | 1.21+ | Build the CLI |
| **Docker** | 20.10+ | Run K3D container nodes |
| **k3d** | 5.x | Local Kubernetes clusters |
| **kubectl** | 1.25+ | Kubernetes interaction |
| **Helm** | 3.x | Chart installation |
| **Git** | 2.x | Repository operations |
| **mkcert** | 1.4+ | Local TLS certificates |

### Optional (for `dev` commands)

| Tool | Purpose |
|------|---------|
| **Telepresence** 2.x | Live traffic intercepts |
| **Skaffold** | Hot-reload dev sessions |
| **jq** 1.6+ | JSON processing |

---

## Development Environment Setup

### 1. Go Toolchain

OpenFrame CLI requires **Go 1.21 or newer**.

```bash
# Verify your Go version
go version
# Expected: go version go1.21.x or higher
```

Configure your Go environment by adding to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.):

```bash
export GOPATH=$HOME/go
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOBIN
```

### 2. Clone the Repository

```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
```

### 3. Install Dependencies

```bash
go mod download
go mod verify
```

### 4. Install Code Quality Tools

```bash
# goimports — import organizer
go install golang.org/x/tools/cmd/goimports@latest

# golangci-lint — multi-linter runner
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOBIN) latest

# Verify
goimports --version
golangci-lint --version
```

### 5. Recommended IDE: VS Code

Install these extensions:

```bash
code --install-extension golang.go
code --install-extension eamodio.gitlens
code --install-extension redhat.vscode-yaml
```

Create `.vscode/settings.json`:

```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "go.formatTool": "goimports",
  "editor.formatOnSave": true,
  "[go]": {
    "editor.defaultFormatter": "golang.go",
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
      "source.organizeImports": "always"
    }
  }
}
```

---

## Building the Project

```bash
# Build the binary to the project root
go build -o openframe ./main.go

# Verify
./openframe --version

# Cross-platform builds
GOOS=linux   GOARCH=amd64 go build -o openframe-linux-amd64   ./main.go
GOOS=darwin  GOARCH=arm64 go build -o openframe-darwin-arm64  ./main.go
GOOS=windows GOARCH=amd64 go build -o openframe-windows-amd64.exe ./main.go
```

### Hot Reload with Air

```bash
go install github.com/air-verse/air@latest
air
```

---

## Running Tests

```bash
# Run all unit tests
go test ./...

# Verbose output
go test ./... -v

# Specific package
go test ./internal/cluster/...

# With race detection
go test -race ./...

# Short tests only (skips integration)
go test ./... -short

# Integration tests (requires Docker, k3d, kubectl, helm)
go test ./tests/integration/...

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Understanding the Mock Executor

All external command execution (k3d, helm, kubectl, git) goes through the `CommandExecutor` interface. Unit tests inject a `MockCommandExecutor` that returns pre-configured responses without running real binaries:

```go
// In test files
testutil.InitializeTestMode()
executor := testutil.NewTestMockExecutor()
flags := testutil.CreateStandardTestFlags()
```

This allows full business logic coverage without a real Kubernetes cluster.

---

## Code Style & Quality

### Formatting and Linting

```bash
# Format all Go files
goimports -w .

# Run the linter
golangci-lint run ./...

# Check for compilation errors
go vet ./...

# Tidy dependencies
go mod tidy
```

### Code Organisation Principles

OpenFrame CLI follows a **layered clean architecture**:

1. **`cmd/` layer** — Thin Cobra command definitions only. No business logic. Parse flags, validate combinations, delegate to services.
2. **`internal/*/service*.go`** — Business logic orchestration. Sequence multi-step operations, coordinate providers, manage UI feedback.
3. **`internal/*/providers/`** — Thin wrappers around external tools and APIs. Each provider implements an interface.
4. **`internal/shared/`** — Cross-cutting utilities (executor, UI, errors, config, files).

**Rules:**
- All external subprocess calls go through `CommandExecutor` — never call `os/exec` directly
- Never interpolate user input into shell strings — always use argument slices
- All providers must have a corresponding interface for mock injection
- No business logic in `cmd/` — it is a wiring layer only

### Security Checklist for PRs

Before submitting a pull request, verify:

- [ ] No credentials or secrets hardcoded in source files
- [ ] All user input validated before use as a command argument
- [ ] External commands use argument slices (not string concatenation)
- [ ] Temporary files are cleaned up in both success and failure paths
- [ ] No new uses of insecure TLS outside K3D-scoped local contexts
- [ ] Error messages do not leak sensitive information (tokens, paths, credentials)
- [ ] `helm-values-tmp.yaml` is gitignored in any new workflows

---

## Architecture Guidelines

### Interface-Based External I/O

```go
// All subprocess execution goes through this interface
type CommandExecutor interface {
    Execute(ctx context.Context, command string, args ...string) (*CommandResult, error)
    ExecuteWithOptions(ctx context.Context, opts ExecuteOptions) (*CommandResult, error)
}
```

Any new provider must implement an interface so tests can inject a mock.

### Shell Injection Prevention

```go
// CORRECT: arguments as a slice — user input never interpolated
result, err := executor.Execute(ctx, "k3d", "cluster", "create", clusterName)

// NEVER do this:
// exec.Command("sh", "-c", "k3d cluster create " + clusterName)
```

### Error Handling

Use the typed errors from `internal/shared/errors/`:

```go
return errors.CreateValidationError("cluster-name", name, err.Error())
return errors.NewBranchNotFoundError(branch, repo)
```

---

## Submitting a Pull Request

### Workflow

1. **Discuss first** — Share your idea in the [OpenMSP Slack](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA) before writing code
2. **Fork and branch** — Create a feature branch from `main`
3. **Write tests** — Unit tests for new business logic; integration tests for new commands
4. **Run quality checks** — All linting and tests must pass
5. **Submit PR** — Write a clear description of what and why

### PR Description Template

```text
## What
Brief description of the change.

## Why
The problem this solves or feature this adds.

## How
Key implementation decisions.

## Testing
How you tested the change (unit tests, manual testing, integration tests).

## Checklist
- [ ] Tests pass: go test ./...
- [ ] Linter passes: golangci-lint run ./...
- [ ] No hardcoded credentials
- [ ] Follows layered architecture (no business logic in cmd/)
- [ ] External commands use argument slices (not shell strings)
```

### Commit Message Format

```text
<type>(<scope>): <subject>

<body>
```

Types: `feat`, `fix`, `docs`, `test`, `refactor`, `chore`

Examples:

```text
feat(cluster): add --timeout flag to cluster create command

fix(chart): handle BranchNotFoundError with retry suggestion

docs(dev): update intercept command flag descriptions
```

---

## Security Issues

For security vulnerabilities, **do not open a public GitHub issue**. Report directly via the [OpenMSP Slack community](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA).

---

## Community

- **OpenMSP Slack:** [Join here](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA) — primary support and discussion channel
- **OpenMSP Website:** [https://www.openmsp.ai/](https://www.openmsp.ai/)
- **OpenFrame Platform:** [https://openframe.ai](https://openframe.ai)
- **Flamingo Platform:** [https://flamingo.run](https://flamingo.run)

---

<div align="center">
  Built with 💛 by the <a href="https://www.flamingo.run/about"><b>Flamingo</b></a> team
</div>
