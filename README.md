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

A modern command-line tool for managing OpenFrame Kubernetes clusters and development workflows. OpenFrame CLI provides interactive wizards for cluster creation, chart installation, and development tools like Telepresence intercepts and Skaffold deployment workflows.

## ✨ Features

- **🎯 Interactive Cluster Creation** - Guided wizard for creating K3d clusters with smart defaults
- **⚡ One-Command Bootstrap** - Complete environment setup with `openframe bootstrap`
- **📊 Real-time Status Monitoring** - Live cluster health and resource monitoring
- **🔧 Smart System Detection** - Automatic prerequisite checking and installation
- **🛠 Developer-Friendly Commands** - Intuitive CLI with clear output and helpful error messages
- **📦 GitOps Ready** - Built-in ArgoCD setup with app-of-apps pattern
- **🚀 Development Workflow Tools** - Integrated Skaffold and Telepresence support
- **🌐 Cross-Platform Support** - Windows, macOS, and Linux compatibility

## 🚀 Quick Start

Get your first OpenFrame cluster running in under 5 minutes:

```bash
# Install OpenFrame CLI
curl -sSL https://install.openframe.dev | bash

# Create cluster with ArgoCD in one command
openframe bootstrap

# Verify everything works
kubectl get pods -A
```

That's it! Your cluster is ready for GitOps workflows and development.

## 📺 Demo

See OpenFrame CLI in action:

[![OpenFrame v0.3.4 - Enhanced Stability & Cross-Platform Support](https://img.youtube.com/vi/h9ZxyeYTBPE/maxresdefault.jpg)](https://www.youtube.com/watch?v=h9ZxyeYTBPE)

## 🔧 Installation

### Automated Installation (Recommended)

```bash
curl -sSL https://install.openframe.dev | bash
```

### Manual Installation

Download the latest release for your platform:

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

# Windows (AMD64)
# Download from: https://github.com/flamingo-stack/openframe-cli/releases/latest
```

### From Source

```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
go build -o openframe .
sudo mv openframe /usr/local/bin/
```

## 🎮 Core Commands

### Bootstrap Complete Environment
```bash
# Interactive setup with wizard
openframe bootstrap

# Non-interactive with specific options
openframe bootstrap my-cluster --deployment-mode=oss-tenant --non-interactive
```

### Cluster Management
```bash
# Create a new cluster
openframe cluster create

# List all clusters
openframe cluster list

# Check cluster status
openframe cluster status

# Delete a cluster
openframe cluster delete my-cluster

# Clean up cluster resources
openframe cluster cleanup
```

### Chart and GitOps Management
```bash
# Install ArgoCD and app-of-apps pattern
openframe chart install

# Install on specific cluster
openframe chart install my-cluster --deployment-mode=oss-tenant
```

### Development Workflows
```bash
# Intercept service traffic with Telepresence
openframe dev intercept my-service --port 8080

# Run Skaffold development workflow
openframe dev scaffold my-cluster
```

## 🏗️ Architecture

OpenFrame CLI is built with a modular architecture supporting:

- **Cluster Providers**: K3d integration with Docker
- **Chart Management**: Helm and ArgoCD orchestration
- **Development Tools**: Telepresence and Skaffold integration
- **Prerequisites**: Automatic tool detection and installation
- **Interactive UI**: Rich terminal experiences with progress tracking

## 🎯 Use Cases

### Local Development
- Spin up lightweight Kubernetes clusters instantly
- Test applications in realistic environments
- Develop with hot-reload using Skaffold

### GitOps Workflows
- Bootstrap ArgoCD with best practices
- Implement app-of-apps pattern out of the box
- Manage multiple environments consistently

### Team Onboarding
- Standardize development environments
- Reduce setup time from hours to minutes
- Provide consistent tooling across the team

### CI/CD Integration
- Create ephemeral test environments
- Validate deployments in realistic clusters
- Integrate with existing pipelines

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details on:

- Setting up your development environment
- Coding standards and best practices
- Submitting pull requests
- Code review process

## 📚 Documentation

📚 See the [Documentation](./docs/README.md) for comprehensive guides including:

- **Getting Started**: Installation, prerequisites, and first steps
- **Development**: Architecture, local development, and contribution guidelines
- **Reference**: Complete command reference and troubleshooting

## 📄 License

This project is licensed under the Flamingo AI Unified License v1.0. See [LICENSE.md](LICENSE.md) for details.

## 🆘 Support

- **📖 Documentation**: Check our [docs](./docs/README.md) for detailed guides
- **🐛 Bug Reports**: Open an issue on [GitHub Issues](https://github.com/flamingo-stack/openframe-cli/issues)
- **💬 Discussions**: Join our [GitHub Discussions](https://github.com/flamingo-stack/openframe-cli/discussions)
- **📧 Contact**: Reach out to the team at [support@flamingo.run](mailto:support@flamingo.run)

---
<div align="center">
  Built with 💛 by the <a href="https://www.flamingo.run/about"><b>Flamingo</b></a> team
</div>