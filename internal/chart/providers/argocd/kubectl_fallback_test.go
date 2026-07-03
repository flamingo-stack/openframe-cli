package argocd

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
)

// These tests cover the kubectl fallback path — the one Windows/WSL always
// takes and every platform falls back to when the native client fails. Until
// now only the native (dynamic-client) path was tested.

func TestGetKubectlArgs(t *testing.T) {
	t.Run("no cluster name passes args through", func(t *testing.T) {
		m := NewManager(executor.NewMockCommandExecutor())
		got := m.getKubectlArgs("get", "pods")
		if !reflect.DeepEqual(got, []string{"get", "pods"}) {
			t.Fatalf("args = %v", got)
		}
	})
	t.Run("cluster name pins the k3d context", func(t *testing.T) {
		m := NewManagerWithCluster(executor.NewMockCommandExecutor(), "demo")
		got := m.getKubectlArgs("get", "pods")
		want := []string{"--context", "k3d-demo", "get", "pods"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("args = %v, want %v", got, want)
		}
	})
}

const twoAppsJSON = `{"items":[
  {"metadata":{"name":"core-api"},
   "status":{"health":{"status":"Healthy"},"sync":{"status":"Synced"}}},
  {"metadata":{"name":"nats"},
   "status":{"health":{"status":"Progressing"},"sync":{"status":"OutOfSync"}}}
]}`

func TestParseApplicationsViaKubectl(t *testing.T) {
	t.Run("parses application list", func(t *testing.T) {
		mock := executor.NewMockCommandExecutor()
		mock.SetResponse("get applications.argoproj.io -o json", &executor.CommandResult{Stdout: twoAppsJSON})
		apps, err := NewManager(mock).parseApplicationsViaKubectl(context.Background(), false)
		if err != nil {
			t.Fatal(err)
		}
		if len(apps) != 2 || apps[0].Name != "core-api" || apps[0].Health != "Healthy" ||
			apps[1].Name != "nats" || apps[1].Sync != "OutOfSync" {
			t.Fatalf("apps = %+v", apps)
		}
	})

	t.Run("executor failure surfaces as kubectl error", func(t *testing.T) {
		mock := executor.NewMockCommandExecutor()
		mock.SetShouldFail(true, "boom")
		_, err := NewManager(mock).parseApplicationsViaKubectl(context.Background(), false)
		if err == nil || !strings.Contains(err.Error(), "kubectl execution failed") {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("connectivity error text means cluster unreachable", func(t *testing.T) {
		mock := executor.NewMockCommandExecutor()
		mock.SetResponse("get applications.argoproj.io -o json",
			&executor.CommandResult{Stdout: "The connection to the server was refused"})
		_, err := NewManager(mock).parseApplicationsViaKubectl(context.Background(), false)
		if err == nil || !strings.Contains(err.Error(), "cluster unreachable") {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("non-JSON non-connectivity output yields empty list", func(t *testing.T) {
		mock := executor.NewMockCommandExecutor()
		mock.SetResponse("get applications.argoproj.io -o json",
			&executor.CommandResult{Stdout: "No resources found in argocd namespace."})
		apps, err := NewManager(mock).parseApplicationsViaKubectl(context.Background(), false)
		if err != nil || len(apps) != 0 {
			t.Fatalf("apps=%v err=%v, want empty+nil", apps, err)
		}
	})
}

func TestGetTotalExpectedApplicationsViaKubectl(t *testing.T) {
	t.Run("counts Applications planned by app-of-apps", func(t *testing.T) {
		mock := executor.NewMockCommandExecutor()
		mock.SetResponse("app-of-apps -o json", &executor.CommandResult{Stdout: `{
		  "status": {"resources": [
		    {"kind": "Application"}, {"kind": "Application"}, {"kind": "ConfigMap"}
		  ]}}`})
		got := NewManager(mock).getTotalExpectedApplicationsViaKubectl(context.Background(), config.ChartInstallConfig{})
		if got != 2 {
			t.Fatalf("planned = %d, want 2 (ConfigMap must not count)", got)
		}
	})

	t.Run("falls back to listing all applications", func(t *testing.T) {
		mock := executor.NewMockCommandExecutor()
		mock.SetResponse("app-of-apps -o json", &executor.CommandResult{Stdout: "not json"})
		mock.SetResponse("get applications.argoproj.io -o json", &executor.CommandResult{Stdout: twoAppsJSON})
		got := NewManager(mock).getTotalExpectedApplicationsViaKubectl(context.Background(), config.ChartInstallConfig{})
		if got != 2 {
			t.Fatalf("total = %d, want 2 from the fallback listing", got)
		}
	})

	t.Run("unknown when both methods fail", func(t *testing.T) {
		mock := executor.NewMockCommandExecutor() // default "mock output" is not JSON
		got := NewManager(mock).getTotalExpectedApplicationsViaKubectl(context.Background(), config.ChartInstallConfig{})
		if got != 0 {
			t.Fatalf("total = %d, want 0 (unknown)", got)
		}
	})
}
