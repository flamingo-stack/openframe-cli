package cluster

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/providers/terraform"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/utils"
	"github.com/flamingo-stack/openframe-cli/internal/k8s"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
)

// writeUseKubeconfig writes a kubeconfig with the given context names and
// points $KUBECONFIG at it.
func writeUseKubeconfig(t *testing.T, current string, contexts ...string) string {
	t.Helper()
	content := "apiVersion: v1\nkind: Config\ncurrent-context: " + current + "\nclusters:\n- name: c\n  cluster:\n    server: https://x.example\ncontexts:\n"
	for _, name := range contexts {
		content += "- name: " + name + "\n  context:\n    cluster: c\n    user: u\n"
	}
	content += "users:\n- name: u\n"
	path := filepath.Join(t.TempDir(), "kubeconfig")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("KUBECONFIG", path)
	return path
}

func setupUse(t *testing.T) *executor.MockCommandExecutor {
	t.Helper()
	t.Setenv("OPENFRAME_CLUSTERS_DIR", t.TempDir())
	utils.InitGlobalFlags()
	mock := executor.NewMockCommandExecutor()
	utils.SetTestExecutor(mock)
	t.Cleanup(utils.ResetGlobalFlags)
	return mock
}

func TestRunUseCluster_K3d(t *testing.T) {
	mock := setupUse(t)
	mock.SetResponse("k3d cluster get dev", &executor.CommandResult{ExitCode: 0, Stdout: "dev"})
	path := writeUseKubeconfig(t, "other", "other", "k3d-dev")

	if err := runUseCluster(getUseCmd(), []string{"dev"}); err != nil {
		t.Fatalf("use k3d: %v", err)
	}
	_, current, _ := k8s.LoadContexts(path)
	if current != "k3d-dev" {
		t.Fatalf("current-context = %q, want k3d-dev", current)
	}
}

func TestRunUseCluster_ManagedGKEAlignsGcloudConfiguration(t *testing.T) {
	mock := setupUse(t)
	mock.SetResponse("gcloud config configurations list", &executor.CommandResult{ExitCode: 0,
		Stdout: `[{"name":"dev-x","properties":{"core":{"project":"proj-x"}}}]`})
	record := terraform.Record{
		Name: "my-gke", Type: models.ClusterTypeGKE, Status: terraform.StatusReady,
		Region: "us-central1", Project: "proj-x", NodeCount: 3,
	}
	reg := terraform.NewRegistry(os.Getenv("OPENFRAME_CLUSTERS_DIR"))
	if err := reg.Workspace("my-gke").Scaffold(record, nil, nil); err != nil {
		t.Fatal(err)
	}
	path := writeUseKubeconfig(t, "other", "other", "my-gke")

	if err := runUseCluster(getUseCmd(), []string{"my-gke"}); err != nil {
		t.Fatalf("use managed gke: %v", err)
	}
	_, current, _ := k8s.LoadContexts(path)
	if current != "my-gke" {
		t.Fatalf("current-context = %q, want my-gke", current)
	}
	if !mock.WasCommandExecuted("gcloud config configurations activate dev-x") {
		t.Fatal("expected the matching gcloud configuration to be activated")
	}
}

func TestRunUseCluster_ExternalGKEWithExistingContext(t *testing.T) {
	mock := setupUse(t)
	path := writeUseKubeconfig(t, "other", "other", "connectgateway_proj-x_us-central1_ext-1")
	mock.SetResponse("gcloud auth list", &executor.CommandResult{ExitCode: 0, Stdout: "me@example.com\n"})
	mock.SetResponse("gcloud config configurations list", &executor.CommandResult{ExitCode: 0,
		Stdout: `[{"name":"dev-x","properties":{"core":{"project":"proj-x"}}}]`})
	mock.SetResponse("clusters list --project proj-x", &executor.CommandResult{ExitCode: 0,
		Stdout: `[{"name":"ext-1","location":"us-central1","status":"RUNNING","currentNodeCount":2}]`})
	// k3d detection must miss so the flow falls through to discovery.
	mock.SetResponse("k3d cluster get ext-1", &executor.CommandResult{ExitCode: 1, Stderr: "not found"})

	if err := runUseCluster(getUseCmd(), []string{"ext-1"}); err != nil {
		t.Fatalf("use external gke: %v", err)
	}
	_, current, _ := k8s.LoadContexts(path)
	if current != "connectgateway_proj-x_us-central1_ext-1" {
		t.Fatalf("current-context = %q", current)
	}
	if mock.WasCommandExecuted("get-credentials") {
		t.Fatal("credentials must not be re-fetched when a context already exists")
	}
}

func TestRunUseCluster_ExternalGKEFetchesCredentials(t *testing.T) {
	mock := setupUse(t)
	writeUseKubeconfig(t, "other", "other")
	mock.SetResponse("gcloud auth list", &executor.CommandResult{ExitCode: 0, Stdout: "me@example.com\n"})
	mock.SetResponse("gcloud config configurations list", &executor.CommandResult{ExitCode: 0,
		Stdout: `[{"name":"dev-x","properties":{"core":{"project":"proj-x"}}}]`})
	mock.SetResponse("clusters list --project proj-x", &executor.CommandResult{ExitCode: 0,
		Stdout: `[{"name":"ext-1","location":"us-central1","status":"RUNNING","currentNodeCount":2}]`})
	mock.SetResponse("k3d cluster get ext-1", &executor.CommandResult{ExitCode: 1, Stderr: "not found"})

	err := runUseCluster(getUseCmd(), []string{"ext-1"})
	// The mock cannot actually write the gke_* context, so the switch fails —
	// but the credentials fetch must have been attempted first.
	if !mock.WasCommandExecuted("gcloud container clusters get-credentials ext-1") {
		t.Fatal("expected a get-credentials attempt for a context-less external cluster")
	}
	if err == nil || !strings.Contains(err.Error(), "no kubeconfig context") {
		t.Fatalf("expected a missing-context error after mock fetch, got: %v", err)
	}
}

func TestRunUseCluster_NotAuthenticatedIsActionable(t *testing.T) {
	mock := setupUse(t)
	writeUseKubeconfig(t, "other", "other")
	mock.SetResponse("gcloud auth list", &executor.CommandResult{ExitCode: 0, Stdout: ""})
	mock.SetResponse("k3d cluster get ghost", &executor.CommandResult{ExitCode: 1, Stderr: "not found"})

	err := runUseCluster(getUseCmd(), []string{"ghost"})
	if err == nil || !strings.Contains(err.Error(), "gcloud auth login") {
		t.Fatalf("expected an auth hint, got: %v", err)
	}
}
