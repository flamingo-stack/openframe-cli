package eks

// Command contract tests: pin the EXACT argv of every aws CLI invocation this
// package issues (cluster create/deploy terraform flows excluded). A drift in
// any flag is a behavior change for the operator's AWS account and must show
// up here as a diff of full command lines, not a missed substring.

import (
	"context"
	"testing"

	tfengine "github.com/flamingo-stack/openframe-cli/internal/cluster/providers/terraform"
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

func TestAWSCommandContract(t *testing.T) {
	t.Setenv("CI", "1") // the orphan sweep must never prompt in tests

	profiled := func(name string) tfengine.Record {
		return tfengine.Record{Name: name, Region: "us-east-1", Profile: "staging"}
	}

	cases := []struct {
		name    string
		prepare func(mock *executor.MockCommandExecutor)
		run     func(t *testing.T, p *Provider)
		want    [][]string
	}{
		{
			name: "credential preflight, default credential chain",
			run: func(t *testing.T, p *Provider) {
				require.NoError(t, p.preflightCredentials(context.Background(), ""))
			},
			want: [][]string{
				{"aws", "sts", "get-caller-identity", "--output", "json"},
			},
		},
		{
			name: "credential preflight, named profile",
			run: func(t *testing.T, p *Provider) {
				require.NoError(t, p.preflightCredentials(context.Background(), "staging"))
			},
			want: [][]string{
				{"aws", "sts", "get-caller-identity", "--output", "json", "--profile", "staging"},
			},
		},
		{
			name: "name-collision preflight, default credential chain",
			run: func(t *testing.T, p *Provider) {
				require.NoError(t, p.preflightNameCollision(context.Background(), eksConfig("demo")))
			},
			want: [][]string{
				{"aws", "eks", "describe-cluster", "--name", "demo", "--region", "us-east-1",
					"--query", "cluster.name", "--output", "text"},
			},
		},
		{
			name: "name-collision preflight, named profile",
			run: func(t *testing.T, p *Provider) {
				config := eksConfig("demo")
				config.Cloud.Profile = "staging"
				require.NoError(t, p.preflightNameCollision(context.Background(), config))
			},
			want: [][]string{
				{"aws", "eks", "describe-cluster", "--name", "demo", "--region", "us-east-1",
					"--query", "cluster.name", "--output", "text", "--profile", "staging"},
			},
		},
		{
			name: "forced orphan sweep: tagged available-volume list, then per-volume delete",
			prepare: func(mock *executor.MockCommandExecutor) {
				mock.SetResponse("ec2 describe-volumes", &executor.CommandResult{
					ExitCode: 0, Stdout: "vol-0abc\tvol-0def\n"})
			},
			run: func(t *testing.T, p *Provider) {
				p.sweepOrphanedVolumes(context.Background(),
					tfengine.Record{Name: "demo", Region: "us-east-1"}, true)
			},
			want: [][]string{
				{"aws", "ec2", "describe-volumes", "--region", "us-east-1",
					"--filters", "Name=tag:openframe:cluster,Values=demo", "Name=status,Values=available",
					"--query", "Volumes[].VolumeId", "--output", "text"},
				{"aws", "ec2", "delete-volume", "--volume-id", "vol-0abc", "--region", "us-east-1"},
				{"aws", "ec2", "delete-volume", "--volume-id", "vol-0def", "--region", "us-east-1"},
			},
		},
		{
			name: "forced orphan sweep, named profile: the profile reaches list and delete",
			prepare: func(mock *executor.MockCommandExecutor) {
				mock.SetResponse("ec2 describe-volumes", &executor.CommandResult{
					ExitCode: 0, Stdout: "vol-0abc\n"})
			},
			run: func(t *testing.T, p *Provider) {
				p.sweepOrphanedVolumes(context.Background(), profiled("demo"), true)
			},
			want: [][]string{
				{"aws", "ec2", "describe-volumes", "--region", "us-east-1",
					"--filters", "Name=tag:openframe:cluster,Values=demo", "Name=status,Values=available",
					"--query", "Volumes[].VolumeId", "--output", "text", "--profile", "staging"},
				{"aws", "ec2", "delete-volume", "--volume-id", "vol-0abc", "--region", "us-east-1",
					"--profile", "staging"},
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

// TestExecArgs_CommandContract pins the kubeconfig exec-plugin argv — every
// kubectl/client-go call authenticates with exactly this command line.
func TestExecArgs_CommandContract(t *testing.T) {
	rec := tfengine.Record{Name: "demo", Region: "us-east-1"}
	assert.Equal(t,
		[]string{"eks", "get-token", "--cluster-name", "demo", "--region", "us-east-1", "--output", "json"},
		execArgs(rec))

	rec.Profile = "staging"
	assert.Equal(t,
		[]string{"eks", "get-token", "--cluster-name", "demo", "--region", "us-east-1", "--output", "json",
			"--profile", "staging"},
		execArgs(rec))
}

// TestExecConfig_PinsPluginContract pins the exec-plugin envelope: the aws
// binary, the v1beta1 client-auth API, and never-interactive mode (a token
// fetch must not hang a headless kubectl on a prompt).
func TestExecConfig_PinsPluginContract(t *testing.T) {
	rec := tfengine.Record{Name: "demo", Region: "us-east-1", Profile: "staging"}
	cfg := execConfig(rec)
	assert.Equal(t, "client.authentication.k8s.io/v1beta1", cfg.APIVersion)
	assert.Equal(t, "aws", cfg.Command)
	assert.Equal(t, execArgs(rec), cfg.Args)
	assert.Equal(t, clientcmdapi.NeverExecInteractiveMode, cfg.InteractiveMode)
}
