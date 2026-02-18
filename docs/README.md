# OpenFrame CLI Documentation

Welcome to the comprehensive documentation for OpenFrame CLI, a modern command-line tool for managing OpenFrame Kubernetes clusters and development workflows.

## 📚 Table of Contents

### Getting Started

New to OpenFrame CLI? Start here for a smooth onboarding experience:

- [Introduction](./getting-started/introduction.md) - Overview and key concepts
- [Prerequisites](./getting-started/prerequisites.md) - System requirements and setup
- [Quick Start](./getting-started/quick-start.md) - Get running in 5 minutes
- [First Steps](./getting-started/first-steps.md) - Explore core features and workflows

### Development

Set up your development environment and learn the contribution workflow:

- [Development Overview](./development/README.md) - Development guide overview
- [Local Development Setup](./development/setup/local-development.md) - Clone, build, and run from source
- [Environment Configuration](./development/setup/environment.md) - Configure your development environment
- [Architecture Guide](./development/architecture/README.md) - Understand the system design

### Reference

Technical reference documentation:

- [Architecture Overview](./architecture/overview.md) - System architecture and component relationships

### Diagrams

Visual documentation and architecture diagrams:

```text
docs/diagrams/architecture/
├── high-level-architecture-diagram.mmd    # System overview
├── service-dependencies-diagram.mmd       # Service relationships  
├── bootstrap-command-sequence.mmd         # Bootstrap workflow
└── README.md                              # Diagram documentation
```

View the Mermaid diagrams in your IDE or convert them using Mermaid CLI for web viewing.

### CLI Tools

The OpenFrame CLI is maintained in this repository and provides comprehensive command-line functionality for:
- Kubernetes cluster management (K3D)
- ArgoCD and Helm chart installation
- Development workflows with service intercepts
- GitOps automation and deployment

For installation and usage, see the [Quick Start guide](./getting-started/quick-start.md) above.

## 🚀 Quick Navigation

### For New Users
1. **[Introduction](./getting-started/introduction.md)** - Learn what OpenFrame CLI does
2. **[Quick Start](./getting-started/quick-start.md)** - Install and bootstrap your first environment
3. **[First Steps](./getting-started/first-steps.md)** - Explore key features

### For Developers  
1. **[Local Development](./development/setup/local-development.md)** - Set up development environment
2. **[Architecture Guide](./development/architecture/README.md)** - Understand the codebase
3. **[Contributing Guidelines](../CONTRIBUTING.md)** - Learn the contribution process

### For Advanced Users
1. **[Architecture Overview](./architecture/overview.md)** - Deep dive into system design
2. **[Command Reference](./getting-started/first-steps.md)** - Complete command documentation
3. **[Troubleshooting](./getting-started/quick-start.md#troubleshooting-quick-issues)** - Common issues and solutions

## 🎯 Key Features Documented

### One-Command Bootstrap
Learn how `openframe bootstrap` creates a complete Kubernetes environment:
- K3D cluster creation with networking
- ArgoCD installation and configuration
- Certificate management with mkcert
- Service deployment with app-of-apps pattern

### Development Workflows
Discover powerful development tools:
- Service intercepts with Telepresence
- Hot reload development with Skaffold
- Code scaffolding for new services
- GitOps workflows with ArgoCD

### Cluster Management
Master Kubernetes cluster operations:
- Create, delete, and list clusters
- Status monitoring and health checks
- Custom configuration options
- Multi-cluster management

## 📖 Documentation Standards

This documentation follows these principles:
- **Comprehensive**: Covers all features and use cases
- **Beginner-Friendly**: Assumes minimal prior knowledge
- **Task-Oriented**: Focuses on practical workflows
- **Well-Tested**: Examples are verified and up-to-date

## 🔗 External Resources

### Community & Support
- **OpenMSP Slack Community**: https://www.openmsp.ai/
- **Slack Invite**: https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA
- **Flamingo Platform**: https://flamingo.run
- **OpenFrame Platform**: https://openframe.ai

### Related Projects
- **[OpenFrame OSS Tenant](https://github.com/flamingo-stack/openframe-oss-tenant)** - Main platform repository
- **[OpenFrame Documentation](https://github.com/flamingo-stack/openframe-oss-tenant/tree/main/docs)** - Platform-wide documentation

## 📝 Contributing to Documentation

Help improve these docs:

1. **Found an Issue?** Report it in our Slack community
2. **Want to Contribute?** See the [Contributing Guidelines](../CONTRIBUTING.md)
3. **Need Clarification?** Ask questions in the #documentation channel

### Documentation Structure
```text
docs/
├── getting-started/     # New user onboarding
├── development/         # Developer workflows  
├── architecture/        # Technical reference
├── diagrams/           # Visual documentation
└── README.md           # This index file
```

## 🎥 Video Resources

Visual learners can watch our video tutorials:

- **Product Walkthrough**: Overview of OpenFrame capabilities
- **Developer Experience**: v0.3.7 feature highlights
- **Getting Started**: Step-by-step installation and setup

## 📊 Quick Links

- [Project README](../README.md) - Main project overview and installation
- [Contributing Guidelines](../CONTRIBUTING.md) - How to contribute code and documentation
- [License Information](../LICENSE.md) - Flamingo AI Unified License v1.0

---

## 🔍 What's Next?

**New to OpenFrame CLI?**  
→ Start with the [Introduction](./getting-started/introduction.md)

**Ready to Install?**  
→ Follow the [Quick Start Guide](./getting-started/quick-start.md)

**Want to Contribute?**  
→ Set up [Local Development](./development/setup/local-development.md)

**Need Architecture Details?**  
→ Read the [Architecture Overview](./architecture/overview.md)

---

*Documentation generated by [OpenFrame Doc Orchestrator](https://github.com/flamingo-stack/openframe-cli)*

**Last Updated**: Generated from CodeWiki analysis  
**Version**: Matches OpenFrame CLI releases