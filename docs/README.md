# OpenFrame CLI Documentation

Welcome to the comprehensive documentation for OpenFrame CLI, a modern interactive command-line tool for managing OpenFrame Kubernetes clusters and development workflows.

## 📚 Table of Contents

### Getting Started
Start here if you're new to OpenFrame CLI:

- [Introduction](./getting-started/introduction.md) - What is OpenFrame CLI and how it works
- [Prerequisites](./getting-started/prerequisites.md) - System requirements and dependencies
- [Quick Start](./getting-started/quick-start.md) - Get running in 5 minutes
- [First Steps](./getting-started/first-steps.md) - Essential commands and workflows

### Development
For contributors and developers:

- [Development Overview](./development/README.md) - Development section index
- [Environment Setup](./development/setup/environment.md) - Set up your development environment
- [Local Development](./development/setup/local-development.md) - Run and test OpenFrame CLI locally
- [Architecture Overview](./development/architecture/overview.md) - System architecture and design
- [Testing Guide](./development/testing/overview.md) - Testing strategies and best practices
- [Contributing Guidelines](./development/contributing/guidelines.md) - How to contribute to the project

### Reference
Technical reference documentation:

- [Architecture Overview](./reference/architecture/overview.md) - Complete technical architecture documentation

### Diagrams
Visual documentation and architecture diagrams:

- [Architecture Diagrams](./diagrams/) - Mermaid diagrams showing system architecture and data flow

## 🚀 Quick Links

### Essential Commands
```bash
# Complete environment setup
openframe bootstrap

# Cluster management
openframe cluster create my-cluster
openframe cluster status my-cluster

# Chart installation
openframe chart install

# Development tools
openframe dev intercept my-service --port 8080
```

### Key Features
- **🎯 Interactive Cluster Creation** - Guided wizard for K3D clusters
- **⚡ Kubernetes Management** - Complete cluster lifecycle operations
- **📦 Chart Installation** - Automated ArgoCD and OpenFrame deployment
- **🔧 Development Tools** - Service intercepts and scaffolding workflows
- **📊 Real-time Monitoring** - Cluster status and health monitoring

### Architecture Highlights
OpenFrame CLI follows a clean hexagonal architecture:
- **Command Layer** - Cobra-based CLI with rich terminal UI
- **Service Layer** - Business logic for cluster, chart, and dev operations
- **Provider Layer** - Integrations with K3D, Helm, ArgoCD, Telepresence
- **Shared Infrastructure** - Common utilities and error handling

## 📖 Navigation Guide

### For New Users
1. Start with [Introduction](./getting-started/introduction.md) to understand what OpenFrame CLI does
2. Check [Prerequisites](./getting-started/prerequisites.md) to ensure your system is ready
3. Follow the [Quick Start](./getting-started/quick-start.md) guide to get running
4. Explore [First Steps](./getting-started/first-steps.md) for essential workflows

### For Developers
1. Read the [Development Overview](./development/README.md) for contribution guidelines
2. Set up your environment with [Environment Setup](./development/setup/environment.md)
3. Understand the architecture with [Architecture Overview](./development/architecture/overview.md)
4. Learn about testing with [Testing Guide](./development/testing/overview.md)

### For System Architects
1. Review [Architecture Overview](./reference/architecture/overview.md) for technical details
2. Examine [Architecture Diagrams](./diagrams/) for visual representations
3. Study component relationships and data flows

## 🔗 External Resources

- **Project Repository**: [GitHub](https://github.com/flamingo-stack/openframe-cli)
- **Main Project README**: [README.md](../README.md)
- **Contributing Guidelines**: [CONTRIBUTING.md](../CONTRIBUTING.md)
- **License**: [LICENSE.md](../LICENSE.md)
- **OpenFrame Website**: [openframe.ai](https://openframe.ai)
- **Flamingo Platform**: [flamingo.run](https://flamingo.run)

## 💬 Community & Support

- **Slack Community**: Join [OpenMSP Slack](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA) for discussions and support
- **Issue Tracking**: Report bugs and request features on [GitHub Issues](https://github.com/flamingo-stack/openframe-cli/issues)
- **Documentation Feedback**: Help us improve these docs by opening issues or pull requests

## 🆕 What's New

OpenFrame CLI provides cutting-edge features for modern Kubernetes development:

- **Bootstrap Workflow** - One command to set up complete development environments
- **Smart Detection** - Automatic prerequisite checking and installation guidance
- **Rich Terminal UI** - Beautiful interactive prompts and real-time status updates
- **Development Integration** - Seamless Telepresence intercepts and Skaffold workflows

## 🗺 Documentation Structure

```
docs/
├── README.md (this file)           # Master index
├── getting-started/                # New user documentation
│   ├── introduction.md
│   ├── prerequisites.md
│   ├── quick-start.md
│   └── first-steps.md
├── development/                    # Contributor documentation
│   ├── README.md
│   ├── setup/
│   ├── architecture/
│   ├── testing/
│   └── contributing/
├── reference/                      # Technical reference
│   └── architecture/
└── diagrams/                      # Visual documentation
```

---
*Documentation generated by [OpenFrame Doc Orchestrator](https://github.com/flamingo-stack/openframe-oss-tenant)*