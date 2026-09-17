# Development Environment Setup

## Required Tooling

| Tool | Version | Purpose |
|---|---|---|
| Go | 1.26.0+ (toolchain pins `1.26.5`) | Building and testing the CLI |
| Git | Any recent version | Cloning the repo, running `go-git`-based tests |
| Docker | Any recent version | Required to exercise `k3d`-backed cluster code paths and integration tests |
| Helm CLI | Latest | Used by chart installation code paths; also downloadable/pinned by the CLI itself |

> **Note:** The module declares `go 1.26.0` with `toolchain go1.26.5` in `go.mod`. This pin exists because Go 1.26.0/1.26.1 have a known crash on `windows-amd64` (corrupted return addresses during GC stack scanning); the fix shipped in 1.26.2. Always let Go resolve the toolchain from `go.mod` rather than overriding it manually.

Install Go from [go.dev/dl](https://go.dev/dl/), then verify:

```bash
go version
```

## IDE Recommendations

Any editor with solid Go tooling works well:

- **VS Code** with the official Go extension (`golang.go`) — provides `gopls`-based autocomplete, `go vet`, and inline test running.
- **GoLand** — full-featured Go IDE with built-in debugging and test runners.
- **Neovim** with `gopls` configured via `nvim-lspconfig`.

Recommended editor settings for this repository:

- Enable `gofmt`/`goimports` on save to match the project's formatting.
- Enable `go vet` and `staticcheck` linting if your IDE supports it, since the codebase relies on structured error wrapping (`errors.As`/`errors.Is`) that benefits from static analysis.

## Environment Variables for Development

These are optional but useful while developing/debugging:

| Variable | Purpose |
|---|---|
| `OPENFRAME_NO_WSL_FORWARD` | On Windows, disables WSL forwarding so you can iterate on Windows-specific code paths without re-entering WSL each time (cluster operations still require WSL to actually succeed) |
| `OPENFRAME_WSL_DISTRO` | Points WSL forwarding at a specific distro during development on Windows |
| `OPENFRAME_UPDATE_INSECURE_SKIP_VERIFY` | Bypasses cosign verification when developing/testing the self-update flow locally (never use in production) |

## Working with the Codebase Layout

```text
cmd/                  Cobra command definitions (thin adapters)
internal/
  bootstrap/          bootstrap.Service - composite create+install flow
  cluster/            ClusterService, providers (k3d/EKS/GKE), UI wizards
  chart/              ChartService, ArgoCD/Helm/Git providers, config builder
  app/                status aggregation, uninstall service
  k8s/                read-only kubeconfig/rest.Config access
  prerequisites/      generic OS-aware prerequisite framework
  platform/           OS detection + install-hint text
  shared/
    executor/         CommandExecutor abstraction (real + mock)
    download/         pinned, checksum-verified tool downloads
    selfupdate/       GitHub-release self-update + cosign verification
    errors/           structured errors, retry policy, friendly hints
    ui/                terminal presentation layer
    wsllauncher/       Windows→WSL2 forwarding
tests/
  testutil/           shared unit-test helpers (mock executor, flag contracts)
  integration/common/  builds and drives the real openframe binary
```

Once your Go toolchain is installed and your editor is configured, continue to [Local Development](local-development.md) to clone, build, and run the CLI.
