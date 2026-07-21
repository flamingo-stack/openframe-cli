package cluster

import (
	"path/filepath"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/utils"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/flamingo-stack/openframe-cli/tests/testutil"
)

func init() {
	testutil.InitializeTestMode()
}

func TestListCommand(t *testing.T) {
	setupFunc := func() {
		utils.SetTestExecutor(testutil.NewTestMockExecutor())
	}
	teardownFunc := func() {
		utils.ResetGlobalFlags()
	}

	testutil.TestClusterCommand(t, "list", getListCmd, setupFunc, teardownFunc)
}

// TestRunListClusters_AllDiscoversExternalGKE drives list --all end to end on
// mocks: k3d listing + gcloud discovery, with the EKS coming-soon notice.
func TestRunListClusters_AllDiscoversExternalGKE(t *testing.T) {
	t.Setenv("OPENFRAME_CLUSTERS_DIR", t.TempDir())
	t.Setenv("KUBECONFIG", filepath.Join(t.TempDir(), "kubeconfig"))
	utils.InitGlobalFlags()
	mock := executor.NewMockCommandExecutor()
	mock.SetResponse("k3d cluster list", &executor.CommandResult{ExitCode: 0, Stdout: "[]"})
	mock.SetResponse("gcloud auth list", &executor.CommandResult{ExitCode: 0, Stdout: "dev@example.com\n"})
	mock.SetResponse("gcloud config configurations list", &executor.CommandResult{ExitCode: 0,
		Stdout: `[{"name":"dev-x","properties":{"core":{"project":"proj-x"}}}]`})
	mock.SetResponse("clusters list --project proj-x", &executor.CommandResult{ExitCode: 0,
		Stdout: `[{"name":"ext-1","location":"us-central1","status":"RUNNING","currentNodeCount":2}]`})
	utils.SetTestExecutor(mock)
	t.Cleanup(utils.ResetGlobalFlags)

	cmd := getListCmd()
	utils.GetGlobalFlags().List.All = true

	if err := runListClusters(cmd, nil); err != nil {
		t.Fatalf("list --all: %v", err)
	}
	// The discovery call chain must have run.
	if !mock.WasCommandExecuted("clusters list --project proj-x") {
		t.Fatal("expected GKE discovery to list clusters in configured projects")
	}
}

// TestRunListClusters_AllNotAuthenticatedDegradesGracefully: a logged-out
// gcloud must not fail the command.
func TestRunListClusters_AllNotAuthenticatedDegradesGracefully(t *testing.T) {
	t.Setenv("OPENFRAME_CLUSTERS_DIR", t.TempDir())
	utils.InitGlobalFlags()
	mock := executor.NewMockCommandExecutor()
	mock.SetResponse("k3d cluster list", &executor.CommandResult{ExitCode: 0, Stdout: "[]"})
	mock.SetResponse("gcloud auth list", &executor.CommandResult{ExitCode: 0, Stdout: ""})
	utils.SetTestExecutor(mock)
	t.Cleanup(utils.ResetGlobalFlags)

	cmd := getListCmd()
	utils.GetGlobalFlags().List.All = true

	if err := runListClusters(cmd, nil); err != nil {
		t.Fatalf("list --all without auth must degrade to a notice, got: %v", err)
	}
	if mock.WasCommandExecuted("clusters list --project") {
		t.Fatal("discovery must not query projects when not authenticated")
	}
}
