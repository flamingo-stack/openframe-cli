## Overview

This section is the entry point for contributing to **OpenFrame CLI**, the Go-based command-line tool that provisions Kubernetes clusters (local k3d, or cloud GKE/EKS via Terraform) and deploys the OpenFrame platform onto them through ArgoCD's app-of-apps pattern.

OpenFrame CLI is published to the Go ecosystem as `github.com/flamingo-stack/openframe-cli`. Per the repository graph, it is a leaf/terminal artifact: it has no recorded upstream dependencies on other repositories in this organization, and no repository is currently recorded as consuming it. Its functional dependencies are third-party Go modules (Cobra, client-go, pterm, huh, bubbletea, sigstore-go) and external CLI tools it shells out to (Docker, k3d, Helm, Terraform, gcloud, aws), not sibling repositories.

> **Note on scope:** This document is the development hub for the `openframe-cli` repository itself. The broader OpenFrame platform that this CLI installs onto a cluster lives in the separate **flamingo-stack/openframe-oss-tenant** repository. If you are looking to develop or modify the platform charts/services deployed by `openframe app install`, refer to that repository directly rather than assuming its code lives in a subdirectory here.

Use this page to find the right doc for the task in front of you — setting up a local environment, understanding the architecture, running tests, following contribution guidelines, or reviewing security practices.

## Development documentation

| Document | What it covers |
|---|---|
| [Environment Setup](./setup/environment.md) | Preparing your machine (Go toolchain, required external tools) before building the CLI |
| [Local Development](./setup/local-development.md) | Building, running, and iterating on the CLI locally |
| [Architecture](./architecture/README.md) | How `cmd/`, `internal/cluster`, `internal/chart`, `internal/app`, and shared packages fit together |
| [Testing](./testing/README.md) | How the test suite under `tests/` and `tests/testutil/` is organized and run |
| [Contributing Guidelines](./contributing/guidelines.md) | Conventions and process for submitting changes |
| [Security](./security/README.md) | Security-relevant practices in the codebase (signature verification, credential handling, etc.) |

## Quick navigation guide

Use this guide to jump straight to what you need:

- **New to the project?** Start with [Introduction](../getting-started/introduction.md), check [Prerequisites](../getting-started/prerequisites.md), then follow the [Quick Start](../getting-started/quick-start.md) and [First Steps](../getting-started/first-steps.md).
- **Setting up your machine to build the CLI?** Go to [Environment Setup](./setup/environment.md).
- **Ready to build and run the code?** Go to [Local Development](./setup/local-development.md).
- **Want to understand how the codebase is structured?** Read [Architecture](./architecture/README.md), and for the full module-level breakdown see [Architecture Overview](../reference/architecture/overview.md) and the [diagrams index](../diagrams/architecture/README.md).
- **Need to see how this repository relates to the rest of the organization?** See [Ecosystem](../reference/architecture/ecosystem.md).
- **Writing or running tests?** Go to [Testing](./testing/README.md).
- **About to open a change?** Review [Contributing Guidelines](./contributing/guidelines.md).
- **Handling credentials, signatures, or sensitive data?** Review [Security](./security/README.md).
- **Trying the CLI end-to-end against a real cluster?** The getting-started how-to guides walk through the full lifecycle: [create a cluster](../getting-started/how-to/create-a-kubernetes-cluster-with-openframe-cli.md), [install the platform](../getting-started/how-to/install-the-openframe-platform-onto-a-cluster.md), [check status](../getting-started/how-to/check-the-status-of-your-openframe-cluster-and-app.md), [update the CLI](../getting-started/how-to/update-the-openframe-cli-to-the-latest-version.md), and [uninstall / delete a cluster](../getting-started/how-to/uninstall-openframe-and-delete-a-cluster.md).

## Repository layout at a glance

The codebase is organized around three core abstractions — **cluster** provisioning, **app** deployment via ArgoCD, and **prerequisites** verification — plus shared infrastructure:

```text
cmd/              Cobra command definitions (bootstrap, cluster, app, prerequisites, update)
internal/cluster/ Cluster lifecycle: service.go, provider/ (k3d, EKS, GKE), discovery/, prerequisites/
internal/chart/   ArgoCD/Helm/Git providers and chart install orchestration (services/)
internal/app/     Platform status aggregation (status/) and uninstall (uninstall/)
internal/k8s/     Kubeconfig context resolution and cluster health/resource access
internal/prerequisites/ OS-aware check/install runner for required tools
internal/shared/  executor, errors, ui, download, selfupdate, wsllauncher
tests/            Integration tests and shared test utilities (tests/testutil/)
main.go           Entry point
go.mod            Module definition: github.com/flamingo-stack/openframe-cli, Go 1.26.0 (toolchain go1.26.5)
```

For the full component-by-component breakdown, dependency diagrams, and sequence diagrams (bootstrap flow, app status aggregation), see [Architecture](./architecture/README.md) and [Architecture Overview](../reference/architecture/overview.md).

## Working agreement

Before opening a pull request:

1. Confirm your change builds and passes the suite described in [Testing](./testing/README.md).
2. Follow the conventions in [Contributing Guidelines](./contributing/guidelines.md).
3. For anything touching credential handling, binary verification, or self-update integrity, check [Security](./security/README.md) first.

This project does not use GitHub Issues or GitHub Discussions for coordination — day-to-day discussion happens on the OpenMSP Slack community.
