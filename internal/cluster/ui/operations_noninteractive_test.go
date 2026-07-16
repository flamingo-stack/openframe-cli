package ui

import (
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
)

// B3 contract guards: destructive cluster operations must either skip the
// prompt (--force) or fail fast with a --force hint in non-interactive
// sessions — never block on a confirm no one can answer.

func TestSelectClusterForCleanup_ForceSkipsPrompt(t *testing.T) {
	t.Setenv("CI", "1") // any prompt attempt would fail fast and flunk the test
	ui := NewOperationsUI()
	clusters := []models.ClusterInfo{{Name: "test-cluster", Type: models.ClusterTypeK3d}}

	name, err := ui.SelectClusterForCleanup(clusters, []string{"test-cluster"}, true)
	if err != nil {
		t.Fatalf("--force must skip the confirmation prompt entirely, got: %v", err)
	}
	if name != "test-cluster" {
		t.Errorf("expected 'test-cluster', got %q", name)
	}
}

func TestSelectClusterForCleanup_NonInteractiveWithoutForceFailsFast(t *testing.T) {
	t.Setenv("CI", "1")
	ui := NewOperationsUI()
	clusters := []models.ClusterInfo{{Name: "test-cluster", Type: models.ClusterTypeK3d}}

	_, err := ui.SelectClusterForCleanup(clusters, []string{"test-cluster"}, false)
	if err == nil {
		t.Fatal("cleanup without --force must fail fast in a non-interactive session")
	}
	if !strings.Contains(err.Error(), "--force") {
		t.Errorf("error %q should hint at --force", err)
	}
}

func TestSelectClusterForDelete_NonInteractiveWithoutForceFailsFast(t *testing.T) {
	t.Setenv("CI", "1")
	ui := NewOperationsUI()
	clusters := []models.ClusterInfo{{Name: "test-cluster", Type: models.ClusterTypeK3d}}

	_, err := ui.SelectClusterForDelete(clusters, []string{"test-cluster"}, false)
	if err == nil {
		t.Fatal("delete without --force must fail fast in a non-interactive session")
	}
	if !strings.Contains(err.Error(), "--force") {
		t.Errorf("error %q should hint at --force", err)
	}
}

func TestSelectClusterForDelete_ForceSkipsPrompt(t *testing.T) {
	t.Setenv("CI", "1")
	ui := NewOperationsUI()
	clusters := []models.ClusterInfo{{Name: "test-cluster", Type: models.ClusterTypeK3d}}

	name, err := ui.SelectClusterForDelete(clusters, []string{"test-cluster"}, true)
	if err != nil {
		t.Fatalf("--force must skip the confirmation prompt entirely, got: %v", err)
	}
	if name != "test-cluster" {
		t.Errorf("expected 'test-cluster', got %q", name)
	}
}
