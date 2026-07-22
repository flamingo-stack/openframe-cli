package provider

import (
	"errors"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	exec := executor.NewMockCommandExecutor()

	t.Run("k3d returns a provider", func(t *testing.T) {
		p, err := New(models.ClusterTypeK3d, exec)
		assert.NoError(t, err)
		assert.NotNil(t, p)
	})

	t.Run("empty type defaults to k3d", func(t *testing.T) {
		p, err := New("", exec)
		assert.NoError(t, err)
		assert.NotNil(t, p)
	})

	t.Run("cloud types return providers", func(t *testing.T) {
		t.Setenv("OPENFRAME_CLUSTERS_DIR", t.TempDir())
		for _, clusterType := range []models.ClusterType{models.ClusterTypeEKS, models.ClusterTypeGKE} {
			p, err := New(clusterType, exec)
			assert.NoError(t, err, "type %s", clusterType)
			assert.NotNil(t, p, "type %s", clusterType)
		}
	})

	t.Run("unknown type is a config error", func(t *testing.T) {
		p, err := New("minikube", exec)
		assert.Nil(t, p)
		var invalid models.ErrInvalidClusterConfig
		assert.True(t, errors.As(err, &invalid), "expected ErrInvalidClusterConfig, got %v", err)
	})
}
