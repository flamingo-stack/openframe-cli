# OpenFrame CLI Documentation

Welcome to the comprehensive documentation for **OpenFrame CLI** - a command-line tool for managing Kubernetes clusters and deploying the OpenFrame platform with streamlined workflows for cluster creation, chart installation, and development tools.

## 📚 Table of Contents

### 🚀 Getting Started
Start here if you're new to OpenFrame CLI:

- **[Introduction](./getting-started/introduction.md)** - What is OpenFrame CLI and why use it?
- **[Prerequisites](./getting-started/prerequisites.md)** - Required tools and system requirements
- **[Quick Start](./getting-started/quick-start.md)** - Get running in 5 minutes with bootstrap command
- **[First Steps](./getting-started/first-steps.md)** - Explore key features and workflows after setup

### 🛠️ Development
For contributors and developers working on OpenFrame CLI:

- **[Development Overview](./development/README.md)** - Complete development section index and navigation
- **[Environment Setup](./development/setup/environment.md)** - Configure your IDE, tools, and development environment
- **[Local Development](./development/setup/local-development.md)** - Clone, build, run, and debug locally
- **[Architecture Overview](./development/architecture/overview.md)** - System architecture, components, and design patterns
- **[Testing Overview](./development/testing/overview.md)** - Test strategy, running tests, and writing new tests
- **[Contributing Guidelines](./development/contributing/guidelines.md)** - Code style, PR process, and review workflow

### 📖 Reference
Technical reference documentation and architecture details:

- **[Architecture Overview](./reference/architecture/overview.md)** - Comprehensive technical architecture documentation
- API References (coming soon)
- Configuration References (coming soon)
- Troubleshooting Guides (coming soon)

### 📊 Diagrams
Visual documentation and architectural diagrams:

- **[Architecture Diagrams](./diagrams/)** - Mermaid diagrams showing system structure and data flows
- Component Interaction Diagrams (coming soon)
- Deployment Flow Diagrams (coming soon)

## 🎯 Documentation Paths by Role

### 🆕 **New Users**
*"I want to try OpenFrame CLI for the first time"*

**Recommended path:**
1. [Introduction](./getting-started/introduction.md) - Understand what OpenFrame CLI does
2. [Prerequisites](./getting-started/prerequisites.md) - Install required tools
3. [Quick Start](./getting-started/quick-start.md) - Bootstrap your first environment
4. [First Steps](./getting-started/first-steps.md) - Learn essential workflows

### 🔧 **Platform Engineers**
*"I want to use OpenFrame CLI in my organization"*

**Recommended path:**
1. [Architecture Overview](./reference/architecture/overview.md) - Understand system design
2. [Quick Start](./getting-started/quick-start.md) - Get hands-on experience
3. [Development > Local Development](./development/setup/local-development.md) - Advanced usage and customization

### 💻 **Contributors**
*"I want to contribute code or documentation"*

**Recommended path:**
1. [Development Overview](./development/README.md) - Development section overview
2. [Environment Setup](./development/setup/environment.md) - Set up development environment
3. [Local Development](./development/setup/local-development.md) - Build and test locally
4. [Contributing Guidelines](./development/contributing/guidelines.md) - Learn the contribution workflow

### 🏗️ **Maintainers**
*"I maintain this project and need comprehensive documentation"*

**Key resources:**
- [Architecture Overview](./development/architecture/overview.md) - System boundaries and extension points
- [Testing Overview](./development/testing/overview.md) - Quality assurance and CI/CD
- [Contributing Guidelines](./development/contributing/guidelines.md) - Review standards and processes

## 🔍 Quick Reference

### Essential Commands
```bash
# Complete environment setup
openframe bootstrap my-cluster

# Cluster management
openframe cluster create my-cluster
openframe cluster list
openframe cluster status my-cluster
openframe cluster delete my-cluster

# Chart operations
openframe chart install my-cluster

# Development tools
openframe dev intercept api-service
openframe dev skaffold my-cluster
```

### Key Concepts

| Concept | Description | Documentation |
|---------|-------------|---------------|
| **Bootstrap** | One-command complete environment setup | [Quick Start](./getting-started/quick-start.md) |
| **Cluster Management** | K8s cluster lifecycle with K3d | [Architecture](./reference/architecture/overview.md#core-components) |
| **Chart Management** | Helm charts and ArgoCD deployment | [Architecture](./reference/architecture/overview.md#component-relationships) |
| **Development Tools** | Telepresence and Skaffold integration | [First Steps](./getting-started/first-steps.md) |
| **Deployment Modes** | OSS-tenant, SaaS-shared, SaaS-tenant | [Quick Start](./getting-started/quick-start.md#available-deployment-modes) |

### Project Structure
```
openframe-cli/
├── cmd/                    # CLI command implementations
│   ├── bootstrap/         # Complete environment setup
│   ├── cluster/          # Cluster management commands
│   ├── chart/            # Helm chart operations
│   └── dev/              # Development tools
├── internal/             # Private application code
│   ├── bootstrap/        # Bootstrap business logic
│   ├── cluster/          # Cluster management services
│   ├── chart/            # Chart installation services
│   ├── dev/              # Development tool integrations
│   └── shared/           # Common utilities and UI
├── docs/                 # Documentation (this directory)
├── examples/             # Usage examples and demos
└── scripts/              # Build and development scripts
```

## 🆘 Troubleshooting

### Common Issues

| Issue | Solution | Documentation |
|-------|----------|---------------|
| Prerequisites not met | Install Docker, kubectl, Helm, K3d | [Prerequisites](./getting-started/prerequisites.md) |
| Cluster creation fails | Check Docker status, clean up partial clusters | [Quick Start Troubleshooting](./getting-started/quick-start.md#troubleshooting-quick-start-issues) |
| Command not found | Verify PATH configuration | [Quick Start Installation](./getting-started/quick-start.md#step-1-install-openframe-cli) |
| Permission denied | Check binary permissions | [Environment Setup](./development/setup/environment.md) |
| Build failures | Verify Go version and dependencies | [Local Development](./development/setup/local-development.md) |

### Getting Help

1. **Search Documentation**: Use browser search (Ctrl/Cmd+F) on relevant pages
2. **Check Issues**: Browse [GitHub Issues](https://github.com/flamingo-stack/openframe-cli/issues) for known problems
3. **Community Support**: Join our [Discord community](https://discord.gg/flamingo) for help
4. **Report Bugs**: Create detailed issue reports with logs and system information

## 📖 Quick Navigation Links

### Main Project Resources
- **[Project README](../README.md)** - Main project overview and features
- **[Contributing Guide](../CONTRIBUTING.md)** - How to contribute to the project
- **[License](../LICENSE.md)** - Flamingo AI Unified License v1.0 details
- **[GitHub Repository](https://github.com/flamingo-stack/openframe-cli)** - Source code and issues

### External Resources
- **[OpenFrame Platform](https://openframe.io)** - Main OpenFrame platform documentation
- **[Flamingo](https://www.flamingo.run)** - Learn about the team behind OpenFrame
- **[Kubernetes](https://kubernetes.io/docs/)** - Kubernetes documentation
- **[Helm](https://helm.sh/docs/)** - Helm package manager documentation
- **[ArgoCD](https://argo-cd.readthedocs.io/)** - GitOps continuous delivery documentation

## 📝 Documentation Maintenance

This documentation is actively maintained and updated. If you find:

- **Outdated Information**: Please open an issue or submit a PR
- **Missing Examples**: We welcome contributions of working examples
- **Unclear Instructions**: Help us improve clarity with feedback
- **Broken Links**: Report broken internal or external links

### Contributing to Documentation

Documentation improvements are always welcome! See our [Contributing Guidelines](../CONTRIBUTING.md) for:

- Writing style guidelines
- Documentation structure standards
- Review process for documentation changes
- How to test documentation locally

---

**Ready to get started?** 🚀

- **New to OpenFrame CLI?** Start with [Introduction](./getting-started/introduction.md)
- **Want to contribute?** Begin with [Development Overview](./development/README.md)
- **Need something specific?** Use the search functionality in your browser

*Documentation generated by [OpenFrame Doc Orchestrator](https://github.com/flamingo-stack/openframe-oss-tenant)*