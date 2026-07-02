package cluster

import (
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
)

func TestClustersToJSON(t *testing.T) {
	in := []models.ClusterInfo{
		{Name: "dev", Type: models.ClusterTypeK3d, Status: "running", NodeCount: 3, K8sVersion: "v1.31.0"},
		{Name: "empty"},
	}
	out := clustersToJSON(in)

	if len(out) != 2 {
		t.Fatalf("len = %d, want 2", len(out))
	}
	if out[0].Name != "dev" || out[0].Type != "k3d" || out[0].Status != "running" || out[0].NodeCount != 3 || out[0].K8sVersion != "v1.31.0" {
		t.Fatalf("first cluster mapped wrong: %+v", out[0])
	}
	if out[1].Name != "empty" || out[1].K8sVersion != "" {
		t.Fatalf("second cluster mapped wrong: %+v", out[1])
	}
}

func TestListCommandHasOutputFlag(t *testing.T) {
	if getListCmd().Flags().Lookup("output") == nil {
		t.Fatal("cluster list is missing the --output flag")
	}
}
