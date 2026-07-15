// Package provider defines the unified cluster-provider abstraction.
//
// A Provider creates and manages Kubernetes clusters. Today only k3d (local) is
// implemented; for the recognized cloud types (GKE, EKS) the factory returns
// ErrProviderNotFound until their backends land. New backends implement the
// same Provider interface, so the rest of the CLI never needs to know which
// backend is used.
package provider

import (
	"context"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/providers/eks"
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

// Compile-time assertions that the backends satisfy Provider.
//
// Backends are selected through New (factory.go). The old decorative factory
// was removed in audit B7 because nothing called it; this one is real —
// ClusterService resolves its backend through it, keyed on ClusterConfig.Type.
var (
	_ Provider = (*k3d.K3dManager)(nil)
	_ Provider = (*eks.Provider)(nil)
)
