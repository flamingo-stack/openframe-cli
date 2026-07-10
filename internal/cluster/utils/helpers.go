package utils

import (
	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
)

// ClusterSelectionResult contains the result of cluster selection (deprecated - use UI types)
type ClusterSelectionResult struct {
	Name string
	Type models.ClusterType
}
