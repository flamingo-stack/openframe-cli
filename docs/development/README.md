# Development Documentation

Welcome to the OpenFrame CLI development documentation. This section covers everything you need to contribute to, extend, and work with the OpenFrame CLI codebase.

---

## Overview

OpenFrame CLI is written in **Go** and uses the **Cobra** framework for CLI command definitions. The project follows a clean layered architecture: commands → services → providers → shared infrastructure.

[![OpenFrame Preview Webinar](https://img.youtube.com/vi/bINdW0CQbvY/hqdefault.jpg)](https://www.youtube.com/watch?v=bINdW0CQbvY)

---

## Documentation Index

| Section | Description |
|---|---|
| [Environment Setup](setup/environment.md) | IDE recommendations, editor plugins, development tools |
| [Local Development](setup/local-development.md) | Cloning, building, running, and debugging locally |
| [Architecture Overview](architecture/README.md) | High-level design, component relationships, data flow diagrams |
| [Security Guidelines](security/README.md) | Authentication, secrets management, secure coding practices |
| [Testing Guide](testing/README.md) | Test structure, running tests, writing new tests |
| [Contributing Guidelines](contributing/guidelines.md) | Code style, PR process, commit conventions |

---

## Quick Navigation

### "I want to..."

| Goal | Go To |
|---|---|
| Set up my dev environment for the first time | [Environment Setup](setup/environment.md) |
| Run the CLI locally from source | [Local Development](setup/local-development.md) |
| Understand how the CLI is structured | [Architecture Overview](architecture/README.md) |
| Add a new CLI command | [Architecture Overview](architecture/README.md) + [Contributing Guidelines](contributing/guidelines.md) |
| Write or run tests | [Testing Guide](testing/README.md) |
| Submit a pull request | [Contributing Guidelines](contributing/guidelines.md) |
| Understand security considerations | [Security Guidelines](security/README.md) |

---

## Technology Stack

| Layer | Technology |
|---|---|
| **Language** | Go 1.22+ |
| **CLI Framework** | [Cobra](https://github.com/spf13/cobra) |
| **Terminal UI** | [pterm](https://github.com/pterm/pterm), [promptui](https://github.com/manifoldco/promptui) |
| **Kubernetes Client** | [client-go](https://github.com/kubernetes/client-go) |
| **GitOps** | [ArgoCD](https://argoproj.github.io/cd/) via native K8s client |
| **Cluster Provider** | [K3D](https://k3d.io) |
| **Package Manager** | [Helm](https://helm.sh) |
| **Dev Intercept** | [Telepresence](https://www.telepresence.io) |
| **Live Reload** | [Skaffold](https://skaffold.dev) |
| **Testing** | [testify](https://github.com/stretchr/testify) |

---

## Project Structure at a Glance

```text
openframe-cli/
├── cmd/                    # Cobra command definitions
│   ├── root.go             # Root command, global flags
│   ├── bootstrap/          # bootstrap command
│   ├── cluster/            # cluster subcommands (create, delete, list, status, cleanup)
│   ├── chart/              # chart subcommands (install)
│   └── dev/                # dev subcommands (intercept, skaffold)
├── internal/               # Internal packages (not exported)
│   ├── bootstrap/          # Bootstrap orchestration service
│   ├── cluster/            # Cluster lifecycle management
│   ├── chart/              # Chart/ArgoCD installation
│   ├── dev/                # Developer workflow services
│   └── shared/             # Shared utilities (executor, ui, errors, config)
├── tests/                  # Test suites
│   ├── integration/        # Integration tests (requires real k3d)
│   ├── mocks/              # Mock implementations
│   └── testutil/           # Test utilities and helpers
└── main.go                 # Application entry point
```

---

## Community & Support

- 💬 **OpenMSP Slack**: [Join here](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)
- 🌐 **OpenMSP Community**: [https://www.openmsp.ai/](https://www.openmsp.ai/)
- 🌐 **OpenFrame**: [https://openframe.ai](https://openframe.ai)
- 🌐 **Flamingo**: [https://flamingo.run](https://flamingo.run)
