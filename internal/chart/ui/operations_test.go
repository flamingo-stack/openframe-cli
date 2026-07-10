package ui

import (
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/tests/testutil"
	"github.com/stretchr/testify/assert"
)

func init() {
	testutil.InitializeTestMode()
}

func TestNewOperationsUI(t *testing.T) {
	ui := NewOperationsUI()
	assert.NotNil(t, ui, "NewOperationsUI should not return nil")
}

func TestSelectClusterForInstall_WithClusterArgument(t *testing.T) {
	ui := NewOperationsUI()

	clusters := []models.ClusterInfo{
		{Name: "cluster1", Status: "running"},
		{Name: "cluster2", Status: "stopped"},
	}

	tests := []struct {
		name         string
		args         []string
		clusters     []models.ClusterInfo
		expectedName string
		expectError  bool
	}{
		{
			name:         "valid cluster name",
			args:         []string{"cluster1"},
			clusters:     clusters,
			expectedName: "cluster1",
			expectError:  false,
		},
		{
			name:         "empty cluster name",
			args:         []string{""},
			clusters:     clusters,
			expectedName: "",
			expectError:  true,
		},
		{
			name:         "whitespace cluster name",
			args:         []string{" \t "},
			clusters:     clusters,
			expectedName: "",
			expectError:  true,
		},
		{
			name:         "non-existent cluster",
			args:         []string{"nonexistent"},
			clusters:     clusters,
			expectedName: "",
			expectError:  true,
		},
		{
			name:         "valid cluster name with multiple args",
			args:         []string{"cluster2", "extra"},
			clusters:     clusters,
			expectedName: "cluster2",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selectedCluster, err := ui.SelectClusterForInstall(tt.clusters, tt.args)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, selectedCluster)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedName, selectedCluster)
			}
		})
	}
}

func TestSelectClusterForInstall_InteractiveMode(t *testing.T) {
	ui := NewOperationsUI()

	clusters := []models.ClusterInfo{
		{Name: "cluster1", Status: "running"},
		{Name: "cluster2", Status: "stopped"},
	}

	// Test interactive mode (no args provided)
	// In test mode, this will fail with user cancellation (^D)
	selectedCluster, err := ui.SelectClusterForInstall(clusters, []string{})

	// Interactive mode should fail in test environment (user cancellation)
	assert.Error(t, err, "Interactive mode should error in test environment due to ^D")
	assert.Contains(t, err.Error(), "cluster selection failed")
	assert.Empty(t, selectedCluster, "Interactive mode should return empty when cancelled")
}

func TestSelectClusterForInstall_EmptyClusterList(t *testing.T) {
	ui := NewOperationsUI()

	// Test with no clusters available - the cluster selector will show a message and return empty
	selectedCluster, err := ui.SelectClusterForInstall([]models.ClusterInfo{}, []string{"cluster1"})

	// Since this delegates to cluster selector, we expect either error or empty string
	if err != nil {
		assert.Contains(t, err.Error(), "cluster")
	}
	// Either way, selected cluster should be empty
	assert.Empty(t, selectedCluster)
}


func TestShowNoClusterMessage(t *testing.T) {
	ui := NewOperationsUI()

	// This method outputs to terminal, we just test it doesn't panic
	assert.NotPanics(t, func() {
		ui.ShowNoClusterMessage()
	})
}






