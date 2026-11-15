# Contributing Guidelines

OpenFrame CLI (`openframe`) is the Go-based command-line tool that provisions Kubernetes clusters and deploys the OpenFrame platform. This guide covers how to propose changes to this repository: code style, branching, commit messages, and what the automated and human review process expects.

The application platform that this CLI deploys lives in the separate [`flamingo-stack/openframe-oss-tenant`](https://github.com/flamingo-stack/openframe-oss-tenant) repository. Contributions to the platform itself (Helm charts, app-of-apps manifests, backend services) belong there, not here. This repository's own architecture is described in [`../../reference/architecture/overview.md`](../../reference/architecture/overview.md) and its dependency graph in [`../../reference/architecture/ecosystem.md`](../../reference/architecture/ecosystem.md). Per the organization's code graph, this repository publishes a single Go module artifact (`github.com/flamingo-stack/openframe-cli`) and has no recorded upstream dependencies or downstream consumers inside the organization — it is a standalone CLI distributed via GitHub Releases.

> **No GitHub Issues or Discussions.** This project does not use GitHub Issues or GitHub Discussions for planning or support. Coordination happens on the OpenMSP Slack community: https://www.openmsp.ai/ (invite: https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA). Open a pull request directly for code changes; use Slack for design discussion, questions, and bug reports before writing code.

## Before you start

Read [`../README.md`](../README.md) for the development documentation index, [`../setup/environment.md`](../setup/environment.md) and [`../setup/local-development.md`](../setup/local-development.md) for how to get a working toolchain, and [`../architecture/README.md`](../architecture/README.md) for how the codebase is organized. [`../testing/README.md`](../testing/README.md) covers how to run the test suite, and [`../security/README.md`](../security/README.md) covers the security posture (verified downloads, cosign signing) that several of the rules below protect.

## Code style and conventions

The repository is Go (module `github.com/flamingo-stack/openframe-cli`, Go 1.26 per `go.mod`) with a small set of shell scripts (for example `scripts/sign-binary.sh`). Static analysis is the primary drift detector: `.golangci.yml` configures `golangci-lint` v2 with the standard rule set plus `gosec` (security: subprocess construction, file paths, file permissions, unverified downloads), `bodyclose`, `errorlint`, and `misspell`. Run the linter locally before opening a pull request; the release workflow (`.github/workflows/release.yml`) also gates on it.

The automated code reviewer enforces a further set of repository-specific conventions. The ones most relevant to day-to-day contributions:

- **Service construction.** Every domain package under `internal/` exposes its logic through an exported `Service` struct (or a domain-prefixed variant such as `ClusterService`) with unexported fields, instantiated only through a `New<Name>Service()` constructor that returns a pointer. Free functions are acceptable only as package-private helpers or thin wrappers that delegate immediately to a service method. The constructor should not derive configuration it could accept as a parameter; external dependencies (executors, sub-services) are injected.
- **Command wiring.** Cobra command groups under `cmd/` are returned by dedicated factory functions — `Get<Group>Cmd()` for exported groups, `get<Sub>Cmd()` for unexported leaf commands — never constructed as inline `cobra.Command` literals. Any command group that defines its own `PersistentPreRunE` must call `ui.ApplyGlobalOutputFlags(cmd)` as its first action, because Cobra only runs the closest ancestor's hook; omitting this silently breaks `--verbose` and `--silent` for that subtree.
- **Error handling.** Errors already displayed to the user (via `sharedErrors.HandleGlobalError` or equivalent) must be returned as, or wrapped in, `*sharederrors.AlreadyHandledError` so `main` does not print them twice. Errors from an external binary (helm, k3d, terraform) must be returned as, or wrappable to, `*executor.CommandError` so the original exit code reaches the OS instead of collapsing to a generic `1`.
- **User-facing output.** All terminal output meant for the user goes through `pterm` printers (`pterm.Success`, `pterm.Info`, `pterm.Warning`, `pterm.DefaultBasicText`, `pterm.DefaultBox`) or the helpers in `internal/shared/ui`, never raw `fmt.Println`/`fmt.Printf` to stdout — this is what makes `--silent` and `--plain` work, and what tests rely on for writer redirection. `fmt.Fprintf(os.Stderr, ...)` is acceptable only in `main`'s error path and the self-update notice.
- **Spinners.** Spinners are for k3d operations only. Cloud cluster operations (EKS/GKE) stream per-resource progress lines from the Terraform engine, and a spinner's in-place redraws garble that output; use `pterm.Info.Printf` for cloud paths instead, guarded by checking `config.Type == models.ClusterTypeK3d`.
- **Root command flags.** The root `cobra.Command` in `cmd/root.go` sets `SilenceErrors: true` and `SilenceUsage: true`; error and usage display is handled explicitly by `sharedErrors.HandleGlobalError` and the `main` sentinel check.
- **Version metadata.** Build-time version fields (`version`, `commit`, `date`) in `cmd/root.go` must stay package-level `string` vars, never struct fields — `-ldflags -X` can only target package-level string variables, and a struct field silently no-ops the injection, shipping every release reporting `"dev"` and disabling self-update.
- **Positional-argument safety.** Commands that target a cluster via a flag (such as `--context`) must explicitly reject positional arguments with a validator that names the correct flag in its error message, rather than relying on Cobra's default `ArbitraryArgs`, which would silently ignore a stray argument on a destructive command.
- **Import aliasing.** When a file imports multiple packages from within `github.com/flamingo-stack/openframe-cli` that collide on their last path segment (two packages both named `config`, `ui`, `errors`, etc.), each import must carry a camelCase alias that encodes the owning subsystem (`sharedErrors`, `chartServices`, `uiCluster`, `utilTypes`). Bare last-segment names are fine only when there is no ambiguity in the file.
- **Error wrapping.** Errors returned up the call stack should be wrapped with a short, lowercase, call-site description using `fmt.Errorf("<description>: %w", err)` so the causal chain is preserved; a bare `return err` at a boundary that would lose context is discouraged.

Before touching a file, it is worth checking the active rule set for the paths you're about to change — the reviewer narrows rules by path glob and language, and the list above is not exhaustive.

## Branch naming and PR process

The repository uses GitHub Flow against `main`. There is no numbered issue tracker to reference in a branch name (see the Slack note above), so branch names should describe the change directly, for example `fix/eks-teardown-orphan-check` or `feat/cluster-status-json-output`. Two prefixes are reserved by the automated pipelines and must not be reused for manual branches:

- `ai-fix/` — created by the automated code-review remediation flow.
- `code-documentation/` — created by the documentation pipeline for generated-docs pull requests.

A branch (other than `main` or `develop`) with no commits for 60 days is deleted automatically by the scheduled cleanup workflow (`.github/workflows/cleanup.yml`), unless it has an open pull request — keep active work pushed, or expect to re-push after a long pause.

Pull request review is governed by `.github/CODEOWNERS`: every path in the repository is owned by `@flamingo-stack/devops-engineers`, so that team's review is required to merge. Expect the automated review bot (🦩 Flamingo Code Review) to also run against your pull request — it reviews on open/ready-for-review/reopened, and otherwise on request: commenting `@flamingo-review` (or checking the box it posts) asks for a pass over the new commits, and `@flamingo-review full` asks it to re-read the whole cumulative diff. A `flamingo-review` label also triggers an on-demand pass and is consumed (removed) automatically once handled; `flamingo-review-always` is a durable label that subscribes the PR to review on every push.

Releases are a separate, manual step from merging: `.github/workflows/release.yml` is dispatched by a maintainer from `main` once changes are ready to ship, and the next version number is computed automatically from Conventional Commits history (see below) — merging a pull request never ships a release by itself.

### Pull request checklist before requesting review

- `golangci-lint` passes locally (the full standard set plus `gosec`, `bodyclose`, `errorlint`, `misspell`; see `.golangci.yml`).
- Unit tests pass (`make test-unit`, which excludes `tests/integration/...`).
- Any new test that creates real k3d clusters, runs real external binaries, or otherwise requires Docker/k3d is gated with `//go:build integration` as the first line of the file and lives under `tests/integration/`.
- Commit messages follow the Conventional Commits format described below — this repository's versioning depends on it.
- No fork-only logic has been introduced without the OPENFRAME sentinel comments, where applicable to files shared with an upstream project.

## Commit message format

This repository's release process (`.github/workflows/release.yml`, driven by `.releaserc.yml`) uses `semantic-release` with the Conventional Commits preset to compute the next version automatically from the commit history since the last tag. Write every commit — and especially the final commit before merge, since squash-merge messages are what `semantic-release` reads — in the form:

```text
<type>(<optional scope>): <short summary>

<optional body>

<optional footer(s)>
```

The type determines the version bump:

| Commit type | Effect on version |
|---|---|
| `fix:` | Patch release |
| `feat:` | Minor release |
| Any type with a `BREAKING CHANGE:` footer (or `!` after the type/scope) | Major release |
| `chore:`, `docs:`, `refactor:`, `test:`, `ci:` | No release triggered by itself |

Example commit messages consistent with this repository's history and conventions:

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

A dispatch of the release workflow with no release-worthy commits since the last tag is a clean no-op — the plan step simply finds nothing to release. Keep commit subjects imperative and under roughly 72 characters where practical, and use the body to explain *why*, not just *what*, especially for anything touching cluster provisioning, self-update, or signing — these are the highest-blast-radius areas in the codebase.

## Review checklist

Use this list when preparing a pull request, and expect the human reviewer (and the automated reviewer) to check the same things:

**Correctness and safety**
- Destructive commands (cluster `delete`, app `uninstall`) still require confirmation or `--force`, and commands using a `--context`-style flag reject stray positional arguments instead of silently ignoring them.
- External command failures propagate their real exit code via `*executor.CommandError` rather than collapsing to a generic failure.
- Any new pinned/downloaded binary (following the pattern in `internal/shared/download/`) is checksum-verified, matching how k3d, Helm, mkcert, Terraform, and infracost are already pinned and verified — never fetched unverified.

**Structure and style**
- New domain logic lives in a `Service` struct behind a `New<Name>Service` constructor, not as loose package functions.
- New Cobra subcommands are wired through a `Get<Group>Cmd()` / `get<Sub>Cmd()` factory, not an inline literal.
- Colliding cross-package imports within this module carry a disambiguating camelCase alias.
- User-facing output goes through `pterm` or `internal/shared/ui`, not raw `fmt.Print*`.
- Spinners are used only for k3d flows; cloud (EKS/GKE) flows print plain `pterm.Info` lines so streamed Terraform output stays readable.

**Error handling**
- Errors already shown to the user are wrapped in `*sharederrors.AlreadyHandledError` before returning to `main`.
- Errors crossing a call boundary are wrapped with a short, lowercase description via `fmt.Errorf("...: %w", err)`.

**Tests and CI**
- `make test-unit` passes; anything requiring Docker/k3d is marked `//go:build integration` and lives under `tests/integration/`.
- `golangci-lint` is clean.
- If the change touches `cmd/root.go`, version metadata (`version`, `commit`, `date`) remains package-level string vars — not struct fields — so `-ldflags -X` injection still works.

**Commit hygiene**
- Commit messages (and the final squash-merge message) use Conventional Commits types (`fix:`, `feat:`, `BREAKING CHANGE:` footer or `!`), since `semantic-release` reads them directly to compute the next version.

**Scope**
- Changes to the platform itself (Helm charts, backend services, app-of-apps manifests) are directed to [`flamingo-stack/openframe-oss-tenant`](https://github.com/flamingo-stack/openframe-oss-tenant), not opened here.

For hands-on setup instructions needed to actually run these checks locally, see [`../setup/local-development.md`](../setup/local-development.md) and [`../testing/README.md`](../testing/README.md).
