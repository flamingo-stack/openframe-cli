package gke

// Command contract tests: pin the EXACT argv of every gcloud CLI invocation
// this package issues (cluster create/deploy terraform flows excluded). A
// drift in any flag is a behavior change for the operator's GCP project and
// must show up here as a diff of full command lines, not a missed substring.

import (
	"context"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// fullArgv flattens the mock's structured log into full command lines
// (program + args, in order) for exact contract assertions.
func fullArgv(commands []executor.RecordedCommand) [][]string {
	out := make([][]string, len(commands))
	for i, c := range commands {
		out[i] = append([]string{c.Name}, c.Args...)
	}
	return out
}

func TestGcloudCommandContract(t *testing.T) {
	t.Setenv("CI", "1") // the orphan sweep must never prompt in tests

	cases := []struct {
		name    string
		prepare func(mock *executor.MockCommandExecutor)
		run     func(t *testing.T, p *Provider)
		want    [][]string
	}{
		{
			name: "credential preflight: token probe, then project access probe",
			run: func(t *testing.T, p *Provider) {
				require.NoError(t, p.preflightCredentials(context.Background(), "my-project"))
			},
			want: [][]string{
				{"gcloud", "auth", "print-access-token", "--quiet"},
				{"gcloud", "projects", "describe", "my-project", "--format=value(projectId)"},
			},
		},
		{
			name: "project services: single idempotent enable of the required APIs",
			run: func(t *testing.T, p *Provider) {
				require.NoError(t, p.ensureProjectServices(context.Background(), "my-project"))
			},
			want: [][]string{
				{"gcloud", "services", "enable", "compute.googleapis.com", "container.googleapis.com",
					"--project", "my-project"},
			},
		},
		{
			name: "project services: denied enable falls back to an enabled-state probe",
			prepare: func(mock *executor.MockCommandExecutor) {
				mock.SetResponse("services enable", &executor.CommandResult{
					ExitCode: 1, Stderr: "PERMISSION_DENIED"})
				mock.SetResponse("services list", &executor.CommandResult{
					ExitCode: 0, Stdout: "compute.googleapis.com\ncontainer.googleapis.com\n"})
			},
			run: func(t *testing.T, p *Provider) {
				// Both APIs are already on — a deploy-only identity may proceed.
				require.NoError(t, p.ensureProjectServices(context.Background(), "my-project"))
			},
			want: [][]string{
				{"gcloud", "services", "enable", "compute.googleapis.com", "container.googleapis.com",
					"--project", "my-project"},
				{"gcloud", "services", "list", "--enabled", "--project", "my-project",
					"--format=value(config.name)"},
			},
		},
		{
			name: "name-collision preflight: project-wide (location-unscoped) name filter",
			run: func(t *testing.T, p *Provider) {
				require.NoError(t, p.preflightNameCollision(context.Background(), gkeConfig("demo")))
			},
			want: [][]string{
				{"gcloud", "container", "clusters", "list", "--project", "my-project",
					"--filter=name=demo", "--format=value(name)"},
			},
		},
		{
			name: "zone resolution for a zonal cluster: deterministic first zone of the region",
			prepare: func(mock *executor.MockCommandExecutor) {
				mock.SetResponse("compute zones list", &executor.CommandResult{
					ExitCode: 0, Stdout: "us-central1-a\n"})
			},
			run: func(t *testing.T, p *Provider) {
				config := gkeConfig("demo")
				require.NoError(t, p.ensureZone(context.Background(), &config))
			},
			want: [][]string{
				{"gcloud", "compute", "zones", "list", "--project", "my-project",
					"--filter", "name~^us-central1-", "--format=value(name)", "--sort-by=name", "--limit=1"},
			},
		},
		{
			name: "forced orphan sweep: labeled disk list, then per-disk zonal delete",
			prepare: func(mock *executor.MockCommandExecutor) {
				// value() output is TAB-separated: name, zone basename, region basename.
				mock.SetResponse("compute disks list", &executor.CommandResult{
					ExitCode: 0, Stdout: "pvc-disk-1\tus-central1-a\t\n"})
			},
			run: func(t *testing.T, p *Provider) {
				p.sweepOrphanedDisks(context.Background(), "my-project", "demo", "us-central1", true)
			},
			want: [][]string{
				{"gcloud", "compute", "disks", "list", "--project", "my-project",
					"--filter", "labels.goog-k8s-cluster-name=demo",
					"--format=value(name,zone.basename(),region.basename())"},
				{"gcloud", "compute", "disks", "delete", "pvc-disk-1",
					"--zone", "us-central1-a", "--project", "my-project", "--quiet"},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := executor.NewMockCommandExecutor()
			if tc.prepare != nil {
				tc.prepare(mock)
			}
			p := NewWithDeps(nil, nil, mock)
			tc.run(t, p)
			assert.Equal(t, tc.want, fullArgv(mock.Commands()))
		})
	}
}

// TestExecConfig_PinsPluginContract pins the kubeconfig exec-plugin envelope:
// the gke-gcloud-auth-plugin binary with no arguments, the v1beta1 client-auth
// API, never-interactive mode (a token fetch must not hang a headless kubectl
// on a prompt), and provideClusterInfo (the plugin needs the cluster entry).
func TestExecConfig_PinsPluginContract(t *testing.T) {
	cfg := execConfig()
	assert.Equal(t, "client.authentication.k8s.io/v1beta1", cfg.APIVersion)
	assert.Equal(t, "gke-gcloud-auth-plugin", cfg.Command)
	assert.Empty(t, cfg.Args, "the plugin takes no arguments")
	assert.Equal(t, clientcmdapi.NeverExecInteractiveMode, cfg.InteractiveMode)
	assert.True(t, cfg.ProvideClusterInfo)
}

// TestEnsureProjectServices_FailsWhenAPIsAreOff: the enabled-state fallback is
// only an escape hatch for deploy-only identities on an already-configured
// project. When a required API is genuinely off and cannot be enabled, create
// must stop with the exact manual command, not proceed into a terraform apply
// that fails minutes later.
func TestEnsureProjectServices_FailsWhenAPIsAreOff(t *testing.T) {
	mock := executor.NewMockCommandExecutor()
	mock.SetResponse("services enable", &executor.CommandResult{ExitCode: 1, Stderr: "PERMISSION_DENIED"})
	mock.SetResponse("services list", &executor.CommandResult{ExitCode: 0, Stdout: "compute.googleapis.com\n"})
	p := NewWithDeps(nil, nil, mock)

	err := p.ensureProjectServices(context.Background(), "my-project")
	require.Error(t, err, "a missing required API must stop the create")
	assert.Contains(t, err.Error(), "gcloud services enable", "the error must carry the manual fix")
}
