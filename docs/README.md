# OpenFrame CLI Documentation

Welcome to the comprehensive documentation for OpenFrame CLI - a modern tool for managing OpenFrame Kubernetes clusters and development workflows.

## 📚 Table of Contents

### Getting Started
Start here if you're new to OpenFrame CLI:
- [Introduction](./getting-started/introduction.md) - What is OpenFrame CLI and why use it?
- [Prerequisites](./getting-started/prerequisites.md) - Required tools and dependencies
- [Quick Start](./getting-started/quick-start.md) - Get running with OpenFrame in 5 minutes
- [First Steps](./getting-started/first-steps.md) - Essential commands and next steps

### Development
For contributors and developers working on OpenFrame CLI:
- [Development Overview](./development/README.md) - Development section index and roadmap
- [Environment Setup](./development/setup/environment.md) - Set up your development environment
- [Local Development](./development/setup/local-development.md) - Run and test OpenFrame CLI locally
- [Architecture Overview](./development/architecture/overview.md) - Technical architecture and design patterns
- [Testing Guide](./development/testing/overview.md) - Unit testing, integration testing, and quality assurance
- [Contributing Guidelines](./development/contributing/guidelines.md) - How to contribute code, documentation, and bug reports

### Reference
Technical reference and architecture documentation:
- [Architecture Overview](./reference/architecture/overview.md) - Complete technical architecture documentation
- CLI Command Reference (coming soon)
- Configuration Reference (coming soon)
- API Documentation (coming soon)

### Diagrams
Visual documentation and system diagrams:
- [Architecture Diagrams](./diagrams/) - Mermaid diagrams showing system architecture
- Component Interaction Flows
- Deployment Workflows

## 🚀 Quick Links

### Essential Resources
- [Project README](../README.md) - Main project overview and quick start
- [Contributing Guide](../CONTRIBUTING.md) - How to contribute to the project
- [License Information](../LICENSE.md) - Project licensing details

### Key Commands
```bash
# Complete environment setup
openframe bootstrap --deployment-mode=oss-tenant

# Cluster management
openframe cluster create    # Interactive cluster creation
openframe cluster list     # List all clusters
openframe cluster status   # Show cluster details

# Development workflows
openframe dev intercept    # Service traffic interception
openframe dev skaffold     # Live development mode
```

### Development Workflows

#### For New Contributors
1. Read the [Introduction](./getting-started/introduction.md)
2. Set up [Prerequisites](./getting-started/prerequisites.md)
3. Follow [Environment Setup](./development/setup/environment.md)
4. Review [Contributing Guidelines](./development/contributing/guidelines.md)

#### For Users
1. Install OpenFrame CLI from [Quick Start](./getting-started/quick-start.md)
2. Complete the [First Steps](./getting-started/first-steps.md) tutorial
3. Explore advanced features in the reference documentation

#### For Developers
1. Understand the [Architecture](./development/architecture/overview.md)
2. Set up [Local Development](./development/setup/local-development.md)
3. Run tests per [Testing Guide](./development/testing/overview.md)
4. Follow [Development Best Practices](./development/contributing/guidelines.md)

## 📖 Documentation Structure

This documentation is organized into four main sections:

- **Getting Started**: User-focused guides for installation and basic usage
- **Development**: Developer-focused guides for contributing and extending the CLI
- **Reference**: Complete technical documentation and architecture details
- **Diagrams**: Visual documentation including Mermaid diagrams and flowcharts

## 🔧 Features Covered

### Cluster Management
- K3d cluster creation with interactive wizard
- Multi-cluster management and monitoring
- Automated prerequisite detection
- Docker resource cleanup and optimization

### Application Deployment
- ArgoCD installation and configuration
- GitOps workflow setup
- OpenFrame application deployment
- Multi-mode support (OSS tenant, SaaS tenant, SaaS shared)

### Development Tools
- Telepresence service interception
- Skaffold live development workflows
- Hot-reload development environments
- Debugging and troubleshooting tools

### System Integration
- Docker and Kubernetes integration
- Helm chart management
- Custom resource definitions
- Service mesh compatibility

## 📝 Contributing to Documentation

Found an issue or want to improve the docs?

1. **Simple edits**: Click "Edit this page" on any documentation page
2. **Larger changes**: Follow the [Contributing Guide](../CONTRIBUTING.md)
3. **New documentation**: Propose new sections in GitHub issues

### Documentation Standards
- Follow [Flamingo Markdown Guidelines](https://github.com/flamingo-stack/openframe-oss-tenant/blob/main/MARKDOWN_FORMATTING.md)
- Include code examples with expected outputs
- Use Mermaid diagrams for complex workflows
- Test all commands and examples before submitting

## 🆘 Getting Help

- **Bug reports**: [Open an issue](https://github.com/flamingo-stack/openframe-cli/issues/new/choose)
- **Feature requests**: [Request a feature](https://github.com/flamingo-stack/openframe-cli/issues/new/choose)
- **Questions**: [Start a discussion](https://github.com/flamingo-stack/openframe-cli/discussions)
- **Documentation issues**: [Report documentation problems](https://github.com/flamingo-stack/openframe-cli/issues/new/choose)

---

*Documentation generated by [OpenFrame Doc Orchestrator](https://github.com/flamingo-stack/openframe-oss-tenant)*