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

	t.Run("recognized cloud types return ErrProviderNotFound", func(t *testing.T) {
		for _, clusterType := range []models.ClusterType{models.ClusterTypeGKE, models.ClusterTypeEKS} {
			p, err := New(clusterType, exec)
			assert.Nil(t, p)
			var notFound models.ErrProviderNotFound
			assert.True(t, errors.As(err, &notFound), "expected ErrProviderNotFound for %s, got %v", clusterType, err)
			assert.Equal(t, clusterType, notFound.ClusterType)
		}
	})

	t.Run("unknown type is a config error", func(t *testing.T) {
		p, err := New("minikube", exec)
		assert.Nil(t, p)
		var invalid models.ErrInvalidClusterConfig
		assert.True(t, errors.As(err, &invalid), "expected ErrInvalidClusterConfig, got %v", err)
	})
}
