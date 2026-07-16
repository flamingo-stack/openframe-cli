package cluster

import (
	"context"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClusterService_getK3dClusterNodes(t *testing.T) {
	tests := []struct {
		name         string
		clusterName  string
		dockerOutput string
		shouldFail   bool
		expected     []string
		expectError  bool
	}{
		{
			name:        "empty cluster name",
			clusterName: "",
			expectError: true,
		},
		{
			name:         "no nodes found",
			clusterName:  "test-cluster",
			dockerOutput: "",
			expected:     []string{},
			expectError:  false,
		},
		{
			name:        "docker command fails",
			clusterName: "test-cluster",
			shouldFail:  true,
			expectError: true,
		},
		{
			name:        "successful node discovery",
			clusterName: "test-cluster",
			dockerOutput: `k3d-test-cluster-server-0
k3d-test-cluster-agent-0
k3d-test-cluster-agent-1
k3d-test-cluster-serverlb
k3d-test-cluster-tools`,
			expected: []string{
				"k3d-test-cluster-server-0",
				"k3d-test-cluster-agent-0",
				"k3d-test-cluster-agent-1",
			},
			expectError: false,
		},
		{
			name:        "mixed valid and invalid nodes",
			clusterName: "my-cluster",
			dockerOutput: `k3d-my-cluster-server-0
k3d-my-cluster-agent-0
k3d-my-cluster-serverlb
k3d-my-cluster-tools
some-other-container`,
			expected: []string{
				"k3d-my-cluster-server-0",
				"k3d-my-cluster-agent-0",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock executor
			mockExec := executor.NewMockCommandExecutor()

			if tt.shouldFail {
				mockExec.SetShouldFail(true, "docker command failed")
			} else if tt.clusterName != "" {
				// Set up expected command call
				mockExec.SetResponse("docker ps", &executor.CommandResult{
					Stdout: tt.dockerOutput,
				})
			}

			service := NewClusterService(mockExec)

			result, err := service.getK3dClusterNodes(context.Background(), tt.clusterName)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestClusterService_filterK3dNodes(t *testing.T) {
	service := NewClusterService(executor.NewMockCommandExecutor())

	tests := []struct {
		name        string
		output      string
		clusterName string
		expected    []string
	}{
		{
			name:        "empty output",
			output:      "",
			clusterName: "test",
			expected:    []string{},
		},
		{
			name:        "whitespace only",
			output:      "   \n  \n  ",
			clusterName: "test",
			expected:    []string{},
		},
		{
			name:        "valid server and agent nodes",
			output:      "k3d-test-server-0\nk3d-test-agent-0\nk3d-test-agent-1",
			clusterName: "test",
			expected:    []string{"k3d-test-server-0", "k3d-test-agent-0", "k3d-test-agent-1"},
		},
		{
			name:        "mixed valid and invalid nodes",
			output:      "k3d-test-server-0\nk3d-test-serverlb\nk3d-test-agent-0\nk3d-test-tools\nother-container",
			clusterName: "test",
			expected:    []string{"k3d-test-server-0", "k3d-test-agent-0"},
		},
		{
			name:        "nodes with extra whitespace",
			output:      "  k3d-test-server-0  \n  k3d-test-agent-0  \n",
			clusterName: "test",
			expected:    []string{"k3d-test-server-0", "k3d-test-agent-0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.filterK3dNodes(tt.output, tt.clusterName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestClusterService_isK3dWorkerNode(t *testing.T) {
	service := NewClusterService(executor.NewMockCommandExecutor())

	tests := []struct {
		name        string
		nodeName    string
		clusterName string
		expected    bool
	}{
		// Valid worker nodes
		{
			name:        "server node",
			nodeName:    "k3d-test-cluster-server-0",
			clusterName: "test-cluster",
			expected:    true,
		},
		{
			name:        "agent node",
			nodeName:    "k3d-test-cluster-agent-0",
			clusterName: "test-cluster",
			expected:    true,
		},
		{
			name:        "agent node with high number",
			nodeName:    "k3d-test-cluster-agent-5",
			clusterName: "test-cluster",
			expected:    true,
		},

		// Invalid nodes (infrastructure containers)
		{
			name:        "load balancer",
			nodeName:    "k3d-test-cluster-serverlb",
			clusterName: "test-cluster",
			expected:    false,
		},
		{
			name:        "tools container",
			nodeName:    "k3d-test-cluster-tools",
			clusterName: "test-cluster",
			expected:    false,
		},

		// Wrong cluster or format
		{
			name:        "wrong cluster prefix",
			nodeName:    "k3d-other-cluster-server-0",
			clusterName: "test-cluster",
			expected:    false,
		},
		{
			name:        "no k3d prefix",
			nodeName:    "test-cluster-server-0",
			clusterName: "test-cluster",
			expected:    false,
		},
		{
			name:        "completely different container",
			nodeName:    "nginx-container",
			clusterName: "test-cluster",
			expected:    false,
		},

		// Edge cases
		{
			name:        "empty node name",
			nodeName:    "",
			clusterName: "test-cluster",
			expected:    false,
		},
		{
			name:        "empty cluster name",
			nodeName:    "k3d-test-server-0",
			clusterName: "",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.isK3dWorkerNode(tt.nodeName, tt.clusterName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestClusterService_cleanupNodeImages_UsesCrictlNotDocker is the regression
// guard for the no-op prune.
//
// k3d nodes are rancher/k3s containers running containerd; they contain no
// docker binary, so `docker exec <node> docker image prune` exited 127 on every
// node of every cleanup. The failures were swallowed, so cleanup reported
// success and reclaimed nothing. The previous version of this test asserted
// that `docker image prune` WAS executed — it pinned the broken behaviour.
func TestClusterService_cleanupNodeImages_UsesCrictlNotDocker(t *testing.T) {
	mockExec := executor.NewMockCommandExecutor()

	// Mock the node discovery
	mockExec.SetResponse("docker ps --filter label=k3d.cluster=test-cluster --filter status=running --format {{.Names}}", &executor.CommandResult{
		Stdout: "k3d-test-cluster-server-0\nk3d-test-cluster-agent-0\nk3d-test-cluster-serverlb",
	})

	service := NewClusterService(mockExec)

	pruned, err := service.cleanupNodeImages(context.Background(), "test-cluster", true)
	require.NoError(t, err)
	assert.Equal(t, 2, pruned, "server and agent nodes are pruned; serverlb is not a cluster node")

	assert.True(t, mockExec.WasCommandExecuted("docker ps"))
	assert.True(t, mockExec.WasCommandExecuted("docker exec k3d-test-cluster-server-0 crictl rmi --prune"))
	assert.True(t, mockExec.WasCommandExecuted("docker exec k3d-test-cluster-agent-0 crictl rmi --prune"))

	// The load balancer is not a cluster node.
	assert.False(t, mockExec.WasCommandExecuted("docker exec k3d-test-cluster-serverlb crictl rmi --prune"))

	// No command may invoke a docker binary INSIDE a node: there isn't one.
	for _, cmd := range mockExec.GetExecutedCommands() {
		assert.NotContains(t, cmd, "docker exec k3d-test-cluster-server-0 docker",
			"k3d nodes run containerd and ship no docker binary")
		assert.NotContains(t, cmd, "container prune", "containerd has no container prune")
		assert.NotContains(t, cmd, "volume prune", "the node's volumes belong to the host daemon")
		assert.NotContains(t, cmd, "network prune", "the node's networks belong to the host daemon")
	}
}

// TestClusterService_cleanupNodeImages_ReportsFailure: a node whose prune fails
// must be counted as failed, not silently as cleaned.
func TestClusterService_cleanupNodeImages_ReportsFailure(t *testing.T) {
	mockExec := executor.NewMockCommandExecutor()
	mockExec.SetResponse("docker ps --filter label=k3d.cluster=test-cluster --filter status=running --format {{.Names}}", &executor.CommandResult{
		Stdout: "k3d-test-cluster-server-0\nk3d-test-cluster-agent-0",
	})
	mockExec.SetResponse("docker exec k3d-test-cluster-agent-0 crictl rmi --prune", &executor.CommandResult{
		ExitCode: 1,
		Stderr:   "connection refused",
	})

	service := NewClusterService(mockExec)

	pruned, err := service.cleanupNodeImages(context.Background(), "test-cluster", false)
	assert.Equal(t, 1, pruned, "only the node that actually pruned may be counted")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "k3d-test-cluster-agent-0")
}
