package services

import (
	"context"
	"fmt"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/chart/models"
	"github.com/flamingo-stack/openframe-cli/internal/chart/providers/argocd"
	"github.com/flamingo-stack/openframe-cli/internal/chart/providers/helm"
	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/errors"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/pterm/pterm"
)

// ArgoCD handles ArgoCD installation logic
type ArgoCD struct {
	helmManager   *helm.HelmManager
	pathResolver  *config.PathResolver
	argoCDManager *argocd.Manager
	executor      executor.CommandExecutor
}

// NewArgoCD creates a new ArgoCD service
func NewArgoCD(helmManager *helm.HelmManager, pathResolver *config.PathResolver, exec executor.CommandExecutor) *ArgoCD {
	// Create a non-verbose executor for ArgoCD operations to reduce command spam
	// We'll handle verbose logging at a higher level in the ArgoCD manager
	argoCDExecutor := executor.NewRealCommandExecutor(false, false) // Never verbose for internal operations

	return &ArgoCD{
		helmManager:   helmManager,
		pathResolver:  pathResolver,
		argoCDManager: argocd.NewManager(argoCDExecutor),
		executor:      exec,
	}
}

// Install installs ArgoCD using Helm
func (a *ArgoCD) Install(ctx context.Context, cfg config.ChartInstallConfig) error {
	// Always install/upgrade ArgoCD

	// Install ArgoCD with progress indication
	err := a.helmManager.InstallArgoCDWithProgress(ctx, cfg)
	if err != nil {
		return errors.WrapAsChartError("installation", "ArgoCD", err).WithCluster(cfg.ClusterName)
	}

	pterm.Success.Println("ArgoCD installed")

	// Run kubectl verification checks
	if err := a.runKubectlVerificationChecks(ctx, cfg); err != nil {
		pterm.Warning.Printf("Kubectl verification checks failed: %v\n", err)
		// Don't fail the installation, just warn
	}

	// Sleep for 10 minutes to allow ArgoCD to fully stabilize
	if err := a.waitForStabilization(ctx, cfg); err != nil {
		return err
	}

	return nil
}

// runKubectlVerificationChecks runs kubectl commands to verify the cluster state after ArgoCD installation
func (a *ArgoCD) runKubectlVerificationChecks(ctx context.Context, cfg config.ChartInstallConfig) error {
	pterm.Info.Println("Running kubectl verification checks...")

	// Build base kubectl args with explicit context if cluster name is provided
	baseArgs := []string{}
	if cfg.ClusterName != "" {
		contextName := fmt.Sprintf("k3d-%s", cfg.ClusterName)
		baseArgs = append(baseArgs, "--context", contextName)
	}

	// Check 1: kubectl get ns
	pterm.Info.Println("kubectl get ns")
	nsArgs := append(baseArgs, "get", "ns")
	result, err := a.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "kubectl",
		Args:    nsArgs,
	})
	if err != nil {
		return fmt.Errorf("failed to get namespaces: %w", err)
	}
	fmt.Println(result.Stdout)

	// Check 2: kubectl get pods -n argocd
	pterm.Info.Println("kubectl get pods -n argocd")
	podsArgs := append(baseArgs, "get", "pods", "-n", "argocd")
	result, err = a.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "kubectl",
		Args:    podsArgs,
	})
	if err != nil {
		return fmt.Errorf("failed to get ArgoCD pods: %w", err)
	}
	fmt.Println(result.Stdout)

	// Check 3: kubectl get deployments -n argocd
	pterm.Info.Println("kubectl get deployments -n argocd")
	deplArgs := append(baseArgs, "get", "deployments", "-n", "argocd")
	result, err = a.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "kubectl",
		Args:    deplArgs,
	})
	if err != nil {
		return fmt.Errorf("failed to get ArgoCD deployments: %w", err)
	}
	fmt.Println(result.Stdout)

	// Check 4: kubectl get svc -n argocd
	pterm.Info.Println("kubectl get svc -n argocd")
	svcArgs := append(baseArgs, "get", "svc", "-n", "argocd")
	result, err = a.executor.ExecuteWithOptions(ctx, executor.ExecuteOptions{
		Command: "kubectl",
		Args:    svcArgs,
	})
	if err != nil {
		return fmt.Errorf("failed to get ArgoCD services: %w", err)
	}
	fmt.Println(result.Stdout)

	pterm.Success.Println("Kubectl verification checks completed")
	return nil
}

// waitForStabilization waits for 10 minutes to allow ArgoCD to fully stabilize
func (a *ArgoCD) waitForStabilization(ctx context.Context, _ config.ChartInstallConfig) error {
	stabilizationTime := 10 * time.Minute

	pterm.Info.Printf("Waiting %v for ArgoCD to stabilize...\n", stabilizationTime)

	// Create a progress bar for the wait
	progressBar, _ := pterm.DefaultProgressbar.
		WithTotal(int(stabilizationTime.Seconds())).
		WithTitle("ArgoCD stabilization").
		WithRemoveWhenDone(false).
		Start()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	elapsed := 0
	for elapsed < int(stabilizationTime.Seconds()) {
		select {
		case <-ctx.Done():
			if progressBar != nil {
				progressBar.Stop()
			}
			return ctx.Err()
		case <-ticker.C:
			elapsed++
			if progressBar != nil {
				progressBar.Increment()
			}
		}
	}

	if progressBar != nil {
		progressBar.Stop()
	}

	pterm.Success.Println("ArgoCD stabilization complete")
	return nil
}

// WaitForApplications waits for all ArgoCD applications to be ready
func (a *ArgoCD) WaitForApplications(ctx context.Context, config config.ChartInstallConfig) error {
	// Silent waiting - show message only in verbose mode
	if config.Verbose {
		pterm.Info.Println("Waiting for ArgoCD applications...")
	}

	err := a.argoCDManager.WaitForApplications(ctx, config)
	if err != nil {
		// Error details handled by caller - no duplicate error message needed
		return errors.NewRecoverableChartError("waiting", "ArgoCD applications", err, 60*time.Second).WithCluster(config.ClusterName)
	}

	// Success message removed - handled by calling service
	return nil
}

// IsInstalled checks if ArgoCD is installed
func (a *ArgoCD) IsInstalled(ctx context.Context) (bool, error) {
	return a.helmManager.IsChartInstalled(ctx, "argo-cd", "argocd")
}

// GetStatus returns the status of ArgoCD installation
func (a *ArgoCD) GetStatus(ctx context.Context) (models.ChartInfo, error) {
	return a.helmManager.GetChartStatus(ctx, "argo-cd", "argocd")
}
