# Contributing to OpenFrame CLI

Thank you for your interest in contributing to **OpenFrame CLI** — the Go command-line tool that provisions Kubernetes clusters (local k3d, or cloud GKE/EKS via Terraform) and deploys the OpenFrame platform onto them through ArgoCD's app-of-apps pattern.

This repository, `flamingo-stack/openframe-cli`, is a single Go module (`github.com/flamingo-stack/openframe-cli`). Per the organization's dependency graph it has no recorded upstream dependencies on other internal repositories and no recorded downstream consumers — it is a terminal artifact whose dependencies are third-party Go modules (Cobra, client-go, pterm, huh/bubbletea, sigstore-go) and external CLI tools it shells out to (Docker, k3d, Helm, Terraform, gcloud, aws).

> **Not the platform repository.** The actual OpenFrame platform that this CLI installs onto a cluster lives in the separate [`flamingo-stack/openframe-oss-tenant`](https://github.com/flamingo-stack/openframe-oss-tenant) repository. Do not look for platform services, charts, or application code inside this repository's subdirectories.

## Before You Start

- This project does not use GitHub Issues or GitHub Discussions. Questions, proposals, and day-to-day coordination happen on the [OpenMSP Slack community](https://www.openmsp.ai/) ([invite link](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)). Pull requests on [`flamingo-stack/openframe-cli`](https://github.com/flamingo-stack/openframe-cli/pulls) remain the mechanism for proposing code changes.
- Read [Architecture](./docs/development/architecture/README.md) to understand how `cmd/`, `internal/cluster`, `internal/chart`, `internal/app`, and the shared packages fit together.
- Set up your development environment following [Environment Setup](./docs/development/setup/environment.md) and [Local Development](./docs/development/setup/local-development.md) before making changes — those documents cover the required Go toolchain (pinned to `go1.26.5` for a Windows-amd64 GC bug), external tools (Docker, k3d, Helm, Terraform, gcloud, aws), linting setup, build commands, and the local run/debug loop. This document does not duplicate those steps.

## Development Workflow

### Branches

- The default/base branches are `main` and `develop`. Most day-to-day work merges into `develop`.
- Releases are cut **only** from `main`, via a manual dispatch of the release workflow — there is no automatic release on push. The release workflow enforces a hard guard that fails if dispatched from anywhere other than `refs/heads/main`, because the self-updater's cosign signing identity is pinned to that workflow running from `main`. Do not rename the release workflow file or remove that guard.
- Stale branches with no commits for 60 days are swept automatically, except `main`, `develop`, and any branch with an open pull request.

### Pull Requests

- Open a pull request against [`flamingo-stack/openframe-cli`](https://github.com/flamingo-stack/openframe-cli/pulls) once your branch is ready.
- CODEOWNERS assigns `@flamingo-stack/devops-engineers` as the default reviewer for the whole repository.
- Before merge, CI must pass: linting (`golangci-lint` + `go mod tidy -diff`), a compile-only cross-platform build matrix across all six supported `GOOS`/`GOARCH` combinations, unit tests on Linux/Windows/macOS (plus the race detector on Linux), and — for PRs targeting `main` from this repository itself — a full end-to-end cluster lifecycle test.
- An automated `🦩 Flamingo Code Review` is available on pull requests. It does not run on every event automatically: a review is triggered by checking a box on the invitation comment, adding the `flamingo-review` label, or commenting `@flamingo-review` (optionally `@flamingo-review full` to re-review the whole cumulative diff).
- PRs are **squash-merged**. Because the PR **title** becomes the squashed commit subject that `semantic-release` analyzes on `main`, the title itself must be a valid Conventional Commit — this is linted automatically on every title edit. A free-form title contributes nothing to any future release version bump and will fail the check.

### Commit Message Format

This repository follows **Conventional Commits**, enforced at the PR-title level (since the repository squash-merges, individual commits within a branch are not separately linted):

```text
type(scope)?: subject
```

Examples:

```text
fix(k3d): tolerate KUBECONFIG lists
feat!: drop v1 config
```

- `type` is one of the conventional types (`fix`, `feat`, `chore`, `docs`, `refactor`, `test`, `ci`, etc.).
- `scope` is optional and parenthesized, naming the affected area (e.g. `k3d`, `cluster`, `app`).
- A `!` after the type/scope (e.g. `feat!:`) marks a breaking change.
- `fix` maps to a patch release, `feat` to a minor release, and a `BREAKING CHANGE` footer or `!` marker to a major release — this is how `semantic-release` computes the next version with no manual version input.

Write your PR title in this format from the start; it is linted on every edit.

### Reviews

Before requesting review, and while reviewing someone else's PR, verify:

- [ ] `golangci-lint run ./...` passes locally with no new findings — this gate is blocking in CI.
- [ ] `go mod tidy -diff` reports no drift.
- [ ] Unit tests pass (`make test-unit`); on Linux, also run the race-detector variant (`make test-race`).
- [ ] Any new test that touches a real k3d cluster or shells out to external binaries is gated with the `integration` build tag and lives under `tests/integration/` — it must not run as part of `go test ./...` or `make test-unit`.
- [ ] New or changed Cobra commands are wired through a `Get<Group>Cmd()`/`get<Sub>Cmd()` factory, not an inline `cobra.Command{}` literal, and any custom `PersistentPreRunE` calls `ui.ApplyGlobalOutputFlags(cmd)`.
- [ ] New service logic lives on an unexported struct behind a `New<Name>Service` constructor, with dependencies injected rather than derived internally.
- [ ] Errors are wrapped with context (`fmt.Errorf("...: %w", err)`) at each call boundary; exit codes from external tools are preserved via `*executor.CommandError`; errors already shown to the user are returned as `*sharederrors.AlreadyHandledError` to avoid double-printing.
- [ ] No raw `fmt.Println`/`fmt.Printf` was added for user-facing output — it goes through `pterm` or `internal/shared/ui`.
- [ ] Spinners are used only for k3d flows, never for EKS/GKE operations that stream Terraform progress.
- [ ] Commands that accept a `--context`/cluster-targeting flag reject stray positional arguments with a named-flag hint, especially for destructive commands.
- [ ] The change does not duplicate an existing exported definition elsewhere in the organization's codebase; if a near-duplicate is intentional, explain why in the PR.
- [ ] Any new or changed exported definition is exercised by at least one test.
- [ ] The PR title is a valid Conventional Commit, since it becomes the squashed commit message `semantic-release` reads from `main`.
- [ ] If the change touches the release workflow, the `refs/heads/main`-only guard and the workflow's filename are both left intact — both are load-bearing for the self-update cosign trust chain.

## Code Style and Conventions

Full conventions and the active, judged review rule catalog are documented in [Contributing Guidelines](./docs/development/contributing/guidelines.md) and enforced live via the organization's code review rule set. The key points to know up front:

### Linting

Linting is done with `golangci-lint` v2, configured in `.golangci.yml`, building on the `standard` default set (`errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`) plus `gosec`, `bodyclose`, `errorlint`, and `misspell`. `govet` runs with `enable-all: true`, excluding only `fieldalignment` and `shadow`. Test files (`_test.go`) are exempt from `errcheck` and `gosec`. Every finding is reported — `max-issues-per-linter` and `max-same-issues` are both `0`.

Run before opening a PR:

```bash
golangci-lint run ./...
go mod tidy -diff
```

### Structural conventions enforced by the automated code reviewer

These are judged, active rules specific to this repository — see [Contributing Guidelines](./docs/development/contributing/guidelines.md) for the full list with examples:

- **Command wiring**: every Cobra command group is returned by a `Get<Group>Cmd()` (exported) or `get<Sub>Cmd()` (unexported) factory function. The root command never constructs `cobra.Command` literals directly. A command group with its own `PersistentPreRunE` must call `ui.ApplyGlobalOutputFlags(cmd)` first, since Cobra runs only the closest parent's hook.
- **Service types**: every service type is an unexported struct, instantiated only through a `New<Name>`/`New<Name>Service` constructor returning a pointer, with dependencies injected rather than derived internally.
- **Import aliasing**: internal packages that collide on their last path segment (e.g. two packages named `config`, `ui`, or `errors`) must be imported with a camelCase alias encoding the owning subsystem (`sharedErrors`, `chartServices`, `uiCluster`, etc.).
- **Error wrapping**: every error returned from a called function is wrapped with a short, lowercase context string via `fmt.Errorf("<context>: %w", err)`. Bare `return err` at a call boundary is not acceptable.
- **Already-displayed errors**: an error already printed via `sharedErrors.HandleGlobalError` must be returned as (or wrap) `*sharederrors.AlreadyHandledError`, so `main` does not double-print it.
- **External command exit codes**: failures from external binaries (helm, k3d, terraform, etc.) must be returned as, or be wrappable to, `*executor.CommandError`, which `main.exitCode()` unwraps to propagate the exact exit code.
- **User-facing output**: goes through `pterm` printers or helpers in `internal/shared/ui`. Raw `fmt.Println`/`fmt.Printf` to stdout is forbidden for user-facing output, since it bypasses `--silent`, `--plain`, and pterm's writer redirection used in tests.
- **Root command flags**: `cmd/root.go`'s root `cobra.Command` must set `SilenceErrors: true` and `SilenceUsage: true`.
- **Version metadata**: build-time version variables (`version`, `commit`, `date`) in `cmd/root.go` must be package-level `string` vars, never struct fields, since the linker's `-ldflags -X` flag can only target package-level string vars.
- **Spinners vs. cloud operations**: spinners are used only for k3d operations; cloud operations (EKS/GKE) stream Terraform progress and must use `pterm.Info.Printf` instead.
- **Positional-argument safety**: commands that target a cluster via a flag (e.g. `--context`) must reject stray positional arguments with a validator naming the correct flag.
- **Production build flags**: all production builds use `CGO_ENABLED=0 go build -trimpath`. The only sanctioned exception is the race-detector target.
- **Integration tests**: any test creating real k3d clusters or running real external binaries must use the `integration` build tag and live under `tests/integration/`.

Two deterministic, organization-wide checks also apply to every PR:

- **No duplicate definitions** (`DUP-001`, warn): a new or changed exported definition whose normalized structure matches an existing one elsewhere in the org is flagged.
- **New exports need a test** (`TEST-001`, info): a new or changed exported definition that no test file references is flagged as shipped unverified.

## Running and Writing Tests

Full testing strategy, test-harness internals, and conventions are documented in [Testing](./docs/development/testing/README.md). In summary:

```bash
# Unit tests (fast, no external services)
go test ./...

# Unit tests with the race detector
go test -race ./...

# Integration tests (real k3d clusters; needs docker + k3d)
go test -tags integration ./tests/integration/...
```

- Prefer the mock executor (`internal/shared/executor.NewMockCommandExecutor()`, or `testutil.NewTestMockExecutor()` / `testutil.CreateStandardTestFlags()`) over real binaries for anything that shells out.
- When a change touches how the CLI constructs an external command, handles credentials, or performs a destructive cluster operation, use the security-focused assertion helpers in `tests/testutil/command_assertions.go` (`AssertNoArgContains`, `AssertArgIsDiscrete`, `AssertNoBashCConcatenation`, `AssertNoDeleteOf`, `AssertNoCurlPipeShell`).
- New Cobra commands should be exercised with the existing Cobra test helpers (`testutil.TestCommandStructure`, `testutil.TestCLICommand`, `testutil.TestClusterCommand`) rather than hand-rolled plumbing.
- When a command or flag (name, shorthand, type, default) is part of the CLI's public surface, lock it with `tests/testutil/flag_contract.go`'s `AssertFlags`/`AssertSubcommands`.
- Any new command needs no extra registration for `--help` coverage — `cmd/help_matrix_test.go` discovers the full command tree automatically — but you must still provide a non-empty `Short` description and flag usage text.

Coverage is enforced through CI gates rather than a single numeric threshold: `golangci-lint run ./...` must report zero issues, `go test ./...` must pass on Linux/Windows/macOS, `go test -race ./...` must pass on Linux, every command must answer `--help` correctly, and the CLI's public surface (persistent flags, subcommands, flag specs) is frozen by contract tests.

## Security-Sensitive Changes

If your change touches credentials, downloads, or self-update code, review [Security](./docs/development/security/README.md) first. In particular:

- New credential-bearing input must be parsed and stripped of secrets before it reaches anything that might be logged or exec'd, following the pattern in `internal/chart/providers/git/auth.go` (secrets carried only in memory, never in argv or on disk).
- Route any new output that might contain a sensitive value through `internal/shared/redact`'s `Redact` helper rather than printing it directly.
- Any new downloaded third-party tool binary must be wired through `internal/shared/download`'s checksum-verification path (`Downloader.FetchVerified` / `VerifyChecksum`), never fetched ad hoc.
- Any new external-tool invocation must use fixed argv (`exec.CommandContext(ctx, name, args...)`), never a shell string; a `#nosec` suppression, if required, must be scoped and justified inline.
- Treat the self-update signature verification path (`internal/shared/selfupdate/cosign.go`) as security-critical — any change that weakens or bypasses it is a regression, not a convenience fix.

### Reporting a Security Vulnerability

**Do not open a public GitHub issue or pull request for a security vulnerability.** Report it by following the process in [SECURITY.md](./SECURITY.md), which covers scope, what to include in a report, and what to expect in response.

## Where to Find More Documentation

The full documentation set for this repository starts at [docs/README.md](./docs/README.md), which indexes getting-started guides, development docs (architecture, setup, testing, security, releasing), and reference material.

Documents referenced above, for convenience:

- [Development hub](./docs/development/README.md)
- [Contributing Guidelines (full detail)](./docs/development/contributing/guidelines.md)
- [Architecture](./docs/development/architecture/README.md)
- [Environment Setup](./docs/development/setup/environment.md)
- [Local Development](./docs/development/setup/local-development.md)
- [Testing](./docs/development/testing/README.md)
- [Security](./docs/development/security/README.md)
- [Releasing](./docs/development/releasing.md)
- [Release Signing](./docs/development/release-signing.md)

## Community

All contribution discussion happens on the [OpenMSP Slack community](https://www.openmsp.ai/) ([invite link](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)) — there are no GitHub Issues or Discussions for this project.

- **OpenFrame platform repository:** [https://github.com/flamingo-stack/openframe-oss-tenant](https://github.com/flamingo-stack/openframe-oss-tenant)
- **Releases:** [https://github.com/flamingo-stack/openframe-cli/releases](https://github.com/flamingo-stack/openframe-cli/releases)
- **Pull Requests:** [https://github.com/flamingo-stack/openframe-cli/pulls](https://github.com/flamingo-stack/openframe-cli/pulls)
