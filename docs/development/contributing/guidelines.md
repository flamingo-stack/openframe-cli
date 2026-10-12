
# Contributing Guidelines

OpenFrame CLI (`github.com/flamingo-stack/openframe-cli`) is a Go command-line tool, built with Cobra, that provisions Kubernetes clusters (k3d locally, GKE/EKS via Terraform) and installs the OpenFrame platform onto them through ArgoCD. This guide documents the conventions enforced in this repository: code style, branching, commit/PR format, and what a reviewer checks before approving a change.

This repository does not use GitHub Issues or GitHub Discussions for coordination — day-to-day discussion happens on the OpenMSP Slack community (https://www.openmsp.ai/, invite: https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA). Pull requests on `https://github.com/flamingo-stack/openframe-cli/pulls` remain the unit of code review.

> For environment setup and local build/test workflows, see `../setup/environment.md` and `../setup/local-development.md`. For how the codebase is organized, see `../architecture/README.md` and `../../reference/architecture/overview.md`.

---

## 1. Code style and conventions

### 1.1 Toolchain

The module is pinned in `go.mod`:

```text
module github.com/flamingo-stack/openframe-cli
go 1.26.0
toolchain go1.26.5
```

The `toolchain` line is intentional, not decorative — Go 1.26.0/1.26.1 have a known Windows-amd64 crash (corrupted return addresses during GC stack scanning), fixed in 1.26.2. Do not downgrade the toolchain directive without checking the comment in `go.mod`.

### 1.2 Linting (`.golangci.yml`)

This repository runs **golangci-lint v2** with the `standard` default set (`errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`) plus explicitly enabled linters:

| Linter | Purpose |
|---|---|
| `gosec` | Security: subprocess construction (G204), file path handling (G304), file permissions (G306), unverified downloads (G107), etc. The full default rule set is kept — `excludes: []` — so do not add exclusions without a strong reason documented in the PR. |
| `bodyclose` | HTTP response bodies must be closed. |
| `errorlint` | Correct `errors.Is`/`errors.As`/`%w` usage. |
| `misspell` | Catches common misspellings. |

`govet` runs with `enable-all: true`, disabling only `fieldalignment` (noise) and `shadow` (too noisy for this codebase currently).

Test files (`_test.go`) are exempted from `errcheck` and `gosec` — fixtures and intentionally-ignored errors in tests are fine, production code is not.

Run it locally before pushing:

```bash
golangci-lint run ./...
```

CI (`.github/workflows/test.yml`, job `lint`) runs the same check as a **blocking gate** — the backlog is already at zero issues, so any new finding fails the build. The same job also checks that `go.mod`/`go.sum` are tidy:

```bash
go mod tidy -diff
```

### 1.3 Error handling

Wrap errors at call boundaries with context, using the standard library's `%w` verb:

```go
if err != nil {
    return fmt.Errorf("create cluster: %w", err)
}
```

A bare `return err` that discards the call site's context is not acceptable practice in this codebase — give the reviewer (and the eventual user-facing error renderer in `internal/shared/errors/`) enough context to explain *what* failed, not just that something did. Keep the prefix short and lowercase, describing the operation, not repeating the underlying error text.

### 1.4 Version metadata

`cmd/root.go` carries build-time version metadata (`version`, `commit`, `date`) injected via `-ldflags -X`. The Go linker's `-X` flag can only set **package-level `string` variables** — never struct fields. If you touch `cmd/root.go`, keep these as package-level `var` declarations; moving them into a struct silently breaks the ldflags injection and ships binaries reporting `"dev"`, which also disables self-update (see `internal/shared/selfupdate/`).

### 1.5 Builds

Production builds use `CGO_ENABLED=0 go build -trimpath` (matching the GoReleaser release config and the `Makefile`): this keeps binaries statically linked (no libc dependency, needed for Linux releases and WSL) and removes build-machine paths for reproducibility. The only sanctioned exception is the race-detector test target, which needs `CGO_ENABLED=1`.

### 1.6 Structure and package boundaries

Follow the existing separation of concerns rather than introducing new cross-cutting packages:

- `cmd/` — Cobra command definitions only (flag parsing, wiring to services). Business logic belongs in `internal/`.
- `internal/cluster/` — cluster lifecycle; `internal/chart/` — ArgoCD/Helm/app-of-apps installation; `internal/app/` — status and uninstall; `internal/k8s/` — read-only access to an *existing* cluster (deliberately separate from cluster *creation* in `internal/cluster/`).
- `internal/shared/` — cross-cutting infrastructure: `executor/` (all shelled-out commands go through the `CommandExecutor` abstraction, real or mock — do not call `os/exec` directly from business logic), `ui/`, `errors/`, `download/` (checksum-verified pinned-tool downloads), `selfupdate/`.

Since the ecosystem graph records no other repository depending on `openframe-cli` (see `../../reference/architecture/ecosystem.md`), `internal/` packages are private implementation detail with no external Go API contract — but that does not relax internal consistency expectations.

### 1.7 Duplicate definitions and test coverage

Two deterministic checks run across the org's code graph on every reviewed change:

- **No duplicate definitions** — a new or changed exported definition whose normalized structure (identifiers/literals abstracted, comments ignored) matches an existing definition elsewhere in the org is flagged. Reuse the existing definition, extract a shared helper, or explain in the PR description why a second copy is necessary.
- **New exported definitions should be tested** — an exported function, type, or method that no test file references is flagged as shipped unverified. Add or extend a test in the relevant `_test.go` file, or `tests/testutil/` helpers, or explain why it isn't testable.

### 1.8 Shell scripts

`scripts/sign-binary.sh` is the one shell script in this repository (used from GoReleaser build hooks). Keep shell scripts POSIX-ish, `set -euo pipefail` where applicable, and avoid introducing new top-level scripts without equivalent care around input handling — the CI workflows in `.github/workflows/` demonstrate the expected defensiveness (e.g. quoting, explicit `trap` cleanup, mktemp for secrets).

---

## 2. Branch naming and PR process

### 2.1 Long-lived branches

`main` and `develop` are the only long-lived branches. Both are tested on every push and pull request (`.github/workflows/test.yml` triggers on `pull_request`/`push` to `[main, develop]`). `develop` is where day-to-day work merges and gets the full cheap-gate treatment (lint, cross-compile matrix, unit tests); the heavy end-to-end cluster lifecycle test (`test-cli` job) is **main-bound only**.

Releases are cut manually from `main` via `workflow_dispatch` on `.github/workflows/release.yml` — merging to main does **not** itself trigger a release.

### 2.2 Branch hygiene

A scheduled job (`.github/workflows/cleanup.yml`) deletes unprotected branches with no commits in the last 60 days, skipping `main` and `develop` and skipping any branch with an open pull request. Keep topic branches short-lived: open a PR promptly, or expect your branch to be swept up after two months of inactivity.

There is no enforced branch-naming prefix convention recorded in this repository's workflows; name branches descriptively for the change they carry (for example, referencing the area touched, such as a cluster provider, a chart service, or the self-updater) so reviewers and the stale-branch cleanup job can tell what they are for.

### 2.3 Pull request process

1. Branch from `develop` for ordinary feature/fix work (or from `main` only for a release-targeted hotfix).
2. Make your changes, keeping to the package boundaries in §1.6.
3. Run `golangci-lint run ./...` and `go mod tidy -diff` locally — both are blocking CI gates.
4. Run unit tests (`make test-unit` is what CI's `unit` job runs) before opening the PR.
5. Open the PR against `develop` (or `main` for a hotfix). The **PR title**, not just commit messages, is linted as a Conventional Commit (see §3) — this is enforced by `.github/workflows/pr-title.yml` on every `opened`, `edited`, `reopened`, and `synchronize` event, because this repository squash-merges, and the PR title becomes the final commit subject that `semantic-release` analyzes on `main`.
6. CI runs automatically: `lint`, `release-build-matrix` (compile-only across all six GoReleaser `GOOS`/`GOARCH` combos), `unit` tests (matrix of Linux/Windows/macOS, plus `-race` on Linux), and — for PRs targeting `main` from this repository itself — the full `test-cli` end-to-end lifecycle job against a real k3d cluster.
7. Address review feedback (see the checklist in §4) and keep the PR focused — one logical change per PR makes both human review and the duplicate/test-coverage checks in §1.7 meaningful.
8. A maintainer squash-merges once checks pass and the PR is approved. Per `.github/CODEOWNERS`, `@flamingo-stack/devops-engineers` owns the whole repository (and `.github/CODEOWNERS` itself) by default.

### 2.4 Automated code review

This repository runs an automated review pipeline (`.github/workflows/flamingo-code-review.yml`) alongside human review. It does **not** run on every push — only when explicitly invited:

- Commenting `@flamingo-review` on an open PR triggers an incremental pass (diff since the last reviewed commit); `@flamingo-review full` re-reviews the whole cumulative diff.
- Checking a box in the bot's own summary comment (e.g. "Review the new commits") re-triggers the same review without typing a command.
- Applying the `flamingo-review` label requests a one-off pass; the label is consumed (removed) immediately so it can be reapplied later.
- Applying `flamingo-review-always` subscribes the PR to a review on every push, for higher-risk branches.

Fork-authored PRs do not get this automated review (it requires write-token access the workflow deliberately withholds from forks), so cross-fork contributions rely on human review alone.

---

## 3. Commit message format

Because this repository **squash-merges pull requests**, the PR title — not individual commit messages on the branch — becomes the commit subject on `main`, and that subject is what `semantic-release` parses to decide the next release version (`.releaserc.yml` + `conventional-changelog-conventionalcommits`). Get the **PR title** right; it is what's linted and what matters for versioning.

### 3.1 Format

Conventional Commits, enforced by `.github/workflows/pr-title.yml` via `@commitlint/cli` with `@commitlint/config-conventional`:

```text
type(scope)?: subject
```

Examples straight from the workflow's own comments:

```text
fix(k3d): tolerate KUBECONFIG lists
feat!: drop v1 config
```

- `type` — one of the conventional types (`feat`, `fix`, `chore`, `docs`, `refactor`, `test`, `build`, `ci`, etc. — standard `@commitlint/config-conventional` types).
- `(scope)` — optional, parenthesized, names the affected area (e.g. a provider, a command group).
- `!` after the type/scope — marks a **breaking change**, which `semantic-release` maps to a major version bump.
- `subject` — concise, imperative description of the change.

### 3.2 Why it matters here

`semantic-release` (triggered manually via the `release` workflow's `workflow_dispatch`, never automatically on push) walks the Conventional Commits since the last tag to compute the next version:

- `fix:` → patch bump
- `feat:` → minor bump
- a breaking-change marker (`!` or a `BREAKING CHANGE:` footer) → major bump

A PR title that doesn't follow this format contributes nothing to the release notes or version calculation once squashed — so the lint is a merge gate, not a style suggestion. If `pr-title.yml` fails, edit the PR title (not just the description) and the check will re-run on the `edited` event.

### 3.3 Individual commits on your branch

Individual commits within your feature branch are not themselves linted by `pr-title.yml` (only the PR title is checked), but writing them with the same discipline makes review easier and gives you a clean history to split into a well-formed PR title. Keep commits scoped and avoid bundling unrelated changes — this also helps the duplicate-definition and missing-test checks (§1.7) evaluate a coherent diff.

---

## 4. Review checklist

Before requesting review, and what a reviewer (human or the automated `flamingo-review` pass) will check:

### Correctness and scope
- [ ] The change does one coherent thing; unrelated refactors are split into separate PRs.
- [ ] New or changed exported Go definitions do not duplicate an existing definition elsewhere in the org's codebase (checked by the deterministic `dup-scan` analyzer — see §1.7). If a near-duplicate is intentional, the PR explains why.
- [ ] New or changed exported definitions are exercised by a test in a `_test.go` file or under `tests/testutil/` / `tests/integration/`. If something genuinely cannot be tested, the PR says so.

### Code style
- [ ] `golangci-lint run ./...` passes with no new findings (the `standard` set plus `gosec`, `bodyclose`, `errorlint`, `misspell` — see §1.2).
- [ ] `go mod tidy -diff` reports no drift.
- [ ] Errors are wrapped with context at call boundaries (`fmt.Errorf("...: %w", err)`), not returned bare.
- [ ] Shelled-out commands go through `internal/shared/executor`'s `CommandExecutor` abstraction rather than calling `os/exec` directly, so they stay mockable in tests.
- [ ] Any file under `internal/shared/download/` that fetches external binaries keeps checksum verification intact — do not relax or bypass `internal/shared/download/verify.go` logic.
- [ ] If `cmd/root.go` is touched, `version`/`commit`/`date` remain package-level `var` strings (not struct fields), preserving `-ldflags -X` injection.

### Commits and PR metadata
- [ ] The PR title is a valid Conventional Commit (`type(scope)?: subject`, `!` for breaking changes) — this becomes the squash-merge commit subject and feeds `semantic-release`.
- [ ] The PR targets `develop` for normal work, or `main` only for a release-targeted hotfix.
- [ ] If the change affects a GitHub Actions workflow, any script it downloads at CI runtime is fetched only through the shared `download_and_verify` function against the hub's manifest digest — never `curl | bash`, and never a hash pinned directly in the workflow YAML.

### CI status
- [ ] `lint` job is green.
- [ ] `release-build-matrix` (compile-only across all six `GOOS`/`GOARCH` combos) is green.
- [ ] `unit` tests pass on all three OS legs, including `-race` on Linux.
- [ ] For a PR touching cluster/chart lifecycle logic and targeting `main` from this repository (not a fork), the `test-cli` end-to-end job against a real k3d cluster is green.

### Documentation
- [ ] Behavior-affecting changes to CLI commands, flags, or installation flow are reflected in relevant docs under `docs/` (for example `../../getting-started/` how-to guides, or `../architecture/README.md` if the architecture changed).
- [ ] If upstream dependencies in `go.mod` changed, the rationale is in the PR description (new transitive dependencies pulled in by `charmbracelet`, `k8s.io/client-go`, `sigstore/sigstore-go`, etc. should be traceable to a real need).

---

## Where to ask questions

This repository does not use GitHub Issues or Discussions. For design questions, clarifications on an in-progress PR, or help getting unblocked, use the OpenMSP Slack community (https://www.openmsp.ai/, invite: https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA) rather than opening an issue.

For related external context: OpenFrame CLI's platform-facing behavior (what it installs onto a cluster) is defined by the **openframe-oss-tenant** repository (https://github.com/flamingo-stack/openframe-oss-tenant), which is maintained separately — it is not a local subdirectory of this repository. If a contribution here needs a corresponding change in the chart/app-of-apps definitions, coordinate across both repositories explicitly in your PR description.
