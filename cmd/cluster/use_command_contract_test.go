package cluster

// Command contract tests for `cluster use`: pin the EXACT argv of every aws
// and gcloud CLI invocation the command issues. A drift in any flag is a
// behavior change against the operator's cloud accounts and must show up here
// as a diff of full command lines, not a missed substring.

import (
	"slices"
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
)

// requireExactCommand fails unless the mock recorded a command whose FULL argv
// (program + args, in order) equals want.
func requireExactCommand(t *testing.T, mock *executor.MockCommandExecutor, want ...string) {
	t.Helper()
	for _, rc := range mock.Commands() {
		got := append([]string{rc.Name}, rc.Args...)
		if slices.Equal(got, want) {
			return
		}
	}
	t.Fatalf("no recorded command matches %q exactly; ran: %v", strings.Join(want, " "), mock.GetExecutedCommands())
}

// The external-GKE credentials fetch: --location (not --region, which would
// reject a zone), no extra flags.
func TestUseCommandContract_GKEGetCredentials(t *testing.T) {
	mock := setupUse(t)
	writeUseKubeconfig(t, "other", "other")
	mock.SetResponse("gcloud auth list", &executor.CommandResult{ExitCode: 0, Stdout: "me@example.com\n"})
	mock.SetResponse("gcloud config configurations list", &executor.CommandResult{ExitCode: 0,
		Stdout: `[{"name":"dev-x","properties":{"core":{"project":"proj-x"}}}]`})
	mock.SetResponse("clusters list --project proj-x", &executor.CommandResult{ExitCode: 0,
		Stdout: `[{"name":"ext-1","location":"us-central1","status":"RUNNING","currentNodeCount":2}]`})
	mock.SetResponse("k3d cluster get ext-1", &executor.CommandResult{ExitCode: 1, Stderr: "not found"})

	// The mock cannot write the gke_* context, so the final switch fails — the
	// fetch and configuration-activate argv are the contract under test.
	_ = runUseCluster(getUseCmd(), []string{"ext-1"})

	requireExactCommand(t, mock,
		"gcloud", "container", "clusters", "get-credentials", "ext-1",
		"--project", "proj-x", "--location", "us-central1")
	requireExactCommand(t, mock,
		"gcloud", "config", "configurations", "activate", "dev-x")
}

// The external-EKS credentials fetch, named profile: --alias names the context
// after the cluster and the discovering profile is passed through.
func TestUseCommandContract_EKSUpdateKubeconfigWithProfile(t *testing.T) {
	mock := setupUse(t)
	writeUseKubeconfig(t, "other", "other")
	mock.SetResponse("gcloud auth list", &executor.CommandResult{ExitCode: 0, Stdout: "me@example.com\n"})
	mock.SetResponse("gcloud config configurations list", &executor.CommandResult{ExitCode: 0, Stdout: `[]`})
	mock.SetResponse("sts get-caller-identity", &executor.CommandResult{ExitCode: 0, Stdout: `{"Account":"123456789012"}`})
	mock.SetResponse("configure list-profiles", &executor.CommandResult{ExitCode: 0, Stdout: "dev\n"})
	mock.SetResponse("configure get region", &executor.CommandResult{ExitCode: 0, Stdout: "us-east-1\n"})
	mock.SetResponse("eks list-clusters", &executor.CommandResult{ExitCode: 0, Stdout: `{"clusters":["ext-eks"]}`})
	mock.SetResponse("eks describe-cluster --name ext-eks", &executor.CommandResult{ExitCode: 0,
		Stdout: `{"cluster":{"name":"ext-eks","arn":"arn:aws:eks:us-east-1:123456789012:cluster/ext-eks","status":"ACTIVE","version":"1.33"}}`})
	mock.SetResponse("k3d cluster get ext-eks", &executor.CommandResult{ExitCode: 1, Stderr: "not found"})

	_ = runUseCluster(getUseCmd(), []string{"ext-eks"})

	requireExactCommand(t, mock,
		"aws", "eks", "update-kubeconfig", "--name", "ext-eks", "--region", "us-east-1",
		"--alias", "ext-eks", "--profile", "dev")
}

// The external-EKS credentials fetch through the default credential chain (no
// named profiles): no --profile flag may appear.
func TestUseCommandContract_EKSUpdateKubeconfigDefaultChain(t *testing.T) {
	mock := setupUse(t)
	writeUseKubeconfig(t, "other", "other")
	mock.SetResponse("gcloud auth list", &executor.CommandResult{ExitCode: 0, Stdout: "me@example.com\n"})
	mock.SetResponse("gcloud config configurations list", &executor.CommandResult{ExitCode: 0, Stdout: `[]`})
	mock.SetResponse("sts get-caller-identity", &executor.CommandResult{ExitCode: 0, Stdout: `{"Account":"123456789012"}`})
	mock.SetResponse("configure list-profiles", &executor.CommandResult{ExitCode: 0, Stdout: ""})
	mock.SetResponse("configure get region", &executor.CommandResult{ExitCode: 0, Stdout: "us-east-1\n"})
	mock.SetResponse("eks list-clusters", &executor.CommandResult{ExitCode: 0, Stdout: `{"clusters":["ext-eks"]}`})
	mock.SetResponse("eks describe-cluster --name ext-eks", &executor.CommandResult{ExitCode: 0,
		Stdout: `{"cluster":{"name":"ext-eks","arn":"arn:aws:eks:us-east-1:123456789012:cluster/ext-eks","status":"ACTIVE","version":"1.33"}}`})
	mock.SetResponse("k3d cluster get ext-eks", &executor.CommandResult{ExitCode: 1, Stderr: "not found"})

	_ = runUseCluster(getUseCmd(), []string{"ext-eks"})

	requireExactCommand(t, mock,
		"aws", "eks", "update-kubeconfig", "--name", "ext-eks", "--region", "us-east-1",
		"--alias", "ext-eks")
	if mock.WasCommandExecuted("--profile") {
		t.Fatalf("the default credential chain must not pass --profile; ran: %v", mock.GetExecutedCommands())
	}
}
