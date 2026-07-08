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

# Build for all platforms
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
make lint          # golangci-lint run ./...
gofmt -w .
goimports -w .
```

## Development Workflow

```bash
# Sync with upstream
git fetch upstream && git checkout main && git merge upstream/main

# Create a branch, make changes, then before committing:
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
# --non-interactive reuses the existing helm-values.yaml.
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
