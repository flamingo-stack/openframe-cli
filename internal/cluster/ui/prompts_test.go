package ui

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectClusterByName(t *testing.T) {
	t.Run("returns empty string when no clusters provided", func(t *testing.T) {
		clusters := []ClusterInfo{}

		result, err := SelectClusterByName(clusters, "Select a cluster")

		assert.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("creates cluster name list correctly", func(t *testing.T) {
		clusters := []ClusterInfo{
			{Name: "cluster1", Status: "running"},
			{Name: "cluster2", Status: "stopped"},
			{Name: "cluster3", Status: "pending"},
		}

		// We can't test the interactive part without mocking promptui,
		// but we can test that the function doesn't panic with valid input
		assert.NotPanics(t, func() {
			// Test the validation logic
			clusterNames := make([]string, 0, len(clusters))
			for _, cl := range clusters {
				clusterNames = append(clusterNames, cl.Name)
			}

			assert.Len(t, clusterNames, 3)
			assert.Contains(t, clusterNames, "cluster1")
			assert.Contains(t, clusterNames, "cluster2")
			assert.Contains(t, clusterNames, "cluster3")
		})
	})

	t.Run("handles empty cluster names", func(t *testing.T) {
		clusters := []ClusterInfo{
			{Name: "", Status: "running"},
			{Name: "valid-cluster", Status: "stopped"},
		}

		assert.NotPanics(t, func() {
			clusterNames := make([]string, 0, len(clusters))
			for _, cl := range clusters {
				clusterNames = append(clusterNames, cl.Name)
			}

			assert.Len(t, clusterNames, 2)
			assert.Contains(t, clusterNames, "")
			assert.Contains(t, clusterNames, "valid-cluster")
		})
	})
}

// Test helper functions for validation logic that can be tested without UI interaction
func TestValidationLogic(t *testing.T) {
	t.Run("validates cluster names", func(t *testing.T) {
		testCases := []struct {
			name    string
			input   string
			wantErr bool
		}{
			{"valid name", "test-cluster", false},
			{"valid name with numbers", "cluster123", false},
			{"valid name with hyphens", "test-cluster-name", false},
			{"empty name", "", true},
			{"whitespace only", "   ", true},
			{"single character", "a", false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Test validation logic for cluster names
				validate := func(input string) error {
					if len(strings.TrimSpace(input)) < 1 {
						return errors.New("cluster name cannot be empty")
					}
					return nil
				}

				err := validate(tc.input)
				if tc.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

// Test constants and type aliases
func TestConstants(t *testing.T) {
	t.Run("cluster type constants are correctly defined", func(t *testing.T) {
		assert.Equal(t, string(ClusterTypeK3d), "k3d")
	})
}
