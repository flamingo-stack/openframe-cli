# Contributing to OpenFrame CLI

Thank you for your interest in contributing to OpenFrame CLI! This document covers everything you need to know to submit high-quality contributions.

---

## 📋 Before You Start

- Join the [OpenMSP Slack community](https://www.openmsp.ai/) to discuss your ideas before starting large features
- Read the [Architecture Overview](./docs/development/architecture/README.md) to understand the codebase
- Set up your [development environment](./docs/development/setup/environment.md) and verify you can [build and run locally](./docs/development/setup/local-development.md)

> **Note:** All contribution discussions happen in the [OpenMSP Slack](https://www.openmsp.ai/). There are no GitHub Issues or Discussions for this project — bring your questions, feature ideas, and bug reports to Slack.
>
> **Slack invite:** [https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)

---

## 🛠️ Development Setup

### Prerequisites

| Tool | Version | Purpose |
|---|---|---|
| **Go** | 1.21+ | Primary language runtime |
| **Git** | 2.30+ | Version control |
| **Docker** | 24.x+ | Container runtime (integration tests) |
| **k3d** | 5.x+ | Local Kubernetes clusters (integration tests) |
| **Helm** | 3.x+ | Kubernetes package manager (integration tests) |

### Clone and Build

```bash
# Clone the repository
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli

# Download dependencies
go mod download

# Build the binary
go build -o openframe .

# Verify
./openframe --version
```

### Code Quality Tools

```bash
# Install goimports (formatting + import management)
go install golang.org/x/tools/cmd/goimports@latest

# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci-lint/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# Format code
goimports -w .

# Run vet
go vet ./...

# Run linter
golangci-lint run
```

---

## 📐 Code Style and Conventions

### Go Style

OpenFrame CLI follows standard Go conventions:

- **`gofmt` / `goimports`** formatting is required — no unformatted code will be merged
- **`go vet`** must pass with no warnings
- Follow [Effective Go](https://go.dev/doc/effective_go) and the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

### Naming Conventions

| Element | Convention | Example |
|---|---|---|
| Package names | Lowercase, single word | `cluster`, `executor`, `redact` |
| Exported types | PascalCase | `ClusterService`, `CommandExecutor` |
| Unexported types | camelCase | `clusterManager`, `mockExecutor` |
| Constants | PascalCase (exported), camelCase (unexported) | `DefaultClusterName`, `maxRetries` |
| Test files | `_test.go` suffix | `service_test.go` |
| Test functions | `Test` prefix + PascalCase | `TestCreateClusterSuccess` |

### Error Handling

Wrap errors with context and use structured types from `shared/errors`:

```go
// GOOD: Wrap with context
if err := mgr.CreateCluster(ctx, cfg); err != nil {
    return fmt.Errorf("creating cluster %q: %w", cfg.Name, err)
}

// BAD: Lost context
if err := mgr.CreateCluster(ctx, cfg); err != nil {
    return err
}
```

### Command Structure

When adding a new Cobra command, follow the established pattern:

```go
func getMyCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "mycommand [name]",
        Short: "One-line description",
        Long: `Multi-line detailed description.

The long description should explain what the command does,
when to use it, and any important caveats.`,
        Args: cobra.MaximumNArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            // Validate input
            // Delegate to service layer
            // Handle errors via sharedErrors.HandleGlobalError
            return nil
        },
    }
    cmd.Flags().StringVar(&flagVar, "flag-name", "default", "Flag description")
    return cmd
}
```

### Service Layer Conventions

- Services must accept interfaces (not concrete types) for all dependencies
- Always accept `context.Context` as the first argument for cancellable operations
- Return descriptive errors, not boolean success flags
- Use the `CommandExecutor` interface for all external binary invocations — **never** `os/exec` directly

---

## 🔐 Security Guidelines

### Secret Redaction

Any credential read from environment variables, config files, or user prompts **must** be registered with the redact package before use:

```go
import "github.com/flamingo-stack/openframe-cli/internal/shared/redact"

redact.RegisterSecret(githubToken)
redact.RegisterSecret(registryPassword)
```

### Command Injection Prevention

Always use argv arrays via `CommandExecutor`, never shell string concatenation:

```go
// SAFE: argv array
result, err := exec.Execute(ctx, "k3d", "cluster", "list", "--output", "json")

// NEVER: shell injection risk
// exec.Execute(ctx, "sh", "-c", "k3d cluster list --output " + userInput)
```

### Security Checklist for New Commands

- [ ] User-supplied cluster names are validated via `ValidateClusterName`
- [ ] Any new credential/token is registered with `redact.RegisterSecret()`
- [ ] External commands use argv arrays, not shell strings
- [ ] New YAML/JSON input is validated via structured types before use
- [ ] Sensitive flags are not printed in error messages

---

## 🌿 Branch Naming

| Type | Pattern | Example |
|---|---|---|
| Feature | `feature/<short-description>` | `feature/add-kind-provider` |
| Bug fix | `fix/<short-description>` | `fix/cluster-delete-timeout` |
| Documentation | `docs/<short-description>` | `docs/update-contributing-guide` |
| Refactor | `refactor/<short-description>` | `refactor/extract-helm-manager` |
| Test | `test/<short-description>` | `test/add-bootstrap-integration` |
| Chore | `chore/<short-description>` | `chore/update-go-dependencies` |

**Rules:** Lowercase and hyphens only. Branch from `main` unless working on a specific release branch.

---

## 💬 Commit Message Format

OpenFrame CLI uses [Conventional Commits](https://www.conventionalcommits.org/):

```text
<type>(<scope>): <short description>

[optional body]

[optional footer(s)]
```

### Types

| Type | When to Use |
|---|---|
| `feat` | A new feature |
| `fix` | A bug fix |
| `docs` | Documentation changes only |
| `refactor` | Code change that neither fixes a bug nor adds a feature |
| `test` | Adding or modifying tests |
| `chore` | Build process, dependency updates, tooling |
| `perf` | Performance improvements |
| `ci` | CI/CD configuration changes |

### Examples

```text
feat(cluster): add --wait flag to cluster create command

fix(argocd): handle stalled sync after ref change

docs(contributing): add commit message guidelines

test(bootstrap): add integration test for non-interactive mode

chore: upgrade go-git to v5.12.0
```

---

## 🧪 Testing

### Running Tests

```bash
# Run all unit tests with race detector (recommended)
go test -race ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run integration tests (requires Docker, k3d, Helm, and 24GB+ RAM)
go test ./tests/integration/... -v -timeout 30m
```

### Writing Unit Tests

Use the `MockCommandExecutor` for isolated unit tests — never invoke real subprocesses:

```go
func TestCreateCluster(t *testing.T) {
    testutil.InitializeTestMode()

    mock := testutil.NewTestMockExecutor()
    mock.SetResponse("k3d cluster create", &executor.CommandResult{
        ExitCode: 0,
        Stdout:   `{"name": "test-cluster"}`,
    })

    svc := cluster.NewClusterService(mock)
    err := svc.CreateCluster(context.Background(), "test-cluster")
    assert.NoError(t, err)
}
```

### Coverage Targets

| Package Type | Target |
|---|---|
| Core services (`internal/`) | ≥ 80% |
| Command layer (`cmd/`) | ≥ 70% |
| Provider implementations | ≥ 75% |
| Shared utilities | ≥ 85% |

---

## 📤 Pull Request Process

### Before Opening a PR

```bash
# 1. Run all tests
go test -race ./...

# 2. Format code
goimports -w .

# 3. Run vet
go vet ./...

# 4. Build successfully
go build -o openframe .
```

### PR Description Template

```text
## Summary
<!-- What does this PR do? -->

## Changes
<!-- List the key changes made -->
-
-

## Testing
<!-- How was this tested? -->
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated (if applicable)
- [ ] Manual testing performed

## Checklist
- [ ] Code follows the style guidelines
- [ ] Self-review completed
- [ ] Tests pass (go test -race ./...)
- [ ] go vet ./... passes
- [ ] goimports formatting applied
- [ ] No secrets or credentials in code
- [ ] Security guidelines followed
```

### PR Size Guidelines

| Size | Lines Changed | Guidance |
|---|---|---|
| Small | < 100 lines | Preferred — fast review |
| Medium | 100–500 lines | Include detailed description |
| Large | 500+ lines | Split into smaller PRs if possible |

---

## ➕ Adding a New Command

Follow these steps when adding a new CLI command:

1. **Create the command file** in `cmd/<group>/<command>.go`
2. **Define a `get<Name>Cmd()` function** returning `*cobra.Command`
3. **Register it** in the parent command group (e.g., `cmd/cluster/cluster.go`)
4. **Create a service** in `internal/<group>/` with injected dependencies
5. **Write unit tests** using `testutil.TestClusterCommand`
6. **Add integration tests** if the command interacts with external systems
7. **Verify `--help` output** is accurate and descriptive

---

## ➕ Adding a New Provider

To add a new cluster provider (e.g., Kind):

1. **Implement the `Provider` interface** in `internal/cluster/providers/<name>/manager.go`
2. **Add prerequisite definitions** in `internal/cluster/prerequisites/`
3. **Register the provider** in the cluster service provider resolution
4. **Add the cluster type** to `internal/cluster/models/cluster.go`
5. **Write unit and integration tests**

---

## 🔍 Review Checklist

When reviewing a PR, check:

**Correctness:**
- [ ] Logic is correct and handles edge cases
- [ ] Error paths are handled and tested
- [ ] Context cancellation is propagated correctly

**Security:**
- [ ] No secrets hardcoded or logged
- [ ] New credentials registered with `redact.RegisterSecret()`
- [ ] User input validated before reaching shell-outs
- [ ] External commands use argv arrays, not shell strings

**Architecture:**
- [ ] Business logic is in the service layer, not command layer
- [ ] Dependencies are injected via interfaces (testable)
- [ ] New command follows the established Cobra pattern

**Tests:**
- [ ] Unit tests cover new code paths
- [ ] Error cases are tested
- [ ] Mock executor used instead of real subprocess calls in unit tests

**Documentation:**
- [ ] Public API has Go doc comments
- [ ] Complex logic has inline comments
- [ ] `--help` text is accurate and helpful

---

## 📦 Release Signing

Release binaries are code-signed automatically during the release workflow:

| Platform | Mechanism |
|---|---|
| macOS | `codesign` (Developer ID Application, hardened runtime) + `notarytool` notarization |
| Windows | Authenticode via Azure Trusted Signing |
| Linux | Integrity via `checksums.txt` + cosign bundle |

All release binaries can be verified using cosign:

```bash
cosign verify-blob --bundle checksums.txt.bundle \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity-regexp '^https://github.com/flamingo-stack/openframe-cli/\.github/workflows/release\.yml@.*$' \
  checksums.txt
```

---

## 💬 Community

- **OpenMSP Slack:** [https://www.openmsp.ai/](https://www.openmsp.ai/)
- **Slack invite:** [https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)
- **OpenFrame platform repo:** [https://github.com/flamingo-stack/openframe-oss-tenant](https://github.com/flamingo-stack/openframe-oss-tenant)
- **Releases:** [https://github.com/flamingo-stack/openframe-cli/releases](https://github.com/flamingo-stack/openframe-cli/releases)

---

<div align="center">
  Built with 💛 by the <a href="https://www.flamingo.run/about"><b>Flamingo</b></a> team
</div>
