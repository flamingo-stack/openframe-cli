<div align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/flamingo-stack/openframe-oss-tenant/main/docs/assets/logo-openframe-full-dark-bg.png">
    <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/flamingo-stack/openframe-oss-tenant/main/docs/assets/logo-openframe-full-light-bg.png">
    <img alt="OpenFrame Logo" src="https://raw.githubusercontent.com/flamingo-stack/openframe-oss-tenant/main/docs/assets/logo-openframe-full-light-bg.png" width="400">
  </picture>
</div>

<p align="center">
  <a href="LICENSE.md"><img alt="License" src="https://img.shields.io/badge/LICENSE-FLAMINGO%20AI%20Unified%20v1.0-%23FFC109?style=for-the-badge&labelColor=white"></a>
</p>

# OpenFrame CLI

A modern command-line tool for managing OpenFrame Kubernetes clusters and development workflows. OpenFrame CLI simplifies local Kubernetes development by providing cluster lifecycle management, GitOps deployment with ArgoCD, and developer-focused tools for traffic interception and live reloading.

## Features

- 🎯 **Interactive Cluster Creation** - Guided wizard for setting up K3d clusters with best practices
- ⚡ **Lightning Fast Setup** - Bootstrap complete environments in minutes with ArgoCD and OpenFrame apps
- 📊 **Real-time Monitoring** - Cluster status and health monitoring with detailed diagnostics
- 🔧 **Smart Configuration** - Automatic system detection and intelligent defaults
- 🛠 **Developer Tools** - Traffic interception with Telepresence and live reloading with Skaffold
- 📦 **GitOps Ready** - Built-in ArgoCD installation and app-of-apps deployment pattern
- 🚀 **Production-grade** - Support for multiple deployment modes (OSS tenant, SaaS tenant, SaaS shared)

## Quick Start

### Prerequisites

Make sure you have Docker, kubectl, Helm, and k3d installed. OpenFrame CLI will check and guide you through installation if any are missing.

### Installation

**Download the latest release:**

```bash
# macOS (ARM64)
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_arm64.tar.gz | tar xz
sudo mv openframe /usr/local/bin/

# Linux (AMD64)
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_linux_amd64.tar.gz | tar xz
sudo mv openframe /usr/local/bin/

# Verify installation
openframe --help
```

**Or build from source:**

```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
go build -o openframe .
sudo mv openframe /usr/local/bin/
```

### Bootstrap Your First Environment

Get up and running in minutes with a complete Kubernetes environment:

```bash
# Interactive setup (recommended for first-time users)
openframe bootstrap

# Quick non-interactive setup
openframe bootstrap my-dev-cluster --deployment-mode=oss-tenant

# Check your cluster status
openframe cluster status
```

### Essential Commands

```bash
# Cluster Management
openframe cluster create my-cluster    # Create new cluster
openframe cluster list                 # List all clusters
openframe cluster status my-cluster    # Check cluster health
openframe cluster delete my-cluster    # Remove cluster

# Development Workflow
openframe dev intercept my-service     # Traffic interception for local dev
openframe dev skaffold my-service      # Live reloading with Skaffold

# Chart Management
openframe chart install               # Install ArgoCD and OpenFrame apps
```

## Architecture

OpenFrame CLI follows a modular architecture with clear separation of concerns:

- **Cluster Management** - K3d cluster lifecycle operations with intelligent health monitoring
- **Chart Installation** - ArgoCD deployment and Helm chart management with app-of-apps pattern
- **Development Tools** - Traffic interception and continuous development workflows
- **Bootstrap Orchestration** - End-to-end environment provisioning combining all components

## Documentation

- **[Getting Started Guide](docs/tutorials/user/getting-started.md)** - Complete setup and installation guide
- **[Common Use Cases](docs/tutorials/user/common-use-cases.md)** - Real-world scenarios and workflows
- **[Architecture Overview](docs/dev/overview.md)** - Technical architecture and component details
- **[Development Guide](docs/tutorials/dev/getting-started-dev.md)** - Contributing and local development setup

## Development Workflows

### Local Service Development

```bash
# Set up traffic interception
openframe dev intercept my-service

# Your local code now receives cluster traffic
npm start  # or go run main.go, python app.py, etc.
```

### Continuous Development

```bash
# Automatic rebuild and redeploy on code changes
openframe dev skaffold my-service
```

### Multi-Environment Management

```bash
# Create project-specific environments
openframe cluster create frontend-dev
openframe cluster create backend-dev
openframe cluster create integration-test

# Switch between environments easily
openframe cluster status frontend-dev
kubectl config use-context k3d-frontend-dev
```

## Supported Deployment Modes

- **OSS Tenant** - Open source multi-tenant deployment
- **SaaS Tenant** - Software-as-a-Service tenant isolation
- **SaaS Shared** - Shared SaaS infrastructure deployment

## Contributing

We welcome contributions! Please read our [Contributing Guidelines](CONTRIBUTING.md) for details on our code of conduct, development process, and how to submit pull requests.

## License

This project is licensed under the Flamingo AI Unified License v1.0. See the [LICENSE](LICENSE.md) file for details.

---
<div align="center">
  Built with 💛 by the <a href="https://www.flamingo.run/about"><b>Flamingo</b></a> team
</div>