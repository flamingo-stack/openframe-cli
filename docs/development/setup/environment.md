## Development Environment Setup

This guide covers how to set up a local development environment for the OpenFrame CLI (`openframe-cli`), a Go-based command-line tool that provisions Kubernetes clusters (local k3d, or cloud GKE/EKS via Terraform) and deploys the OpenFrame platform onto them via ArgoCD's app-of-apps pattern.

> **Note:** The OpenFrame platform itself (the services the CLI deploys) lives in the separate [`flamingo-stack/openframe-oss-tenant`](https://github.com/flamingo-stack/openframe-oss-tenant) repository. This document covers setup for developing the CLI tool in **this** repository (`flamingo-stack/openframe-cli`) only — do not assume platform source lives in a subdirectory here.

### Scope

According to the repository ecosystem graph, `openframe-cli` is a **service** published to the `go` ecosystem as `github.com/flamingo-stack/openframe-cli`. It has no recorded upstream dependencies on other internal repositories and no recorded downstream consumers in the graph — its functional dependencies are entirely external, third-party Go modules and CLI tools (Docker, k3d, Helm, Terraform, gcloud, aws) invoked through its own executor abstraction.

---

## Required Development Tools

The CLI is a single Go module (see `go.mod`). The toolchain and the external binaries it shells out to must both be present locally.

### Go toolchain

| Requirement | Value | Source |
|---|---|---|
| Go module path | `github.com/flamingo-stack/openframe-cli` | `go.mod` |
| Minimum Go version | `go 1.26.0` | `go.mod` |
| Pinned toolchain | `go1.26.5` | `go.mod` |

> **Important:** `go.mod` pins the toolchain to `go1.26.5` deliberately. Go `1.26.0`/`1.26.1` crash intermittently on `windows-amd64` due to return addresses being corrupted during GC stack scanning (a known upstream Go issue), fixed in `1.26.2`. CI and release builds resolve the toolchain from this file, so do not downgrade or remove the `toolchain` directive — it is what keeps shipped Windows binaries off the buggy runtimes. If your local Go installation is older than `1.26.5`, Go's own toolchain management will download the pinned version automatically when you build.

Install Go from your platform's standard distribution channel, then verify:

```bash
go version
```

### External CLI tools invoked by the binary

The CLI does not link these as libraries — it shells out to them through `internal/shared/executor`. Having them installed (or letting the CLI's own `prerequisites` command install them) is required to actually *run* cluster/app commands end-to-end during development, though they are not required to compile the Go module itself:

| Tool | Used for |
|---|---|
| Docker | Backing runtime for local k3d clusters |
| k3d | Local Kubernetes cluster provisioning |
| Helm | Installing ArgoCD and chart releases |
| Terraform | Provisioning EKS/GKE clusters |
| gcloud CLI | GKE cluster discovery and auth |
| aws CLI | EKS cluster discovery, auth (`aws eks get-token`) |
| mkcert / certificates tooling | Local TLS certificates for ingress |
| infracost | Cost estimation for cloud cluster plans |

The CLI's own `openframe prerequisites check` and `openframe prerequisites install` commands (see `cmd/prerequisites/prerequisites.go`, `internal/prerequisites/runner.go`) can verify or auto-install most of these per OS — auto-install is supported on macOS/Linux, with docs-only guidance on Windows (see `internal/platform/platform.go`). Binaries managed this way are pinned and checksum-verified via `internal/shared/download` (see `internal/shared/download/pins.go`, `internal/shared/download/verify.go`) rather than fetched ad hoc.

### Linting

This repository is linted with `golangci-lint` v2, configured in `.golangci.yml`. The configuration:

- Uses the `standard` default linter set (`errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`)
- Additionally enables `gosec` (security — subprocess execution, file paths/perms, unverified downloads), `bodyclose`, `errorlint`, and `misspell`
- Runs `govet` with `enable-all: true`, disabling only `fieldalignment` (noise) and `shadow` (too noisy for this codebase currently)
- Excludes `errcheck` and `gosec` findings in `_test.go` files, since tests legitimately ignore some errors and use weak fixtures
- Reports every issue (`max-issues-per-linter: 0`, `max-same-issues: 0`) rather than capping output

Install `golangci-lint` locally and run it before opening a pull request:

```bash
golangci-lint run ./...
```

> `gosec`'s rule set is intentionally kept complete (`excludes: []`) because its findings map directly to real remediation items in this codebase — subprocess construction (G204), file path/permission handling (G304/G306), and unverified downloads (G107). Do not silently add exclusions when a finding looks noisy; fix the underlying code or get the exclusion reviewed.

### Code signing tooling (release-adjacent)

`scripts/sign-binary.sh` is the one shell script present in this repository, used for signing release binaries. It is not part of day-to-day `go build`/`go test` work but is worth knowing about if you touch the release pipeline.

---

## IDE Recommendations and Setup

The codebase is pure Go (plus a small amount of embedded Terraform under `internal/cluster/providers/{eks,gke}/templates/`), so any editor with solid Go tooling works well. The two most common choices in Go ecosystems:

- **GoLand** — full-featured Go IDE with built-in debugging, test runners, and refactoring support out of the box.
- **Visual Studio Code** with the Go extension — lightweight, widely used, and the easiest to align with the extensions list below.

Whichever editor you choose, make sure it is configured to use `gofmt`/`goimports` on save and can invoke `go vet` and `golangci-lint`, since those are the checks this repository is actually linted against (see `.golangci.yml`).

### Recommended editor configuration

- Enable format-on-save using `gofmt` (or `goimports`, which is a superset that also manages import grouping).
- Point your editor's Go language server at the module root (where `go.mod` lives) so package resolution for `github.com/flamingo-stack/openframe-cli/...` imports works correctly.
- Configure your editor to surface `golangci-lint` diagnostics inline, using the repository's `.golangci.yml` rather than each tool's built-in defaults — the project enables `gosec`, `bodyclose`, `errorlint`, and `misspell` on top of the standard set, and disables `fieldalignment`/`shadow` from `govet`'s `enable-all`. An editor using only its own defaults will show a different (looser) set of warnings than CI enforces.
- If your editor supports Terraform syntax highlighting, enable it for `*.tf` files under `internal/cluster/providers/eks/templates/` and `internal/cluster/providers/gke/templates/` — these are real Terraform root modules the CLI generates and manages, not throwaway examples.

### Editor extensions / plugins

For VS Code specifically:

| Extension | Purpose |
|---|---|
| Go (`golang.go`) | Official Go extension — language server (`gopls`), debugging (`delve`), test lens, formatting |
| golangci-lint integration (via the Go extension's lint settings) | Surfaces the same linter configuration enforced by `.golangci.yml` |
| YAML | Useful for editing Helm values templates referenced by `internal/chart/ui/templates` and `internal/chart/utils/config` |
| HashiCorp Terraform | Syntax/validation support for the embedded `main.tf` files under the EKS/GKE provider template directories |

For GoLand, the equivalent capabilities (Go language support, linter integration, Terraform plugin) are available as built-in or JetBrains-marketplace plugins; no repository-specific GoLand configuration is checked in.

> This repository does not ship a committed editor configuration directory (no `.vscode/` or `.idea/` settings are part of the material covered here) — the recommendations above are tooling choices to align with the lint rules actually enforced, not a mandated project-level setting.

---

## Environment Variables for Development

The CLI is designed to run without requiring custom environment variables for a basic build/test cycle (`go build`, `go test` need none beyond a working Go toolchain). The environment variables below become relevant depending on which commands you are exercising during development:

| Variable | Relevance |
|---|---|
| `KUBECONFIG` | Honored when merging or removing cluster contexts (see `internal/cluster/providers/eks/kubeconfig.go`, which merges EKS cluster access into the default kubeconfig and refuses to overwrite an existing context pointing at a different server). Point this at a scratch kubeconfig file while testing cluster lifecycle commands if you do not want your real kubeconfig touched. |
| AWS credentials (via the `aws` CLI's standard resolution: environment variables, profiles, SSO) | Needed only when exercising the EKS provider path (`internal/cluster/providers/eks`) or `cmd/cluster/aws_identity.go`. Kubeconfig auth for EKS clusters goes through the `aws eks get-token` exec plugin rather than static tokens, so no long-lived EKS token needs to be exported. |
| `gcloud` application-default credentials | Needed only when exercising the GKE provider path (`internal/cluster/providers/gke`). |
| Docker daemon availability (no single env var, but `DOCKER_HOST` if you use a non-default socket) | Needed for the k3d provider path (`internal/cluster/providers/k3d`), since k3d clusters run as Docker containers. |

Secrets collected interactively during chart configuration (Docker registry password, Ngrok API key/auth token — see `internal/chart/ui/configuration/docker.go` and `internal/chart/ui/configuration/ingress.go`) are registered with the redaction helper (`internal/shared/redact`) immediately after input, specifically so they never leak into verbose logs or error output. When developing or debugging around these code paths, be aware that any new secret-collecting prompt should follow the same `redact.RegisterSecret(...)` pattern rather than introducing an unredacted log line.

> Refer to your environment configuration for any cluster-specific credentials (cloud project IDs, regions, registry credentials) — none are hardcoded or defaulted in this repository, and no default credentials are shipped.

### Global CLI flags relevant during development

While not environment variables, the root command (`cmd/root.go`) exposes global flags that are useful while developing and debugging:

- `--verbose` — enables debug-level output across commands.
- `--silent` — suppresses interactive UI elements, useful when scripting repeated test runs.
- `--plain` — disables colored/decorated terminal output, useful when piping CLI output to logs or files during test iteration.

These are applied consistently via `ui.ApplyGlobalOutputFlags(cmd)`, including inside command groups like `app` (see `cmd/app/app.go`) that define their own `PersistentPreRunE` and must therefore re-apply the global flag contract explicitly.

---

## Building and Running Locally

From the module root (where `go.mod` lives):

```bash
# Build the binary
go build -o openframe .

# Run directly without building a binary
go run . --help

# Run the test suite
go test ./...
```

The entry point is `main.go`, which wires `cmd.Execute()` and preserves a failed external command's exit code for automation (so CI scripts calling `openframe` see the same exit code the underlying `helm`/`terraform`/`k3d` call produced, when applicable).

Typical local development loop commands, once built:

```bash
# Local k3d cluster + platform in one step
openframe bootstrap

# Check cluster status
openframe cluster status

# Verify or install required external tools for a given provider type
openframe prerequisites check --type k3d
```

---

## Where to Go Next

For day-to-day local development workflow details (running the CLI against a local cluster, iterating on chart configuration, etc.), see the local development guide. For the broader architecture this CLI implements — command groups, domain services, and providers — see the architecture overview and the ecosystem/dependency diagrams. For contribution conventions, see the contributing guidelines.

- Local development workflow: [`./local-development.md`](./local-development.md)
- Architecture overview: [`../architecture/README.md`](../architecture/README.md)
- Reference architecture (components, data flow, dependencies): [`../../reference/architecture/overview.md`](../../reference/architecture/overview.md)
- Ecosystem graph (published artifacts, consumers): [`../../reference/architecture/ecosystem.md`](../../reference/architecture/ecosystem.md)
- Architecture diagrams index: [`../../diagrams/architecture/README.md`](../../diagrams/architecture/README.md)
- Contribution guidelines: [`../contributing/guidelines.md`](../contributing/guidelines.md)
- Security practices: [`../security/README.md`](../security/README.md)
- Testing practices: [`../testing/README.md`](../testing/README.md)
