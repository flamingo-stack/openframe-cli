# Development Documentation

This section covers everything you need to contribute to and work on OpenFrame CLI itself — the Go-based command-line tool that provisions Kubernetes clusters and deploys the OpenFrame platform.

> If you're looking to **use** OpenFrame CLI rather than develop it, start with the [Getting Started](../getting-started/introduction.md) documentation instead.

## Contents

| Section | Description |
|---|---|
| [Environment Setup](setup/environment.md) | Toolchain, editor setup, and environment variables for development |
| [Local Development](setup/local-development.md) | Cloning the repo, building, running, and debugging locally |
| [Architecture](architecture/README.md) | High-level architecture, core components, and data flow |
| [Security](security/README.md) | Security patterns, secret handling, and secure-by-default practices |
| [Testing](testing/README.md) | Test structure, running tests, and coverage expectations |
| [Contributing Guidelines](contributing/guidelines.md) | Code style, branching, commit conventions, and review checklist |

## Quick Navigation

- New to the codebase? Start with [Architecture](architecture/README.md) to understand how `cmd/`, `internal/cluster`, `internal/chart`, and `internal/shared` fit together.
- Setting up your machine? See [Environment Setup](setup/environment.md) and [Local Development](setup/local-development.md).
- Writing a change? Check [Testing](testing/README.md) and the [Contributing Guidelines](contributing/guidelines.md) before opening a PR.
- Touching credentials, downloads, or self-update code? Read [Security](security/README.md) first.

## Project at a Glance

OpenFrame CLI is a Go service (module `github.com/flamingo-stack/openframe-cli`) built with [Cobra](https://github.com/spf13/cobra) for command routing and `client-go` for native Kubernetes API access. It shells out to Docker, k3d, Helm, Terraform, gcloud, and the AWS CLI via a testable `CommandExecutor` abstraction, and renders its terminal UI with `pterm`, `huh`, and `bubbletea`.
