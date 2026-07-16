package services

import (
	"context"
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/types"
	clusterDomain "github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/shared/files"
	"k8s.io/client-go/rest"
)

// recordingLister wraps MockClusterLister and records whether cluster
// selection was consulted at all.
type recordingLister struct {
	listCalled bool
}

func (r *recordingLister) ListClusters() ([]clusterDomain.ClusterInfo, error) {
	r.listCalled = true
	return []clusterDomain.ClusterInfo{{Name: "some-cluster"}}, nil
}

func (r *recordingLister) GetRestConfig(name string) (*rest.Config, error) {
	return &rest.Config{Host: "https://127.0.0.1:1"}, nil
}

// TestExecuteWithContext_ExplicitConfigSkipsClusterSelection is the N2 guard
// (0.4.7 verification report): an explicit rest.Config from --context IS the
// install target — the workflow must not run k3d cluster selection on top of
// it, which used to fail non-interactive runs with "requires a cluster name".
// The context is pre-cancelled so the workflow stops right after the target
// resolution step; the assertions are about which path it took to get there.
func TestExecuteWithContext_ExplicitConfigSkipsClusterSelection(t *testing.T) {
	t.Chdir(t.TempDir()) // no stray openframe-helm-values.yaml

	lister := &recordingLister{}
	svc, err := NewChartServiceDeferred(lister, false, false)
	if err != nil {
		t.Fatal(err)
	}
	w := &InstallationWorkflow{chartService: svc, clusterService: lister, fileCleanup: files.NewFileCleanup()}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // stop at the first ctx check after target resolution

	req := types.InstallationRequest{
		NonInteractive: true,
		KubeConfig:     &rest.Config{Host: "https://127.0.0.1:1"},
		KubeContext:    "k3d-explicit",
	}
	err = w.ExecuteWithContext(ctx, req)

	if lister.listCalled {
		t.Error("cluster selection must be skipped when an explicit rest.Config is provided")
	}
	if err != nil {
		for _, banned := range []string{"cluster name", "no cluster selected"} {
			if strings.Contains(err.Error(), banned) {
				t.Errorf("explicit --context run failed on cluster selection: %v", err)
			}
		}
	}
}

// TestExecuteWithContext_NoConfigStillSelectsCluster is the control case: the
// old contract stands when no explicit target is given — non-interactive
// without a cluster name fails fast on selection.
func TestExecuteWithContext_NoConfigStillSelectsCluster(t *testing.T) {
	t.Chdir(t.TempDir())

	lister := &recordingLister{}
	svc, err := NewChartServiceDeferred(lister, false, false)
	if err != nil {
		t.Fatal(err)
	}
	w := &InstallationWorkflow{chartService: svc, clusterService: lister, fileCleanup: files.NewFileCleanup()}

	req := types.InstallationRequest{NonInteractive: true} // no name, no KubeConfig
	err = w.ExecuteWithContext(context.Background(), req)

	if !lister.listCalled {
		t.Error("without an explicit rest.Config, cluster selection must run")
	}
	if err == nil || !strings.Contains(err.Error(), "cluster name") {
		t.Errorf("non-interactive without a name must fail fast on selection, got: %v", err)
	}
}
