<!-- source-hash: d4f31502ba5b44700be2e2259703919e -->
Defines comprehensive interfaces for a Helm chart installation system with ArgoCD integration, providing contracts for service orchestration, cluster management, and GitOps workflows.

## Key Components

**Core Services:**
- `ChartInstaller` - Main orchestrator for chart installation processes
- `ArgoCDService` - Manages ArgoCD lifecycle and application monitoring  
- `AppOfAppsService` - Handles app-of-apps pattern deployments

**Providers:**
- `HelmProvider` - Helm chart operations and status checking
- `GitProvider` - Git repository cloning and cleanup
- `ClusterLister` - Cluster discovery and listing

**Configuration & UI:**
- `ConfigBuilder` - Constructs installation configurations
- `OperationsUI` - User interaction and confirmation prompts
- `ServiceFactory` - Dependency injection and service creation

**Orchestration:**
- `ServiceOrchestrator` - Coordinates service interactions
- `WorkflowExecutor` - Executes multi-step workflows with tracking

## Usage Example

```go
// Create and execute an installation workflow
factory := NewServiceFactory()
installer := factory.CreateInstaller()

request := InstallationRequest{
    Args:           []string{"my-cluster"},
    Force:          false,
    DryRun:         false,
    GitHubRepo:     "https://github.com/org/charts",
    GitHubBranch:   "main",
    DeploymentMode: "oss-tenant",
}

orchestrator := factory.CreateOrchestrator()
err := orchestrator.ExecuteInstallation(request)
```