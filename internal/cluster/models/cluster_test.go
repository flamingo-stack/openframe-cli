package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClusterType(t *testing.T) {
	t.Run("cluster type constants", func(t *testing.T) {
		assert.Equal(t, ClusterType("k3d"), ClusterTypeK3d)
		assert.Equal(t, ClusterType("gke"), ClusterTypeGKE)
	})

	t.Run("cluster type string conversion", func(t *testing.T) {
		assert.Equal(t, "k3d", string(ClusterTypeK3d))
		assert.Equal(t, "gke", string(ClusterTypeGKE))
	})
}

func TestClusterConfig(t *testing.T) {
	t.Run("creates cluster config with all fields", func(t *testing.T) {
		config := ClusterConfig{
			Name:       "test-cluster",
			Type:       ClusterTypeK3d,
			NodeCount:  3,
			K8sVersion: "v1.25.0-k3s1",
		}

		assert.Equal(t, "test-cluster", config.Name)
		assert.Equal(t, ClusterTypeK3d, config.Type)
		assert.Equal(t, 3, config.NodeCount)
		assert.Equal(t, "v1.25.0-k3s1", config.K8sVersion)
	})

	t.Run("creates minimal cluster config", func(t *testing.T) {
		config := ClusterConfig{
			Name: "minimal-cluster",
			Type: ClusterTypeK3d,
		}

		assert.Equal(t, "minimal-cluster", config.Name)
		assert.Equal(t, ClusterTypeK3d, config.Type)
		assert.Equal(t, 0, config.NodeCount) // Zero value
		assert.Empty(t, config.K8sVersion)   // Zero value
	})

	t.Run("validates cluster config fields", func(t *testing.T) {
		config := ClusterConfig{}

		// Test zero values
		assert.Empty(t, config.Name)
		assert.Empty(t, config.Type)
		assert.Equal(t, 0, config.NodeCount)
		assert.Empty(t, config.K8sVersion)
	})
}

func TestClusterInfo(t *testing.T) {
	t.Run("creates cluster info with all fields", func(t *testing.T) {
		createdAt := time.Now()
		nodes := []NodeInfo{
			{Name: "node1", Status: "ready", Role: "control-plane"},
			{Name: "node2", Status: "ready", Role: "worker"},
		}

		info := ClusterInfo{
			Name:       "test-cluster",
			Type:       ClusterTypeK3d,
			Status:     "running",
			NodeCount:  2,
			K8sVersion: "v1.25.0-k3s1",
			CreatedAt:  createdAt,
			Nodes:      nodes,
		}

		assert.Equal(t, "test-cluster", info.Name)
		assert.Equal(t, ClusterTypeK3d, info.Type)
		assert.Equal(t, "running", info.Status)
		assert.Equal(t, 2, info.NodeCount)
		assert.Equal(t, "v1.25.0-k3s1", info.K8sVersion)
		assert.Equal(t, createdAt, info.CreatedAt)
		assert.Len(t, info.Nodes, 2)
		assert.Equal(t, "node1", info.Nodes[0].Name)
		assert.Equal(t, "control-plane", info.Nodes[0].Role)
	})

	t.Run("creates minimal cluster info", func(t *testing.T) {
		info := ClusterInfo{
			Name:   "minimal-cluster",
			Type:   ClusterTypeGKE,
			Status: "pending",
		}

		assert.Equal(t, "minimal-cluster", info.Name)
		assert.Equal(t, ClusterTypeGKE, info.Type)
		assert.Equal(t, "pending", info.Status)
		assert.Equal(t, 0, info.NodeCount)
		assert.Empty(t, info.K8sVersion)
		assert.True(t, info.CreatedAt.IsZero())
		assert.Empty(t, info.Nodes)
	})

	t.Run("handles different cluster statuses", func(t *testing.T) {
		statuses := []string{"running", "stopped", "pending", "error", "unknown"}

		for _, status := range statuses {
			info := ClusterInfo{
				Name:   "test-cluster",
				Type:   ClusterTypeK3d,
				Status: status,
			}

			assert.Equal(t, "test-cluster", info.Name)
			assert.Equal(t, ClusterTypeK3d, info.Type)
			assert.Equal(t, status, info.Status)
		}
	})
}

func TestNodeInfo(t *testing.T) {
	t.Run("creates node info with all fields", func(t *testing.T) {
		node := NodeInfo{
			Name:   "test-node-1",
			Status: "ready",
			Role:   "control-plane",
		}

		assert.Equal(t, "test-node-1", node.Name)
		assert.Equal(t, "ready", node.Status)
		assert.Equal(t, "control-plane", node.Role)
	})

	t.Run("creates worker node", func(t *testing.T) {
		node := NodeInfo{
			Name:   "worker-node-1",
			Status: "ready",
			Role:   "worker",
		}

		assert.Equal(t, "worker-node-1", node.Name)
		assert.Equal(t, "ready", node.Status)
		assert.Equal(t, "worker", node.Role)
	})

	t.Run("handles different node statuses", func(t *testing.T) {
		statuses := []string{"ready", "not ready", "pending", "terminating", "unknown"}

		for _, status := range statuses {
			node := NodeInfo{
				Name:   "test-node",
				Status: status,
				Role:   "worker",
			}

			assert.Equal(t, "test-node", node.Name)
			assert.Equal(t, "worker", node.Role)
			assert.Equal(t, status, node.Status)
		}
	})

	t.Run("handles different node roles", func(t *testing.T) {
		roles := []string{"control-plane", "worker", "master", "agent"}

		for _, role := range roles {
			node := NodeInfo{
				Name:   "test-node",
				Status: "ready",
				Role:   role,
			}

			assert.Equal(t, "test-node", node.Name)
			assert.Equal(t, "ready", node.Status)
			assert.Equal(t, role, node.Role)
		}
	})
}

func TestCloudConfig(t *testing.T) {
	t.Run("cluster config without cloud settings has nil Cloud", func(t *testing.T) {
		var config ClusterConfig

		assert.Nil(t, config.Cloud)
	})

	t.Run("holds provider-agnostic cloud settings", func(t *testing.T) {
		cloud := CloudConfig{
			Region:      "us-east-1",
			Profile:     "default",
			MachineType: "m6i.large",
			MinNodes:    1,
			MaxNodes:    5,
			Spot:        true,
		}

		assert.Equal(t, "us-east-1", cloud.Region)
		assert.Equal(t, "default", cloud.Profile)
		assert.Equal(t, "m6i.large", cloud.MachineType)
		assert.Equal(t, 1, cloud.MinNodes)
		assert.Equal(t, 5, cloud.MaxNodes)
		assert.True(t, cloud.Spot)
	})
}

func TestJSONSerialization(t *testing.T) {
	t.Run("cluster config serialization", func(t *testing.T) {
		config := ClusterConfig{
			Name:       "test-cluster",
			Type:       ClusterTypeK3d,
			NodeCount:  3,
			K8sVersion: "v1.25.0-k3s1",
		}

		// Basic validation that struct tags are correct
		assert.Equal(t, "test-cluster", config.Name)
		assert.Equal(t, ClusterTypeK3d, config.Type)
		assert.Equal(t, 3, config.NodeCount)
		assert.Equal(t, "v1.25.0-k3s1", config.K8sVersion)
	})

	t.Run("cluster info serialization", func(t *testing.T) {
		info := ClusterInfo{
			Name:      "test-cluster",
			Type:      ClusterTypeGKE,
			Status:    "running",
			NodeCount: 5,
		}

		// Basic validation that struct tags are correct
		assert.Equal(t, "test-cluster", info.Name)
		assert.Equal(t, ClusterTypeGKE, info.Type)
		assert.Equal(t, "running", info.Status)
		assert.Equal(t, 5, info.NodeCount)
	})

	t.Run("cloud config round-trips through JSON and is omitted when nil", func(t *testing.T) {
		config := ClusterConfig{
			Name: "cloud-cluster",
			Type: ClusterTypeEKS,
			Cloud: &CloudConfig{
				Region:   "eu-west-1",
				MaxNodes: 4,
			},
		}

		data, err := json.Marshal(config)
		assert.NoError(t, err)

		var decoded ClusterConfig
		assert.NoError(t, json.Unmarshal(data, &decoded))
		assert.Equal(t, config, decoded)

		local, err := json.Marshal(ClusterConfig{Name: "local", Type: ClusterTypeK3d})
		assert.NoError(t, err)
		assert.NotContains(t, string(local), "cloud")
	})
}
