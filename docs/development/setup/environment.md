## Development Environment Setup

This guide walks through setting up a local development environment for **OpenFrame CLI** — the Go-based command-line tool that provisions Kubernetes clusters (k3d, GKE, EKS) and deploys the OpenFrame platform onto them via ArgoCD's app-of-apps pattern.

> OpenFrame CLI is a **service** repository published to the `go` ecosystem as `github.com/flamingo-stack/openframe-cli`. Per the organization's repository graph, it has no recorded upstream dependencies on other internal repositories and no recorded downstream consumers — its `internal/` packages are private implementation detail, not a published API surface for other modules to import.

If you are looking to set up the broader OpenFrame platform (the Kubernetes workloads this CLI installs), see the external repository **flamingo-stack/openframe-oss-tenant** at `https://github.com/flamingo-stack/openframe-oss-tenant` — this document covers only the CLI codebase itself.

---

### 1. Required Development Tools

The CLI is a single Go module (`github.com/flamingo-stack/openframe-cli`, see `go.mod`). The tools below are required to build, lint, and exercise it locally.

| Tool | Why it's needed | Notes |
|---|---|---|
| **Go** | Compiles and runs the CLI | `go.mod` pins `go 1.26.0` with `toolchain go1.26.5`. The toolchain line exists specifically to avoid a GC-related crash on windows-amd64 present in 1.26.0/1.26.1 (fixed in 1.26.2) — `go build`/`go test` will auto-select `go1.26.5` via the toolchain directive, so install at least Go 1.26.0 and let the toolchain mechanism do the rest. |
| **git** | Source control, and used at runtime by `internal/chart/providers/git` to clone the app-of-apps chart repository | Required dependency of `go-git/go-git/v5`, but a native `git` binary is also expected on `$PATH` for everyday repo work. |
| **Docker** | Backs local `k3d` clusters created by `openframe cluster create` / `openframe bootstrap` | Exercised through `internal/cluster/prerequisites/docker` and the k3d provider in `internal/cluster/providers/k3d`. Needed if you intend to run end-to-end flows locally rather than only unit tests. |
| **Helm** | Installs ArgoCD and the app-of-apps chart | Shelled out to via `internal/chart/providers/helm`; also checked/installed via `internal/chart/prerequisites/helm` and `internal/cluster/prerequisites/helm`. |
| **k3d** | Creates local Kubernetes clusters | Managed by `internal/cluster/providers/k3d`; the CLI downloads a pinned, checksum-verified binary through `internal/shared/download` if not already present. |
| **Terraform** | Provisions cloud clusters (EKS/GKE) | Used via `hashicorp/terraform-exec`/`terraform-json` (see `go.mod`) and the Terraform root modules in `internal/cluster/providers/eks/templates/main.tf` and `internal/cluster/providers/gke/templates/main.tf`. Only required if you work on the cloud-provider code paths. |
| **gcloud** / **aws** CLIs | Cloud provider authentication and cluster discovery | Used by `internal/cluster/prerequisites/gcloud`, `internal/cluster/prerequisites/aws`, and the discovery logic in `internal/cluster/discovery/{eks,gke}.go`. Only required for GKE/EKS work. |
| **golangci-lint (v2)** | Static analysis gate used in CI and locally before pushing | Configuration lives in `.golangci.yml` at the repo root. See below for what it enforces. |

All of the CLI-managed binaries (k3d, Helm, mkcert, Terraform, infracost) are downloaded and SHA256-verified by the CLI itself at runtime via `internal/shared/download/verify.go` and pinned in `internal/shared/download/pins.go` — you do not need to pre-install these system-wide for the CLI to function, though having them already on `$PATH` speeds up iteration.

#### Cloning the repository

```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
```

#### Building and running

```bash
go build -o openframe .
./openframe --help
```

Because `main.go` is the single entry point (see `main.go`), `go run .` also works for quick iteration:

```bash
go run . cluster list
```

#### Running the linter

`.golangci.yml` configures `golangci-lint` v2 with the standard default linter set (`errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`) plus `gosec`, `bodyclose`, and `errorlint` enabled explicitly. `gosec`'s full default rule set is kept (no exclusions) because it catches the classes of issue the project specifically watches for: subprocess construction (G204), file path handling (G304), file permissions (G306), and unverified downloads (G107) — all relevant given how much of this CLI shells out to external tools and downloads binaries.

```bash
golangci-lint run
```

Test files (`_test.go`) are exempted from `errcheck` and `gosec` findings, since they legitimately ignore errors and use weak fixtures — see the `issues.exclusions.rules` block in `.golangci.yml`.

#### Running tests

The repository has dedicated test-support packages worth knowing about before writing or running tests:

- `tests/testutil/` — shared assertions, cluster test helpers, command-assertion helpers, flag-contract checks, and setup utilities (`tests/testutil/assertions.go`, `tests/testutil/cluster.go`, `tests/testutil/command_assertions.go`, `tests/testutil/flag_contract.go`, `tests/testutil/setup.go`, `tests/testutil/utilities.go`, `tests/testutil/patterns.go`).
- `tests/integration/common/` — a CLI runner, cluster-management helpers, and dependency helpers for integration-style tests (`tests/integration/common/cli_runner.go`, `tests/integration/common/cluster_management.go`, `tests/integration/common/dependencies.go`).
- `internal/shared/executor/mock.go` — a mock `CommandExecutor`, used so unit tests don't need Docker/Helm/Terraform actually installed to exercise command-construction logic.

```bash
go test ./...
```

---

### 2. IDE Recommendations and Setup

The codebase has no IDE-specific configuration committed (no `.vscode/`, `.idea/`, or editor config files were found in the repository), so any editor with solid Go tooling works. Given the module is plain Go with no frontend or JVM build, the practical requirement is: an editor that can run `gopls` (the Go language server) and invoke `golangci-lint`.

**Recommended setup, in order of how directly it maps to this codebase:**

- **GoLand / IntelliJ with the Go plugin** — strong support for Go modules, built-in test runner integration for `go test ./...`, and native `go.mod`/toolchain awareness that matches the pinned `toolchain go1.26.5` directive.
- **VS Code with the official Go extension** — lightweight, uses `gopls` under the hood, and can be configured to run `golangci-lint` as the linter backend (matching `.golangci.yml` exactly, rather than VS Code's bundled default linters).
- **Vim/Neovim with `gopls` via an LSP client** — viable for contributors already in a terminal-first workflow; pairs well with the CLI's own terminal-UI-heavy codebase (`internal/shared/ui`, `internal/app/status/tui`).

Whichever editor you choose, point its linter integration at `.golangci.yml` rather than relying on default settings, since the project intentionally diverges from golangci-lint's defaults (e.g., `govet.disable: [fieldalignment, shadow]`, full `gosec` with no excludes).

#### Editor extensions/plugins

| Editor | Extension/Plugin | Purpose |
|---|---|---|
| VS Code | Go (`golang.go`) | `gopls` integration, go-to-definition, inline diagnostics, test lens |
| VS Code | YAML | Editing Helm values files and GitHub Actions workflows cleanly |
| VS Code | HashiCorp Terraform | Syntax highlighting/validation for `internal/cluster/providers/eks/templates/main.tf` and `internal/cluster/providers/gke/templates/main.tf` |
| GoLand/IntelliJ | Go plugin (bundled in GoLand) | Same role as above, built in |
| Any | golangci-lint integration (via LSP or extension) | Surfaces the same findings `golangci-lint run` reports, inline |

> No `.editorconfig` or formatting-tool configuration beyond standard `gofmt`/`goimports` conventions was found in the repository. Rely on `gofmt`-on-save (most Go tooling defaults to this) to keep formatting consistent with the existing codebase.

---

### 3. Environment Variables for Development

The CLI reads a handful of environment variables at runtime, primarily to control output mode and to interoperate with the ambient kubeconfig. None are required simply to build or unit-test the project — they matter when you exercise real commands against a cluster.

| Variable | Purpose |
|---|---|
| `$KUBECONFIG` | Standard Kubernetes convention honored by `internal/cluster/providers/eks/kubeconfig.go` when merging a newly created EKS cluster's context into your default kubeconfig (`mergeIntoDefaultKubeconfig`) and when removing it on teardown (`removeFromDefaultKubeconfig`). If unset, the usual `~/.kube/config` default applies. |
| `$HOME` | Used to resolve the certificate directory. `internal/chart/utils/config/paths.go` (`GetCertificateDirectory`) stores generated certificates under `~/.config/openframe/certs`, falling back to a path under the working directory if `$HOME` cannot be resolved. |

Beyond these, most "configuration" is passed through CLI flags rather than environment variables — for example, `--context` to target a specific kube-context (see `cmd/app/app.go`'s `noPositionalArgs` helper, which explicitly rejects positional cluster-name arguments in favor of `--context`), `--silent`/`--verbose`/`--plain` as global output flags applied via `ui.ApplyGlobalOutputFlags`, and `--type k3d|eks|gke` for prerequisite checks.

When working on code that touches `internal/shared/config/envbool.go`, note that the pattern for any new boolean environment-variable toggle in this codebase goes through that shared helper rather than ad hoc `os.Getenv` checks — look there before adding a new one.

> Do not commit real Docker registry credentials, Ngrok auth tokens/API keys, or any other secret into a Helm values file or environment file. The interactive configuration flow (`internal/chart/ui/configuration/docker.go`, `internal/chart/ui/configuration/ingress.go`) deliberately masks these inputs and registers them with `internal/shared/redact` so they never leak into verbose logs — treat your local shell environment and `openframe-helm-values.yaml` with the same care.

---

### 4. Working With the Codebase: Structure You'll Touch Most

For day-to-day development, the areas you are most likely to be editing are:

- `cmd/` — Cobra command definitions and flag wiring (`cmd/root.go`, `cmd/cluster/`, `cmd/app/`, `cmd/bootstrap/`, `cmd/prerequisites/`, `cmd/update/`).
- `internal/cluster/` — cluster lifecycle logic, provider dispatch (k3d/EKS/GKE), and the prerequisite checkers per cluster type.
- `internal/chart/` — ArgoCD/Helm/Git providers, the chart installation service, and the interactive configuration wizard (`internal/chart/ui/configuration/`).
- `internal/shared/` — cross-cutting infrastructure: command execution (`internal/shared/executor`), error handling (`internal/shared/errors`), terminal UI (`internal/shared/ui`), download/verification (`internal/shared/download`), and self-update (`internal/shared/selfupdate`).
- `internal/k8s/` — kubeconfig/context resolution and cluster health checks, deliberately separated from `internal/cluster` (creation) so `app install` can target any reachable cluster, not just ones the CLI created.

A quick sanity-check command once your environment is set up:

```bash
go build -o openframe . && ./openframe prerequisites check --type k3d
```

This exercises the Cobra command tree, the prerequisites framework (`internal/prerequisites/runner.go`), and the terminal UI output path without requiring a live cluster.

---

For first contact with the CLI's commands and a working cluster, see the introduction and quick-start guides; for the wider system this CLI deploys onto, see the architecture overview and ecosystem documentation. For local day-to-day workflow details beyond environment setup (running against a live cluster, iterating on chart installs, etc.), see the local development guide. Contribution conventions are documented separately in the contributing guidelines.
