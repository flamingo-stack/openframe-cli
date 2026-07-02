package helm

import (
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
)

func TestArgoCDInstallArgs(t *testing.T) {
	args := argoCDInstallArgs(config.ChartInstallConfig{}, "/tmp/values.yaml")
	s := strings.Join(args, " ")

	for _, want := range []string{
		"upgrade --install argo-cd argo/argo-cd",
		"--version=10.1.0",
		"--namespace argocd",
		"--create-namespace",
		"--wait",
		"--timeout 7m",
		"-f /tmp/values.yaml",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("args missing %q\ngot: %s", want, s)
		}
	}
	// CRDs are chart-managed now — must NOT disable them.
	if strings.Contains(s, "crds.install=false") {
		t.Errorf("args must not set crds.install=false:\n%s", s)
	}
	// No cluster / no dry-run → neither flag.
	if strings.Contains(s, "--kube-context") {
		t.Errorf("no ClusterName should mean no --kube-context:\n%s", s)
	}
	if strings.Contains(s, "--dry-run") {
		t.Errorf("no DryRun should mean no --dry-run:\n%s", s)
	}
}

func TestArgoCDInstallArgs_ClusterContextAndDryRun(t *testing.T) {
	args := argoCDInstallArgs(config.ChartInstallConfig{ClusterName: "demo", DryRun: true}, "v.yaml")
	s := strings.Join(args, " ")

	if !strings.Contains(s, "--kube-context k3d-demo") {
		t.Errorf("expected --kube-context k3d-demo:\n%s", s)
	}
	if !strings.Contains(s, "--dry-run") {
		t.Errorf("expected --dry-run:\n%s", s)
	}
}
