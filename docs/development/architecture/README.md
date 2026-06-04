# Architecture Overview

OpenFrame CLI is a Go-based command-line tool built with a strict layered architecture. Commands at the top delegate to internal services, which use provider abstractions over external tools. All cross-cutting concerns (command execution, error handling, UI, configuration) live in a shared infrastructure layer.

> For the full architecture reference, see [./reference/architecture/overview.md](./reference/architecture/overview.md).

---

## High-Level Architecture

```mermaid
graph TB
    subgraph CLI["CLI Layer (cmd/)"]
        Root["root.go"]
        BootstrapCmd["bootstrap"]
        ClusterCmd["cluster/*"]
        ChartCmd["chart/*"]
        DevCmd["dev/*"]
    end

    subgraph Services["Internal Services (internal/)"]
        BootstrapSvc["bootstrap.Service"]
        ClusterSvc["cluster.ClusterService"]
        ChartSvc["chart.ChartService"]
        InterceptSvc["intercept.Service"]
        ScaffoldSvc["scaffold.Service"]
    end

    subgraph Providers["External Tool Providers"]
        K3dProv["k3d.Manager"]
        HelmProv["helm.HelmManager"]
        ArgoCDProv["argocd.Manager"]
        GitProv["git.Repository"]
        KubectlProv["kubectl.Provider"]
        TeleProv["telepresence.Provider"]
    end

    subgraph Shared["Shared Infrastructure (internal/shared/)"]
        Executor["executor (cmd abstraction)"]
        Errors["errors (structured types)"]
        UI["ui (pterm / promptui)"]
        Config["config (TLS / credentials)"]
    end

    subgraph External["External Tools"]
        K3D["K3D"]
        Helm["Helm"]
        ArgoCD["ArgoCD"]
        Git["Git"]
        Telepresence["Telepresence"]
        Skaffold["Skaffold"]
        Docker["Docker"]
        Kubectl["kubectl"]
    end

    Root --> BootstrapCmd
    Root --> ClusterCmd
    Root --> ChartCmd
    Root --> DevCmd

    BootstrapCmd --> BootstrapSvc
    ClusterCmd --> ClusterSvc
    ChartCmd --> ChartSvc
    DevCmd --> InterceptSvc
    DevCmd --> ScaffoldSvc

    BootstrapSvc --> ClusterSvc
    BootstrapSvc --> ChartSvc

    ClusterSvc --> K3dProv
    ChartSvc --> HelmProv
    ChartSvc --> ArgoCDProv
    ChartSvc --> GitProv
    InterceptSvc --> KubectlProv
    InterceptSvc --> TeleProv
    ScaffoldSvc --> KubectlProv
    ScaffoldSvc --> ChartSvc

    K3dProv --> K3D
    HelmProv --> Helm
    ArgoCDProv --> ArgoCD
    GitProv --> Git
    TeleProv --> Telepresence
    ScaffoldSvc --> Skaffold
    K3dProv --> Docker
    KubectlProv --> Kubectl

    ClusterSvc --> Shared
    ChartSvc --> Shared
    InterceptSvc --> Shared
    ScaffoldSvc --> Shared
```

---

## Core Components

| Package | Path | Responsibility |
|---|---|---|
| **Root Command** | `cmd/root.go` | Entry point; wires all subcommands, version info, global flags |
| **Bootstrap Command** | `cmd/bootstrap/` | Orchestrates cluster creation + chart installation in one command |
| **Cluster Command** | `cmd/cluster/` | Subcommands: create, delete, list, status, cleanup |
| **Chart Command** | `cmd/chart/` | Subcommands: install (ArgoCD + app-of-apps) |
| **Dev Command** | `cmd/dev/` | Subcommands: intercept, skaffold |
| **Bootstrap Service** | `internal/bootstrap/service.go` | Sequentially calls cluster create then chart install; handles Windows WSL init |
| **Cluster Service** | `internal/cluster/service.go` | Business logic for cluster lifecycle; wraps K3D manager |
| **K3D Manager** | `internal/cluster/providers/k3d/manager.go` | Low-level K3D operations via CLI; produces `rest.Config` |
| **Chart Service** | `internal/chart/services/chart_service.go` | Orchestrates ArgoCD + app-of-apps installation workflow |
| **Helm Manager** | `internal/chart/providers/helm/manager.go` | Helm operations using native Go clients + kubectl fallback |
| **ArgoCD Manager** | `internal/chart/providers/argocd/applications.go` | Watches ArgoCD application health/sync via native K8s clients |
| **Git Repository** | `internal/chart/providers/git/repository.go` | Shallow-clones app-of-apps chart repo to temp dir |
| **Intercept Service** | `internal/dev/services/intercept/service.go` | Manages Telepresence intercept lifecycle |
| **Scaffold Service** | `internal/dev/services/scaffold/service.go` | Runs Skaffold dev workflow with cluster bootstrap |
| **Configuration Wizard** | `internal/chart/ui/configuration/wizard.go` | Interactive Helm values configuration wizard |
| **Shared Executor** | `internal/shared/executor/executor.go` | Command execution abstraction (real + mock); handles WSL on Windows |
| **Shared Errors** | `internal/shared/errors/errors.go` | Structured error types, retry policies, user-friendly display |
| **Shared UI** | `internal/shared/ui/` | Prompts, tables, logo, progress tracking via pterm/promptui |

---

## Bootstrap Command Data Flow

The most important flow is the `bootstrap` command, which orchestrates the full environment setup:

```mermaid
sequenceDiagram
    participant User
    participant CLI as "openframe bootstrap"
    participant Bootstrap as "bootstrap.Service"
    participant ClusterSvc as "cluster.ClusterService"
    participant K3dMgr as "k3d.Manager"
    participant ChartSvc as "chart.ChartService"
    participant HelmMgr as "helm.HelmManager"
    participant ArgoCDMgr as "argocd.Manager"
    participant GitRepo as "git.Repository"

    User->>CLI: openframe bootstrap [name]
    CLI->>Bootstrap: Execute(cmd, args)
    Bootstrap->>Bootstrap: Validate flags

    Bootstrap->>ClusterSvc: CreateClusterWithPrerequisites()
    ClusterSvc->>ClusterSvc: CheckPrerequisites (Docker, k3d, kubectl, helm)
    ClusterSvc->>K3dMgr: CreateCluster(config)
    K3dMgr-->>ClusterSvc: rest.Config
    ClusterSvc-->>Bootstrap: rest.Config

    Bootstrap->>ChartSvc: InstallChartsWithConfig(request)
    ChartSvc->>ChartSvc: CheckPrerequisites (git, helm, mkcert)
    ChartSvc->>ChartSvc: ConfigureHelmValues (wizard)

    ChartSvc->>HelmMgr: InstallArgoCD(ctx, config)
    HelmMgr-->>ChartSvc: ArgoCD installed

    ChartSvc->>GitRepo: CloneChartRepository(appOfAppsConfig)
    GitRepo-->>ChartSvc: CloneResult (tempDir, chartPath)

    ChartSvc->>HelmMgr: InstallAppOfApps(ctx, config)
    HelmMgr-->>ChartSvc: app-of-apps installed

    ChartSvc->>ArgoCDMgr: WaitForApplications(ctx, config)
    ArgoCDMgr->>ArgoCDMgr: Poll ArgoCD Application CRDs
    ArgoCDMgr-->>ChartSvc: All apps Healthy + Synced

    ChartSvc-->>Bootstrap: success
    Bootstrap-->>CLI: success
    CLI-->>User: Bootstrap complete
```

---

## Configuration Wizard Flow

The interactive chart installation flow guides operators through Helm values configuration:

```mermaid
sequenceDiagram
    participant User
    participant Workflow as "InstallationWorkflow"
    participant Wizard as "ConfigurationWizard"
    participant Modifier as "HelmValuesModifier"
    participant Installer as "chart.Installer"

    User->>Workflow: ExecuteWithContext(ctx, req)
    Workflow->>Workflow: SelectCluster (interactive or arg)
    Workflow->>Wizard: ConfigureHelmValues()
    Wizard->>User: Select deployment mode (oss-tenant / saas-tenant / saas-shared)
    User-->>Wizard: oss-tenant
    Wizard->>User: Select config mode (default / interactive)
    User-->>Wizard: interactive
    Wizard->>Modifier: LoadOrCreateBaseValues()
    Modifier-->>Wizard: map of values
    Wizard->>User: Configure branch / Docker / Ingress
    User-->>Wizard: selections
    Wizard->>Modifier: ApplyConfiguration(values, config)
    Wizard->>Modifier: CreateTemporaryValuesFile(values)
    Modifier-->>Wizard: helm-values-tmp.yaml
    Wizard-->>Workflow: ChartConfiguration

    Workflow->>Installer: InstallChartsWithContext(ctx, config)
    Installer->>Installer: ArgoCD install
    Installer->>Installer: app-of-apps install
    Installer->>Installer: WaitForApplications
    Installer-->>Workflow: success
```

---

## Dependency Injection Pattern

The CLI uses constructor injection throughout. The `CommandExecutor` interface is the central abstraction that makes the entire CLI unit-testable without running real tools:

```mermaid
graph LR
    TestCode["Test Code"] --> MockExec["MockCommandExecutor"]
    ProdCode["Production Code"] --> RealExec["RealCommandExecutor"]
    MockExec --> ExecInterface["CommandExecutor (interface)"]
    RealExec --> ExecInterface
    ExecInterface --> Services["All Services"]
    Services --> Providers["All Providers"]
```

All services accept a `CommandExecutor` via their constructors. In production this is `executor.NewRealExecutor()`. In tests this is `testutil.NewTestMockExecutor()`.

---

## Error Handling Architecture

Errors in the CLI follow a structured pattern using custom error types:

| Error Type | When Used |
|---|---|
| `ValidationError` | Field validation failures (flag values, cluster names, ports) |
| `CommandError` | External tool execution failures (k3d, helm, kubectl) |
| `BranchNotFoundError` | Git branch not found in chart repositories |
| `AlreadyHandledError` | Errors already displayed — prevents double-showing |

The `ErrorHandler` routes errors to appropriate display logic based on type, using pterm for colored, user-friendly output with troubleshooting guidance.

---

## Prerequisite System

Each command group has its own prerequisite checker:

```mermaid
graph TD
    ClusterOps["Cluster Operations"] --> ClusterPrereqs["Docker, kubectl, k3d, Helm"]
    ChartOps["Chart Operations"] --> ChartPrereqs["Git, Helm, mkcert, Memory check"]
    DevOps["Dev Operations"] --> DevPrereqs["Telepresence, jq, Skaffold"]
    ClusterPrereqs --> PrereqChecker["PrerequisiteChecker"]
    ChartPrereqs --> PrereqChecker
    DevPrereqs --> PrereqChecker
    PrereqChecker --> InstallGuide["Platform-specific install guidance"]
```

Prerequisite checkers detect CI environments and skip interactive prompts automatically.

---

## Key Design Decisions

1. **Interface-first providers** — All external tool interactions go through interfaces defined in `internal/*/utils/types/interfaces.go`, enabling full mockability in tests.

2. **Shared executor abstraction** — The `CommandExecutor` interface centralizes all shell command execution, with WSL2 wrapping built in for Windows.

3. **Structured error types** — Custom error types provide rich context for user-friendly display and enable programmatic error handling.

4. **Deferred Helm initialization** — The Helm manager defers Go client initialization until first use, preventing startup overhead for commands that don't need it.

5. **App-of-apps GitOps pattern** — The CLI shallow-clones the chart repository to a temp directory and installs locally, avoiding the need for a running chart server.

---

## Reference Documentation

For detailed per-component documentation, see the auto-generated reference:

- [Architecture Reference](./reference/architecture/overview.md) — Full CodeWiki output with complete component details
