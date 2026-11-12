# Contributing Guidelines

OpenFrame CLI is the Go command-line tool that provisions Kubernetes clusters (local k3d, or cloud GKE/EKS via Terraform) and deploys the OpenFrame platform onto them via ArgoCD's app-of-apps pattern. This document covers the conventions, branch/PR process, commit format, and review checklist for contributing to this repository, `flamingo-stack/openframe-cli`.

> This repository does not use GitHub Issues or GitHub Discussions. Questions, proposals, and coordination happen on the OpenMSP Slack community (https://www.openmsp.ai/, invite: https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA). Pull requests on `flamingo-stack/openframe-cli` (https://github.com/flamingo-stack/openframe-cli/pulls) remain the mechanism for proposing code changes.

For environment setup before you start contributing, see `../setup/environment.md` and `../setup/local-development.md`. For how the codebase is organized, see `../architecture/README.md` and `../../reference/architecture/overview.md`.

## Before you start

OpenFrame CLI has no `package.json` or `pom.xml` — it is a single Go module (`go.mod`) declaring `module github.com/flamingo-stack/openframe-cli`, built with Go 1.26 (toolchain `go1.26.5`, pinned because 1.26.0/1.26.1 have a Windows-amd64 GC crash). The ecosystem graph records no upstream dependencies on other internal repositories and no downstream consumers inside this organization's graph — the CLI is a terminal artifact whose dependencies are third-party Go modules (Cobra, client-go, pterm, huh/bubbletea, sigstore-go) and external CLI tools it shells out to (Docker, k3d, Helm, Terraform, gcloud, aws).

This is distinct from the main platform code, which lives in the separate **flamingo-stack/openframe-oss-tenant** repository (https://github.com/flamingo-stack/openframe-oss-tenant). Do not look for platform services inside this CLI repository's subdirectories — they are maintained externally.

## Code style and conventions

The repository enforces a set of structural and stylistic rules through `golangci-lint` (configured in `.golangci.yml`) and through the automated code reviewer's active rule catalog. Key conventions to follow:

### Linting

`.golangci.yml` builds on golangci-lint v2's `standard` default set (`errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`) and additionally enables:

- **`gosec`** — security findings (G204 subprocess construction, G304 file paths, G306 file permissions, G107 unverified downloads). The full default rule set is kept; nothing is excluded.
- **`bodyclose`** — flags unclosed HTTP response bodies.
- **`errorlint`** — flags non-wrapping error comparisons.
- **`misspell`** — catches common misspellings.

`govet` runs with `enable-all: true`, excluding only `fieldalignment` (noise) and `shadow` (too noisy for this codebase currently). Test files (`_test.go`) are exempted from `errcheck` and `gosec`, since tests legitimately ignore errors and use weak fixtures. `max-issues-per-linter` and `max-same-issues` are both set to `0` — every finding is shown, nothing is silently capped.

Run the linter locally before opening a PR:

```bash
golangci-lint run ./...
```

Also keep `go.mod`/`go.sum` tidy — CI checks this explicitly:

```bash
go mod tidy -diff
```

### Structural conventions (enforced by the code reviewer)

These are active, judged rules specific to this repository. Follow them in any new or changed code:

- **Command wiring**: every Cobra command group must be returned by a `Get<Group>Cmd()` (exported) or `get<Sub>Cmd()` (unexported, leaf) factory function. The root command (`cmd/root.go`) only calls these factories — it never constructs `cobra.Command` literals for subcommands directly. A command group that defines its own `PersistentPreRunE` must call `ui.ApplyGlobalOutputFlags(cmd)` as its first action, because Cobra only runs the closest parent's hook; omitting this silently breaks `--verbose` and `--silent` for that whole subtree.
- **Service types**: every service type is an unexported struct, instantiated only through a `New<Name>` (or `New<Name>Service`) constructor returning a pointer. Struct fields stay unexported; business logic lives on methods of that struct, not free functions (free functions are acceptable only as package-private helpers or thin delegating wrappers). The constructor must not accept configuration it can derive internally — dependencies like executors or sub-services are injected as parameters.
- **Import aliasing**: when importing multiple internal packages that collide on their last path segment (e.g. two packages both named `config`, `ui`, or `errors`), each import must carry a camelCase alias encoding the owning subsystem — e.g. `sharedErrors`, `chartServices`, `uiCluster`, `utilTypes`. Bare last-segment names are fine only when there's no ambiguity in the file.
- **Error wrapping**: every error returned from a called function is wrapped with a short, lowercase context string via `fmt.Errorf("<context>: %w", err)`. Bare `return err` at a call boundary loses context and is not acceptable.
- **Already-displayed errors**: when a command's `Execute` method calls `sharedErrors.HandleGlobalError` (which prints the error to the terminal), it must return that call's result, which is (or wraps) `*sharederrors.AlreadyHandledError`. `main` checks for this sentinel with `errors.As` and suppresses a second print. Any new error-display path that skips this sentinel will double-print on stderr.
- **External command exit codes**: when an external binary (helm, k3d, terraform, etc.) fails, the error must be returned as, or be wrappable to, `*executor.CommandError`, which carries the original exit code. `main.exitCode()` unwraps this type to propagate the exact code to the OS — never collapse an external failure to a generic error string.
- **User-facing output**: all user-visible output (progress, success, warnings, summaries) goes through `pterm` printers (`pterm.Success`, `pterm.Info`, `pterm.Warning`, `pterm.DefaultBasicText`, `pterm.DefaultBox`) or helpers in `internal/shared/ui`. Raw `fmt.Println`/`fmt.Printf` to stdout is forbidden for user-facing output because it bypasses `--silent`, `--plain`, and the pterm writer-redirection used in tests. `fmt.Fprintf(os.Stderr, ...)` is acceptable only in `main`'s error path and the post-command self-update notice.
- **Root command flags**: the root `cobra.Command` in `cmd/root.go` must set `SilenceErrors: true` and `SilenceUsage: true`, since error display is handled explicitly by `sharedErrors.HandleGlobalError` and the `main` sentinel check.
- **Version metadata**: build-time version variables (`version`, `commit`, `date`) in `cmd/root.go` must be package-level `string` vars, never struct fields — the linker's `-ldflags -X` flag can only target package-level string vars, and a struct field will silently no-op the injection, shipping every release binary reporting `"dev"` and disabling self-update.
- **Spinners vs. cloud operations**: spinners are only started for k3d operations. Cloud operations (EKS/GKE) stream per-resource progress lines from the Terraform engine, and a spinner's in-place redraws would garble that output — use `pterm.Info.Printf` for cloud flows instead. Any method wrapping a potentially long-running provider call must guard on `config.Type == models.ClusterTypeK3d` before starting a spinner.
- **Positional-argument safety**: commands that target a cluster via a flag (e.g. `--context`) must reject stray positional arguments using a validator like `noPositionalArgs` that names the correct flag in the error — otherwise Cobra's default `ArbitraryArgs` silently ignores them, which is especially dangerous for destructive commands like `app uninstall`.
- **Production build flags**: all production builds use `CGO_ENABLED=0 go build -trimpath` for static, reproducible binaries. The only sanctioned exception is the race-detector target, which sets `CGO_ENABLED=1`.
- **Integration tests**: any test creating real k3d clusters or running real external binaries must start with the `//go:build integration` build tag and live under `tests/integration/`. The Makefile's unit-test target excludes that directory; the integration target adds `-tags integration`.

Two deterministic, organization-wide checks also apply to every PR:

- **No duplicate definitions** (`DUP-001`, warn): a new or changed exported definition whose normalized structure matches an existing one elsewhere in the org is flagged. Reuse the existing definition, extract a shared helper, or explain in the PR why a second copy is needed.
- **New exports need a test** (`TEST-001`, info): a new or changed exported definition that no test file references is flagged as shipped unverified.

Before writing new code, you can check the live rule set that applies to the files you're touching — `get_code_rules` (scoped by `languages`/`files`) returns the active, enforced catalog rather than a static snapshot.

## Branch naming and PR process

- Default/base branches are `main` and `develop`. The test workflow (`.github/workflows/test.yml`) runs its cheap gates (lint, cross-compile matrix, unit tests) against pull requests targeting **either** branch, since `develop` is where most work actually merges. The heavier end-to-end lifecycle job (`test-cli`) is restricted to pull requests whose base is `main`, or pushes to `main`, to avoid monopolizing self-hosted runners on `develop` traffic.
- Releases are cut **only** from `main` via a manual `workflow_dispatch` of `.github/workflows/release.yml` — there is no automatic release on push. The release workflow also has a hard guard that fails if dispatched from anything other than `refs/heads/main`, because the self-updater's cosign signing identity is pinned to `release.yml@refs/heads/main` (`internal/shared/selfupdate/cosign.go`). Do not rename `release.yml` or remove that guard.
- Stale branches are swept automatically: `.github/workflows/cleanup.yml` deletes non-protected branches with no commits for 60 days, skipping `main`, `develop`, and any branch with an open pull request.
- Open a pull request against `flamingo-stack/openframe-cli` (https://github.com/flamingo-stack/openframe-cli/pulls) once your branch is ready. CODEOWNERS (`.github/CODEOWNERS`) assigns `@flamingo-stack/devops-engineers` as the default reviewer for the whole repository, and specifically for `.github/CODEOWNERS` itself.
- Before merge, CI (`.github/workflows/test.yml`) must pass: `lint` (golangci-lint + `go mod tidy -diff`), `release-build-matrix` (compile-only check across all six GoReleaser GOOS/GOARCH combinations: linux/amd64, linux/arm64, windows/amd64, windows/arm64, darwin/amd64, darwin/arm64), `unit` (unit tests on Linux/Windows/macOS, plus `-race` on Linux), and — for PRs targeting `main` from this repository itself — `test-cli` (full e2e cluster lifecycle across all three OSes).
- An automated review pass is also available on pull requests via the `🦩 Flamingo Code Review` workflow (`.github/workflows/flamingo-code-review.yml`). It does not run automatically on every event: a draft or freshly opened PR gets an invitation comment, and a review is actually triggered by checking a box in that comment, adding the `flamingo-review` label, or commenting `@flamingo-review` (optionally `@flamingo-review full` to re-review the whole cumulative diff instead of just the incremental delta).
- PRs are squash-merged. Because the PR **title** becomes the squashed commit subject that `semantic-release` analyzes on `main`, the title itself must be a valid Conventional Commit — this is enforced by `.github/workflows/pr-title.yml`, which lints the PR title with `commitlint` against `@commitlint/config-conventional` on every `opened`, `edited`, `reopened`, or `synchronize` event. A free-form title contributes nothing to any future release version bump and will fail this check.

## Commit message format

This repository follows **Conventional Commits**, enforced at the PR-title level (not necessarily on every individual commit, since the repo squash-merges). The format is:

```text
type(scope)?: subject
```

Examples straight from the release workflow's own comments:

```text
fix(k3d): tolerate KUBECONFIG lists
feat!: drop v1 config
```

- `type` is one of the conventional types (`fix`, `feat`, `chore`, `docs`, `refactor`, `test`, `ci`, etc.) recognized by `@commitlint/config-conventional`.
- `scope` is optional and parenthesized, naming the affected area (e.g. `k3d`, `cluster`, `app`).
- A `!` after the type/scope (e.g. `feat!:`) marks a breaking change, which `semantic-release` maps to a major version bump.
- `fix` commits map to a patch release, `feat` to a minor release, and a `BREAKING CHANGE` footer or `!` marker to a major release — this is how `.releaserc.yml`-driven `semantic-release` computes the next version with no manual version input.

Write your PR title (which becomes the merged commit subject) in this format from the start — the `pr-title.yml` workflow lints it on every title edit, and a non-conforming title fails the check.

## Review checklist

Before requesting review, and while reviewing someone else's PR against this repository, verify:

- [ ] `golangci-lint run ./...` passes locally with no new findings (the gate is blocking in CI; the audit backlog is already clean, so any new issue fails the pipeline).
- [ ] `go mod tidy -diff` reports no drift.
- [ ] Unit tests pass: `make test-unit` covers the root, `cmd/`, `internal/`, and `tests/testutil/` packages. On Linux, also run the race-detector variant (`make test-race`).
- [ ] Any new test that touches a real k3d cluster or shells out to external binaries is gated with `//go:build integration` and lives under `tests/integration/` — it must not run as part of `go test ./...` or `make test-unit`.
- [ ] New or changed Cobra commands are wired through a `Get<Group>Cmd()`/`get<Sub>Cmd()` factory, not an inline `cobra.Command{}` literal, and any custom `PersistentPreRunE` calls `ui.ApplyGlobalOutputFlags(cmd)`.
- [ ] New service logic lives on an unexported struct behind a `New<Name>Service` constructor, with dependencies injected rather than derived internally.
- [ ] Errors are wrapped with context (`fmt.Errorf("...: %w", err)`) at each call boundary, exit codes from external tools are preserved via `*executor.CommandError`, and errors already shown to the user are returned as `*sharederrors.AlreadyHandledError` to avoid double-printing.
- [ ] No raw `fmt.Println`/`fmt.Printf` was added for user-facing output — it goes through `pterm` or `internal/shared/ui`.
- [ ] Spinners are used only for k3d flows, never for EKS/GKE operations that stream Terraform progress.
- [ ] Commands that accept a `--context`/cluster-targeting flag reject stray positional arguments with a named-flag hint, especially for destructive commands.
- [ ] The change does not duplicate an existing exported definition elsewhere in the organization's codebase (the `dup-scan` check); if a near-duplicate is intentional, the PR explains why.
- [ ] Any new or changed exported definition is exercised by at least one test.
- [ ] The PR title is a valid Conventional Commit (`type(scope)?: subject`), since it becomes the squashed commit message `semantic-release` reads from `main`.
- [ ] If the change touches `.github/workflows/release.yml`, the `refs/heads/main`-only guard and the `release.yml` filename are both left intact — both are load-bearing for the self-update cosign trust chain.
- [ ] If the change touches a GitHub Actions workflow that downloads a script from the Flamingo hub, the download goes through the shared `download_and_verify` function after `load_script_manifest`, verified against the hub's manifest digest — never a hash pinned in the workflow YAML, and never `curl | bash`.

For guidance on setting up a local build and test environment before you start, see `../setup/environment.md` and `../setup/local-development.md`. For a deeper look at package responsibilities and how commands route into services and providers, see `../architecture/README.md` and `../../reference/architecture/overview.md`.
