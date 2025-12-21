# OpenFrame CLI Documentation

Welcome to the comprehensive documentation for OpenFrame CLI - a Kubernetes cluster management tool that transforms complex local development workflows into simple, one-command operations.

## 📚 Table of Contents

### Getting Started
Start here if you're new to OpenFrame CLI:

- **[Introduction](./getting-started/introduction.md)** - What is OpenFrame CLI and why use it?
- **[Prerequisites](./getting-started/prerequisites.md)** - System requirements and setup
- **[Quick Start](./getting-started/quick-start.md)** - Get running in 5 minutes
- **[First Steps](./getting-started/first-steps.md)** - Essential workflows and next steps

### Development
For contributors and developers who want to extend OpenFrame CLI:

- **[Development Overview](./development/README.md)** - Development section index and navigation
- **[Environment Setup](./development/setup/environment.md)** - Set up your development environment
- **[Local Development](./development/setup/local-development.md)** - Build and run OpenFrame locally
- **[Architecture Overview](./development/architecture/overview.md)** - System architecture and design
- **[Testing Guide](./development/testing/overview.md)** - Testing strategies and practices
- **[Contributing Guidelines](./development/contributing/guidelines.md)** - How to contribute code and documentation

### Reference
Technical reference documentation:

- **[Architecture Overview](./reference/architecture/overview.md)** - Detailed technical architecture documentation
- **[CLI Commands Reference](./reference/cli-commands.md)** - Complete command reference (if available)
- **[Configuration Reference](./reference/configuration.md)** - Configuration options and formats (if available)

### Diagrams
Visual documentation and system diagrams:

- **[Architecture Diagrams](./diagrams/architecture/)** - Mermaid diagrams showing system structure and data flow

## 🚀 Quick Navigation

### 👋 **New to OpenFrame?**
→ Start with the **[Introduction](./getting-started/introduction.md)** to understand what OpenFrame CLI does and how it can help you.

### ⚡ **Want to Get Started Fast?**
→ Jump to the **[Quick Start Guide](./getting-started/quick-start.md)** for a 5-minute setup walkthrough.

### 🛠️ **Want to Contribute?**
→ Check out **[Development Overview](./development/README.md)** and **[Contributing Guidelines](./development/contributing/guidelines.md)**.

### 🏗️ **Need Technical Details?**
→ Review the **[Architecture Overview](./reference/architecture/overview.md)** for comprehensive system documentation.

## 📖 Documentation Sections

### 🎯 Getting Started Section

Perfect for new users who want to understand and start using OpenFrame CLI:

| Guide | Purpose | Time Required |
|-------|---------|---------------|
| **Introduction** | Understand OpenFrame's purpose and benefits | 5 minutes |
| **Prerequisites** | Install required tools and dependencies | 10 minutes |
| **Quick Start** | Complete your first successful deployment | 5 minutes |
| **First Steps** | Learn essential workflows and patterns | 15 minutes |

### 🔧 Development Section

Comprehensive guides for contributors and developers:

| Guide | Purpose | Audience |
|-------|---------|----------|
| **Environment Setup** | Configure development tools and IDE | All developers |
| **Local Development** | Build, run, and debug OpenFrame locally | Contributors |
| **Architecture Overview** | Understand system design and patterns | All developers |
| **Testing Guide** | Writing and running tests effectively | Contributors |
| **Contributing Guidelines** | Code standards and contribution process | Contributors |

### 📚 Reference Section

Technical documentation and specifications:

| Document | Content | Use Case |
|----------|---------|----------|
| **Architecture Overview** | Complete system architecture | Understanding internals |
| **CLI Commands** | Command reference and examples | Daily usage reference |
| **Configuration** | Configuration options and formats | Advanced customization |

## 🎯 Common Use Cases

### **Local Kubernetes Development**
1. **Start here:** [Prerequisites](./getting-started/prerequisites.md) → [Quick Start](./getting-started/quick-start.md)
2. **Learn more:** [Introduction](./getting-started/introduction.md) → [First Steps](./getting-started/first-steps.md)

### **Contributing to OpenFrame**
1. **Understand the system:** [Architecture Overview](./reference/architecture/overview.md)
2. **Set up development:** [Environment Setup](./development/setup/environment.md) → [Local Development](./development/setup/local-development.md)
3. **Follow the process:** [Contributing Guidelines](./development/contributing/guidelines.md)

### **Platform Integration**
1. **Technical understanding:** [Architecture Overview](./reference/architecture/overview.md)
2. **Development setup:** [Environment Setup](./development/setup/environment.md)
3. **Testing approach:** [Testing Guide](./development/testing/overview.md)

### **Learning Kubernetes**
1. **Get familiar:** [Introduction](./getting-started/introduction.md)
2. **Hands-on practice:** [Quick Start](./getting-started/quick-start.md) → [First Steps](./getting-started/first-steps.md)
3. **Understand the internals:** [Architecture Overview](./reference/architecture/overview.md)

## 🛠️ Key Features Documentation

### **One-Command Bootstrap**
- **Overview:** [Introduction - Key Features](./getting-started/introduction.md#key-features--benefits)
- **Usage:** [Quick Start - Bootstrap](./getting-started/quick-start.md)
- **Technical:** [Architecture - Bootstrap Flow](./reference/architecture/overview.md#bootstrap-command-execution-flow)

### **K3d Integration**
- **Getting Started:** [Prerequisites - K3d Setup](./getting-started/prerequisites.md)
- **Commands:** [First Steps - Cluster Management](./getting-started/first-steps.md)
- **Architecture:** [Reference - Cluster Operations](./reference/architecture/overview.md#cluster-operation-flow)

### **ArgoCD Automation**
- **Overview:** [Introduction - GitOps Features](./getting-started/introduction.md)
- **Setup:** [Quick Start - Chart Installation](./getting-started/quick-start.md)
- **Technical:** [Architecture - Chart Management](./reference/architecture/overview.md#chart-management-commands)

### **Development Tools**
- **Concepts:** [Introduction - Development Tools](./getting-started/introduction.md)
- **Usage:** [First Steps - Development Workflows](./getting-started/first-steps.md)
- **Implementation:** [Architecture - Development Commands](./reference/architecture/overview.md#development-commands)

## 📋 Quick Reference

### Essential Commands

```bash
# Complete setup in one command
openframe bootstrap my-cluster

# Step-by-step approach
openframe cluster create my-cluster
openframe chart install my-cluster

# Management commands
openframe cluster list
openframe cluster status my-cluster
openframe cluster delete my-cluster
```

### Key Concepts

| Concept | Description | Documentation |
|---------|-------------|---------------|
| **Bootstrap** | One-command complete environment setup | [Architecture - Bootstrap](./reference/architecture/overview.md#bootstrap-commands) |
| **Cluster** | K3d-based Kubernetes cluster management | [Quick Start](./getting-started/quick-start.md) |
| **Charts** | ArgoCD and GitOps application management | [First Steps](./getting-started/first-steps.md) |
| **Dev Tools** | Traffic interception and development workflows | [Introduction](./getting-started/introduction.md) |

### Global Options

| Flag | Purpose | Example |
|------|---------|---------|
| `--verbose` | Detailed logging | `openframe bootstrap --verbose` |
| `--non-interactive` | CI/CD mode | `openframe bootstrap --non-interactive` |
| `--deployment-mode` | Specify deployment type | `openframe bootstrap --deployment-mode=oss-tenant` |
| `--dry-run` | Preview without execution | `openframe cluster create --dry-run` |

## 📖 Quick Links

### 🏠 **Project Resources**
- **[Main README](../README.md)** - Project overview and quick start
- **[Contributing Guide](../CONTRIBUTING.md)** - How to contribute to the project
- **[License](../LICENSE.md)** - License information and terms

### 🔗 **External Resources**
- **[Kubernetes Documentation](https://kubernetes.io/docs/)** - Official Kubernetes docs
- **[K3d Documentation](https://k3d.io/)** - K3d cluster management
- **[ArgoCD Documentation](https://argo-cd.readthedocs.io/)** - GitOps with ArgoCD
- **[Helm Documentation](https://helm.sh/docs/)** - Kubernetes package management

### 💬 **Community & Support**
- **[GitHub Issues](https://github.com/flamingo-stack/openframe-cli/issues)** - Bug reports and feature requests
- **[GitHub Discussions](https://github.com/flamingo-stack/openframe-cli/discussions)** - Questions and community discussions
- **[OpenFrame Website](https://openframe.run/)** - Official project website

## 🎓 Learning Paths

### **Beginner Path** (30 minutes)
1. Read [Introduction](./getting-started/introduction.md) (5 min)
2. Complete [Prerequisites](./getting-started/prerequisites.md) (15 min)
3. Follow [Quick Start](./getting-started/quick-start.md) (5 min)
4. Explore [First Steps](./getting-started/first-steps.md) (5 min)

### **Developer Path** (2 hours)
1. Complete Beginner Path (30 min)
2. Read [Architecture Overview](./reference/architecture/overview.md) (30 min)
3. Set up [Development Environment](./development/setup/environment.md) (30 min)
4. Try [Local Development](./development/setup/local-development.md) (30 min)

### **Contributor Path** (4 hours)
1. Complete Developer Path (2 hours)
2. Read [Contributing Guidelines](./development/contributing/guidelines.md) (30 min)
3. Review [Testing Guide](./development/testing/overview.md) (30 min)
4. Make your first contribution (1 hour)

## 🆘 Getting Help

### **For Users**
- **Command help:** Add `--help` to any command
- **Interactive modes:** Most commands offer guided prompts
- **Troubleshooting:** Use `--verbose` flag for detailed output

### **For Developers**
- **Architecture questions:** Check [Architecture Overview](./reference/architecture/overview.md)
- **Development setup:** See [Development Environment](./development/setup/environment.md)
- **Contribution process:** Follow [Contributing Guidelines](./development/contributing/guidelines.md)

### **For Everyone**
- **General questions:** Use [GitHub Discussions](https://github.com/flamingo-stack/openframe-cli/discussions)
- **Bug reports:** Create [GitHub Issues](https://github.com/flamingo-stack/openframe-cli/issues)
- **Feature requests:** Start with [GitHub Discussions](https://github.com/flamingo-stack/openframe-cli/discussions)

---

*Documentation generated by [OpenFrame Doc Orchestrator](https://github.com/flamingo-stack/openframe-oss-tenant)*