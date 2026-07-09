// Package provider defines the unified cluster-provider abstraction.
//
// A Provider creates and manages Kubernetes clusters. Today only k3d (local) is
// implemented; cloud providers (GKE, EKS) are placeholders that return a
// friendly "coming soon" error. New backends implement the same Provider
// interface, so the rest of the CLI never needs to know which backend is used.
package provider

import (
	"context"
	"fmt"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/providers/k3d"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"k8s.io/client-go/rest"
)

// Target is where a cluster runs.
type Target string

const (
	// TargetLocal is a cluster on the local machine (e.g. k3d).
	TargetLocal Target = "local"
	// TargetCloud is a managed cluster in a cloud provider (e.g. GKE/EKS). Not yet implemented.
	TargetCloud Target = "cloud"
)

// Provider is the unified contract every cluster backend implements. The k3d
// manager satisfies it today (see the compile-time assertion below); GKE/EKS
// will implement the same interface when added.
type Provider interface {
	// CreateCluster creates a cluster and returns a rest.Config for reaching it.
	CreateCluster(ctx context.Context, config models.ClusterConfig) (*rest.Config, error)
	// DeleteCluster removes a cluster.
	DeleteCluster(ctx context.Context, name string, clusterType models.ClusterType, force bool) error
	// StartCluster starts a stopped cluster.
	StartCluster(ctx context.Context, name string, clusterType models.ClusterType) error
	// ListClusters returns the clusters managed by this provider.
	ListClusters(ctx context.Context) ([]models.ClusterInfo, error)
	// ListAllClusters returns all clusters visible to this provider.
	ListAllClusters(ctx context.Context) ([]models.ClusterInfo, error)
	// GetClusterStatus returns detailed status for a single cluster.
	GetClusterStatus(ctx context.Context, name string) (models.ClusterInfo, error)
	// DetectClusterType reports what kind of cluster a given name is.
	DetectClusterType(ctx context.Context, name string) (models.ClusterType, error)
	// GetRestConfig returns a rest.Config for an existing cluster.
	GetRestConfig(ctx context.Context, name string) (*rest.Config, error)
	// GetKubeconfig returns the kubeconfig for a cluster.
	GetKubeconfig(ctx context.Context, name string, clusterType models.ClusterType) (string, error)
}

// Compile-time assertion that the k3d manager satisfies Provider.
var _ Provider = (*k3d.K3dManager)(nil)

// New returns the Provider for the given cluster type and target.
//
// Only (k3d, local) is implemented. Cloud providers return a clear "coming
// soon" error so callers can surface a friendly message instead of failing
// obscurely. This is the single seam through which new providers are added.
func New(clusterType models.ClusterType, target Target, exec executor.CommandExecutor, verbose bool) (Provider, error) {
	switch clusterType {
	case models.ClusterTypeK3d:
		if target != TargetLocal {
			return nil, fmt.Errorf("the k3d provider only supports the local target, not %q", target)
		}
		return k3d.NewK3dManager(exec, verbose), nil
	case models.ClusterTypeGKE, models.ClusterTypeEKS:
		return nil, fmt.Errorf("the %s cluster provider is not implemented yet — coming soon", clusterType)
	default:
		return nil, fmt.Errorf("unknown cluster provider %q", clusterType)
	}
}
