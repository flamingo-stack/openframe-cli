package types

import (
	"context"

	"github.com/flamingo-stack/openframe-cli/internal/chart/models"
	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
	clusterDomain "github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"k8s.io/client-go/rest"
)

// This file keeps ONLY the interfaces that are actually implemented and
// consumed. It used to declare 13 — a speculative service-locator layer
// (ServiceFactory, ServiceOrchestrator, WorkflowExecutor, OperationsUI, ...)
// with zero implementations or callers (audit B7).

// ClusterLister provides cluster listing capabilities
type ClusterLister interface {
	ListClusters() ([]clusterDomain.ClusterInfo, error)
}

// ClusterAccess provides the read-only cluster capabilities the app subsystem
// needs — listing clusters and resolving a rest.Config for one — WITHOUT
// depending on cluster-creation code. Keeping the app (chart) subsystem behind
// this interface isolates it from the cluster subsystem (req 18/19); the
// concrete implementation is injected by the composition root (cmd / bootstrap),
// which is allowed to import both.
type ClusterAccess interface {
	ClusterLister
	GetRestConfig(name string) (*rest.Config, error)
}

// ArgoCDService manages ArgoCD installation and lifecycle
type ArgoCDService interface {
	Install(ctx context.Context, config config.ChartInstallConfig) error
	IsInstalled(ctx context.Context) (bool, error)
	GetStatus(ctx context.Context) (models.ChartInfo, error)
	WaitForApplications(ctx context.Context, config config.ChartInstallConfig) error
}

// AppOfAppsService manages app-of-apps installation and lifecycle
type AppOfAppsService interface {
	Install(ctx context.Context, config config.ChartInstallConfig) error
	IsInstalled(ctx context.Context, namespace string) (bool, error)
	GetStatus(ctx context.Context, namespace string) (models.ChartInfo, error)
}

// InstallationRequest contains all parameters for chart installation
type InstallationRequest struct {
	Args           []string
	Force          bool
	DryRun         bool
	Verbose        bool
	GitHubRepo     string
	GitHubBranch   string
	// GitHubRefExplicit is true when the operator explicitly set --ref.
	// When set, GitHubBranch is pinned into the helm values (repository.branch) so it
	// wins over the values-file branch and both the app-of-apps clone and the child
	// Applications' targetRevision track that ref.
	GitHubRefExplicit bool
	CertDir           string
	NonInteractive    bool // Skip all prompts, use existing openframe-helm-values.yaml
	// RequireExistingValues makes a missing openframe-helm-values.yaml a hard
	// error instead of "deploy chart defaults". Set by upgrade (Mode 1): an
	// upgrade with an empty values map would replace the release values with
	// chart defaults, silently wiping registry credentials and ingress settings
	// when run from the wrong directory (audit F3/T1-2). Fresh installs and
	// bootstrap keep the defaults-with-warning behavior — a clean machine has no
	// values file yet.
	RequireExistingValues bool
	// SyncStragglersOnStall: on the upgrade (ref-change) path, let the
	// application wait sync OutOfSync-but-healthy stragglers once progress
	// stalls (children with autoSync off never pick a new ref up themselves).
	SyncStragglersOnStall bool
	KubeConfig            *rest.Config // Kubernetes REST config for cluster communication
	// KubeContext is the kube-context name KubeConfig was resolved from
	// (--context or the interactive target selector). When set, every helm CLI
	// call targets it too, so the helm CLI, the native client checks, and the
	// ArgoCD wait all watch the SAME cluster (audit F4: three different targets
	// could be used within a single install).
	KubeContext string
	// ClusterAccess resolves clusters and their rest.Config for the install
	// target. Injected by the composition root so the app subsystem never imports
	// cluster-creation code (req 18/19). Required for interactive/named-cluster
	// installs; may be nil when KubeConfig is supplied directly.
	ClusterAccess ClusterAccess
}
