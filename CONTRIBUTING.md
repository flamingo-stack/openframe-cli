# Contributing to OpenFrame CLI

Thank you for your interest in contributing to **OpenFrame CLI** — the Go command-line tool that provisions Kubernetes clusters (local k3d, or cloud GKE/EKS via Terraform) and deploys the OpenFrame platform onto them through ArgoCD's app-of-apps pattern.

This repository, `flamingo-stack/openframe-cli`, publishes a single Go module, `github.com/flamingo-stack/openframe-cli`. Per the organization's dependency graph (coverage: full) it has no recorded upstream dependencies on other internal repositories and no recorded downstream consumers — it is a standalone binary distributed via GitHub Releases and `go install`, not a library other services import.

> **Not the platform repository.** The actual OpenFrame platform that this CLI installs onto a cluster lives in the separate [`flamingo-stack/openframe-oss-tenant`](https://github.com/flamingo-stack/openframe-oss-tenant) repository ("OpenFrame Main code Repository"). Do not look for platform services, charts, or application code inside this repository's subdirectories — the relationship between the two repositories is operational (this CLI clones the app-of-apps chart at install time), not a code dependency. Contributions to the platform itself belong in that repository, not here.

## Before you start

- This project does not use GitHub Issues or GitHub Discussions. Questions, proposals, and day-to-day coordination happen on the [OpenMSP Slack community](https://www.openmsp.ai/) ([invite link](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)). Open a pull request directly on [`flamingo-stack/openframe-cli`](https://github.com/flamingo-stack/openframe-cli/pulls) for code changes; use Slack for design discussion, questions, and bug reports before writing code.
- Read the [Development documentation index](./docs/development/README.md) and [Architecture](./docs/development/architecture/README.md) to understand how `cmd/`, `internal/cluster`, `internal/chart`, `internal/app`, and the shared packages fit together.

## Setting up a development environment

Do not duplicate the environment setup here — follow the dedicated guides, which cover the required Go toolchain (pinned to `go1.26.5`, because Go `1.26.0`/`1.26.1` crash intermittently on windows-amd64 during GC stack scanning), the external tools you need available (Docker, k3d, Helm, Terraform, gcloud, aws), IDE/linting setup, and the build/run/debug loop:

- [Environment Setup](./docs/development/setup/environment.md) — toolchain, required tools, IDE recommendations, static analysis configuration, environment variables.
- [Local Development](./docs/development/setup/local-development.md) — cloning, building (`make build`, `make build-all`), running (`go run . --help`, `openframe bootstrap`), debugging, and test targets.

## Development workflow

### Branches

Keep branch names short, descriptive, and prefixed by the kind of change, mirroring the Conventional Commit types used for commits and PR titles, for example `fix/eks-teardown-orphan-check` or `feat/cluster-status-json-output`. There is no numbered issue tracker to reference in a branch name (see the Slack note above), so branch names should describe the change directly.

Two prefixes are reserved by automated pipelines and must not be reused for manual branches:

- `ai-fix/` — created by the automated code-review remediation flow.
- `code-documentation/` — created by the documentation pipeline for generated-docs pull requests.

- Feature and fix branches should be based on, and merged back into, `main`.
- Releases are a separate, manual step from merging: `.github/workflows/release.yml` is dispatched by a maintainer from `main` once changes are ready to ship, and the next version number is computed automatically from Conventional Commits history (see below) — merging a pull request never ships a release by itself.
- A branch (other than `main` or `develop`) with no commits for 60 days is deleted automatically by the scheduled cleanup workflow, unless it has an open pull request — keep active work pushed, or expect to re-push after a long pause.

### Pull requests

1. Branch from `main`, following the naming convention above.
2. Keep a PR scoped to one logical change.
3. Run the local gates before pushing:
   ```bash
   golangci-lint run ./...
   make test-unit
   ```
4. Open the PR against `main` at [flamingo-stack/openframe-cli/pulls](https://github.com/flamingo-stack/openframe-cli/pulls).
5. Pull request review is governed by `.github/CODEOWNERS`: every path in the repository is owned by `@flamingo-stack/devops-engineers`, so that team's review is required to merge.
6. Expect the automated review bot (🦩 Flamingo Code Review) to also run against your pull request — it reviews on open/ready-for-review/reopened, and otherwise on request: commenting `@flamingo-review` (incremental, since the last reviewed commit) or `@flamingo-review full` (re-reads the whole cumulative diff) asks for a pass. A `flamingo-review` label also triggers an on-demand pass and is consumed (removed) automatically once handled; `flamingo-review-always` is a durable label that subscribes the PR to review on every push. Review only runs on PRs from the same repository, not forks.
7. Address findings before merge.
8. Merge once checks pass and findings are addressed. A release is dispatched separately by a maintainer when the team decides to ship.

### Commit message format

This repository's release process uses `semantic-release` with the Conventional Commits preset to compute the next version automatically from the commit history since the last tag. Write every commit — and especially the final commit before merge, since squash-merge messages are what `semantic-release` reads — in the form:

```text
<type>(<optional scope>): <short summary>

<optional body>

<optional footer(s)>
```

| Commit type | Effect on version |
|---|---|
| `fix:` | Patch release |
| `feat:` | Minor release |
| Any type with a `BREAKING CHANGE:` footer (or `!` after the type/scope) | Major release |
| `chore:`, `docs:`, `refactor:`, `test:`, `ci:` | No release triggered by itself |

Examples:

```text
fix(cluster/eks): skip orphan check when kubeconfig context is missing
```

```text
feat(app/status): add --interactive k9s-style TUI for ArgoCD application health
```

```text
fix(selfupdate)!: require cosign verification before binary replacement

BREAKING CHANGE: `openframe update` now refuses to install an update whose
Sigstore signature cannot be verified against the pinned release.yml
identity, instead of warning and proceeding.
```

A dispatch of the release workflow with no release-worthy commits since the last tag is a clean no-op. Keep commit subjects imperative and under roughly 72 characters where practical, and use the body to explain *why*, not just *what* — especially for anything touching cluster provisioning, self-update, or signing, the highest-blast-radius areas in the codebase.

### Reviews

Use this checklist when preparing a pull request, and expect the human reviewer (and the automated reviewer) to check the same things. The full rule catalog with examples lives in [Contributing Guidelines](./docs/development/contributing/guidelines.md); the highlights:

**Correctness and safety**
- Destructive commands (cluster `delete`, app `uninstall`) still require confirmation or `--force`, and commands using a `--context`-style flag reject stray positional arguments instead of silently ignoring them.
- External command failures propagate their real exit code via `*executor.CommandError` rather than collapsing to a generic failure.
- Any new pinned/downloaded binary follows the pattern in `internal/shared/download/` and is checksum-verified — never fetched unverified.

**Structure and style**
- New domain logic lives in a `Service` struct behind a `New<Name>Service` constructor, not as loose package functions.
- New Cobra subcommands are wired through a `Get<Group>Cmd()` / `get<Sub>Cmd()` factory, not an inline literal, and any custom `PersistentPreRunE` calls `ui.ApplyGlobalOutputFlags(cmd)` first (Cobra only runs the closest parent's hook).
- Colliding cross-package imports within this module carry a disambiguating camelCase alias (`sharedErrors`, `chartServices`, `uiCluster`, `utilTypes`).
- User-facing output goes through `pterm` or `internal/shared/ui`, never raw `fmt.Print*`.
- Spinners are used only for k3d flows; cloud (EKS/GKE) flows print plain `pterm.Info` lines so streamed Terraform output stays readable.

**Error handling**
- Errors already shown to the user are wrapped in `*sharederrors.AlreadyHandledError` before returning to `main`.
- Errors crossing a call boundary are wrapped with a short, lowercase description via `fmt.Errorf("...: %w", err)`.

**Tests and CI**
- `make test-unit` passes; anything requiring Docker/k3d is marked `//go:build integration` and lives under `tests/integration/`.
- `golangci-lint` is clean.
- If the change touches `cmd/root.go`, version metadata (`version`, `commit`, `date`) remains package-level string vars, not struct fields, so `-ldflags -X` injection still works.

**Commit hygiene**
- Commit messages (and the final squash-merge message) use Conventional Commits types, since `semantic-release` reads them directly to compute the next version.

**Scope**
- Changes to the platform itself (Helm charts, backend services, app-of-apps manifests) are directed to [`flamingo-stack/openframe-oss-tenant`](https://github.com/flamingo-stack/openframe-oss-tenant), not opened here.

## Code style and conventions

The repository is Go (module `github.com/flamingo-stack/openframe-cli`, Go 1.26 per `go.mod`) with a small set of shell scripts, for example `scripts/sign-binary.sh`. It is organized around `cmd/` (thin Cobra command wrappers) and `internal/` (domain services: `cluster`, `chart`, `app`, `prerequisites`, `k8s`, `shared`). New code should follow the same layering: a command under `cmd/` should delegate immediately into a service under `internal/`, never embed business logic itself.

Static analysis is the primary drift detector: `.golangci.yml` configures `golangci-lint` v2 with the standard rule set (`errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`) plus `gosec` (security: subprocess construction, file paths, file permissions, unverified downloads), `bodyclose`, `errorlint`, and `misspell`. `govet` runs with `enable-all: true`, excluding only `fieldalignment` and `shadow`. Test files (`_test.go`) are exempted from `errcheck` and `gosec`. Every finding is shown — `max-issues-per-linter` and `max-same-issues` are both `0`. Run before opening a PR:

```bash
golangci-lint run ./...
```

The automated code reviewer enforces a further set of repository-specific conventions beyond what `golangci-lint` checks mechanically. The full catalog with examples is in [Contributing Guidelines](./docs/development/contributing/guidelines.md); the ones most relevant day-to-day:

- **Service construction.** Every domain package under `internal/` exposes its logic through an exported `Service` struct (or a domain-prefixed variant such as `ClusterService`) with unexported fields, instantiated only through a `New<Name>Service()` constructor that returns a pointer. External dependencies (executors, sub-services) are injected, not derived internally.
- **Command wiring.** Cobra command groups under `cmd/` are returned by dedicated factory functions — `Get<Group>Cmd()` for exported groups, `get<Sub>Cmd()` for unexported leaf commands — never constructed as inline `cobra.Command` literals.
- **Error handling.** Errors already displayed to the user must be returned as, or wrapped in, `*sharederrors.AlreadyHandledError` so `main` does not print them twice. Errors from an external binary (helm, k3d, terraform) must be returned as, or wrappable to, `*executor.CommandError` so the original exit code reaches the OS.
- **User-facing output.** All terminal output meant for the user goes through `pterm` printers or the helpers in `internal/shared/ui`, never raw `fmt.Println`/`fmt.Printf` to stdout.
- **Root command flags.** The root `cobra.Command` in `cmd/root.go` sets `SilenceErrors: true` and `SilenceUsage: true`.
- **Version metadata.** Build-time version fields (`version`, `commit`, `date`) in `cmd/root.go` must stay package-level `string` vars, never struct fields — `-ldflags -X` can only target package-level string variables.
- **Positional-argument safety.** Commands that target a cluster via a flag (such as `--context`) must explicitly reject positional arguments with a validator that names the correct flag in its error message.
- **Import aliasing.** When a file imports multiple packages from within this module that collide on their last path segment (two packages both named `config`, `ui`, `errors`, etc.), each import must carry a camelCase alias that encodes the owning subsystem (`sharedErrors`, `chartServices`, `uiCluster`, `utilTypes`).
- **Error wrapping.** Errors returned up the call stack should be wrapped with a short, lowercase, call-site description using `fmt.Errorf("<description>: %w", err)`.

Before touching a file, it is worth checking the active rule set for the paths you're about to change — see [Contributing Guidelines](./docs/development/contributing/guidelines.md) and [Security](./docs/development/security/README.md) for the conventions that protect checksum verification and cosign signing specifically.

## Running and writing tests

Full testing strategy and test-harness internals are documented in [Testing](./docs/development/testing/README.md). The suite is organized in three layers: unit tests co-located beside the code in `cmd/` and `internal/`, root-level CLI matrix tests (`main_test.go`, `cli_matrix_test.go`) that build the real binary and exercise it as a black box, and integration tests under `tests/integration/` gated behind the `integration` build tag.

```bash
# Unit tests (fast, hermetic — no docker/k3d/network)
make test-unit

# Equivalent, narrower invocation while iterating
go test ./cmd/... ./internal/... ./tests/testutil/...

# Unit tests with the race detector (requires CGO)
make test-race

# Integration tests (real k3d clusters; needs docker + k3d)
make test-integration
```

`make test-unit` is the same gate the release workflow runs before any tag is cut, and it deliberately excludes `./tests/integration/...`. Any new test that creates real k3d clusters, runs real external binaries, or otherwise requires Docker/k3d must be gated with `//go:build integration` as the first line of the file and must live under `tests/integration/`.

Prefer the shared helpers over ad hoc assertions:

- `testutil.CreateStandardTestFlags()` (from `tests/testutil/setup.go`) to get a `cluster.FlagContainer` with a `MockCommandExecutor` already wired, instead of constructing executor mocks by hand.
- `testutil.TestClusterCommand(t, name, cmdFunc, setupFunc, teardownFunc)` for a new cluster subcommand, to get structure, flag, CLI-arg, and execution checks automatically.
- `testutil.AssertFlags(t, cmd, []testutil.FlagSpec{...})` whenever a command's flags should be frozen as a contract.
- For anything touching command construction (argv, env, stdin) — especially code shelling out to `docker`/`k3d`/`helm`/`terraform`/`gcloud`/`aws` — assert against the structured log via `testutil.AssertNoArgContains`, `AssertNoBashCConcatenation`, `AssertNoCurlPipeShell`, `AssertArgIsDiscrete`, or `AssertNoDeleteOf`, rather than a flattened command string.

Because `cmd/help_matrix_test.go` walks the live command tree, any new `cobra.Command` you add is automatically checked for a working `--help`, a non-empty `Short` description, and non-empty usage text on every local flag — there is nothing extra to wire up for this coverage; just give the command a `Short` description and real flag `Usage` strings.

See [Testing](./docs/development/testing/README.md) for the full breakdown of `tests/testutil/`, `tests/integration/common/`, and the mock command executor.

## Reporting security issues

See [SECURITY.md](./SECURITY.md) for the organization-wide vulnerability disclosure process — how to report a vulnerability, response SLAs, and scope. Do not open a public GitHub issue, pull request, or community channel post for a security problem; email the address given in that policy. [Security](./docs/development/security/README.md) also documents this repository's security-sensitive conventions (checksum-verified downloads, cosign/Sigstore self-update signing, secret redaction) that several of the review rules above protect.

## Where to find more documentation

The full documentation index is at [`docs/README.md`](./docs/README.md).
