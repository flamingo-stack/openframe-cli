# Contributing Guidelines

Thank you for contributing to OpenFrame CLI! This guide covers everything you need to know to submit high-quality contributions.

---

## Before You Start

- Join the [OpenMSP Slack community](https://www.openmsp.ai/) to discuss your ideas before starting large features
- Read the [Architecture Overview](../architecture/README.md) to understand the codebase
- Set up your [development environment](../setup/environment.md) and verify you can [build and run locally](../setup/local-development.md)

---

## Code Style and Conventions

### Go Style

OpenFrame CLI follows standard Go conventions:

- **`gofmt` / `goimports`** formatting is required — no unformatted code will be merged
- **`go vet`** must pass with no warnings
- Follow [Effective Go](https://go.dev/doc/effective_go) and the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

```bash
# Format and organize imports
goimports -w .

# Run vet
go vet ./...

# Run linter (if golangci-lint is configured)
golangci-lint run
```

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

- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Use `shared/errors` types for structured errors (`CommandError`, `AlreadyHandledError`)
- Use `friendlyHint` patterns for user-facing errors
- Never swallow errors silently — return them or log them

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
        Long:  `Multi-line detailed description.

The long description should explain what the command does,
when to use it, and any important caveats.`,
        Args:  cobra.MaximumNArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            // Validate input
            // Delegate to service layer
            // Handle errors via sharedErrors.HandleGlobalError
            return nil
        },
    }

    // Add flags
    cmd.Flags().StringVar(&flagVar, "flag-name", "default", "Flag description")

    return cmd
}
```

### Service Layer Conventions

- Services must accept interfaces (not concrete types) for all dependencies
- Always accept `context.Context` as the first argument for cancellable operations
- Return descriptive errors, not boolean success flags
- Use the `CommandExecutor` interface for all external binary invocations — never `os/exec` directly

---

## Branch Naming

| Type | Pattern | Example |
|---|---|---|
| Feature | `feature/<short-description>` | `feature/add-kind-provider` |
| Bug fix | `fix/<short-description>` | `fix/cluster-delete-timeout` |
| Documentation | `docs/<short-description>` | `docs/update-contributing-guide` |
| Refactor | `refactor/<short-description>` | `refactor/extract-helm-manager` |
| Test | `test/<short-description>` | `test/add-bootstrap-integration` |
| Chore | `chore/<short-description>` | `chore/update-go-dependencies` |

**Rules:**
- Use lowercase and hyphens only (no underscores, no uppercase)
- Keep descriptions short and meaningful
- Branch from `main` unless working on a specific release branch

---

## Commit Message Format

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

### Scopes (Optional but Recommended)

| Scope | Area |
|---|---|
| `cluster` | Cluster commands and services |
| `app` | App commands and chart services |
| `bootstrap` | Bootstrap command and service |
| `prereq` | Prerequisites system |
| `update` | Self-update mechanism |
| `executor` | Command executor |
| `k8s` | Kubernetes client package |
| `argocd` | ArgoCD provider |
| `helm` | Helm provider |
| `ui` | Terminal UI and wizards |
| `errors` | Error handling |
| `redact` | Secret redaction |

### Examples

```text
feat(cluster): add --wait flag to cluster create command

Adds a --wait flag that blocks until all cluster nodes are Ready.
Useful for CI pipelines that need the cluster immediately after creation.

fix(argocd): handle stalled sync after ref change

When ArgoCD silently fails to adopt a new ref (ArgoCD v3 regression),
the stall detector now surfaces a diagnostic message after 90s.

docs(contributing): add commit message guidelines

test(bootstrap): add integration test for non-interactive mode

chore: upgrade go-git to v5.12.0
```

---

## Pull Request Process

### Before Opening a PR

```bash
# 1. Ensure all tests pass
go test -race ./...

# 2. Format code
goimports -w .

# 3. Run vet
go vet ./...

# 4. Build successfully
go build -o openframe .

# 5. Test your changes manually
./openframe --help
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
- [ ] Tests pass (`go test -race ./...`)
- [ ] `go vet ./...` passes
- [ ] `goimports` formatting applied
- [ ] No secrets or credentials in code
- [ ] Security guidelines followed (secrets registered with redact, no shell injection)
```

### PR Size Guidelines

| Size | Lines Changed | Guidance |
|---|---|---|
| Small | < 100 lines | Preferred — fast review |
| Medium | 100–500 lines | Include detailed description |
| Large | 500+ lines | Split into smaller PRs if possible |

---

## Review Checklist

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

## Adding a New Command

Follow these steps when adding a new CLI command:

1. **Create the command file** in `cmd/<group>/<command>.go`
2. **Define a `get<Name>Cmd()` function** returning `*cobra.Command`
3. **Register it** in the parent command group (e.g., `cmd/cluster/cluster.go`)
4. **Create a service** in `internal/<group>/` with injected dependencies
5. **Write unit tests** using `testutil.TestClusterCommand`
6. **Add integration tests** if the command interacts with external systems
7. **Verify `--help` output** is accurate and descriptive

---

## Adding a New Provider

To add a new cluster provider (e.g., Kind):

1. **Implement the `Provider` interface** in `internal/cluster/providers/<name>/manager.go`
2. **Add prerequisite definitions** in `internal/cluster/prerequisites/`
3. **Register the provider** in the cluster service provider resolution
4. **Add the cluster type** to `internal/cluster/models/cluster.go`
5. **Write unit and integration tests**

---

## Community

All contribution discussions happen in the [OpenMSP Slack](https://www.openmsp.ai/). There are no GitHub Issues or Discussions for this project — bring your questions, feature ideas, and bug reports to Slack.

- **Slack invite:** [https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)
- **OpenFrame platform repo:** [https://github.com/flamingo-stack/openframe-oss-tenant](https://github.com/flamingo-stack/openframe-oss-tenant)
- **Releases:** [https://github.com/flamingo-stack/openframe-cli/releases](https://github.com/flamingo-stack/openframe-cli/releases)
