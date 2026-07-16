package ui

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	sharedUI "github.com/flamingo-stack/openframe-cli/internal/shared/ui"
	"github.com/pterm/pterm"
)

// Selector handles cluster selection logic across different commands
type Selector struct {
	operation string
}

// NewSelector creates a new cluster selector for the given operation
func NewSelector(operation string) *Selector {
	return &Selector{
		operation: operation,
	}
}

// SelectCluster handles cluster selection with consistent logic
// Supports both argument-based and interactive selection
func (s *Selector) SelectCluster(clusters []models.ClusterInfo, args []string) (string, error) {
	// Validate input
	if len(clusters) == 0 {
		s.showNoClusterMessage()
		return "", nil
	}

	// If cluster name provided as argument, validate and use it
	if len(args) > 0 {
		clusterName := strings.TrimSpace(args[0])
		if clusterName == "" {
			return "", fmt.Errorf("cluster name cannot be empty")
		}

		// Validate that the cluster exists
		for _, cluster := range clusters {
			if cluster.Name == clusterName {
				return clusterName, nil
			}
		}
		return "", fmt.Errorf("cluster '%s' not found", clusterName)
	}

	// Use interactive selection
	clusterNames := make([]string, len(clusters))
	for i, cluster := range clusters {
		clusterNames[i] = cluster.Name
	}

	prompt := fmt.Sprintf("Select cluster for %s", s.operation)
	_, selectedCluster, err := sharedUI.SelectFromList(prompt, clusterNames)
	if err != nil {
		return "", fmt.Errorf("cluster selection failed: %w", err)
	}

	if selectedCluster == "" {
		s.showOperationCancelled()
		return "", nil
	}

	return selectedCluster, nil
}

// showNoClusterMessage displays a message when no clusters are available
func (s *Selector) showNoClusterMessage() {
	pterm.Error.Println("No clusters found. Create a cluster first with: openframe cluster create")
}

// showOperationCancelled displays a cancellation message
func (s *Selector) showOperationCancelled() {
	pterm.Info.Printf("No cluster selected. %s cancelled.\n", capitalizeFirst(s.operation))
}

// capitalizeFirst upper-cases the first rune of a single-word label.
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
