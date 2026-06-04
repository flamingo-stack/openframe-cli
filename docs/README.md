# OpenFrame CLI — Documentation

Welcome to the OpenFrame CLI documentation. This index provides navigation to all available guides, references, and architecture documentation.

[![OpenFrame Product Walkthrough (Beta Access)](https://img.youtube.com/vi/awc-yAnkhIo/maxresdefault.jpg)](https://www.youtube.com/watch?v=awc-yAnkhIo)

---

## 📚 Table of Contents

- [Getting Started](#-getting-started)
- [Development](#-development)
- [Reference Architecture](#-reference-architecture)
- [Diagrams](#-architecture-diagrams)
- [Quick Links](#-quick-links)

---

## 🚀 Getting Started

Everything you need to install the CLI and bootstrap your first OpenFrame environment.

| Guide | Description |
|---|---|
| [Introduction](./getting-started/introduction.md) | What is OpenFrame CLI, key features, target audience, and architecture overview |
| [Prerequisites](./getting-started/prerequisites.md) | Hardware requirements, required software, OS support, and binary downloads |
| [Quick Start](./getting-started/quick-start.md) | Install the CLI and bootstrap a complete environment in under 5 minutes |
| [First Steps](./getting-started/first-steps.md) | Explore your new cluster, set up local dev workflows, and learn the commands |

### Quick Install

**macOS (Apple Silicon)**

```bash
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_arm64.tar.gz | tar xz
chmod +x openframe && sudo mv openframe /usr/local/bin/
```

**macOS (Intel)**

```bash
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_amd64.tar.gz | tar xz
chmod +x openframe && sudo mv openframe /usr/local/bin/
```

**Linux (AMD64)**

```bash
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_linux_amd64.tar.gz | tar xz
chmod +x openframe && sudo mv openframe /usr/local/bin/
```

**Windows (AMD64)**
Download [openframe-cli_windows_amd64.zip](https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_windows_amd64.zip), extract, and add `openframe.exe` to your `PATH`.

---

## 🛠️ Development

Guides for contributors and developers working on the OpenFrame CLI codebase.

| Guide | Description |
|---|---|
| [Development Overview](./development/README.md) | Overview of the development workflow, technology stack, and repository structure |
| [Environment Setup](./development/setup/environment.md) | IDE configuration, Go toolchain, recommended extensions, and Docker resources |
| [Local Development](./development/setup/local-development.md) | Clone, build, run, debug, and iterate on the CLI locally |
| [Architecture Overview](./development/architecture/README.md) | High-level architecture, component relationships, bootstrap flow, and design decisions |
| [Testing Guide](./development/testing/README.md) | Test organization, running tests, writing unit and integration tests, mock executor |
| [Security Best Practices](./development/security/README.md) | Credential management, TLS, secrets handling, and secure development guidelines |
| [Contributing Guidelines](./development/contributing/guidelines.md) | Code style, branch naming, commit format, PR process, and review checklist |

### Technology Stack

| Component | Technology |
|---|---|
| Language | Go 1.21+ |
| CLI framework | [Cobra](https://github.com/spf13/cobra) |
| Terminal UI | [pterm](https://github.com/pterm/pterm) |
| Interactive prompts | [promptui](https://github.com/manifoldco/promptui) |
| Kubernetes client | [client-go](https://pkg.go.dev/k8s.io/client-go) |
| ArgoCD client | [argo-cd/v2](https://pkg.go.dev/github.com/argoproj/argo-cd/v2) |
| Testing | [testify](https://github.com/stretchr/testify) |
| Cluster provider | K3D |
| GitOps | ArgoCD + Helm app-of-apps |

---

## 📖 Reference Architecture

Detailed technical reference generated from CodeWiki source analysis.

| Document | Description |
|---|---|
| [Architecture Overview](./reference/architecture/overview.md) | Complete component documentation: CLI layer, internal services, providers, shared infrastructure, CLI commands, and dependency graph |

### Deployment Modes

| Mode | Repository | Use Case |
|---|---|---|
| `oss-tenant` | `openframe-oss-tenant` | Default self-hosted OpenFrame deployment |
| `saas-tenant` | `openframe-saas-tenant` | SaaS managed tenant deployment |
| `saas-shared` | `openframe-saas-shared` | Shared SaaS infrastructure deployment |

---

## 🗺️ Architecture Diagrams

Visual documentation of the CLI architecture and data flows. Diagrams are in [Mermaid](https://mermaid.js.org/) format.

| Diagram | Description |
|---|---|
| [High-Level Architecture](./diagrams/architecture/high-level-architecture-diagram.mmd) | Top-level view of the CLI layer, services, providers, and target environment |
| [Dependency Graph](./diagrams/architecture/dependency-graph.mmd) | How commands, services, providers, and shared infrastructure relate |
| [Bootstrap Command Sequence](./diagrams/architecture/bootstrap-command-sequence.mmd) | Step-by-step sequence of the `openframe bootstrap` command |
| [Chart Install Configuration Flow](./diagrams/architecture/chart-install-interactive-configuration-flow.mmd) | Interactive wizard flow for Helm values configuration |
| [How Dependencies Are Used](./diagrams/architecture/how-dependencies-are-used.mmd) | Library dependency usage across CLI components |

---

## ⚡ CLI Command Reference

| Command | Description |
|---|---|
| `openframe bootstrap [name]` | Full environment setup: cluster create + chart install |
| `openframe cluster create [name]` | Create a K3D cluster |
| `openframe cluster delete [name]` | Delete a cluster |
| `openframe cluster list` | List all managed clusters |
| `openframe cluster status [name]` | Show cluster health and ArgoCD app status |
| `openframe cluster cleanup [name]` | Remove unused Docker images and resources |
| `openframe chart install [name]` | Install ArgoCD + app-of-apps on a cluster |
| `openframe dev intercept [service]` | Start a Telepresence intercept for local development |
| `openframe dev skaffold [cluster]` | Run a Skaffold hot-reload workflow |

---

## 🔗 Quick Links

| Resource | Link |
|---|---|
| [Project README](../README.md) | Main project overview and quick start |
| [Contributing Guide](../CONTRIBUTING.md) | How to contribute to OpenFrame CLI |
| [OpenMSP Community](https://www.openmsp.ai/) | Join the Slack community for support |
| [Slack Invite](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA) | Direct invite link |
| [OpenFrame Platform](https://openframe.ai) | The full OpenFrame platform |
| [Flamingo](https://flamingo.run) | Built by Flamingo |

---

> **Support happens in Slack, not GitHub Issues.**
> All questions, bug reports, and feature requests are handled in the [OpenMSP Slack community](https://www.openmsp.ai/).

---

*Documentation generated by [OpenFrame Doc Orchestrator](https://github.com/flamingo-stack/openframe-oss-tenant)*
