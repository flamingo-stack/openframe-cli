package provider

import (
	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/providers/k3d"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
)

// New returns the Provider for the given cluster type. An empty type defaults
// to k3d (the local development default). GKE/EKS are recognized types whose
// backends are not implemented yet, so they return ErrProviderNotFound; an
// unrecognized type is a configuration error.
func New(clusterType models.ClusterType, exec executor.CommandExecutor) (Provider, error) {
	switch clusterType {
	case models.ClusterTypeK3d, "":
		return k3d.CreateClusterManagerWithExecutor(exec), nil
	case models.ClusterTypeGKE, models.ClusterTypeEKS:
		return nil, models.NewProviderNotFoundError(clusterType)
	default:
		return nil, models.NewInvalidConfigError("type", clusterType, "unknown cluster type")
	}
}
