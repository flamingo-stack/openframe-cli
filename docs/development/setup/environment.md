# Development Environment Setup

Tools and editor configuration for working on OpenFrame CLI (a Go/Cobra CLI). This is generic Go tooling — adapt it to your own preferences.

## Prerequisites

- Go 1.26 or later
- Docker plus Kubernetes tooling: `kubectl`, `helm`, `k3d`
- Git

Install the Kubernetes tools with your package manager, e.g. on macOS:

```bash
brew install kubectl helm k3d
```

## Go Tools

```bash
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/go-delve/delve/cmd/dlv@latest   # debugger
```

## VS Code

The Go extension (`golang.Go`) covers most needs. A reasonable `.vscode/settings.json`:

```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.formatTool": "goimports",
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": "explicit"
  },
  "files.eol": "\n"
}
```

Debug configurations in `.vscode/launch.json`. `--non-interactive` reuses the existing `helm-values.yaml`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch OpenFrame CLI",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/main.go",
      "args": ["--help"]
    },
    {
      "name": "Debug Bootstrap",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/main.go",
      "args": ["bootstrap", "--verbose", "--non-interactive"]
    },
    {
      "name": "Debug Cluster Status",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/main.go",
      "args": ["cluster", "status"]
    }
  ]
}
```

## Command-line Debugging with Delve

```bash
# Debug the binary
dlv debug . -- cluster status

# Debug a package's tests
dlv test ./internal/bootstrap/
```

## Optional Shell Aliases

```bash
alias oft="go test ./..."
alias ofl="golangci-lint run ./..."
alias k="kubectl"
alias kof="kubectl config use-context k3d-openframe-local"
```

## Next Steps

- **[Local Development](local-development.md)** - Clone, build, run, and test the CLI
- **[Architecture Overview](../architecture/README.md)** - Understand the system design
