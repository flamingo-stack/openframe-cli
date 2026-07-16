// Package provider defines the unified cluster-provider abstraction.
//
// A Provider creates and manages Kubernetes clusters. Today only k3d (local) is
// implemented; cloud providers (GKE, EKS) are placeholders that return a
// friendly "coming soon" error. New backends implement the same Provider
// interface, so the rest of the CLI never needs to know which backend is used.
package provider

import (
	"context"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/providers/k3d"
	"k8s.io/client-go/rest"
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
//
// NOTE: there is deliberately NO factory here. The old New(clusterType,
// target, ...) "single seam" was never called from production — every
// constructor hard-coded the k3d manager, so the factory was decorative
// (audit B7). The interface itself is the real seam: it is what
// ClusterService depends on and what tests mock. When a second backend
// (GKE/EKS) actually lands, reintroduce a factory alongside its first caller.
var _ Provider = (*k3d.K3dManager)(nil)
