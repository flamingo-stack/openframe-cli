package discovery

// Command contract tests: pin the EXACT argv of every aws and gcloud CLI
// invocation discovery issues. A drift in any flag is a behavior change
// against the operator's cloud accounts and must show up here as a diff of
// full command lines, not a missed substring.

import (
	"context"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestDiscoveryCommandContract(t *testing.T) {
	// Discover matches clusters against kubeconfig contexts — isolate it.
	kubeconfigWith(t, map[string]string{})

	cases := []struct {
		name    string
		prepare func(mock *executor.MockCommandExecutor)
		run     func(t *testing.T, mock *executor.MockCommandExecutor)
		want    [][]string
	}{
		{
			name: "EKS auth status probe",
			run: func(t *testing.T, mock *executor.MockCommandExecutor) {
				NewEKSDiscoverer(mock).AuthStatus(context.Background())
			},
			want: [][]string{
				{"aws", "sts", "get-caller-identity", "--output", "json"},
			},
		},
		{
			name: "EKS profile listing",
			run: func(t *testing.T, mock *executor.MockCommandExecutor) {
				_, err := NewEKSDiscoverer(mock).Profiles(context.Background())
				require.NoError(t, err)
			},
			want: [][]string{
				{"aws", "configure", "list-profiles"},
			},
		},
		{
			name: "EKS region listing, default credential chain",
			run: func(t *testing.T, mock *executor.MockCommandExecutor) {
				_, err := NewEKSDiscoverer(mock).Regions(context.Background(), "")
				require.NoError(t, err)
			},
			want: [][]string{
				{"aws", "ec2", "describe-regions", "--query", "Regions[].RegionName", "--output", "text"},
			},
		},
		{
			name: "EKS region listing, named profile",
			run: func(t *testing.T, mock *executor.MockCommandExecutor) {
				_, err := NewEKSDiscoverer(mock).Regions(context.Background(), "dev")
				require.NoError(t, err)
			},
			want: [][]string{
				{"aws", "ec2", "describe-regions", "--query", "Regions[].RegionName", "--output", "text",
					"--profile", "dev"},
			},
		},
		{
			name: "EKS discover, named profile: profiles, profile region, list, describe",
			prepare: func(mock *executor.MockCommandExecutor) {
				mock.SetResponse("configure list-profiles", &executor.CommandResult{ExitCode: 0, Stdout: "dev\n"})
				mock.SetResponse("configure get region", &executor.CommandResult{ExitCode: 0, Stdout: "us-east-1\n"})
				mock.SetResponse("eks list-clusters", &executor.CommandResult{ExitCode: 0, Stdout: `{"clusters": ["demo"]}`})
				mock.SetResponse("eks describe-cluster", &executor.CommandResult{ExitCode: 0, Stdout: demoClusterJSON})
			},
			run: func(t *testing.T, mock *executor.MockCommandExecutor) {
				_, err := NewEKSDiscoverer(mock).Discover(context.Background())
				require.NoError(t, err)
			},
			want: [][]string{
				{"aws", "configure", "list-profiles"},
				{"aws", "configure", "get", "region", "--profile", "dev"},
				{"aws", "eks", "list-clusters", "--region", "us-east-1", "--output", "json", "--profile", "dev"},
				{"aws", "eks", "describe-cluster", "--name", "demo", "--region", "us-east-1", "--output", "json",
					"--profile", "dev"},
			},
		},
		{
			name: "EKS discover, default credential chain: no --profile anywhere",
			prepare: func(mock *executor.MockCommandExecutor) {
				mock.SetResponse("configure list-profiles", &executor.CommandResult{ExitCode: 0, Stdout: ""})
				mock.SetResponse("configure get region", &executor.CommandResult{ExitCode: 0, Stdout: "us-east-1\n"})
				mock.SetResponse("eks list-clusters", &executor.CommandResult{ExitCode: 0, Stdout: `{"clusters": ["demo"]}`})
				mock.SetResponse("eks describe-cluster", &executor.CommandResult{ExitCode: 0, Stdout: demoClusterJSON})
			},
			run: func(t *testing.T, mock *executor.MockCommandExecutor) {
				_, err := NewEKSDiscoverer(mock).Discover(context.Background())
				require.NoError(t, err)
			},
			want: [][]string{
				{"aws", "configure", "list-profiles"},
				{"aws", "configure", "get", "region"},
				{"aws", "eks", "list-clusters", "--region", "us-east-1", "--output", "json"},
				{"aws", "eks", "describe-cluster", "--name", "demo", "--region", "us-east-1", "--output", "json"},
			},
		},
		{
			name: "GKE auth status probe",
			run: func(t *testing.T, mock *executor.MockCommandExecutor) {
				NewGKEDiscoverer(mock).AuthStatus(context.Background())
			},
			want: [][]string{
				{"gcloud", "auth", "list", "--filter=status:ACTIVE", "--format=value(account)"},
			},
		},
		{
			name: "GKE project listing via configurations",
			prepare: func(mock *executor.MockCommandExecutor) {
				mock.SetResponse("config configurations list", &executor.CommandResult{ExitCode: 0, Stdout: configurationsJSON})
			},
			run: func(t *testing.T, mock *executor.MockCommandExecutor) {
				_, err := NewGKEDiscoverer(mock).Projects(context.Background())
				require.NoError(t, err)
			},
			want: [][]string{
				{"gcloud", "config", "configurations", "list", "--format=json"},
			},
		},
		{
			name: "GKE account-wide project listing",
			run: func(t *testing.T, mock *executor.MockCommandExecutor) {
				_, err := NewGKEDiscoverer(mock).AllProjects(context.Background())
				require.NoError(t, err)
			},
			want: [][]string{
				{"gcloud", "projects", "list", "--format=value(projectId)"},
			},
		},
		{
			name: "GKE region listing",
			run: func(t *testing.T, mock *executor.MockCommandExecutor) {
				_, err := NewGKEDiscoverer(mock).Regions(context.Background(), "proj-x")
				require.NoError(t, err)
			},
			want: [][]string{
				{"gcloud", "compute", "regions", "list", "--project", "proj-x", "--format=value(name)"},
			},
		},
		{
			name: "GKE discover: configurations, then per-project clusters list",
			prepare: func(mock *executor.MockCommandExecutor) {
				mock.SetResponse("config configurations list", &executor.CommandResult{ExitCode: 0,
					Stdout: `[{"name": "dev-x", "properties": {"core": {"project": "proj-x"}}}]`})
				mock.SetResponse("container clusters list", &executor.CommandResult{ExitCode: 0, Stdout: `[]`})
			},
			run: func(t *testing.T, mock *executor.MockCommandExecutor) {
				_, err := NewGKEDiscoverer(mock).Discover(context.Background())
				require.NoError(t, err)
			},
			want: [][]string{
				{"gcloud", "config", "configurations", "list", "--format=json"},
				{"gcloud", "container", "clusters", "list", "--project", "proj-x", "--format=json"},
			},
		},
		{
			name: "auth flow ADC probe",
			run: func(t *testing.T, mock *executor.MockCommandExecutor) {
				f := NewAuthFlowWithSeams(mock, nil, nil, nil)
				assert.True(t, f.hasADC(context.Background()))
			},
			want: [][]string{
				{"gcloud", "auth", "application-default", "print-access-token"},
			},
		},
		{
			name: "auth flow fresh-login probe (forces a token refresh)",
			run: func(t *testing.T, mock *executor.MockCommandExecutor) {
				f := NewAuthFlowWithSeams(mock, nil, nil, nil)
				assert.True(t, f.hasFreshLogin(context.Background()))
			},
			want: [][]string{
				{"gcloud", "auth", "print-access-token", "--quiet"},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := executor.NewMockCommandExecutor()
			if tc.prepare != nil {
				tc.prepare(mock)
			}
			tc.run(t, mock)
			assert.Equal(t, tc.want, fullArgv(mock.Commands()))
		})
	}
}

// TestNewAuthFlow_WiresProductionSeams pins the production constructor: the
// given executor plus non-nil interactive/confirm/runLogin seams.
func TestNewAuthFlow_WiresProductionSeams(t *testing.T) {
	mock := executor.NewMockCommandExecutor()
	f := NewAuthFlow(mock)
	require.NotNil(t, f)
	assert.Same(t, mock, f.exec.(*executor.MockCommandExecutor))
	assert.NotNil(t, f.interactive)
	assert.NotNil(t, f.confirm)
	assert.NotNil(t, f.runLogin)
}
