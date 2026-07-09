package cluster

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
)

func TestStatusCommandHasOutputFlag(t *testing.T) {
	f := getStatusCmd().Flags().Lookup("output")
	if f == nil {
		t.Fatal("cluster status is missing the --output flag")
	}
	if f.Shorthand != "o" {
		t.Fatalf("--output shorthand = %q, want %q", f.Shorthand, "o")
	}
}

// captureStdout runs fn with os.Stdout redirected to a pipe and returns what it wrote.
func captureStdout(t *testing.T, fn func() error) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	runErr := fn()
	_ = w.Close()
	os.Stdout = old
	if runErr != nil {
		t.Fatalf("printClusterStatus: %v", runErr)
	}
	b, _ := io.ReadAll(r)
	return string(b)
}

func sampleStatus() models.ClusterInfo {
	return models.ClusterInfo{
		Name:       "dev",
		Type:       models.ClusterTypeK3d,
		Status:     "running",
		NodeCount:  2,
		K8sVersion: "v1.31.0",
		Nodes: []models.NodeInfo{
			{Name: "dev-server-0", Status: "Ready", Role: "control-plane"},
		},
	}
}

func TestPrintClusterStatus_JSON(t *testing.T) {
	out := captureStdout(t, func() error { return printClusterStatus(sampleStatus(), "json") })

	// Must be valid JSON round-tripping the ClusterInfo (including nested nodes).
	var got models.ClusterInfo
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}
	if got.Name != "dev" || got.Status != "running" || got.NodeCount != 2 {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
	if len(got.Nodes) != 1 || got.Nodes[0].Role != "control-plane" {
		t.Fatalf("nodes not serialized: %+v", got.Nodes)
	}
}

func TestPrintClusterStatus_YAML(t *testing.T) {
	out := captureStdout(t, func() error { return printClusterStatus(sampleStatus(), "yaml") })

	// sigs.k8s.io/yaml reuses the json tags, so field names are the JSON ones.
	for _, want := range []string{"name: dev", "status: running", "control-plane"} {
		if !strings.Contains(out, want) {
			t.Fatalf("yaml missing %q:\n%s", want, out)
		}
	}
}
