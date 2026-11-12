# Contributing to OpenFrame CLI

Thank you for your interest in contributing to **OpenFrame CLI** — the Go command-line tool that provisions Kubernetes clusters (local k3d, or cloud GKE/EKS via Terraform) and deploys the OpenFrame platform onto them through ArgoCD's app-of-apps pattern.

This repository, `flamingo-stack/openframe-cli`, publishes a single Go module, `github.com/flamingo-stack/openframe-cli`. Per the organization's dependency graph (coverage: full) it has no recorded upstream dependencies on other internal repositories and no recorded downstream consumers — it is a terminal artifact whose dependencies are third-party Go modules (Cobra, client-go, pterm, huh/bubbletea, sigstore-go) and external CLI tools it shells out to (Docker, k3d, Helm, Terraform, gcloud, aws).

> **Not the platform repository.** The actual OpenFrame platform that this CLI installs onto a cluster lives in the separate [`flamingo-stack/openframe-oss-tenant`](https://github.com/flamingo-stack/openframe-oss-tenant) repository ("OpenFrame Main code Repository"). Do not look for platform services, charts, or application code inside this repository's subdirectories — the relationship between the two repositories is operational (this CLI clones the app-of-apps chart at install time), not a code dependency.

## Before You Start

- This project does not use GitHub Issues or GitHub Discussions. Questions, proposals, and day-to-day coordination happen on the [OpenMSP Slack community](https://www.openmsp.ai/) ([invite link](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)). Pull requests on [`flamingo-stack/openframe-cli`](https://github.com/flamingo-stack/openframe-cli/pulls) remain the mechanism for proposing code changes.
- Read [Architecture](./docs/development/architecture/README.md) to understand how `cmd/`, `internal/cluster`, `internal/chart`, `internal/app`, and the shared packages fit together.
- Set up your development environment following [Environment Setup](./docs/development/setup/environment.md) and [Local Development](./docs/development/setup/local-development.md) **before** making changes. Those documents cover the required Go toolchain (pinned to `go1.26.5`, because Go `1.26.0`/`1.26.1` crash intermittently on windows-amd64 during GC stack scanning), the external tools you need installed (Docker, k3d, Helm, Terraform, gcloud, aws), linting setup, build commands, and the local run/debug loop. This document does not duplicate those steps — follow the linked guides for environment setup.

## Development Workflow

### Branches

Keep branch names short, descriptive, and prefixed by the kind of change, mirroring the Conventional Commit types used for commits/PR titles:

```text
feat/cluster-status-json-output
fix/eks-teardown-orphan-cleanup
docs/contributing-guidelines
chore/golangci-lint-bump
```

- Feature and fix branches should be based on, and merged back into, `main`.
- Releases are cut **only** from `main`, via a manual dispatch of the release workflow — there is no automatic release on push or merge. The release workflow enforces a hard guard that fails if dispatched from anywhere other than `refs/heads/main`, because the self-update signing identity (`internal/shared/selfupdate/cosign.go`) is pinned to the release workflow running from `main`. Do not rename the release workflow file or remove that guard.
- Stale branches with no commits in the last 60 days are swept automatically, except `main`, `develop`, and any branch with an open pull request — don't let a long-lived branch go quiet for two months if you intend to keep it.
- Automated documentation PRs use the `code-documentation/` branch prefix; avoid naming a human-authored branch with that prefix.

### Pull Requests

1. Branch from `main`, following the naming convention above.
2. Keep a PR scoped to one logical change — this keeps the Conventional Commit history meaningful and keeps review tractable.
3. Run the local gates before pushing:
   ```bash
   golangci-lint run
   make test-unit
   ```
   `make test-unit` is the same gate the release workflow runs before any tag is cut.
4. Open the PR against `main` at [flamingo-stack/openframe-cli/pulls](https://github.com/flamingo-stack/openframe-cli/pulls). Use Conventional Commit style in the PR title, since this repository squash-merges PRs — the PR title becomes the commit subject `semantic-release` analyzes on `main`. A free-form title contributes nothing to any future release version bump.
5. `@flamingo-stack/devops-engineers` owns the entire repository per `.github/CODEOWNERS` and is the default reviewer on every PR.
6. An automated `🦩 Flamingo Code Review` is available on pull requests, but it does not run automatically the moment a PR is opened — it posts an invitation comment with a checkbox instead. Trigger a review pass by:
   - Checking the box in the bot's invitation/summary comment, or
   - Commenting `@flamingo-review` (incremental, since the last reviewed commit) or `@flamingo-review full` (re-reads the whole cumulative diff), or
   - Applying the `flamingo-review` label (consumed immediately, so it can be reapplied to ask again), or
   - Applying `flamingo-review-always` to subscribe the PR to review on every push.

   Review only runs on PRs from the same repository, not forks.
7. Address findings. Review findings are keyed to the active rule catalog (see [Code Style and Conventions](#code-style-and-conventions) below); `error`-severity findings should be resolved before merge.
8. Merge once checks pass and findings are addressed. A release is a separate, manual step dispatched when the team decides to ship — it does not happen automatically on merge.

### Commit Message Format

Versioning is fully automatic: `semantic-release` computes the next version from **Conventional Commits** since the last release tag. Because this repository squash-merges PRs, the commit subjects `semantic-release` analyzes on `main` are the **PR titles**, so the PR title itself must be a valid Conventional Commit:

```text
<type>[optional scope]: <short summary>

[optional body]

[optional footer(s)]
```

| Type | Effect on version | Use for |
|---|---|---|
| `fix:` | patch bump | Bug fixes |
| `feat:` | minor bump | New functionality |
| `feat!:` / footer `BREAKING CHANGE:` | major bump | Breaking changes |
| `docs:`, `chore:`, `refactor:`, `test:`, `ci:`, `build:` | no release triggered alone | Non-release-worthy changes |

Examples:

```text
fix: skip stale-branch deletion when the commit-date API call fails

feat(cluster): add --spot flag for EKS node groups

docs: add contributing guidelines covering branch and commit conventions
```

A dispatch of the release workflow with no release-worthy commits since the last tag is a clean no-op. This means a PR consisting only of `docs:`/`chore:` commits will never trigger a version bump on its own, which is expected.

### Reviews

Before requesting review, and while reviewing someone else's PR, verify:

- [ ] `golangci-lint run` passes with no new findings.
- [ ] `make test-unit` passes locally — the same gate the release workflow enforces before tagging.
- [ ] Any new test that touches a real k3d cluster or shells out to external binaries carries `//go:build integration` as the first line and lives under `tests/integration/` — it must never run as part of a bare `go test ./...` or `make test-unit`.
- [ ] New or changed Cobra commands are wired through a `Get<Group>Cmd()`/`get<Sub>Cmd()` factory, not an inline `cobra.Command{}` literal, and any custom `PersistentPreRunE` calls `ui.ApplyGlobalOutputFlags(cmd)` first (Cobra only runs the closest parent's hook).
- [ ] New service logic lives on an unexported-fields struct behind a `New<Name>`/`New<Name>Service` constructor returning a pointer, with external dependencies injected rather than derived internally.
- [ ] Errors already displayed to the user are returned as, or wrapping, `*sharederrors.AlreadyHandledError`, so `main` does not double-print.
- [ ] External-command failures (helm, k3d, terraform, etc.) are returned as, or wrappable to, `*executor.CommandError`, carrying the original exit code.
- [ ] No raw `fmt.Println`/`fmt.Printf` was added for user-facing output — it goes through `pterm` printers or `internal/shared/ui`.
- [ ] Spinners are used only for k3d operations; cloud (EKS/GKE) flows use `pterm.Info.Printf` instead, since a spinner's in-place redraws garble Terraform's streamed progress output.
- [ ] Commands that target a cluster via a flag (e.g. `--context`) reject stray positional arguments with a named-flag hint, rather than relying on Cobra's default `ArbitraryArgs`.
- [ ] Cross-package imports that would collide on their last path segment (two packages both named `config`, `ui`, `errors`, etc.) carry a disambiguating camelCase alias (`sharedErrors`, `chartServices`, `uiCluster`, `utilTypes`).
- [ ] `cmd/root.go`'s `version`, `commit`, and `date` remain package-level `string` vars, never struct fields, since the linker's `-ldflags -X` flag can only set package-level string variables.
- [ ] The PR title is a valid Conventional Commit, since it becomes the squashed commit message read from `main`.

## Code Style and Conventions

The repository is organized around `cmd/` (thin Cobra command wrappers) and `internal/` (domain services: `cluster`, `chart`, `app`, `prerequisites`, `k8s`, `shared`). New code should follow the same layering: a command under `cmd/` should delegate immediately into a service under `internal/`, never embed business logic itself. The full convention set and the active, judged review rule catalog are documented in [Contributing Guidelines](./docs/development/contributing/guidelines.md); the highlights are summarized below.

### Linting

Static analysis is the primary drift detector for this codebase. The configuration lives in `.golangci.yml` (golangci-lint v2). It enables the standard default set (`errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`) plus `gosec` (the full default rule set is kept, nothing is silently excluded), `bodyclose`, `errorlint`, and `misspell`. `govet` runs with `enable-all: true`, excluding only `fieldalignment` (noise) and `shadow` (too noisy for this codebase currently). Test files (`_test.go`) are exempted from `errcheck` and `gosec`. Every finding is shown — `max-issues-per-linter` and `max-same-issues` are both `0`.

Run before opening a PR:

```bash
golangci-lint run
```

### Structural conventions enforced by the automated code reviewer

These are the organization's actively enforced, Go-specific rules for this repository; a PR that violates one of the `error`-severity rules will be flagged in review. See [Contributing Guidelines](./docs/development/contributing/guidelines.md) for the full rule catalog with examples.

- **Import aliasing.** Imports of same-module packages that would otherwise collide on their last path segment (two packages both named `config`, `ui`, `errors`, etc.) must carry a camelCase alias encoding the owning subsystem — e.g. `sharedErrors`, `chartServices`, `uiCluster`, `utilTypes`. A bare last-segment name is only acceptable when there is no ambiguity in the file.
- **Service construction.** Every service type is an unexported-fields struct exposed only through a `New<Name>` (or `New<Name>Service`) constructor returning a pointer. Direct struct literals of service types must not appear outside the constructor; external dependencies (executors, sub-services) are injected as constructor parameters rather than derived internally.
- **Cobra command wiring.** Each command group is returned by a `Get<Group>Cmd()` (exported) or `get<Sub>Cmd()` (unexported, leaf) factory function. The root command calls only these factories — it never constructs `cobra.Command` literals inline for subcommands. A command group that defines its own `PersistentPreRunE` must call `ui.ApplyGlobalOutputFlags(cmd)` as its first action, or `--verbose`/`--silent` silently stop working for that whole subtree.
- **Error handling.** Errors already displayed to the user must be returned as, or wrapping, `*sharederrors.AlreadyHandledError`, so `main` does not double-print. External-command failures must be returned as, or wrappable to, `*executor.CommandError`, carrying the original exit code. Root `cobra.Command` must set `SilenceErrors: true` and `SilenceUsage: true`.
- **Output.** All user-facing terminal output must go through `pterm` printers or helpers in `internal/shared/ui` — never raw `fmt.Println`/`fmt.Printf` to stdout, since that bypasses `--silent`, `--plain`, and the writer-redirection tests rely on.
- **Spinners.** Spinners are only started for k3d operations; cloud operations (EKS/GKE) stream per-resource progress lines from Terraform instead, since a spinner's in-place redraws garble that output.
- **Build-time version metadata.** `version`, `commit`, and `date` in `cmd/root.go` must remain package-level `string` vars, never struct fields.
- **Positional-argument guards.** Commands that target a cluster via a flag (e.g. `--context`) must explicitly reject positional arguments with a validator that names the correct flag, instead of relying on Cobra's default `ArbitraryArgs`.
- **Integration tests.** Any test that creates real k3d clusters, runs real external binaries, or otherwise needs Docker/k3d must carry `//go:build integration` as the first line and live under `tests/integration/`. `make test-unit` excludes `./tests/integration/...`.

## Running and Writing Tests

Full testing strategy and test-harness internals are documented in [Testing](./docs/development/testing/README.md). In summary, the suite is built entirely on the standard `go test` toolchain plus `testify` assertions, organized in three layers: unit tests beside the code in `cmd/` and `internal/`, CLI matrix/root-level tests at the repository root (`main_test.go`, `cli_matrix_test.go`), and integration tests under `tests/integration/`.

```bash
# Unit tests (fast, hermetic — no docker/k3d/network)
make test-unit

# Equivalent, narrower invocation while iterating
go test ./cmd/... ./internal/... ./tests/testutil/...

# Root-level CLI matrix and exit-code tests
go test .

# Unit tests with the race detector (requires CGO)
make test-race

# Integration tests (real k3d clusters; needs docker + k3d)
make test-integration
```

- Prefer the mock executor (`internal/shared/executor.NewMockCommandExecutor()`, or the `testutil` helpers `testutil.CreateStandardTestFlags()` / `testutil.NewTestMockExecutor()` in `tests/testutil/setup.go`) over real binaries for anything that shells out.
- When a change touches how the CLI constructs an external command, handles credentials, or performs a destructive cluster operation, use the structured-log assertions in `tests/testutil/command_assertions.go` — `AssertNoArgContains`, `AssertArgIsDiscrete`, `AssertNoBashCConcatenation`, `AssertNoDeleteOf`, `AssertNoCurlPipeShell` — rather than the legacy flattened-string command log.
- New Cobra commands should be exercised with the existing helpers — `testutil.TestCommandStructure`, `testutil.TestCLICommand`, and (for cluster commands specifically) `testutil.TestClusterCommand`, which runs all four standard sub-tests (`Structure`, `Flags`, `CLI`, `Execution`) in one call.
- When a command or flag (name, shorthand, type, default) is part of the CLI's public, scriptable surface, lock it down with `tests/testutil/flag_contract.go`'s `AssertFlag`/`AssertFlags`/`AssertSubcommands`, as `cmd/contract_test.go` already does for the root command's `--verbose`/`--silent` flags and top-level subcommand list.
- Any new command needs no extra registration for `--help` coverage: `cmd/help_matrix_test.go` walks the entire command tree returned by `GetRootCmd` automatically. It still must have a non-empty `Short` description and fully-described flag usage strings, or the sweep fails.
- Call `testutil.InitializeTestMode()` at the start of any test that touches `internal/shared/ui`, so UI components render in a test-safe way.

A release is not published unless `make test-unit` passes — this is the hard floor the release workflow enforces. In practice, a change is review-ready when the unit-test scope passes, `go test .` passes at the repository root, any new or changed command still satisfies the whole-tree help sweep, and any flag/subcommand forming part of the CLI's scriptable contract has a corresponding `flag_contract.go` assertion.

## Security-Sensitive Changes

If your change touches credentials, downloads, or self-update code, review [Security](./docs/development/security/README.md) first. In particular:

- New credential-bearing input must be parsed and stripped of secrets before it reaches anything that might be logged or executed, following the pattern in `internal/chart/providers/git/auth.go` (secrets carried only in memory via an in-memory `transport.AuthMethod`, never written into a URL, argv, or disk).
- Route any new output that might contain a sensitive value through `internal/shared/redact`'s `Redact` helper rather than printing it directly.
- Any new downloaded third-party tool binary must be wired through `internal/shared/download`'s pinned-version, checksum-verified path, never fetched ad hoc or via `curl | bash`.
- Any new external-tool invocation must use a fixed argv slice assembled by the CLI's own code, routed through `internal/shared/executor`'s `CommandExecutor` interface — never a shell string built from untrusted input. A `#nosec` suppression, if required, must carry a specific, scoped justification comment (see `internal/cluster/discovery/auth.go`'s example), never a blanket suppression.
- Treat the self-update signature verification path (`internal/shared/selfupdate/cosign.go`) as security-critical — any change that weakens or bypasses checksum or cosign/Sigstore signature verification is a regression, not a convenience fix.

### Reporting a Security Vulnerability

**Do not open a public GitHub issue, pull request, or discussion for a security vulnerability**, and do not post it in a community channel — a public report puts every user at risk before a fix exists. Report it by following the process in [SECURITY.md](./SECURITY.md), which covers what to include in a report, response-time commitments, and scope.

## Where to Find More Documentation

The full documentation set for this repository starts at [docs/README.md](./docs/README.md). From there you can reach the getting-started guides (introduction, prerequisites, quick start), the development hub (environment setup, local development, architecture, testing, security), and the how-to guides for the full cluster/platform lifecycle.
