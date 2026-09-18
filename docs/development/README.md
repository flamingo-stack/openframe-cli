# Development Documentation

Welcome to the OpenFrame CLI development documentation. This section covers everything you need to contribute to, extend, and understand the internals of the `openframe` CLI.

OpenFrame CLI is written in **Go** and uses [Cobra](https://github.com/spf13/cobra) for command-line parsing. It orchestrates K3D clusters, ArgoCD GitOps deployments, and Helm chart management through a layered service/provider architecture.

---

## Documentation Index

| Document | Description |
|---|---|
| [Environment Setup](setup/environment.md) | IDE configuration, Go toolchain, editor extensions |
| [Local Development](setup/local-development.md) | Clone, build, run, and debug the CLI locally |
| [Architecture Overview](architecture/README.md) | High-level design, component breakdown, data flows |
| [Security Guidelines](security/README.md) | Auth patterns, secret handling, vulnerability mitigations |
| [Testing Guide](testing/README.md) | Unit tests, integration tests, test utilities |
| [Contributing Guidelines](contributing/guidelines.md) | Code style, PR process, commit messages |

---

## Quick Navigation

### I want to...

**Build and run the CLI locally**
→ See [Local Development](setup/local-development.md)

**Understand how the codebase is structured**
→ See [Architecture Overview](architecture/README.md)

**Add a new command or feature**
→ Start with [Architecture Overview](architecture/README.md), then [Contributing Guidelines](contributing/guidelines.md)

**Write or run tests**
→ See [Testing Guide](testing/README.md)

**Handle secrets or security concerns**
→ See [Security Guidelines](security/README.md)

**Set up my development environment**
→ See [Environment Setup](setup/environment.md)

---

## Repository Structure

```text
openframe-cli/
├── cmd/                    # Cobra command definitions (entry points)
│   ├── root.go             # Root command, wires all subcommands
│   ├── bootstrap/          # openframe bootstrap
│   ├── cluster/            # openframe cluster (create/delete/list/status/cleanup)
│   ├── app/                # openframe app (install/upgrade/status/access/uninstall)
│   ├── prerequisites/      # openframe prerequisites (check/install)
│   └── update/             # openframe update (self-update/rollback)
├── internal/               # All internal business logic
│   ├── bootstrap/          # Bootstrap service (cluster + chart orchestration)
│   ├── cluster/            # Cluster service + K3D provider
│   ├── chart/              # Chart services, ArgoCD/Helm/Git providers
│   ├── app/                # App status and uninstall services
│   ├── k8s/                # Kubernetes client utilities
│   ├── platform/           # OS detection and platform hints
│   ├── prerequisites/      # Prerequisite framework
│   └── shared/             # Cross-cutting: executor, UI, errors, config, selfupdate
├── tests/
│   ├── integration/        # Integration tests (requires running cluster)
│   └── testutil/           # Shared test utilities and patterns
├── scripts/
│   └── sign-binary.sh      # Binary signing helper
└── main.go                 # Entry point
```

---

## Tech Stack

| Technology | Role |
|---|---|
| **Go** | Primary language |
| **Cobra** | CLI framework (command/flag parsing) |
| **K3D** | Local Kubernetes cluster provider |
| **ArgoCD** | GitOps deployment engine (via client-go dynamic client) |
| **Helm** | Kubernetes package manager (CLI wrapper) |
| **go-git** | Git operations (no `git` binary dependency) |
| **client-go** | Kubernetes API client |
| **pterm** | Terminal UI rendering (spinners, prompts, colors) |
| **Sigstore/cosign** | Binary signature verification for self-updates |

---

## External Dependencies

The OpenFrame platform chart lives in a separate repository:

- **openframe-oss-tenant:** [https://github.com/flamingo-stack/openframe-oss-tenant](https://github.com/flamingo-stack/openframe-oss-tenant)
- Documentation: [https://github.com/flamingo-stack/openframe-oss-tenant/tree/main/docs](https://github.com/flamingo-stack/openframe-oss-tenant/tree/main/docs)

---

## Getting Help

- **OpenMSP Slack:** [https://www.openmsp.ai/](https://www.openmsp.ai/)
- **CLI Source:** [https://github.com/flamingo-stack/openframe-cli](https://github.com/flamingo-stack/openframe-cli)
