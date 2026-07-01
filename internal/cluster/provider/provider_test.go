package provider

import (
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_K3dLocal(t *testing.T) {
	p, err := New(models.ClusterTypeK3d, TargetLocal, executor.NewMockCommandExecutor(), false)
	require.NoError(t, err)
	assert.NotNil(t, p, "k3d/local must return a Provider")
}

func TestNew_K3dRejectsCloudTarget(t *testing.T) {
	_, err := New(models.ClusterTypeK3d, TargetCloud, executor.NewMockCommandExecutor(), false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "local target")
}

func TestNew_CloudProvidersComingSoon(t *testing.T) {
	for _, ct := range []models.ClusterType{models.ClusterTypeGKE, models.ClusterTypeEKS} {
		_, err := New(ct, TargetCloud, executor.NewMockCommandExecutor(), false)
		require.Errorf(t, err, "%s should not be implemented yet", ct)
		assert.Containsf(t, err.Error(), "coming soon", "%s should return a friendly not-implemented message", ct)
	}
}

func TestNew_UnknownProvider(t *testing.T) {
	_, err := New(models.ClusterType("bogus"), TargetLocal, executor.NewMockCommandExecutor(), false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown cluster provider")
}
