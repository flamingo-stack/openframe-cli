package manager

import (
	"context"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"k8s.io/client-go/rest"
)

// ClusterManager defines the contract for cluster provider implementations
// This interface abstracts the differences between k3d (Linux/macOS) and kind (Windows)
type ClusterManager interface {
	// CreateCluster creates a new cluster with the given configuration
	// Returns the *rest.Config for the created cluster that can be used to interact with it
	CreateCluster(ctx context.Context, config models.ClusterConfig) (*rest.Config, error)

	// DeleteCluster removes a cluster by name
	DeleteCluster(ctx context.Context, name string, clusterType models.ClusterType, force bool) error

	// StartCluster starts a stopped cluster
	StartCluster(ctx context.Context, name string, clusterType models.ClusterType) error

	// ListClusters returns all clusters managed by this provider
	ListClusters(ctx context.Context) ([]models.ClusterInfo, error)

	// ListAllClusters returns all clusters (alias for backward compatibility)
	ListAllClusters(ctx context.Context) ([]models.ClusterInfo, error)

	// GetClusterStatus returns detailed status information for a specific cluster
	GetClusterStatus(ctx context.Context, name string) (models.ClusterInfo, error)

	// DetectClusterType checks if this provider manages the given cluster
	DetectClusterType(ctx context.Context, name string) (models.ClusterType, error)

	// GetKubeconfig returns the kubeconfig for accessing the cluster
	GetKubeconfig(ctx context.Context, name string, clusterType models.ClusterType) (string, error)

	// GetRestConfig returns the rest.Config for an existing cluster
	GetRestConfig(ctx context.Context, clusterName string) (*rest.Config, error)
}
