//go:build !windows

package k3d

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestForceDelete_FallbackSelectsByClusterLabel is the T0-2 regression guard:
// when `k3d cluster delete` fails and --force falls back to direct Docker
// cleanup, containers must be selected by the exact-match k3d.cluster label.
// A `name=k3d-<name>` filter is an unanchored regex — force-deleting cluster
// "dev" would also remove the containers of "dev-2", "dev-old", etc.
func TestForceDelete_FallbackSelectsByClusterLabel(t *testing.T) {
	mock := executor.NewMockCommandExecutor()
	// k3d delete fails -> triggers the force fallback.
	mock.SetResponse("cluster delete", &executor.CommandResult{
		ExitCode: 1,
		Stderr:   "simulated k3d failure",
		Duration: time.Millisecond,
	})
	mock.SetResponse("docker ps", &executor.CommandResult{
		ExitCode: 0,
		Stdout:   "id1\nid2\n",
		Duration: time.Millisecond,
	})
	manager := NewK3dManager(mock, false)

	err := manager.DeleteCluster(context.Background(), "dev", models.ClusterTypeK3d, true)
	require.NoError(t, err, "force delete must succeed via the Docker fallback")

	var sawPs bool
	var removed []string
	for _, rc := range mock.Commands() {
		if rc.Name != "docker" || len(rc.Args) == 0 {
			continue
		}
		switch rc.Args[0] {
		case "ps":
			sawPs = true
			assert.Truef(t, hasArgPair(rc.Args, "--filter", "label=k3d.cluster=dev"),
				"container selection must use the exact-match cluster label, got: %v", rc.Args)
			for _, a := range rc.Args {
				assert.NotContainsf(t, a, "name=", "must not select containers by name regex: %v", rc.Args)
			}
		case "rm":
			removed = append(removed, rc.Args[len(rc.Args)-1])
		}
	}
	assert.True(t, sawPs, "fallback must list containers")
	assert.Equal(t, []string{"id1", "id2"}, removed, "exactly the listed containers are removed")
}

// TestForceDelete_FallbackFailurePropagates: when both k3d delete and the
// Docker fallback fail, the caller gets an error (not a silent success).
func TestForceDelete_FallbackFailurePropagates(t *testing.T) {
	mock := executor.NewMockCommandExecutor()
	mock.SetResponse("cluster delete", &executor.CommandResult{ExitCode: 1, Duration: time.Millisecond})
	mock.SetResponse("docker ps", &executor.CommandResult{ExitCode: 1, Duration: time.Millisecond})
	manager := NewK3dManager(mock, false)

	err := manager.DeleteCluster(context.Background(), "dev", models.ClusterTypeK3d, true)
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "dev"))
}

// hasArgPair reports whether argv contains flag immediately followed by value.
func hasArgPair(argv []string, flag, value string) bool {
	for i := 0; i+1 < len(argv); i++ {
		if argv[i] == flag && argv[i+1] == value {
			return true
		}
	}
	return false
}
