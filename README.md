# OpenFrame CLI

A modern CLI tool for managing OpenFrame Kubernetes clusters and development workflows.

## Installation

### From Release

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
```

## Quick Start

```bash
# Create a cluster
openframe cluster create

# List clusters
openframe cluster list

# Check cluster status
openframe cluster status

# Bootstrap OpenFrame on cluster
openframe bootstrap --deployment-mode=oss-tenant

# Get help
openframe --help
```

## Features

- 🎯 Interactive cluster creation with guided wizard
- ⚡ K3d cluster management for local development
- 📊 Real-time cluster status and monitoring
- 🔧 Smart system detection and configuration
- 🛠 Developer-friendly commands and clear output
- 📦 Chart installation and ArgoCD management
- 🚀 Development workflow tools (Skaffold, Telepresence)

## Documentation

For detailed documentation, see the [OpenFrame documentation](https://github.com/flamingo-stack/openframe-oss-tenant/tree/main/docs).

## Commands

### Cluster Management

- `openframe cluster create` - Create a new K3d cluster
- `openframe cluster list` - List all clusters
- `openframe cluster status` - Show cluster details
- `openframe cluster delete` - Delete a cluster
- `openframe cluster start` - Start a stopped cluster
- `openframe cluster cleanup` - Clean up cluster resources

### Chart Management

- `openframe chart install` - Install Helm charts and ArgoCD
- `openframe bootstrap` - Bootstrap full OpenFrame installation

### Development

- `openframe dev scaffold` - Run Skaffold for service development
- `openframe dev intercept` - Intercept service traffic with Telepresence

## Contributing

Contributions are welcome! Please see the [contributing guidelines](https://github.com/flamingo-stack/openframe-oss-tenant/blob/main/CONTRIBUTING.md).

## License

This project is licensed under the MIT License - see the LICENSE file for details.
