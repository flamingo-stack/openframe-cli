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

	t.Run("eks returns a provider", func(t *testing.T) {
		t.Setenv("OPENFRAME_CLUSTERS_DIR", t.TempDir())
		p, err := New(models.ClusterTypeEKS, exec)
		assert.NoError(t, err)
		assert.NotNil(t, p)
	})

	t.Run("gke has no backend yet and returns ErrProviderNotFound", func(t *testing.T) {
		p, err := New(models.ClusterTypeGKE, exec)
		assert.Nil(t, p)
		var notFound models.ErrProviderNotFound
		assert.True(t, errors.As(err, &notFound), "expected ErrProviderNotFound for gke, got %v", err)
		assert.Equal(t, models.ClusterTypeGKE, notFound.ClusterType)
	})

	t.Run("unknown type is a config error", func(t *testing.T) {
		p, err := New("minikube", exec)
		assert.Nil(t, p)
		var invalid models.ErrInvalidClusterConfig
		assert.True(t, errors.As(err, &invalid), "expected ErrInvalidClusterConfig, got %v", err)
	})
}
