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

A modern CLI tool for managing OpenFrame Kubernetes clusters and development workflows. OpenFrame CLI provides developers with a streamlined experience for creating, managing, and deploying applications to local K3d clusters with integrated ArgoCD GitOps workflows.

## ✨ Features

- 🎯 **Interactive Cluster Creation** - Guided wizard with smart defaults for quick setup
- ⚡ **K3d Cluster Management** - Lightweight Kubernetes clusters for local development
- 📊 **Real-time Monitoring** - Cluster status, health checks, and resource usage
- 🔧 **Smart System Detection** - Automatic prerequisite checking and installation guidance
- 🛠 **Developer-Friendly** - Clear output, helpful error messages, and verbose logging
- 📦 **GitOps Ready** - Integrated ArgoCD installation and app-of-apps pattern
- 🚀 **Development Tools** - Built-in support for Skaffold and Telepresence workflows
- ♻️ **Resource Management** - Intelligent cleanup and resource optimization

## 🚀 Quick Start

### One-Command Setup

Get a complete OpenFrame environment running in minutes:

```bash
# Download and install OpenFrame CLI
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_arm64.tar.gz | tar xz
sudo mv openframe /usr/local/bin/

# Bootstrap your first environment
openframe bootstrap

# That's it! 🎉
```

### Step-by-Step Installation

#### 1. Install Prerequisites

OpenFrame CLI requires these tools (it will check and guide you through installation):

- **Docker** (20.10+) - Container runtime
- **kubectl** (1.20+) - Kubernetes CLI
- **Helm** (3.0+) - Package manager
- **K3d** (5.0+) - Lightweight Kubernetes

#### 2. Download OpenFrame CLI

Choose your platform:

```bash
# macOS (ARM64)
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_arm64.tar.gz | tar xz
sudo mv openframe /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_amd64.tar.gz | tar xz
sudo mv openframe /usr/local/bin/

# Linux (AMD64)
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_linux_amd64.tar.gz | tar xz
sudo mv openframe /usr/local/bin/

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_windows_amd64.zip" -OutFile "openframe.zip"
Expand-Archive -Path "openframe.zip" -DestinationPath "."
```

#### 3. Create Your First Cluster

```bash
# Interactive setup with guided configuration
openframe bootstrap

# Or create just a cluster
openframe cluster create my-dev-cluster

# Check status
openframe cluster status
```

## 🛠 Core Commands

| Command | Description | Example |
|---------|-------------|---------|
| `openframe bootstrap` | Complete environment setup | `openframe bootstrap my-env` |
| `openframe cluster create` | Create a new K3d cluster | `openframe cluster create dev-cluster` |
| `openframe cluster list` | List all clusters | `openframe cluster list` |
| `openframe cluster status` | Show detailed cluster info | `openframe cluster status my-cluster` |
| `openframe cluster delete` | Remove a cluster | `openframe cluster delete old-cluster` |
| `openframe cluster cleanup` | Clean unused resources | `openframe cluster cleanup` |
| `openframe chart install` | Install ArgoCD and charts | `openframe chart install --deployment-mode=oss-tenant` |

### Advanced Workflows

```bash
# Development with live reload
openframe dev scaffold

# Traffic interception for debugging
openframe dev intercept my-service

# Automated setup for CI/CD
openframe bootstrap --non-interactive --deployment-mode=oss-tenant

# Verbose logging for troubleshooting
openframe cluster create debug-env --verbose
```

## 🏗 Architecture

OpenFrame CLI follows a layered architecture with clear separation of concerns:

- **Command Layer** - CLI interface and user interaction
- **Service Layer** - Business logic and orchestration
- **Provider Layer** - External tool integrations (K3d, Helm, ArgoCD)

Key integrations:
- **K3d** - Local Kubernetes clusters in Docker
- **ArgoCD** - GitOps continuous deployment
- **Helm** - Kubernetes package management
- **Telepresence** - Local development against remote clusters
- **Skaffold** - Continuous development workflows

## 📚 Documentation

| Resource | Description |
|----------|-------------|
| [Getting Started Guide](docs/tutorials/user/getting-started.md) | Complete setup walkthrough |
| [Common Use Cases](docs/tutorials/user/common-use-cases.md) | Daily workflows and patterns |
| [Architecture Overview](docs/dev/overview.md) | Technical deep dive |
| [Developer Guide](docs/tutorials/dev/getting-started-dev.md) | Contributing to the project |

## 🔧 Configuration Options

### Bootstrap Modes

- **OSS Tenant** (`oss-tenant`) - Open source deployment
- **SaaS Tenant** (`saas-tenant`) - Single tenant SaaS
- **SaaS Shared** (`saas-shared`) - Multi-tenant SaaS

### Interactive vs. Non-Interactive

```bash
# Interactive mode (default) - guided prompts
openframe bootstrap

# Non-interactive mode - for automation
openframe bootstrap \
  --deployment-mode=oss-tenant \
  --non-interactive \
  --cluster-name=ci-cluster
```

## 🚨 Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| Docker not running | `sudo systemctl start docker` or start Docker Desktop |
| Port conflicts | Stop conflicting services or use different ports |
| Permission denied | Add user to docker group: `sudo usermod -aG docker $USER` |
| Insufficient memory | Ensure at least 4GB RAM available for Docker |

### Debug Commands

```bash
# Check prerequisites
openframe bootstrap --help

# Verbose output
openframe cluster create test --verbose

# Clean slate
openframe cluster cleanup
```

## 🤝 Contributing

We welcome contributions! Whether you're fixing bugs, adding features, or improving documentation, please see our [Contributing Guidelines](CONTRIBUTING.md).

### Quick Development Setup

```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
go mod download
make build
```

## 📄 License

This project is licensed under the Flamingo AI Unified License v1.0. See [LICENSE.md](LICENSE.md) for details.

---

<div align="center">
  Built with 💛 by the <a href="https://www.flamingo.run/about"><b>Flamingo</b></a> team
</div>