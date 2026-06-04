# Contributing to OpenFrame CLI

Thank you for your interest in contributing to OpenFrame CLI! This document describes the code style conventions, branching strategy, commit message format, and PR review process.

---

## Community First

OpenFrame is a community-driven project. All discussions, questions, bug reports, and feature requests happen in the **OpenMSP Slack**:

👉 [https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)

> We do **not** use GitHub Issues or GitHub Discussions. Please bring questions and bug reports to the Slack community at [https://www.openmsp.ai/](https://www.openmsp.ai/).

---

## Getting Started as a Contributor

### 1. Set Up Your Development Environment

Review the [Development Environment Setup](./docs/development/setup/environment.md) guide to install Go 1.22+, Docker, K3D, kubectl, Helm, mkcert, and the development linting tools.

### 2. Clone and Build

```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
go mod download
go build -o openframe main.go
./openframe --version
```

### 3. Run the Tests

```bash
# All unit tests
go test ./...

# With race detection
go test -race ./...

# Lint
golangci-lint run ./...
```

---

## Code Style and Conventions

### Go Formatting

All Go code must be formatted with `goimports` before committing:

```bash
goimports -w .
```

### Naming Conventions

| Construct | Convention | Example |
|---|---|---|
| Packages | lowercase, no underscores | `cluster`, `executor` |
| Exported types | PascalCase | `ClusterService`, `K3dManager` |
| Unexported types | camelCase | `clusterConfig`, `helmValues` |
| Interfaces | PascalCase, noun or adjective | `ClusterLister`, `HelmProvider` |
| Functions/Methods | PascalCase (exported), camelCase (unexported) | `CreateCluster`, `validateInputs` |
| Constants | PascalCase (exported), camelCase (unexported) | `DeploymentModeOSS` |
| Test files | `_test.go` suffix | `service_test.go` |
| Mock files | `mock.go` or `_mock.go` | `mock.go` |

### Layered Architecture

The CLI strictly follows a layered architecture. **Never skip a layer.**

```text
cmd/         → Command definitions only (flag parsing, delegation to services)
internal/    → All business logic
  └── <domain>/
      ├── service.go           # Main service struct and methods
      ├── models/              # Domain types and flag structs
      ├── providers/<tool>/    # External tool wrappers
      ├── prerequisites/       # Tool prerequisite checkers
      └── ui/                  # Interactive prompts and display
```

**Rules:**
- Commands in `cmd/` must not call external tools directly — delegate to service layer
- Services must not call external tools directly — delegate to providers
- Providers must use `shared/executor` for all subprocess execution
- Interfaces must be defined in `utils/types/interfaces.go` for testability

### Error Handling

Use the typed error constructors from `internal/shared/errors/`:

```go
// For user input validation failures
return errors.CreateValidationError("cluster-name", name, "must be alphanumeric")

// For external command failures
return errors.CreateCommandError("k3d", args, originalErr)
```

Do not return raw `fmt.Errorf` from user-facing code paths.

### Interface Compliance

Always add a compile-time interface assertion when implementing an interface:

```go
var _ types.ClusterLister = (*ClusterService)(nil)
```

### Command Injection Prevention

Always pass external tool arguments as slices — never concatenate user input into shell strings:

```go
// CORRECT
executor.Execute("k3d", []string{"cluster", "create", clusterName, "--agents", "2"})

// WRONG — injection risk
exec.Command("sh", "-c", "k3d cluster create " + clusterName)
```

---

## Branch Naming

| Type | Pattern | Example |
|---|---|---|
| Feature | `feature/<short-description>` | `feature/add-cluster-resize` |
| Bug fix | `fix/<short-description>` | `fix/k3d-wsl2-ip-detection` |
| Documentation | `docs/<short-description>` | `docs/update-intercept-guide` |
| Refactor | `refactor/<short-description>` | `refactor/executor-abstraction` |
| Chore | `chore/<short-description>` | `chore/update-dependencies` |

- Use lowercase and hyphens (no underscores)
- Keep descriptions concise (2–5 words)
- Branch from `main` for features and fixes

---

## Commit Message Format

OpenFrame CLI follows the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```text
<type>(<scope>): <short summary>

[optional body]

[optional footer(s)]
```

### Types

| Type | When to Use |
|---|---|
| `feat` | A new feature |
| `fix` | A bug fix |
| `docs` | Documentation only changes |
| `refactor` | Code change that neither fixes a bug nor adds a feature |
| `test` | Adding or correcting tests |
| `chore` | Build process, dependency updates, tooling |
| `perf` | Performance improvement |

### Scopes

Use the top-level package or area name: `cluster`, `chart`, `bootstrap`, `dev`, `executor`, `ui`, `errors`, `config`, `prereqs`, `ci`

### Examples

```text
feat(cluster): add support for multi-node k3d clusters

fix(dev): resolve WSL2 IP detection for Telepresence intercepts

docs(chart): update configuration wizard documentation

test(executor): add mock response patterns for helm upgrade

chore(deps): update k8s.io/client-go to v0.30.0
```

### Breaking Changes

Add `BREAKING CHANGE:` in the commit footer:

```text
feat(bootstrap): change default deployment mode to oss-tenant

BREAKING CHANGE: The default --deployment-mode flag value has changed
from "saas-tenant" to "oss-tenant". Update your CI/CD pipelines to
explicitly specify --deployment-mode=saas-tenant if needed.
```

---

## Pull Request Process

### Before Opening a PR

- [ ] Run `go test -race ./...` — all tests pass
- [ ] Run `golangci-lint run ./...` — no lint errors
- [ ] Run `goimports -w .` — code is formatted
- [ ] Run `govulncheck ./...` — no new vulnerabilities
- [ ] Add tests for any new functionality
- [ ] Update relevant documentation if behavior changes

### PR Title

Follow the same Conventional Commits format:

```text
feat(cluster): add cluster resize command
fix(chart): handle missing ArgoCD CRD during wait
```

### PR Description Template

```markdown
## Summary
Brief description of what this PR does.

## Changes
- Added X
- Fixed Y
- Updated Z

## Testing
How you tested this change (unit tests added, manual testing steps).

## Breaking Changes
Any breaking changes and migration steps.
```

### PR Size Guidelines

- **Small PRs** (< 200 lines): Preferred — easier to review and merge quickly
- **Medium PRs** (200–500 lines): Acceptable — include a detailed description
- **Large PRs** (> 500 lines): Split into smaller PRs where possible

---

## Writing Tests

### Unit Test Pattern

```go
func TestMyNewFeature(t *testing.T) {
    testutil.InitializeTestMode() // Disables interactive UI
    executor := testutil.NewTestMockExecutor()

    executor.SetResponse("some-tool --arg value", &executor.CommandResult{
        ExitCode: 0,
        Stdout:   "expected output",
    })

    svc := NewMyService(executor, false)
    result, err := svc.DoSomething(context.Background(), "input")
    require.NoError(t, err)
    assert.Equal(t, "expected result", result)
}
```

Always test failure scenarios alongside happy paths. Coverage targets:

| Package | Minimum Coverage |
|---|---|
| `internal/cluster/` | 70% |
| `internal/chart/` | 70% |
| `internal/dev/` | 60% |
| `internal/shared/` | 80% |
| `cmd/` | 50% |

---

## Code Review Checklist

Reviewers should check:

- [ ] Code follows the layered architecture (cmd → service → provider → executor)
- [ ] Error types use the shared error constructors
- [ ] External commands use argument slices (no shell string interpolation)
- [ ] User input is validated before use in system commands
- [ ] New public types and functions have Go doc comments
- [ ] Tests cover the happy path and at least one error path
- [ ] No hardcoded secrets, tokens, or passwords
- [ ] Interface compliance assertions are present for new interface implementations
- [ ] `--non-interactive` mode works correctly for any new wizard prompts

---

## Security Guidelines Summary

- **Never commit secrets** — credentials are always prompted at runtime or passed via environment variables
- **Never use shell string interpolation** for external commands — always use argument slices via the `shared/executor`
- **`InsecureTLSConfig` is only for local K3D clusters** — never apply it to production cluster connections
- **Validate all user input** before passing to system commands using `ValidateClusterName()` and similar validators

Report security vulnerabilities privately via [OpenMSP Slack](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA) — do **not** open a public GitHub issue.

---

## Developer Certificate of Origin

By contributing to this project, you certify that your contribution is your original work (or that you have the right to submit it) and that you agree to license it under the project's open-source license.

---

## Quick Reference: Common Development Commands

| Task | Command |
|---|---|
| Build binary | `go build -o openframe main.go` |
| Run all tests | `go test ./...` |
| Run with race detection | `go test -race ./...` |
| Run specific test | `go test -v -run TestClusterCreate ./internal/cluster/...` |
| Lint | `golangci-lint run ./...` |
| Format code | `goimports -w .` |
| Check vulnerabilities | `govulncheck ./...` |
| Download deps | `go mod download` |
| Tidy deps | `go mod tidy` |

---

<div align="center">
  Built with 💛 by the <a href="https://www.flamingo.run/about"><b>Flamingo</b></a> team
</div>
