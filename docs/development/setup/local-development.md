# Local Development Guide

Clone, build, run, test, and debug OpenFrame CLI locally.

## Prerequisites

- **[Environment Setup](environment.md)** - Go toolchain, editor, and Kubernetes tools

## Clone the Repository

Fork on GitHub (recommended for contributors), then:

```bash
git clone https://github.com/YOUR-USERNAME/openframe-cli.git
cd openframe-cli
git remote add upstream https://github.com/flamingo-stack/openframe-cli.git
```

Or clone directly for read-only use:

```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
```

## Project Structure

```text
openframe-cli/
├── main.go        # Entry point
├── Makefile       # build / test / lint targets
├── cmd/           # Command definitions: bootstrap, cluster, app, prerequisites, update, root.go
├── internal/      # Private packages: bootstrap, cluster, chart, app, k8s, platform, prerequisites, shared
├── tests/         # integration/ and testutil/
└── docs/          # Documentation
```

Unit tests are colocated as `*_test.go` next to the code they cover.

## Build and Run

```bash
# Build for your current platform (produces openframe-<os>-<arch>)
make build

# Cross-compile all six release platforms (matches .goreleaser.yml)
make build-all

# Or build directly
go build -o openframe .
./openframe --version
```

Run without building during development:

```bash
go run . --help
go run . cluster status
go run . app status
```

## Run Tests

```bash
make test              # unit + integration
make test-unit         # ./cmd/... ./internal/...
make test-race         # unit tests with the race detector (needs CGO)
make test-integration  # ./tests/integration/...

# Or with go directly
go test ./...
go test -run TestClusterCreate ./internal/cluster/...
go test -cover ./...
```

Integration tests may require a running cluster:

```bash
k3d cluster create openframe-test
go test ./tests/integration/...
k3d cluster delete openframe-test
```

## Lint and Format

```bash
make fmt     # gofmt -w over the tree
make vet     # go vet ./...
make lint    # golangci-lint run ./...
make tidy    # fail if `go mod tidy` would change go.mod/go.sum
```

These mirror the CI gates — run them before pushing.

## Development Workflow

```bash
# Sync with upstream
git fetch upstream && git checkout main && git merge upstream/main

# Create a branch, make changes, then before committing:
make fmt vet tidy
make test
make lint

# Commit (conventional commits) and push
git commit -m "feat(cluster): add support for custom node labels"
git push origin feature/your-feature-name
```

## Debugging

### VS Code

Use the launch configurations from [Environment Setup](environment.md), set breakpoints, and press F5.

### Delve

```bash
go install github.com/go-delve/delve/cmd/dlv@latest

dlv debug . -- bootstrap --verbose --non-interactive
dlv test ./internal/bootstrap/
```

## Manually Testing Your Changes

The CLI's top-level commands are `bootstrap`, `cluster`, `app`, `prerequisites`, and `update`.

```bash
# Prerequisites
go run . prerequisites check

# Cluster lifecycle
go run . cluster create test-cluster
go run . cluster status test-cluster
go run . cluster list
go run . cluster delete test-cluster

# App-of-apps: clones openframe-oss-tenant and installs ArgoCD + the app-of-apps chart.
# --non-interactive reuses the existing openframe-helm-values.yaml.
go run . app install --non-interactive
go run . app status
go run . app access

# Full bootstrap (cluster + app-of-apps)
go run . bootstrap --non-interactive
```

Verify against a real cluster:

```bash
kubectl get pods --all-namespaces
kubectl get applications -n argocd
```

### Overriding ArgoCD chart values

The CLI installs ArgoCD from a built-in baseline (embedded
`internal/chart/providers/argocd/argocd-values.yaml`), which is separate from
the app-of-apps values. To change an ArgoCD chart value without rebuilding the
CLI, add a top-level `argocd:` section to `openframe-helm-values.yaml`:

```yaml
# openframe-helm-values.yaml
repository:
  branch: main            # (app-of-apps settings, as before)

argocd:                   # deep-merged over the built-in ArgoCD baseline
  dex:
    enabled: true         # e.g. re-enable dex (disabled by default)
  server:
    replicas: 2
```

Only the `argocd:` subtree is applied to the ArgoCD install — the rest of the
file targets the app-of-apps chart, and keeping them separate stops secrets
(e.g. the docker registry password) from leaking into the ArgoCD release. The
merge follows Helm semantics (maps merge, scalars/lists replace), and the CLI
prints a warning listing the keys you overrode, since a bad override can break
the ArgoCD install. Without an `argocd:` section the baseline is used unchanged.

## Cross-platform Builds

```bash
GOOS=linux   GOARCH=amd64 go build -o openframe-linux-amd64 .
GOOS=darwin  GOARCH=arm64 go build -o openframe-darwin-arm64 .
GOOS=windows GOARCH=amd64 go build -o openframe-windows-amd64.exe .
```

On Windows the CLI forwards into WSL2 and runs the Linux binary; that launch is handled by `internal/shared/wsllauncher`.

## Troubleshooting

```bash
# Module issues
go clean -modcache && go mod tidy

# Build cache
go clean -cache

# Kubernetes context
kubectl config current-context
kubectl config use-context k3d-openframe-local
```

## Next Steps

- **[Architecture Overview](../architecture/README.md)** - Understand the system design

## Getting Help

Search existing GitHub issues, or ask in the [OpenMSP community](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA).
