# Development Environment Setup

This page covers the tools and editor setup recommended for working on OpenFrame CLI.

## Required Development Tools

| Tool | Purpose |
|---|---|
| Go | Primary language toolchain — the entire CLI is a Go module (`github.com/flamingo-stack/openframe-cli`) |
| Git | Source control; also required at runtime by `internal/chart/providers/git` to clone the app-of-apps chart repo |
| Docker | Required to exercise `k3d`-backed cluster code paths locally |
| k3d | Required to run/test local cluster provisioning end-to-end |
| Helm | Required to run/test ArgoCD and app-of-apps install/upgrade code paths |
| Terraform (>= 1.15.0) | Required to exercise EKS/GKE provider code paths |

> You can install most of these automatically using the CLI's own tooling once you have a first build: `openframe prerequisites install --type k3d` (or `eks`/`gke`).

## Recommended IDE Setup

Any editor with solid Go tooling works well for this codebase. Recommended setup:

- **VS Code** with the official Go extension (`golang.go`) — provides `gopls`-powered autocomplete, go-to-definition, and inline test running.
- **GoLand / IntelliJ with the Go plugin** — strong refactoring and debugging support for larger Go codebases.

Useful editor features/extensions for this project:

- Go language server (`gopls`) for navigation across the many `internal/` packages (`cluster`, `chart`, `shared`, `k8s`).
- `gofmt`/`goimports` on save, to match the existing formatting conventions.
- A Cobra/CLI-aware snippet or outline view helps when navigating the many `cmd/*` subcommand files.

## Environment Variables for Development

| Variable | Purpose |
|---|---|
| `OPENFRAME_UPDATE_INSECURE_SKIP_VERIFY` | Bypasses cosign signature verification in `internal/shared/selfupdate` — useful only when testing the self-update flow against unsigned local builds |
| `OPENFRAME_AUTO_UPDATE` | Set to `1` to exercise the opt-in automatic update-check code path |

Beyond these, the CLI reads standard cloud provider environment/config (AWS CLI config/credentials, gcloud CLI config) when exercising EKS/GKE provider code — no OpenFrame-specific cloud credentials are required.

## Verifying Your Setup

Once your toolchain is installed, confirm the module builds and its own prerequisite checks pass:

```bash
go build ./...
go vet ./...
```

Continue to [Local Development](local-development.md) for clone, build, and run instructions.
