package utils

import (
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/stretchr/testify/assert"
)

func TestClusterSelectionResult(t *testing.T) {
	t.Run("creates cluster selection result", func(t *testing.T) {
		result := ClusterSelectionResult{
			Name: "test-cluster",
			Type: models.ClusterTypeK3d,
		}

		assert.Equal(t, "test-cluster", result.Name)
		assert.Equal(t, models.ClusterTypeK3d, result.Type)
	})

	t.Run("creates cluster selection result with different types", func(t *testing.T) {
		tests := []struct {
			name        string
			clusterType models.ClusterType
		}{
			{"k3d-cluster", models.ClusterTypeK3d},
			{"gke-cluster", models.ClusterTypeGKE},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := ClusterSelectionResult{
					Name: tt.name,
					Type: tt.clusterType,
				}

				assert.Equal(t, tt.name, result.Name)
				assert.Equal(t, tt.clusterType, result.Type)
			})
		}
	})
}

func TestTypeAliases(t *testing.T) {
	t.Run("cluster type aliases work correctly", func(t *testing.T) {
		// Test that the type aliases are correctly set up
		var ct = models.ClusterTypeK3d
		assert.Equal(t, "k3d", string(ct))

		ct = models.ClusterTypeGKE
		assert.Equal(t, "gke", string(ct))
	})

	t.Run("cluster info alias works correctly", func(t *testing.T) {
		info := models.ClusterInfo{
			Name: "test-cluster",
			Type: models.ClusterTypeK3d,
		}

		assert.Equal(t, "test-cluster", info.Name)
		assert.Equal(t, models.ClusterTypeK3d, info.Type)
	})

	t.Run("node info alias works correctly", func(t *testing.T) {
		node := models.NodeInfo{
			Name:   "test-node",
			Status: "ready",
			Role:   "worker",
		}

		assert.Equal(t, "test-node", node.Name)
		assert.Equal(t, "ready", node.Status)
		assert.Equal(t, "worker", node.Role)
	})
}

func TestConstants(t *testing.T) {
	t.Run("cluster type constants are correctly re-exported", func(t *testing.T) {
		// Verify that the constants match the expected string values
		assert.Equal(t, "k3d", string(models.ClusterTypeK3d))
		assert.Equal(t, "gke", string(models.ClusterTypeGKE))
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("handles various whitespace combinations in cluster name validation", func(t *testing.T) {
		testCases := []struct {
			name    string
			input   string
			wantErr bool
			errMsg  string
		}{
			{"normal name", "test-cluster", false, ""},
			{"name with spaces around", "  test-cluster  ", false, ""},
			{"empty string", "", true, "cluster name cannot be empty"},
			{"only spaces", "   ", true, "cluster name cannot be empty or contain only whitespace"},
			{"only tabs", "\t\t\t", true, "cluster name cannot be empty or contain only whitespace"},
			{"only newlines", "\n\n", true, "cluster name cannot be empty or contain only whitespace"},
			{"mixed whitespace", " \t\n ", true, "cluster name cannot be empty or contain only whitespace"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := models.ValidateClusterName(tc.input)
				if tc.wantErr {
					assert.Error(t, err)
					if tc.errMsg != "" {
						assert.Contains(t, err.Error(), tc.errMsg)
					}
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

}
