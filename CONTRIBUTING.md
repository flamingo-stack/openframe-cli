<div align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://shdrojejslhgnojzkzak.supabase.co/storage/v1/object/public/public/doc-orchestrator/logos/1771384772513-n227fc-logo-openframe-full-dark-bg.png">
    <source media="(prefers-color-scheme: light)" srcset="https://shdrojejslhgnojzkzak.supabase.co/storage/v1/object/public/public/doc-orchestrator/logos/1771384777189-nbcwbo-logo-openframe-full-light-bg.png">
    <img alt="OpenFrame" src="https://shdrojejslhgnojzkzak.supabase.co/storage/v1/object/public/public/doc-orchestrator/logos/1771384777189-nbcwbo-logo-openframe-full-light-bg.png" width="300">
  </picture>
</div>

# Contributing to OpenFrame CLI

Thank you for your interest in contributing to OpenFrame CLI! This guide covers everything you need to know about code style, branching strategy, pull requests, and the review process.

> **Community support happens in Slack, not GitHub Issues.**
> Join the [OpenMSP Slack](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA) to discuss contributions, ask questions, or report bugs before opening a PR.

---

## Table of Contents

- [Getting Started](#getting-started)
- [Code Style and Conventions](#code-style-and-conventions)
- [Branch Naming Convention](#branch-naming-convention)
- [Commit Message Format](#commit-message-format)
- [Pull Request Process](#pull-request-process)
- [Adding New Commands](#adding-new-commands)
- [Local Validation Before Submitting](#local-validation-before-submitting)
- [Getting Help](#getting-help)

---

## Getting Started

Before contributing, set up your development environment:

1. **Install Go 1.21+** — the primary language runtime
2. **Clone the repository**:

```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
```

3. **Install dependencies**:

```bash
go mod download
go mod tidy
```

4. **Install development tools**:

```bash
# Linter
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Import formatter
go install golang.org/x/tools/cmd/goimports@latest

# Vulnerability scanner
go install golang.org/x/vuln/cmd/govulncheck@latest
```

5. **Build and verify**:

```bash
go build -o openframe ./main.go
./openframe --help
```

For detailed environment setup, see the [Development Documentation](./docs/README.md).

---

## Code Style and Conventions

### Go Style Guide

OpenFrame CLI follows standard Go conventions with these project-specific rules:

- **Format with `gofmt`**: All code must be formatted with `gofmt` before committing
- **Organize imports with `goimports`**: Use `goimports` for import grouping (stdlib, external, internal)
- **Lint with `golangci-lint`**: All linter rules must pass before a PR can merge
- **Error wrapping**: Use `fmt.Errorf("operation failed: %w", err)` for wrapping errors
- **Context propagation**: All service methods that call external tools must accept and propagate `context.Context`
- **Interface-first**: New external tool integrations must be defined as interfaces before implementation

### Naming Conventions

| Element | Convention | Example |
|---|---|---|
| Packages | lowercase, single word | `cluster`, `executor`, `argocd` |
| Exported types | PascalCase | `ClusterService`, `K3dManager` |
| Unexported types | camelCase | `clusterConfig`, `helmValues` |
| Interfaces | Noun or `er` suffix | `ClusterManager`, `CommandExecutor` |
| Test functions | `Test<FunctionUnderTest>` | `TestCreateCluster` |
| Test files | `<file>_test.go` | `service_test.go` |
| Constructor functions | `New<TypeName>()` | `NewK3dManager()`, `NewClusterService()` |

### Error Handling Conventions

```go
// ✅ CORRECT — wrap with context
if err := someOperation(); err != nil {
    return fmt.Errorf("creating cluster %s: %w", name, err)
}

// ✅ CORRECT — use structured error types for user-facing errors
return errors.CreateCommandError("k3d", args, originalErr)

// ❌ WRONG — discarding context
if err := someOperation(); err != nil {
    return err
}
```

### Logging and Output

- **Never use `fmt.Println` in service/provider code** — all user-facing output must go through `internal/shared/ui`
- **Verbose output**: Wrap detailed logs in verbose-mode checks
- **No secrets in logs**: Never log credential values, tokens, or passwords — even in verbose mode

---

## Branch Naming Convention

| Branch Type | Format | Example |
|---|---|---|
| Feature | `feature/<short-description>` | `feature/add-cluster-pause-command` |
| Bug fix | `fix/<short-description>` | `fix/k3d-timeout-on-slow-machines` |
| Documentation | `docs/<short-description>` | `docs/add-telepresence-guide` |
| Refactor | `refactor/<short-description>` | `refactor/executor-interface-cleanup` |
| Release | `release/v<major>.<minor>.<patch>` | `release/v1.2.0` |

**Rules:**
- Use lowercase only
- Use hyphens as word separators (not underscores or spaces)
- Keep descriptions short (2–5 words)
- Branch off `main` for all new work

```bash
git checkout main
git pull origin main
git checkout -b feature/my-new-feature
```

---

## Commit Message Format

OpenFrame CLI uses the [Conventional Commits](https://www.conventionalcommits.org/) format:

```text
<type>(<scope>): <short description>

[optional body]

[optional footer]
```

### Commit Types

| Type | When to Use |
|---|---|
| `feat` | A new feature or capability |
| `fix` | A bug fix |
| `docs` | Documentation changes only |
| `style` | Formatting, missing semicolons, no logic change |
| `refactor` | Code change that neither fixes a bug nor adds a feature |
| `test` | Adding or updating tests |
| `chore` | Build process, dependency updates, tooling changes |
| `perf` | Performance improvements |

### Scope Examples

```text
feat(cluster): add pause/resume cluster commands
fix(chart): resolve ArgoCD wait timeout on slow machines
docs(bootstrap): update non-interactive flag documentation
test(executor): add mock response for k3d list command
refactor(shared): consolidate error display in ErrorHandler
chore(deps): update client-go to v0.29.0
```

### Commit Message Rules

- Subject line: 72 characters max, imperative mood ("add X", not "added X" or "adds X")
- No period at the end of the subject line
- Body: wrap at 72 characters; explain *what* and *why*, not *how*

---

## Pull Request Process

### Before Opening a PR

Complete this checklist:

- [ ] Code is formatted with `gofmt` and `goimports`
- [ ] All linter rules pass: `golangci-lint run ./...`
- [ ] All tests pass: `go test ./...`
- [ ] New code has corresponding unit tests
- [ ] Vulnerability check passes: `govulncheck ./...`
- [ ] PR description explains *what* changed and *why*

### PR Title

Follow the same Conventional Commits format as commit messages:

```text
feat(cluster): add cluster pause/resume subcommands
fix(chart): handle ArgoCD sync timeout gracefully
docs(dev): improve Telepresence intercept documentation
```

### PR Description Template

```markdown
## Summary
Brief description of what this PR does and why.

## Changes
- Added `openframe cluster pause` command
- Added `openframe cluster resume` command
- Updated ClusterService with pause/resume methods
- Added unit tests for pause/resume operations

## Testing
Describe how you tested these changes:
- Unit tests: `go test ./internal/cluster/...`
- Manual test: `./openframe cluster pause test-cluster`

## Related
Link to Slack discussion or related context (if applicable).
```

### Review Checklist

When reviewing a PR, verify:

**Code Quality**
- [ ] Logic is clear and follows existing patterns
- [ ] No unnecessary complexity or over-engineering
- [ ] Error messages are user-friendly and actionable
- [ ] No hardcoded values that should be configurable

**Architecture**
- [ ] New external tools are abstracted behind interfaces
- [ ] Dependencies are injected (not instantiated inline)
- [ ] Commands delegate to services; services delegate to providers
- [ ] Shared infrastructure is used (don't re-implement logging, prompts, errors)

**Testing**
- [ ] Unit tests cover the happy path and error cases
- [ ] Mock executor is used correctly (no real tool calls in unit tests)

**Security**
- [ ] No credentials or secrets in source code
- [ ] All user inputs are validated
- [ ] External commands use `exec.Command` with arg arrays (not shell strings)
- [ ] Temporary files are cleaned up with `defer`

**Documentation**
- [ ] Exported functions and types have Go doc comments
- [ ] New commands have `Short` and `Long` descriptions in the Cobra command

---

## Adding New Commands

When adding a new command, follow this pattern:

1. **Create the command file**: `cmd/<group>/<command>.go`
2. **Register in group file**: Add to `cmd/<group>/<group>.go`'s `AddCommand()` call
3. **Create service logic**: `internal/<group>/services/<command>.go`
4. **Define interface**: Add to `internal/<group>/utils/types/interfaces.go`
5. **Add prerequisite checks**: Update `internal/<group>/prerequisites/checker.go` if needed
6. **Write unit tests**: `internal/<group>/services/<command>_test.go`
7. **Update inline docs**: Exported functions and types must have doc comments

---

## Local Validation Before Submitting

Run this sequence before opening a PR:

```bash
# Format code
gofmt -w .
goimports -w .

# Lint
golangci-lint run ./...

# Test
go test ./...

# Vulnerability check
govulncheck ./...

# Build check
go build -o /tmp/openframe-test ./main.go
```

---

## Repository Structure

```text
openframe-cli/
├── main.go                          # Binary entry point
├── cmd/                             # Cobra CLI commands
│   ├── root.go                      # Root command wiring
│   ├── bootstrap/bootstrap.go       # openframe bootstrap
│   ├── cluster/                     # openframe cluster *
│   ├── chart/                       # openframe chart *
│   └── dev/                         # openframe dev *
├── internal/                        # Business logic (not exported)
│   ├── bootstrap/                   # Bootstrap service
│   ├── cluster/                     # Cluster service, models, providers
│   ├── chart/                       # Chart service, models, providers, UI
│   ├── dev/                         # Dev service, intercept, scaffold
│   └── shared/                      # Cross-cutting concerns
│       ├── executor/                # Command execution abstraction
│       ├── errors/                  # Error types and handlers
│       ├── ui/                      # Prompts, tables, logo, progress
│       └── config/                  # TLS, credentials, system init
└── tests/                           # Test utilities and integration tests
    ├── testutil/                    # Mock executors, flag factories
    ├── integration/                 # End-to-end CLI integration tests
    └── mocks/                       # Mock implementations
```

---

## Getting Help

Stuck on something? The best place to ask is the **OpenMSP Slack**:

| Resource | Link |
|---|---|
| OpenMSP Community | [openmsp.ai](https://www.openmsp.ai/) |
| Slack Invite | [Join OpenMSP Slack](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA) |
| OpenFrame Platform | [openframe.ai](https://openframe.ai) |
| Flamingo | [flamingo.run](https://flamingo.run) |

Discuss your contribution idea in Slack before starting large changes — this avoids duplicated effort and ensures alignment with the project's direction.

---

<div align="center">
  Built with 💛 by the <a href="https://www.flamingo.run/about"><b>Flamingo</b></a> team
</div>
