# OpenFrame CLI — Documentation

Welcome to the official documentation for **OpenFrame CLI**, the command-line tool that bootstraps and manages [OpenFrame](https://openframe.ai) Kubernetes environments.

> **OpenFrame** is the unified platform from [Flamingo](https://flamingo.run) that integrates multiple MSP tools into a single AI-driven interface, automating IT support operations across the stack.

---

## 📚 Table of Contents

- [Getting Started](#-getting-started)
- [Development](#-development)
- [Reference](#-reference)
- [Architecture Diagrams](#-architecture-diagrams)
- [External Repositories](#-external-repositories)
- [Quick Links](#-quick-links)

---

## 🚀 Getting Started

New to OpenFrame CLI? Start here:

| Guide | Description |
|---|---|
| [Introduction](./getting-started/introduction.md) | What is OpenFrame CLI and what can it do? |
| [Prerequisites](./getting-started/prerequisites.md) | System requirements, required tools, and environment setup |
| [Quick Start](./getting-started/quick-start.md) | Download, install, and run your first `openframe bootstrap` in 5 minutes |
| [First Steps](./getting-started/first-steps.md) | Explore key features and commands after your first installation |

---

## 🛠️ Development

Documentation for contributors and developers working on the CLI:

| Guide | Description |
|---|---|
| [Development Overview](./development/README.md) | Repository structure, tech stack, and navigation index |
| [Environment Setup](./development/setup/environment.md) | Go toolchain, IDE configuration, and linting tools |
| [Local Development](./development/setup/local-development.md) | Clone, build, run, debug, and iterate on the CLI |
| [Architecture Overview](./development/architecture/README.md) | High-level design, component breakdown, and data flows |
| [Security Guidelines](./development/security/README.md) | Secret handling, input validation, and security patterns |
| [Testing Guide](./development/testing/README.md) | Unit tests, integration tests, mocks, and coverage targets |
| [Contributing Guidelines](./development/contributing/guidelines.md) | Code style, branch naming, commit messages, and PR process |
| [Release Signing](./development/release-signing.md) | Binary signing for macOS, Windows, and cosign verification |

---

## 📖 Reference

Technical reference documentation generated from source code analysis:

| Document | Description |
|---|---|
| [Architecture Reference](./reference/architecture/overview.md) | Full component reference — commands, services, providers, shared infrastructure, CLI command reference, and dependency table |

---

## 🗺️ Architecture Diagrams

Visual documentation for the OpenFrame CLI architecture. View Mermaid (`.mmd`) diagrams using any Mermaid-compatible renderer (VS Code with the Mermaid extension, GitHub, or the [Mermaid Live Editor](https://mermaid.live)):

| Diagram | Description |
|---|---|
| [High-Level System Design](./diagrams/architecture/high-level-system-design.mmd) | Overview of CLI entry point, command layer, service layer, providers, and external tools |
| [Dependency Flowchart](./diagrams/architecture/dependency-flowchart.mmd) | How commands wire into services, providers, and shared infrastructure |
| [Bootstrap Sequence Diagram](./diagrams/architecture/bootstrap-sequence-diagram.mmd) | Full sequence of `openframe bootstrap` from user invocation to platform ready |
| [App Install/Upgrade Data Flow](./diagrams/architecture/app-install-upgrade-data-flow.mmd) | Sequence diagram for `openframe app install` and context selection |

---

## 📦 External Repositories

### OpenFrame Platform (openframe-oss-tenant)

The OpenFrame platform configuration — Helm charts and values — lives in a separate repository maintained independently:

- **Repository:** [flamingo-stack/openframe-oss-tenant](https://github.com/flamingo-stack/openframe-oss-tenant)
- **Documentation:** [https://github.com/flamingo-stack/openframe-oss-tenant/tree/main/docs](https://github.com/flamingo-stack/openframe-oss-tenant/tree/main/docs)

> **Note:** The `openframe-oss-tenant` repository is **not** located in a subdirectory of this CLI repository. Always reference the external repository for platform chart installation and usage.

---

## 🔗 Quick Links

| Resource | Link |
|---|---|
| Project README | [../README.md](../README.md) |
| Contributing Guide | [../CONTRIBUTING.md](../CONTRIBUTING.md) |
| Releases | [https://github.com/flamingo-stack/openframe-cli/releases](https://github.com/flamingo-stack/openframe-cli/releases) |
| OpenMSP Slack | [https://www.openmsp.ai/](https://www.openmsp.ai/) |
| OpenFrame Platform | [https://openframe.ai](https://openframe.ai) |
| Flamingo | [https://flamingo.run](https://flamingo.run) |

---

## 💬 Community & Support

All questions, feature discussions, and bug reports are handled in the **OpenMSP Slack community** — there are no GitHub Issues or Discussions for this project.

- **Join:** [https://www.openmsp.ai/](https://www.openmsp.ai/)
- **Invite link:** [https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)

---

*Documentation generated by [🦩 Flamingo AI Technical Writer](https://flamingo.run)*
