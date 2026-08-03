package discovery

import (
	"context"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const demoClusterJSON = `{"cluster": {
  "name": "demo",
  "arn": "arn:aws:eks:us-east-1:123456789012:cluster/demo",
  "status": "ACTIVE",
  "version": "1.33"
}}`

func TestEKSAuthStatus(t *testing.T) {
	t.Run("usable credentials", func(t *testing.T) {
		mock := executor.NewMockCommandExecutor()
		mock.SetResponse("sts get-caller-identity", &executor.CommandResult{ExitCode: 0, Stdout: `{"Account": "123456789012"}`})
		assert.Equal(t, Authenticated, NewEKSDiscoverer(mock).AuthStatus(context.Background()))
	})

	t.Run("aws errors map to not-authenticated", func(t *testing.T) {
		mock := executor.NewMockCommandExecutor()
		mock.SetShouldFail(true, "InvalidClientTokenId")
		assert.Equal(t, NotAuthenticated, NewEKSDiscoverer(mock).AuthStatus(context.Background()))
	})
}

func TestEKSProfiles_SortsAndDedupes(t *testing.T) {
	mock := executor.NewMockCommandExecutor()
	mock.SetResponse("configure list-profiles", &executor.CommandResult{
		ExitCode: 0,
		Stdout:   "staging\ndefault\nstaging\n\n  prod  \n",
	})

	profiles, err := NewEKSDiscoverer(mock).Profiles(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"default", "prod", "staging"}, profiles)
}

func TestEKSRegions_SortsAndDedupes(t *testing.T) {
	mock := executor.NewMockCommandExecutor()
	// `--output text` is tab-separated on one line.
	mock.SetResponse("ec2 describe-regions", &executor.CommandResult{
		ExitCode: 0,
		Stdout:   "us-east-1\teu-west-1\tus-east-1\n",
	})

	regions, err := NewEKSDiscoverer(mock).Regions(context.Background(), "staging")
	require.NoError(t, err)
	assert.Equal(t, []string{"eu-west-1", "us-east-1"}, regions)
}

func TestEKSRegions_ErrorsWhenAwsFails(t *testing.T) {
	mock := executor.NewMockCommandExecutor()
	mock.SetShouldFail(true, "ExpiredToken")

	_, err := NewEKSDiscoverer(mock).Regions(context.Background(), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "listing AWS regions")
}

func TestEKSDiscover_ListsClustersAcrossProfiles(t *testing.T) {
	kubeconfigWith(t, map[string]string{
		"arn:aws:eks:us-east-1:123456789012:cluster/demo": "",
	})

	mock := executor.NewMockCommandExecutor()
	mock.SetResponse("configure list-profiles", &executor.CommandResult{ExitCode: 0, Stdout: "dev\nprod\n"})
	mock.SetResponse("configure get region --profile dev", &executor.CommandResult{ExitCode: 0, Stdout: "us-east-1\n"})
	mock.SetResponse("configure get region --profile prod", &executor.CommandResult{ExitCode: 0, Stdout: "eu-west-1\n"})
	mock.SetResponse("eks list-clusters --region us-east-1", &executor.CommandResult{ExitCode: 0, Stdout: `{"clusters": ["demo"]}`})
	mock.SetResponse("eks describe-cluster --name demo", &executor.CommandResult{ExitCode: 0, Stdout: demoClusterJSON})
	// prod profile: an expired session must not break discovery of the rest.
	mock.SetResponse("eks list-clusters --region eu-west-1", &executor.CommandResult{ExitCode: 1, Stderr: "ExpiredToken"})

	result, err := NewEKSDiscoverer(mock).Discover(context.Background())
	require.NoError(t, err)

	require.Len(t, result.Clusters, 1)
	c := result.Clusters[0]
	assert.Equal(t, "demo", c.Name)
	assert.Equal(t, models.ClusterTypeEKS, c.Type)
	assert.Equal(t, models.SourceExternal, c.Source)
	assert.Equal(t, "Active", c.Status)
	assert.Equal(t, "1.33", c.K8sVersion)
	assert.Equal(t, "us-east-1", c.Region)
	assert.Equal(t, "arn:aws:eks:us-east-1:123456789012:cluster/demo", c.Context)

	require.Len(t, result.Warnings, 1)
	assert.Contains(t, result.Warnings[0], "prod")
}

// Without named profiles, discovery must fall back to the default credential
// chain instead of silently finding nothing.
func TestEKSDiscover_DefaultCredentialChain(t *testing.T) {
	kubeconfigWith(t, map[string]string{})

	mock := executor.NewMockCommandExecutor()
	mock.SetResponse("configure list-profiles", &executor.CommandResult{ExitCode: 0, Stdout: ""})
	mock.SetResponse("configure get region", &executor.CommandResult{ExitCode: 0, Stdout: "us-east-1\n"})
	mock.SetResponse("eks list-clusters --region us-east-1", &executor.CommandResult{ExitCode: 0, Stdout: `{"clusters": ["demo"]}`})
	mock.SetResponse("eks describe-cluster --name demo", &executor.CommandResult{ExitCode: 0, Stdout: demoClusterJSON})

	result, err := NewEKSDiscoverer(mock).Discover(context.Background())
	require.NoError(t, err)
	require.Len(t, result.Clusters, 1)
	assert.Equal(t, "demo", result.Clusters[0].Name)
}

// A profile without a default region cannot be discovered against: it must
// surface as a warning, never as a guessed region or a hard error.
func TestEKSDiscover_ProfileWithoutRegionWarns(t *testing.T) {
	kubeconfigWith(t, map[string]string{})

	mock := executor.NewMockCommandExecutor()
	mock.SetResponse("configure list-profiles", &executor.CommandResult{ExitCode: 0, Stdout: "dev\n"})
	mock.SetResponse("configure get region --profile dev", &executor.CommandResult{ExitCode: 1, Stderr: ""})

	result, err := NewEKSDiscoverer(mock).Discover(context.Background())
	require.NoError(t, err)
	assert.Empty(t, result.Clusters)
	require.Len(t, result.Warnings, 1)
	assert.Contains(t, result.Warnings[0], "no default region")
}

func TestMatchEKSContext(t *testing.T) {
	contexts := []string{
		"k3d-openframe-dev",
		"arn:aws:eks:us-east-1:123456789012:cluster/alpha",
		"beta",
	}
	arn := func(name string) string { return "arn:aws:eks:us-east-1:123456789012:cluster/" + name }
	assert.Equal(t, arn("alpha"), matchEKSContext(contexts, arn("alpha"), "alpha"))
	assert.Equal(t, "beta", matchEKSContext(contexts, arn("beta"), "beta"))
	assert.Equal(t, "", matchEKSContext(contexts, arn("gamma"), "gamma"))
}
