# OpenFrame CLI Documentation

Welcome to the comprehensive documentation for OpenFrame CLI - a modern, interactive command-line tool for managing OpenFrame Kubernetes clusters and development workflows.

## 📚 Table of Contents

### Getting Started

New to OpenFrame CLI? Start here to get up and running quickly:

- [Introduction](./getting-started/introduction.md) - Overview of OpenFrame CLI and key concepts
- [Prerequisites](./getting-started/prerequisites.md) - System requirements and dependency installation
- [Quick Start](./getting-started/quick-start.md) - Get OpenFrame running in 5 minutes
- [First Steps](./getting-started/first-steps.md) - Explore key features and workflows

### Development

Resources for developers working on or with OpenFrame CLI:

- [Development Overview](./development/README.md) - Developer documentation index
- [Environment Setup](./development/setup/environment.md) - IDE configuration, tools, and development environment
- [Local Development](./development/setup/local-development.md) - Clone, build, and run OpenFrame CLI locally
- [Architecture Overview](./development/architecture/README.md) - System design and component relationships

### Reference

Technical reference documentation generated from source code analysis:

- [Architecture Documentation](./architecture/overview.md) - Comprehensive system architecture and component design

### Diagrams

Visual documentation to understand system architecture and workflows:

- [Architecture Diagrams](./diagrams/architecture/README.md) - Mermaid diagrams showing system design, data flows, and component relationships

### CLI Tools

The OpenFrame Main code Repository is maintained in a separate repository:
- **Repository**: [flamingo-stack/openframe-oss-tenant](https://github.com/flamingo-stack/openframe-oss-tenant)
- **Documentation**: [CLI Documentation](https://github.com/flamingo-stack/openframe-oss-tenant/tree/main/docs)

**Note**: CLI tools are NOT located in this repository. Always refer to the external repository for installation and usage.

## 🚀 Quick Navigation

### For New Users
1. Check [Prerequisites](./getting-started/prerequisites.md)
2. Follow [Quick Start](./getting-started/quick-start.md)
3. Explore [First Steps](./getting-started/first-steps.md)

### For Developers
1. Set up [Development Environment](./development/setup/environment.md)
2. Review [Architecture Documentation](./architecture/overview.md)
3. Check [Contributing Guidelines](../CONTRIBUTING.md)

### For System Administrators
1. Review [Prerequisites](./getting-started/prerequisites.md) for system requirements
2. Follow [Bootstrap Guide](./getting-started/quick-start.md#step-3-bootstrap-your-first-environment)
3. Explore cluster management commands in [First Steps](./getting-started/first-steps.md)

## 🎯 Key Features Covered

This documentation covers all major OpenFrame CLI capabilities:

- **Complete Environment Bootstrapping**: One-command setup with `openframe bootstrap`
- **Cluster Management**: Create, delete, and monitor K3D Kubernetes clusters
- **Chart & Application Management**: Helm charts and ArgoCD GitOps workflows
- **Development Tools**: Service intercepts, scaffolding, and local development workflows
- **Multi-mode Deployment**: OSS tenant, SaaS tenant, and SaaS shared configurations

## 📖 Quick Links

- [Project README](../README.md) - Main project overview and installation
- [Contributing](../CONTRIBUTING.md) - How to contribute to OpenFrame CLI
- [License](../LICENSE.md) - Project license information

## 🏗️ System Requirements

Before using OpenFrame CLI, ensure your system meets these requirements:

| Resource | Minimum | Recommended |
|----------|---------|-------------|
| **RAM** | 24GB | 32GB |
| **CPU Cores** | 6 cores | 12 cores |
| **Disk Space** | 50GB free | 100GB free |

## 🛠️ Core Dependencies

Required tools that must be installed:
- [Docker](https://docs.docker.com/get-docker/) 20.10+
- [kubectl](https://kubernetes.io/docs/tasks/tools/) 1.25+
- [Helm](https://helm.sh/docs/intro/install/) 3.10+
- [K3D](https://k3d.io/v5.4.6/#installation) 5.0+

## 🌟 What Makes OpenFrame CLI Different

- **One-Command Bootstrap**: Complete environment setup with single command
- **Developer-Friendly**: Interactive prompts, clear error messages, rich terminal UI
- **GitOps Native**: Built-in ArgoCD integration for modern deployment practices
- **Local Development**: Telepresence service intercepts for debugging
- **Multi-Platform**: Linux, macOS, and Windows (WSL2) support
- **Open Source**: Complete transparency and community-driven development

## 🤝 Community and Support

Need help or want to contribute?

- **Primary Support**: [OpenMSP Slack](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)
- **Website**: [https://flamingo.run](https://flamingo.run)
- **OpenFrame Platform**: [https://openframe.ai](https://openframe.ai)

> **Note**: We don't monitor GitHub Issues for support. All community support and discussion happens in our Slack workspace.

## 📝 Documentation Maintenance

This documentation is automatically generated and maintained by the OpenFrame development team. If you find errors or have suggestions for improvement, please reach out in the OpenMSP Slack community.

---
*Documentation generated by [OpenFrame Doc Orchestrator](https://github.com/flamingo-stack/openframe-oss-tenant)*