package k3d

import (
	"context"
	"net"
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

	used := m.getUsedPortsByExistingClusters()

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
		if used := NewK3dManager(mock, false).getUsedPortsByExistingClusters(); len(used) != 0 {
			t.Fatalf("want empty map on executor error, got %v", used)
		}
	})
	t.Run("malformed JSON", func(t *testing.T) {
		mock := executor.NewMockCommandExecutor()
		mock.SetResponse("k3d cluster list", &executor.CommandResult{Stdout: "FATAL: not json"})
		if used := NewK3dManager(mock, false).getUsedPortsByExistingClusters(); len(used) != 0 {
			t.Fatalf("want empty map on malformed JSON, got %v", used)
		}
	})
}

// --- wsl.go: convertWindowsPathToWSL (expandShortPath is a no-op off Windows) ---

func TestConvertWindowsPathToWSL(t *testing.T) {
	m := NewK3dManager(executor.NewMockCommandExecutor(), false)

	cases := []struct{ in, want string }{
		{`C:\Users\foo\file.txt`, "/mnt/c/Users/foo/file.txt"},
		{`D:\data`, "/mnt/d/data"},
		{`relative\path`, "relative/path"}, // no drive letter → slashes only
	}
	for _, c := range cases {
		got, err := m.convertWindowsPathToWSL(c.in)
		if err != nil {
			t.Fatalf("convertWindowsPathToWSL(%q): %v", c.in, err)
		}
		if got != c.want {
			t.Errorf("convertWindowsPathToWSL(%q) = %q, want %q", c.in, got, c.want)
		}
	}

	if _, err := m.convertWindowsPathToWSL(""); err == nil {
		t.Error("empty path must error")
	}
}

// --- wsl.go: getWSLUser ---

func TestGetWSLUser(t *testing.T) {
	t.Run("runner user exists", func(t *testing.T) {
		mock := executor.NewMockCommandExecutor()
		mock.SetResponse("-u runner whoami", &executor.CommandResult{Stdout: "runner\n"})
		got, err := NewK3dManager(mock, false).getWSLUser(context.Background())
		if err != nil || got != "runner" {
			t.Fatalf("got (%q, %v), want (runner, nil)", got, err)
		}
	})

	t.Run("falls back to first home-dir user", func(t *testing.T) {
		mock := executor.NewMockCommandExecutor()
		mock.SetResponse("-u runner whoami", &executor.CommandResult{Stdout: ""}) // no runner
		mock.SetResponse("getent passwd", &executor.CommandResult{Stdout: "ubuntu\n"})
		mock.SetResponse("-u ubuntu whoami", &executor.CommandResult{Stdout: "ubuntu\n"})
		got, err := NewK3dManager(mock, false).getWSLUser(context.Background())
		if err != nil || got != "ubuntu" {
			t.Fatalf("got (%q, %v), want (ubuntu, nil)", got, err)
		}
	})

	t.Run("defaults to runner when detection fails", func(t *testing.T) {
		mock := executor.NewMockCommandExecutor()
		// Every probe returns empty output → no user detected → default "runner".
		mock.SetDefaultResult(&executor.CommandResult{Stdout: ""})
		got, err := NewK3dManager(mock, false).getWSLUser(context.Background())
		if err != nil || got != "runner" {
			t.Fatalf("got (%q, %v), want (runner, nil)", got, err)
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
