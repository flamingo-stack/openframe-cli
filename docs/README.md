# OpenFrame CLI Documentation

Welcome to the comprehensive documentation for OpenFrame CLI - a modern command-line tool for managing OpenFrame Kubernetes clusters and development workflows.

## 📚 Table of Contents

### Getting Started
Start here if you're new to OpenFrame CLI:

- [Introduction](./getting-started/introduction.md) - What is OpenFrame CLI and how it fits into the OpenFrame ecosystem
- [Prerequisites](./getting-started/prerequisites.md) - System requirements and tool dependencies
- [Quick Start](./getting-started/quick-start.md) - Get running with OpenFrame CLI in 5 minutes
- [First Steps](./getting-started/first-steps.md) - Your first cluster, chart installation, and development workflow

### Development
For contributors and developers working with OpenFrame CLI:

- [Development Overview](./development/README.md) - Development section index and overview
- [Environment Setup](./development/setup/environment.md) - Set up your development environment
- [Local Development](./development/setup/local-development.md) - Run and test OpenFrame CLI locally
- [Architecture Overview](./development/architecture/overview.md) - Technical architecture and component design
- [Testing Guide](./development/testing/overview.md) - Unit testing, integration testing, and test strategies
- [Contributing Guidelines](./development/contributing/guidelines.md) - How to contribute code, documentation, and bug reports

### Reference
Technical reference documentation:

- [Architecture Overview](./reference/architecture/overview.md) - Detailed technical architecture, component relationships, and data flows

### Diagrams
Visual documentation and architecture diagrams:

- [Architecture Diagrams](./diagrams/) - Mermaid diagrams showing system architecture and workflows

## 🚀 Quick Navigation

### Common Tasks
- **Create your first cluster**: [Quick Start Guide](./getting-started/quick-start.md#create-your-first-cluster)
- **Install OpenFrame charts**: [Chart Installation](./getting-started/first-steps.md#chart-installation)
- **Set up development workflow**: [Development Setup](./getting-started/first-steps.md#development-workflow)
- **Contribute to the project**: [Contributing Guidelines](./development/contributing/guidelines.md)

### Key Concepts
- **Bootstrap Workflow**: Complete cluster setup and chart installation in one command
- **K3D Integration**: Lightweight Kubernetes clusters for local development
- **ArgoCD Management**: GitOps-based application deployment and management
- **Development Tools**: Telepresence intercepts and Skaffold workflows

## 📖 Quick Links

- [Project README](../README.md) - Main project overview and installation
- [Contributing Guide](../CONTRIBUTING.md) - How to contribute to OpenFrame CLI
- [License Information](../LICENSE.md) - Flamingo AI Unified License v1.0

## 🏗 Architecture at a Glance

OpenFrame CLI is built with a modular architecture:

```
┌─────────────────┐
│   CLI Commands  │  ← Cobra-based command interface
└─────────┬───────┘
          │
┌─────────▼───────┐
│  Service Layer  │  ← Business logic (Cluster, Chart, Dev services)
└─────────┬───────┘
          │
┌─────────▼───────┐
│ Provider Layer  │  ← External tool integrations (K3D, Helm, ArgoCD)
└─────────────────┘
```

**Key Technologies:**
- **Go + Cobra**: CLI framework and command structure
- **K3D**: Lightweight Kubernetes clusters
- **ArgoCD**: GitOps application deployment
- **Telepresence**: Service intercepts for local development
- **pterm**: Rich terminal UI and interactive prompts

## 🎯 Use Cases

### For Platform Engineers
- Standardize local development environments
- Automate cluster provisioning and configuration
- Implement GitOps workflows with ArgoCD
- Provide self-service developer tools

### For Application Developers
- Quickly spin up development clusters
- Intercept service traffic for local debugging
- Run continuous development workflows with Skaffold
- Test changes against realistic environments

### For DevOps Teams
- Streamline onboarding processes
- Ensure consistent tooling across teams
- Reduce manual setup and configuration
- Enable rapid prototyping and testing

## 🔧 System Requirements

### Minimum Requirements
- **RAM**: 24GB
- **CPU**: 6 cores
- **Disk**: 50GB available space
- **OS**: macOS, Linux, or Windows with WSL2

### Recommended Requirements
- **RAM**: 32GB
- **CPU**: 12 cores
- **Disk**: 100GB available space
- **Network**: Stable internet connection for image pulls

### Tool Dependencies
- **Docker**: Container runtime for K3D clusters
- **kubectl**: Kubernetes command-line tool (auto-installed)
- **helm**: Package manager for Kubernetes (auto-installed)
- **jq**: JSON processor for parsing outputs (auto-installed)

## 📊 Command Reference

### Core Commands Overview

| Command Category | Purpose | Key Commands |
|-----------------|---------|--------------|
| **Bootstrap** | Complete setup | `openframe bootstrap` |
| **Cluster** | K3D cluster management | `create`, `list`, `status`, `delete` |
| **Chart** | ArgoCD and Helm operations | `install`, `upgrade`, `status` |
| **Development** | Developer workflows | `intercept`, `scaffold` |

### Example Workflows

**Complete Setup:**
```bash
openframe bootstrap --deployment-mode=oss-tenant
```

**Manual Step-by-Step:**
```bash
openframe cluster create my-cluster
openframe chart install my-cluster
openframe dev intercept api-service --port 8080
```

**Development Cycle:**
```bash
openframe dev scaffold --cluster my-cluster
# Make code changes
openframe dev intercept my-service --port 3000
# Test changes locally
```

## 📝 Documentation Conventions

### File Organization
- **getting-started/**: New user onboarding and tutorials
- **development/**: Contributor and developer information
- **reference/**: Technical specifications and architecture
- **diagrams/**: Visual documentation and Mermaid diagrams

### Markdown Standards
- Use descriptive headings with emoji prefixes
- Include code examples with proper language highlighting
- Provide cross-references to related documentation
- Use tables for structured information
- Include diagrams for complex concepts

## 🆘 Need Help?

### Documentation Issues
If you find errors or gaps in the documentation:
1. Check existing [GitHub Issues](https://github.com/flamingo-stack/openframe-cli/issues)
2. Create a new issue with the `documentation` label
3. Submit a pull request with fixes (see [Contributing](../CONTRIBUTING.md))

### Usage Questions
For usage questions and community support:
1. Check the [Getting Started](./getting-started/) guides first
2. Search existing [GitHub Discussions](https://github.com/flamingo-stack/openframe-cli/discussions)
3. Open a new discussion with your question

### Bug Reports
For bugs and feature requests:
1. Review the [Issue Templates](https://github.com/flamingo-stack/openframe-cli/issues/new/choose)
2. Provide detailed reproduction steps
3. Include environment information and logs

---

*Documentation generated by [OpenFrame Doc Orchestrator](https://github.com/flamingo-stack/openframe-oss-tenant)*