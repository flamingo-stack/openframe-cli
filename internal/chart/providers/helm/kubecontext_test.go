package helm

import (
	"path/filepath"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
)

// TestHelmKubeContext is the F4 one-target guard: an explicit KubeContext
// (from --context / the interactive target selector) must win over the
// ClusterName-derived context, so the helm CLI targets the same cluster as the
// native clients built from the resolved rest.Config.
func TestHelmKubeContext(t *testing.T) {
	// Point context resolution at a nonexistent kubeconfig so the ClusterName
	// fallback is deterministic (k3d-<name>) regardless of the host machine.
	t.Setenv("KUBECONFIG", filepath.Join(t.TempDir(), "absent"))

	t.Run("explicit context wins over cluster name", func(t *testing.T) {
		got := helmKubeContext(config.ChartInstallConfig{KubeContext: "prod-ctx", ClusterName: "my-cluster"})
		if got != "prod-ctx" {
			t.Fatalf("got %q, want prod-ctx", got)
		}
	})

	t.Run("cluster name resolves to its k3d context", func(t *testing.T) {
		got := helmKubeContext(config.ChartInstallConfig{ClusterName: "my-cluster"})
		if got != "k3d-my-cluster" {
			t.Fatalf("got %q, want k3d-my-cluster", got)
		}
	})

	t.Run("neither set yields empty (current context)", func(t *testing.T) {
		if got := helmKubeContext(config.ChartInstallConfig{}); got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})
}

// TestArgoCDInstallArgs_ExplicitContext: the resolved context reaches the
// actual helm argv.
func TestArgoCDInstallArgs_ExplicitContext(t *testing.T) {
	args := argoCDInstallArgs(config.ChartInstallConfig{KubeContext: "prod-ctx", ClusterName: "my-cluster"}, "-")
	for i := 0; i+1 < len(args); i++ {
		if args[i] == "--kube-context" {
			if args[i+1] != "prod-ctx" {
				t.Fatalf("--kube-context %q, want prod-ctx", args[i+1])
			}
			return
		}
	}
	t.Fatal("argv must carry --kube-context")
}
