package provider

import (
	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/providers/eks"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/providers/k3d"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/pterm/pterm"
)

// New returns the Provider for the given cluster type. An empty type defaults
// to k3d (the local development default). GKE is a recognized type without a
// backend yet, so it returns ErrProviderNotFound; an unrecognized type is a
// configuration error.
func New(clusterType models.ClusterType, exec executor.CommandExecutor) (Provider, error) {
	switch clusterType {
	case models.ClusterTypeK3d, "":
		return k3d.CreateClusterManagerWithExecutor(exec), nil
	case models.ClusterTypeEKS:
		// pterm's debug switch is the CLI-wide --verbose signal; it makes the
		// terraform engine stream terraform's own output during long applies.
		return eks.New(exec, pterm.PrintDebugMessages)
	case models.ClusterTypeGKE:
		return nil, models.NewProviderNotFoundError(clusterType)
	default:
		return nil, models.NewInvalidConfigError("type", clusterType, "unknown cluster type")
	}
}
