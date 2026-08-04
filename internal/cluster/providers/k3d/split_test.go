package k3d

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
)

// --- ports.go: getUsedPortsByExistingClusters ---

const clusterListJSON = `[
  {
    "name": "c1",
    "nodes": [
      {"name": "s0", "role": "server", "runtimeLabels": {"k3d.server.api.port": "6550"}},
      {"name": "lb", "role": "loadbalancer", "portMappings": {
        "80/tcp":  [{"HostIp": "0.0.0.0", "HostPort": "80"}],
        "443/tcp": [{"HostIp": "", "HostPort": "443"}]
      }},
      {"name": "agent0", "role": "agent", "runtimeLabels": {"k3d.server.api.port": "9999"}}
    ]
  }
]`

func TestGetUsedPortsByExistingClusters(t *testing.T) {
	mock := executor.NewMockCommandExecutor()
	mock.SetResponse("k3d cluster list", &executor.CommandResult{Stdout: clusterListJSON})
	m := NewK3dManager(mock, false)

	used := m.getUsedPortsByExistingClusters(context.Background())

	for _, want := range []int{6550, 80, 443} {
		if !used[want] {
			t.Errorf("port %d should be marked used, got %v", want, used)
		}
	}
	// Agent nodes are ignored — their labels must not count.
	if used[9999] {
		t.Errorf("agent-node port must be ignored, got %v", used)
	}
}

func TestGetUsedPortsByExistingClusters_ErrorsYieldEmpty(t *testing.T) {
	// Documented (if debatable) behavior: on k3d failure or malformed JSON the
	// function falls back to an EMPTY used-set and relies on the later
	// isPortAvailable dial check. If this ever changes, update findPort too.
	t.Run("executor error", func(t *testing.T) {
		mock := executor.NewMockCommandExecutor()
		mock.SetShouldFail(true, "k3d unavailable")
		if used := NewK3dManager(mock, false).getUsedPortsByExistingClusters(context.Background()); len(used) != 0 {
			t.Fatalf("want empty map on executor error, got %v", used)
		}
	})
	t.Run("malformed JSON", func(t *testing.T) {
		mock := executor.NewMockCommandExecutor()
		mock.SetResponse("k3d cluster list", &executor.CommandResult{Stdout: "FATAL: not json"})
		if used := NewK3dManager(mock, false).getUsedPortsByExistingClusters(context.Background()); len(used) != 0 {
			t.Fatalf("want empty map on malformed JSON, got %v", used)
		}
	})
}

// --- verify.go: waitForTCPPort ---

func TestWaitForTCPPort(t *testing.T) {
	m := NewK3dManager(executor.NewMockCommandExecutor(), false)

	t.Run("open port succeeds", func(t *testing.T) {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		defer ln.Close()
		_, port, _ := net.SplitHostPort(ln.Addr().String())

		if err := m.waitForTCPPort(context.Background(), "127.0.0.1", port, 3, 10*time.Millisecond); err != nil {
			t.Fatalf("expected success on open port: %v", err)
		}
	})

	t.Run("closed port exhausts retries", func(t *testing.T) {
		// Grab a free port, close it, then expect connection refusals.
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		_, port, _ := net.SplitHostPort(ln.Addr().String())
		_ = ln.Close()

		if err := m.waitForTCPPort(context.Background(), "127.0.0.1", port, 2, time.Millisecond); err == nil {
			t.Fatal("expected an error for a closed port")
		}
	})

	t.Run("cancelled context aborts", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := m.waitForTCPPort(ctx, "127.0.0.1", "1", 5, time.Millisecond)
		if err == nil {
			t.Fatal("expected cancellation error")
		}
	})
}

// --- verify.go: getKubeconfigPath / firstKubeconfigPath ---

// KUBECONFIG may hold a kubectl-style path list; passing it verbatim to
// clientcmd.LoadFromFile fails, surfacing as "cluster created but not
// reachable" right after a successful create.
func TestGetKubeconfigPath_PathList(t *testing.T) {
	m := NewK3dManager(executor.NewMockCommandExecutor(), false)
	dir := t.TempDir()
	missing := filepath.Join(dir, "missing")
	existing := filepath.Join(dir, "existing")
	if err := os.WriteFile(existing, []byte("apiVersion: v1"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Run("single path is returned verbatim", func(t *testing.T) {
		t.Setenv("KUBECONFIG", missing)
		if got := m.getKubeconfigPath(); got != missing {
			t.Fatalf("single-path KUBECONFIG must pass through unchanged, got %q", got)
		}
	})

	t.Run("list resolves to the first existing file", func(t *testing.T) {
		t.Setenv("KUBECONFIG", missing+string(os.PathListSeparator)+existing)
		if got := m.getKubeconfigPath(); got != existing {
			t.Fatalf("want first existing entry %q, got %q", existing, got)
		}
	})

	t.Run("list with no existing file falls back to the first entry", func(t *testing.T) {
		other := filepath.Join(dir, "also-missing")
		t.Setenv("KUBECONFIG", missing+string(os.PathListSeparator)+other)
		if got := m.getKubeconfigPath(); got != missing {
			t.Fatalf("want first entry %q when nothing exists yet, got %q", missing, got)
		}
	})

	t.Run("leading empty entry is skipped", func(t *testing.T) {
		t.Setenv("KUBECONFIG", string(os.PathListSeparator)+existing)
		if got := m.getKubeconfigPath(); got != existing {
			t.Fatalf("empty list entries must be ignored, got %q", got)
		}
	})
}
