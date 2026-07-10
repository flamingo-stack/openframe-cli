package ui

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDisplayService(t *testing.T) {
	t.Run("creates new display service", func(t *testing.T) {
		service := NewDisplayService()
		assert.NotNil(t, service)
		assert.IsType(t, &DisplayService{}, service)
	})
}

func TestDisplayService_ShowClusterList(t *testing.T) {
	t.Run("displays cluster list with multiple clusters", func(t *testing.T) {
		service := NewDisplayService()
		var buf bytes.Buffer

		clusters := []ClusterDisplayInfo{
			{
				Name:      "cluster1",
				Type:      "k3d",
				Status:    "running",
				NodeCount: 3,
				CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			{
				Name:      "cluster2",
				Type:      "gke",
				Status:    "stopped",
				NodeCount: 5,
				CreatedAt: time.Date(2023, 1, 2, 14, 30, 0, 0, time.UTC),
			},
		}

		service.ShowClusterList(clusters, &buf)

		output := buf.String()
		assert.Contains(t, output, "cluster1")
		assert.Contains(t, output, "cluster2")
		assert.Contains(t, output, "k3d")
		assert.Contains(t, output, "gke")
		assert.Contains(t, output, "running")
		assert.Contains(t, output, "stopped")
		assert.Contains(t, output, "3")
		assert.Contains(t, output, "5")
	})

	t.Run("displays no clusters message when list is empty", func(t *testing.T) {
		service := NewDisplayService()
		var buf bytes.Buffer

		clusters := []ClusterDisplayInfo{}

		service.ShowClusterList(clusters, &buf)

		output := buf.String()
		assert.Contains(t, output, "No clusters found.")
	})

	t.Run("handles single cluster", func(t *testing.T) {
		service := NewDisplayService()
		var buf bytes.Buffer

		clusters := []ClusterDisplayInfo{
			{
				Name:      "single-cluster",
				Type:      "k3d",
				Status:    "pending",
				NodeCount: 1,
				CreatedAt: time.Date(2023, 6, 15, 10, 0, 0, 0, time.UTC),
			},
		}

		service.ShowClusterList(clusters, &buf)

		output := buf.String()
		assert.Contains(t, output, "single-cluster")
		assert.Contains(t, output, "pending")
		assert.Contains(t, output, "1")
	})

	t.Run("formats table headers correctly", func(t *testing.T) {
		service := NewDisplayService()
		var buf bytes.Buffer

		clusters := []ClusterDisplayInfo{
			{
				Name:      "test",
				Type:      "k3d",
				Status:    "running",
				NodeCount: 2,
				CreatedAt: time.Now(),
			},
		}

		service.ShowClusterList(clusters, &buf)

		output := buf.String()
		lines := strings.Split(output, "\n")
		if len(lines) > 0 {
			headerLine := lines[0]
			assert.Contains(t, headerLine, "NAME")
			assert.Contains(t, headerLine, "TYPE")
			assert.Contains(t, headerLine, "STATUS")
			assert.Contains(t, headerLine, "NODES")
			assert.Contains(t, headerLine, "CREATED")
		}
	})
}

func TestClusterDisplayInfo(t *testing.T) {
	t.Run("creates cluster display info with all fields", func(t *testing.T) {
		createdAt := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		nodes := []NodeDisplayInfo{
			{Name: "node1", Role: "control-plane", Status: "ready"},
			{Name: "node2", Role: "worker", Status: "ready"},
		}

		info := ClusterDisplayInfo{
			Name:      "test-cluster",
			Type:      "k3d",
			Status:    "running",
			NodeCount: 2,
			CreatedAt: createdAt,
			Nodes:     nodes,
		}

		assert.Equal(t, "test-cluster", info.Name)
		assert.Equal(t, "k3d", info.Type)
		assert.Equal(t, "running", info.Status)
		assert.Equal(t, 2, info.NodeCount)
		assert.Equal(t, createdAt, info.CreatedAt)
		assert.Len(t, info.Nodes, 2)
		assert.Equal(t, "node1", info.Nodes[0].Name)
		assert.Equal(t, "control-plane", info.Nodes[0].Role)
		assert.Equal(t, "ready", info.Nodes[0].Status)
	})
}

func TestNodeDisplayInfo(t *testing.T) {
	t.Run("creates node display info", func(t *testing.T) {
		node := NodeDisplayInfo{
			Name:   "worker-node-1",
			Role:   "worker",
			Status: "ready",
		}

		assert.Equal(t, "worker-node-1", node.Name)
		assert.Equal(t, "worker", node.Role)
		assert.Equal(t, "ready", node.Status)
	})
}
