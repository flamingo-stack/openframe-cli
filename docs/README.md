
# Project Documentation

Welcome to the documentation for **OpenFrame CLI**.

OpenFrame CLI is the command-line tool for provisioning Kubernetes clusters (local k3d, or cloud GKE/EKS via Terraform) and deploying the [OpenFrame](https://openframe.ai) platform onto them.

## 📚 Table of Contents

### Getting Started

- [Introduction](./getting-started/introduction.md) - OpenFrame CLI is the command-line tool for provisioning Kubernetes clusters and deploying the OpenFrame platform onto them
- [Prerequisites](./getting-started/prerequisites.md) - What your machine needs before installing and running `openframe`
- [Quick Start](./getting-started/quick-start.md) - Provision a Kubernetes cluster (local k3d, or cloud GKE/EKS via Terraform) and deploy the OpenFrame platform onto it
- [First Steps](./getting-started/first-steps.md) - The first things to do right after installing the OpenFrame CLI
- [Cloud Clusters (EKS / GKE)](./getting-started/cloud-clusters.md) - Provisioning managed clusters beyond local k3d
- [GKE Workflow — from zero to a running cluster](./getting-started/gke-workflow.md) - The complete, in-order flow for working with Google Kubernetes Engine

**How-to guides:**
- [How to check the status of your OpenFrame cluster and app](./getting-started/how-to/check-the-status-of-your-openframe-cluster-and-app.md) - Confirm the Kubernetes cluster is reachable and that the OpenFrame app is healthy
- [How to create a Kubernetes cluster with OpenFrame CLI](./getting-started/how-to/create-a-kubernetes-cluster-with-openframe-cli.md) - Use the `openframe` command to provision a new Kubernetes cluster
- [How to install the OpenFrame platform onto a cluster](./getting-started/how-to/install-the-openframe-platform-onto-a-cluster.md) - Deploy the OpenFrame platform (ArgoCD plus the app-of-apps) onto a Kubernetes cluster
- [How to uninstall OpenFrame and delete a cluster](./getting-started/how-to/uninstall-openframe-and-delete-a-cluster.md) - Remove the OpenFrame platform from a cluster and, if needed, delete the cluster itself
- [How to update the OpenFrame CLI to the latest version](./getting-started/how-to/update-the-openframe-cli-to-the-latest-version.md) - Check for a newer release and update the installed CLI in place

### Development

- [Development Overview](./development/README.md) - Entry point for contributing to OpenFrame CLI, the Go-based command-line tool
- [Architecture Overview](./development/architecture/README.md) - OpenFrame CLI is a Go command-line tool, published as the `go` module `github.com/flamingo-stack/openframe-cli`
- [Contributing Guidelines](./development/contributing/guidelines.md) - Guidelines for contributing to the OpenFrame CLI
- [Release Signing](./development/release-signing.md) - Release binaries are code-signed during `goreleaser release`
- [Releasing](./development/releasing.md) - Versioning is fully automated
- [Security](./development/security/README.md) - OpenFrame CLI is a local operator tool that provisions Kubernetes clusters and deploys the OpenFrame platform
- [Environment Setup](./development/setup/environment.md) - Setting up a local development environment for the OpenFrame CLI
- [Local Development](./development/setup/local-development.md) - Cloning and building the OpenFrame CLI, running it against a local cluster, and configuring a debugger
- [Testing](./development/testing/README.md) - OpenFrame CLI's test suite, the primary safety net for a tool that provisions real Kubernetes infrastructure

### Reference

Reference docs are in `./reference/architecture/`:

- [Ecosystem: OpenFrame CLI](./reference/architecture/ecosystem.md) - OpenFrame CLI is a service (go), and how it fits the wider dependency graph
- [openframe-cli Module Documentation](./reference/architecture/overview.md) - OpenFrame CLI is a modern, interactive command-line tool for provisioning Kubernetes clusters and deploying the OpenFrame platform

### Diagrams

Visual documentation - View Mermaid diagrams in: `./diagrams/architecture/`

- [Diagrams Overview](./diagrams/architecture/README.md) - Index of the architecture diagrams below
- `./diagrams/architecture/app-status-aggregation.mmd` - App status aggregation flow
- `./diagrams/architecture/architecture-diagram.mmd` - Overall architecture diagram
- `./diagrams/architecture/bootstrap-sequence.mmd` - Bootstrap sequence diagram
- `./diagrams/architecture/dependency-diagram.mmd` - Dependency diagram

### CLI Tools

The OpenFrame Main code Repository is maintained in a separate repository:
- **Repository**: [flamingo-stack/openframe-oss-tenant](https://github.com/flamingo-stack/openframe-oss-tenant)
- **Documentation**: [CLI Documentation](https://github.com/flamingo-stack/openframe-oss-tenant/tree/main/docs)

**Note**: CLI tools (the main OpenFrame platform source) are NOT located in this repository. Always refer to the external repository above for the platform's own installation and usage details. This repository, `flamingo-stack/openframe-cli`, publishes the `openframe` command-line tool itself (Go module `github.com/flamingo-stack/openframe-cli`) that provisions clusters and deploys that platform.

### Other Documentation

Documentation coming soon.

## 📖 Quick Links
- [Project README](../README.md) - Main project README
- [Contributing](../CONTRIBUTING.md) - How to contribute
- [License](../LICENSE.md) - License information

---
*Documentation generated by [🦩 Flamingo Code Documentation](https://flamingo.run)*
