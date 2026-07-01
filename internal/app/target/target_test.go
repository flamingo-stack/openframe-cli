package target

import (
	"context"
	"errors"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/rest"
)

type acceptCurrentPrompter struct{}

func (acceptCurrentPrompter) Confirm(string, bool) (bool, error)   { return true, nil }
func (acceptCurrentPrompter) Choose(string, []string) (int, error) { return 0, nil }

type fakeChecker struct {
	health    k8s.Health
	healthErr error
	res       k8s.Resources
	ok        bool
	resErr    error
}

func (f fakeChecker) CheckHealth(context.Context) (k8s.Health, error) { return f.health, f.healthErr }
func (f fakeChecker) CheckResources(context.Context, k8s.Requirements) (k8s.Resources, bool, error) {
	return f.res, f.ok, f.resErr
}

// selectorWith builds a Selector whose seams are stubbed for testing.
func selectorWith(t *testing.T, checker clusterChecker) *Selector {
	t.Helper()
	return &Selector{
		Prompter:       acceptCurrentPrompter{},
		Requirements:   k8s.Requirements{CPUMillis: 4000, MemBytes: 8 << 30},
		KubeconfigPath: "ignored",
		loadContexts: func(string) ([]k8s.ContextInfo, string, error) {
			return []k8s.ContextInfo{{Name: "ctx-a", Current: true}}, "ctx-a", nil
		},
		buildConfig: func(string, string) (*rest.Config, error) {
			return &rest.Config{Host: "https://x"}, nil
		},
		newChecker: func(*rest.Config) (clusterChecker, error) { return checker, nil },
	}
}

func healthy() k8s.Health { return k8s.Health{Reachable: true, NodesTotal: 1, NodesReady: 1} }

func TestSelect_HappyPath(t *testing.T) {
	s := selectorWith(t, fakeChecker{health: healthy(), ok: true, res: k8s.Resources{AllocatableCPUMillis: 8000, AllocatableMemBytes: 16 << 30}})
	res, err := s.Select(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "ctx-a", res.Context)
	assert.Equal(t, "https://x", res.Config.Host)
	assert.True(t, res.Health.Ready())
}

func TestSelect_Unreachable(t *testing.T) {
	s := selectorWith(t, fakeChecker{healthErr: errors.New("connection refused")})
	_, err := s.Select(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not reachable")
}

func TestSelect_NoReadyNodes(t *testing.T) {
	s := selectorWith(t, fakeChecker{health: k8s.Health{Reachable: true, NodesTotal: 1, NodesReady: 0}})
	_, err := s.Select(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no ready nodes")
}

func TestSelect_InsufficientResourcesIsAdvisory(t *testing.T) {
	s := selectorWith(t, fakeChecker{health: healthy(), ok: false, res: k8s.Resources{AllocatableCPUMillis: 1000, AllocatableMemBytes: 1 << 30}})
	res, err := s.Select(context.Background())
	require.NoError(t, err, "insufficient resources must not block — only warn")
	assert.False(t, res.ResourcesSufficient, "caller can warn based on this flag")
	assert.Equal(t, "ctx-a", res.Context)
}

func TestSelect_ContextLoadError(t *testing.T) {
	s := selectorWith(t, fakeChecker{health: healthy(), ok: true})
	s.loadContexts = func(string) ([]k8s.ContextInfo, string, error) {
		return nil, "", errors.New("no kubeconfig")
	}
	_, err := s.Select(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "kubeconfig")
}
