package ui

import (
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
)

// The picker rows must carry enough columns to tell a local k3d apart from a
// same-named cloud cluster; row order must mirror the input so the selected
// index maps back to the right cluster.
func TestClusterPickerRows(t *testing.T) {
	rows := clusterPickerRows([]ClusterInfo{
		{Name: "dev", Type: models.ClusterTypeK3d, NodeCount: 3, Status: "Ready"},
		{Name: "dev-gke", Type: models.ClusterTypeGKE, Region: "us-central1", NodeCount: 6, Status: "Creating"},
	})
	if len(rows) != 2 {
		t.Fatalf("rows = %v", rows)
	}
	if !strings.Contains(rows[0], "dev") || !strings.Contains(rows[0], "K3D") ||
		!strings.Contains(rows[0], "local") || !strings.Contains(rows[0], "Ready") {
		t.Fatalf("k3d row lacks columns: %q", rows[0])
	}
	if !strings.Contains(rows[1], "GKE") || !strings.Contains(rows[1], "us-central1") ||
		!strings.Contains(rows[1], "6 node(s)") || !strings.Contains(rows[1], "Creating") {
		t.Fatalf("gke row lacks columns: %q", rows[1])
	}
}
