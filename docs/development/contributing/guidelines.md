
# Contributing Guidelines

OpenFrame CLI (`github.com/flamingo-stack/openframe-cli`) is the bootstrap and lifecycle-management tool for the [OpenFrame](https://openframe.ai) platform built by [Flamingo](https://flamingo.run). It is a standalone Go binary — per the organization's dependency graph (coverage: **full**), it has no recorded upstream dependencies on other managed repositories and no recorded downstream consumers. It deploys the platform whose source lives in the separate [`flamingo-stack/openframe-oss-tenant`](https://github.com/flamingo-stack/openframe-oss-tenant) repository, but that relationship is operational (cloning the app-of-apps chart at install time), not a code dependency. Keep that boundary in mind: changes here affect the CLI only, never the platform's own code.

This guide covers code style, branching, commits, the PR process, and the review checklist enforced on this repository. For environment setup and local workflows, see [Local Development](../setup/local-development.md) and [Environment Setup](../setup/environment.md). For how the codebase is organized, see [Architecture Overview](../architecture/README.md) and the [reference architecture overview](../../reference/architecture/overview.md).

> **No GitHub Issues or Discussions.** This project does not use GitHub Issues or GitHub Discussions for planning or support. Discussion happens on the OpenMSP Slack community: [openmsp.ai](https://www.openmsp.ai/) ([join the Slack](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)). Pull requests on [flamingo-stack/openframe-cli](https://github.com/flamingo-stack/openframe-cli/pulls) are still the mechanism for proposing code changes.

## Table of Contents

- [Code Style and Conventions](#code-style-and-conventions)
- [Branch Naming](#branch-naming)
- [Commit Message Format](#commit-message-format)
- [Pull Request Process](#pull-request-process)
- [Review Checklist](#review-checklist)

## Code Style and Conventions

The repository is organized around `cmd/` (thin Cobra command wrappers) and `internal/` (domain services: `cluster`, `chart`, `app`, `prerequisites`, `k8s`, `shared`). New code should follow the same layering: a command under `cmd/` should delegate immediately into a service under `internal/`, never embed business logic itself.

### Linting

Static analysis is the primary drift detector for this codebase. The configuration lives in `.golangci.yml` (golangci-lint v2). It enables the standard default set (`errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`) plus:

- `gosec` — security findings (subprocess construction, file paths, file permissions, unverified downloads). The full default rule set is kept; nothing is silently excluded.
- `bodyclose`, `errorlint`, `misspell`

`govet` runs with `enable-all: true`, excluding only `fieldalignment` (noise) and `shadow` (too noisy for this codebase currently). Test files (`_test.go`) are exempted from `errcheck` and `gosec`, since they legitimately ignore errors and use weak fixtures. `issues.max-issues-per-linter` and `max-same-issues` are both set to `0` — every finding is shown, nothing is capped.

Run the linter locally before opening a PR:

```bash
golangci-lint run
```

### Conventions enforced by the active code-review rules

These are the organization's actively enforced rules for this repository (Go-specific); a PR that violates one of the `error`-severity rules will be flagged in review.

**Import aliasing.** When importing multiple packages from within `github.com/flamingo-stack/openframe-cli` that would otherwise collide on their last path segment (two packages both named `config`, `ui`, `errors`, etc.), each import must carry a camelCase alias encoding the owning subsystem — e.g. `sharedErrors`, `chartServices`, `uiCluster`, `utilTypes`. A bare last-segment name is only acceptable when there is no ambiguity in the file.

**Service construction.** Every service type is an unexported-fields struct exposed only through a `New<Name>` (or `New<Name>Service`) constructor returning a pointer — e.g. `Service`, `ClusterService`. Direct struct literals of service types must not appear outside the constructor. Business logic lives on methods of that struct; free functions are acceptable only as package-private helpers or thin delegators. The constructor must not accept configuration it can derive internally — external dependencies (executors, sub-services) are injected as constructor parameters.

**Cobra command wiring.** Each command group is returned by a `Get<Group>Cmd()` factory function (exported) or `get<Sub>Cmd()` (unexported, for leaf commands). The root command (`cmd/root.go`) calls only these factories — it never constructs `cobra.Command` literals inline for subcommands. A command group that defines its own `PersistentPreRunE` must call `ui.ApplyGlobalOutputFlags(cmd)` as its first action, because Cobra only runs the closest parent's hook; omitting this silently breaks `--verbose` and `--silent` for that entire subtree.

**Error handling.**
- Errors already displayed to the user (via `sharedErrors.HandleGlobalError` or equivalent) must be returned as, or wrapping, `*sharederrors.AlreadyHandledError`. `main` checks for this sentinel with `errors.As` and suppresses re-printing; any display path that returns a plain error instead causes double-printing on stderr.
- External-command failures (helm, k3d, terraform, etc.) must be returned as, or wrappable to, `*executor.CommandError`, carrying the original exit code, so `main`'s exit-code logic can propagate it to the OS exactly. Don't swallow or replace the exit code with a generic error string.
- Root `cobra.Command` must set `SilenceErrors: true` and `SilenceUsage: true` — error display is handled explicitly by `sharedErrors.HandleGlobalError` and the `main` sentinel check.

**Output.** All user-facing terminal output must go through pterm printers (`pterm.Success`, `pterm.Info`, `pterm.Warning`, `pterm.DefaultBasicText`, `pterm.DefaultBox`) or helpers in `internal/shared/ui` — never raw `fmt.Println`/`fmt.Printf` to stdout, since that bypasses `--silent`, `--plain`, and the writer-redirection tests rely on. `fmt.Fprintf(os.Stderr, ...)` is acceptable only in `main`'s error path and the post-command self-update notice.

**Spinners.** Spinners are only started for k3d operations. Cloud operations (EKS/GKE) stream per-resource progress lines from the Terraform engine, and a spinner's in-place redraws garble that output; use `pterm.Info.Printf` for cloud flows instead. The guard checked is `config.Type == models.ClusterTypeK3d` (and, where applicable, `!s.suppressUI`) before starting a spinner.

**Build-time version metadata.** `version`, `commit`, and `date` in `cmd/root.go` must remain package-level `string` vars, never struct fields — the linker's `-ldflags -X` flag can only set package-level string variables, and a struct field would silently no-op the injection, shipping every release binary reporting `"dev"` and disabling self-update.

**Positional-argument guards.** Commands that target a cluster via a flag (e.g. `--context`) must explicitly reject positional arguments with a validator that names the correct flag in the error message, instead of relying on Cobra's default `ArbitraryArgs` (which silently ignores stray positional arguments — risky for a destructive command like uninstall).

**Integration tests.** Any test that creates real k3d clusters, runs real external binaries, or otherwise needs Docker/k3d must carry `//go:build integration` as the first line of the file and live under `tests/integration/`. The `test-unit` Make target excludes `./tests/integration/...`; `test-integration` adds `-tags integration`.

## Branch Naming

The repository has no issue tracker to reference in branch names (see the Slack note above), so branch names should be short, descriptive, and scoped by purpose. Keep them lowercase, hyphen-separated, and prefixed by the kind of change, mirroring the Conventional Commit types used for commit messages:

```text
feat/cluster-status-json-output
fix/eks-teardown-orphan-cleanup
docs/contributing-guidelines
chore/golangci-lint-bump
```

Two automated workflows interact with branches directly, which shapes a few practical rules:

- **Stale-branch cleanup.** `.github/workflows/cleanup.yml` runs daily and deletes branches with no commits in the last 60 days, except `main` and `develop`, and except any branch with an open pull request. Don't let a long-lived feature branch go quiet for two months if you intend to keep it — either push a commit or keep the PR open.
- **Documentation branches.** Automated documentation PRs use the `code-documentation/` branch prefix (`DOCS_BRANCH_PREFIX` in `.github/workflows/code-documentation.yml`). Avoid naming a human-authored branch with that prefix.

Releases are cut from `main` only — `.github/workflows/release.yml` is manually dispatched (`workflow_dispatch`) and explicitly guards that the triggering ref is `refs/heads/main`, because the self-update signing identity (`internal/shared/selfupdate/cosign.go`) is pinned to `release.yml@refs/heads/main`. Feature and fix branches should be based on, and merged back into, `main`.

## Commit Message Format

Versioning is fully automatic: `.releaserc.yml`-driven `semantic-release` computes the next version from **Conventional Commits** since the last release tag. There is no manual version bump — the commit messages *are* the release mechanism, so the format matters functionally, not just stylistically.

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

Examples, consistent with the style already used in this repository's own workflow history:

```text
fix: skip stale-branch deletion when the commit-date API call fails

feat(cluster): add --spot flag for EKS node groups

docs: add contributing guidelines covering branch and commit conventions
```

A dispatch of the release workflow with no release-worthy commits since the last tag is a clean no-op — the plan job resolves no version and the release job is skipped. This means a PR consisting only of `docs:`/`chore:` commits will never trigger a version bump on its own, which is expected.

## Pull Request Process

1. **Branch from `main`**, following the naming convention above.
2. **Make focused changes.** Keep a PR scoped to one logical change; this keeps the Conventional Commit history meaningful for `semantic-release` and keeps review tractable.
3. **Run the local gates before pushing:**
   ```bash
   golangci-lint run
   make test-unit
   ```
   `make test-unit` is the same gate the release workflow runs (`.github/workflows/release.yml`, job `release`, step "Gate: unit tests") before any tag is cut.
4. **Open the PR** against `main` at [flamingo-stack/openframe-cli/pulls](https://github.com/flamingo-stack/openframe-cli/pulls). Use Conventional Commit style in the PR title as well, since squash-merge titles typically become the commit `semantic-release` reads.
5. **Automated checks run on the PR:**
   - `flamingo-code-graph`/`code-graph` (from `.github/workflows/code-documentation.yml`) re-indexes the code graph used by code review and documentation tooling. It runs on pushes to the default branch and on PR-related change-set events; it ignores `docs/**` and `**.md` changes.
   - `🦩 Flamingo Code Review` (`.github/workflows/flamingo-code-review.yml`) reviews pull requests on demand. It does **not** run automatically the moment a PR is opened or marked ready — it posts an invitation comment with a checkbox instead. To trigger a review pass, either:
     - Check the box in the bot's invitation/summary comment, or
     - Comment `@flamingo-review` (incremental, since the last reviewed commit) or `@flamingo-review full` (re-reads the whole cumulative diff) on the PR, or
     - Apply the `flamingo-review` label (consumed immediately by the `consume-review-request` job, so it can be re-applied to ask again), or
     - Apply `flamingo-review-always` to subscribe the PR to review on every push.
   - Review only runs on PRs from the same repository (not forks) due to the token/secrets model described in the workflow.
6. **Address findings.** Review findings are keyed to the active rule catalog (see Review Checklist below); `error`-severity findings should be resolved before merge.
7. **Merge.** Once checks pass and review findings are addressed, merge to `main`. A release is a separate, manual step — see `.github/workflows/release.yml` — dispatched when the team decides to ship, not automatically on merge.

### CODEOWNERS

Per `.github/CODEOWNERS`, `@flamingo-stack/devops-engineers` owns the entire repository (`*`) and specifically `.github/CODEOWNERS` itself. Expect review from that team on every PR.

## Review Checklist

Use this checklist before requesting review, mirroring the rules the automated reviewer enforces on this repository:

- [ ] **Imports**: Any import of a same-module package whose last path segment collides with another import (e.g. two `config` packages) carries a disambiguating camelCase alias (`sharedErrors`, `chartServices`, etc.).
- [ ] **Services**: New service types are unexported-field structs built only via a `New<Name>`/`New<Name>Service` constructor; no bare struct literals outside that constructor.
- [ ] **Commands**: New Cobra subcommands are wired through a `Get<Group>Cmd()`/`get<Sub>Cmd()` factory, not constructed inline in the root command.
- [ ] **PersistentPreRunE**: Any command group defining its own `PersistentPreRunE` calls `ui.ApplyGlobalOutputFlags(cmd)` first.
- [ ] **Error display**: Errors already printed to the user are returned as/wrapped in `*sharederrors.AlreadyHandledError` so `main` does not double-print.
- [ ] **Exit codes**: External-command failures preserve their original exit code via `*executor.CommandError`, not a generic error.
- [ ] **Output**: User-facing output goes through `pterm` printers or `internal/shared/ui`, never raw `fmt.Println`/`fmt.Printf` to stdout.
- [ ] **Spinners**: Spinners are gated to k3d operations only; cloud (EKS/GKE) flows use `pterm.Info.Printf` instead.
- [ ] **Version vars**: `cmd/root.go`'s `version`/`commit`/`date` remain package-level string vars (never struct fields), so `-ldflags -X` injection keeps working.
- [ ] **Positional args**: Flag-targeted commands (`--context`, etc.) reject stray positional arguments with a named-flag error rather than silently ignoring them.
- [ ] **Integration tests**: Any test touching real Docker/k3d/external binaries has `//go:build integration` as its first line and lives under `tests/integration/`.
- [ ] **Lint clean**: `golangci-lint run` passes with no new findings (gosec's default rule set is never silently excluded).
- [ ] **Unit tests pass**: `make test-unit` passes locally — the same gate the release workflow enforces before tagging.
- [ ] **Commit messages**: Follow Conventional Commits (`fix:`, `feat:`, `feat!:`/`BREAKING CHANGE:`, `docs:`, `chore:`, etc.) so `semantic-release` computes the correct next version.
- [ ] **Scope boundary**: Changes stay within this repository's responsibility — cluster provisioning, platform install/upgrade/uninstall, prerequisite management, self-update — and do not attempt to modify the platform code that lives in the separate `flamingo-stack/openframe-oss-tenant` repository.

For the full, current rule catalog (including severities and exact file-path scoping), consult the organization's active code-review rules for this repository; the points above reflect the Go-language rules in force at the time of writing.
