# Development Documentation

Welcome to the OpenFrame CLI development documentation. This section covers everything you need to contribute to, extend, or understand the internals of the OpenFrame CLI codebase.

---

## What Is OpenFrame CLI?

OpenFrame CLI is a Go application built with [Cobra](https://github.com/spf13/cobra) that orchestrates Kubernetes environment management. It follows a layered clean architecture: thin command handlers delegate to service layers, which compose providers and infrastructure utilities. All external I/O is abstracted behind interfaces to maximize testability.

---

## Documentation Index

| Document | Description |
|----------|-------------|
| [Environment Setup](setup/environment.md) | IDE setup, editor plugins, Go toolchain configuration |
| [Local Development Guide](setup/local-development.md) | Cloning, building, running, and debugging locally |
| [Architecture Overview](architecture/README.md) | High-level architecture, component diagram, data flow |
| [Security Best Practices](security/README.md) | Auth patterns, secrets management, input validation |
| [Testing Guide](testing/README.md) | Test structure, running tests, writing new tests |
| [Contributing Guidelines](contributing/guidelines.md) | Code style, PR process, commit format, review checklist |

---

## Technology Stack

| Component | Technology |
|-----------|-----------|
| **Language** | Go 1.21+ |
| **CLI Framework** | [Cobra](https://github.com/spf13/cobra) |
| **Terminal UI** | [pterm](https://github.com/pterm/pterm) + [promptui](https://github.com/manifoldco/promptui) |
| **Kubernetes Client** | [client-go](https://github.com/kubernetes/client-go) |
| **ArgoCD Client** | [argo-cd/v2](https://github.com/argoproj/argo-cd) generated clientset |
| **YAML Parsing** | [gopkg.in/yaml.v3](https://pkg.go.dev/gopkg.in/yaml.v3) |
| **Cluster Provider** | K3D (K3s-in-Docker) |
| **GitOps** | ArgoCD with App-of-Apps pattern |
| **Dev Workflows** | Telepresence + Skaffold |

---

## Repository Structure

```text
openframe-cli/
├── cmd/                        # CLI entry layer (Cobra commands)
│   ├── root.go                 # Root command, global flags
│   ├── bootstrap/              # bootstrap command
│   ├── cluster/                # cluster subcommands
│   ├── chart/                  # chart subcommands
│   └── dev/                    # dev subcommands
├── internal/                   # Business logic (not importable externally)
│   ├── bootstrap/              # Bootstrap service
│   ├── cluster/                # Cluster service, K3D provider, UI, models
│   ├── chart/                  # Chart service, Helm/ArgoCD/Git providers
│   ├── dev/                    # Intercept & scaffold services
│   └── shared/                 # Cross-cutting: executor, UI, errors, config
├── tests/
│   ├── integration/            # Integration tests (require real tools)
│   ├── mocks/                  # Test doubles
│   └── testutil/               # Test helpers and assertion utilities
└── main.go                     # Binary entrypoint
```

---

## Quick Navigation

**New to the codebase?** Start with the [Architecture Overview](architecture/README.md) to understand how the layers fit together.

**Setting up your machine?** Go to [Environment Setup](setup/environment.md).

**Ready to code?** Follow the [Local Development Guide](setup/local-development.md).

**Writing tests?** See the [Testing Guide](testing/README.md).

**Submitting a PR?** Check the [Contributing Guidelines](contributing/guidelines.md) first.

---

## Community

All development discussions, bug reports, and feature requests are managed in the **OpenMSP Slack community**:

- [Join OpenMSP Slack](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)
- [OpenMSP Website](https://www.openmsp.ai/)
