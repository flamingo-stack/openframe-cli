package helm

import (
	"context"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/chart/providers/argocd"
	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
)

// findHelmUpgrade returns the recorded `helm upgrade --install` command.
func findHelmUpgrade(t *testing.T, cmds []executor.RecordedCommand) executor.RecordedCommand {
	t.Helper()
	for _, c := range cmds {
		if c.Name == "helm" && len(c.Args) > 0 && c.Args[0] == "upgrade" {
			return c
		}
	}
	t.Fatalf("no `helm upgrade` command was recorded; got %d commands", len(cmds))
	return executor.RecordedCommand{}
}

func hasAdjacentFlag(args []string, flag, value string) bool {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == flag && args[i+1] == value {
			return true
		}
	}
	return false
}

// TestInstallArgoCDHelm_PipesValuesViaStdin locks the no-temp-file contract: the
// ArgoCD values are piped to helm through `-f -` (stdin), not written to the
// filesystem.
func TestInstallArgoCDHelm_PipesValuesViaStdin(t *testing.T) {
	mock := executor.NewMockCommandExecutor()
	m, _ := NewHelmManager(mock, nil, false)

	if _, err := m.installArgoCDHelm(context.Background(), config.ChartInstallConfig{ClusterName: "test"}); err != nil {
		t.Fatalf("installArgoCDHelm: %v", err)
	}

	up := findHelmUpgrade(t, mock.Commands())

	if !hasAdjacentFlag(up.Args, "-f", "-") {
		t.Errorf("expected `-f -` (stdin) in args, got %v", up.Args)
	}
	if len(up.Stdin) == 0 {
		t.Fatal("values stdin is empty — nothing was piped to helm")
	}
	if string(up.Stdin) != argocd.GetArgoCDValues() {
		t.Errorf("helm stdin does not match the embedded ArgoCD values (got %d bytes)", len(up.Stdin))
	}
	// No temp values file must be left behind anywhere the caller could see.
	for _, c := range mock.Commands() {
		for _, a := range c.Args {
			if a != "-" && (len(a) > 5 && a[len(a)-5:] == ".yaml") {
				t.Errorf("unexpected values file path in args: %q", a)
			}
		}
	}
}
