package ui

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/pterm/pterm"
)

func TestOperationsUI_SelectClusterForOperation(t *testing.T) {
	ui := NewOperationsUI()

	t.Run("returns cluster name from args when provided", func(t *testing.T) {
		clusters := []models.ClusterInfo{
			{Name: "test-cluster", Type: models.ClusterTypeK3d},
		}
		args := []string{"test-cluster"}

		result, err := ui.SelectClusterForOperation(clusters, args, "cleanup")

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if result != "test-cluster" {
			t.Errorf("expected 'test-cluster', got %s", result)
		}
	})

	t.Run("returns error when cluster name is empty", func(t *testing.T) {
		clusters := []models.ClusterInfo{
			{Name: "test-cluster", Type: models.ClusterTypeK3d},
		}
		args := []string{""}

		_, err := ui.SelectClusterForOperation(clusters, args, "cleanup")

		if err == nil {
			t.Error("expected error for empty cluster name")
		}
	})

	t.Run("returns empty string when no clusters available", func(t *testing.T) {
		clusters := []models.ClusterInfo{}
		args := []string{}

		result, err := ui.SelectClusterForOperation(clusters, args, "cleanup")

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if result != "" {
			t.Errorf("expected empty string, got %s", result)
		}
	})

	t.Run("handles whitespace-only cluster name", func(t *testing.T) {
		clusters := []models.ClusterInfo{
			{Name: "test-cluster", Type: models.ClusterTypeK3d},
		}
		args := []string{"   "}

		_, err := ui.SelectClusterForOperation(clusters, args, "cleanup")

		if err == nil {
			t.Error("expected error for whitespace-only cluster name")
		}
	})
}

func TestNewOperationsUI(t *testing.T) {
	ui := NewOperationsUI()

	if ui == nil {
		t.Fatal("NewOperationsUI should not return nil")
	}
}

func TestOperationsUI_ShowOperationStart(t *testing.T) {
	ui := NewOperationsUI()

	t.Run("shows cleanup operation start without panicking", func(t *testing.T) {
		// This test verifies the function doesn't panic
		// The actual output is tested manually since it involves UI rendering
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ShowOperationStart panicked: %v", r)
			}
		}()

		ui.ShowOperationStart("cleanup", "test-cluster")
	})

	t.Run("shows start operation start without panicking", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ShowOperationStart panicked: %v", r)
			}
		}()

		ui.ShowOperationStart("cleanup", "test-cluster")
	})

	t.Run("shows generic operation start without panicking", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ShowOperationStart panicked: %v", r)
			}
		}()

		ui.ShowOperationStart("unknown", "test-cluster")
	})
}

func TestOperationsUI_ShowOperationSuccess(t *testing.T) {
	ui := NewOperationsUI()

	// "cleanup" is deliberately absent: cleanup reports through
	// ShowCleanupSummary, which prints the counts the run actually produced
	// rather than a fixed list of accomplishments.
	t.Run("shows k3d delete success without panicking", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ShowOperationSuccess panicked: %v", r)
			}
		}()

		ui.ShowOperationSuccess("delete", "test-cluster", models.ClusterTypeK3d)
	})

	t.Run("shows gke delete success without panicking", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ShowOperationSuccess panicked: %v", r)
			}
		}()

		ui.ShowOperationSuccess("delete", "test-cluster", models.ClusterTypeGKE)
	})

	t.Run("shows generic success without panicking", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ShowOperationSuccess panicked: %v", r)
			}
		}()

		ui.ShowOperationSuccess("unknown", "test-cluster", models.ClusterTypeK3d)
	})
}

// TestShowOperationSuccess_ResourcesRowIsBackendHonest: a cloud delete's box
// must not claim "Cleaned up" — PVC-provisioned disks live outside terraform
// state, and the orphan sweep printed right above the box may just have
// reported survivors. Only k3d, where delete removes everything the cluster
// owned, keeps the unqualified claim.
func TestShowOperationSuccess_ResourcesRowIsBackendHonest(t *testing.T) {
	captureBox := func(fn func()) string {
		var buf bytes.Buffer
		box := pterm.DefaultBox
		defer func() { pterm.DefaultBox = box }()
		pterm.DefaultBox = *pterm.DefaultBox.WithWriter(&buf)
		fn()
		return buf.String()
	}

	gke := captureBox(func() {
		NewOperationsUI().ShowOperationSuccess("delete", "dev", models.ClusterTypeGKE)
	})
	if strings.Contains(gke, "Cleaned up") {
		t.Errorf("a cloud delete box must not claim full resource cleanup; got:\n%s", gke)
	}
	if !strings.Contains(gke, "reported above") {
		t.Errorf("a cloud delete box must point at the sweep report; got:\n%s", gke)
	}

	k3d := captureBox(func() {
		NewOperationsUI().ShowOperationSuccess("delete", "dev", models.ClusterTypeK3d)
	})
	if !strings.Contains(k3d, "Cleaned up") {
		t.Errorf("a k3d delete box keeps the cleaned-up claim; got:\n%s", k3d)
	}
}

func TestOperationsUI_ShowOperationError(t *testing.T) {
	ui := NewOperationsUI()
	testErr := errors.New("test error message")

	t.Run("shows cleanup error without panicking", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ShowOperationError panicked: %v", r)
			}
		}()

		ui.ShowOperationError("cleanup", "test-cluster", testErr)
	})

	t.Run("shows start error without panicking", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ShowOperationError panicked: %v", r)
			}
		}()

		ui.ShowOperationError("cleanup", "test-cluster", testErr)
	})

	t.Run("shows generic error without panicking", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ShowOperationError panicked: %v", r)
			}
		}()

		ui.ShowOperationError("unknown", "test-cluster", testErr)
	})
}

func TestOperationsUI_ShowNoResourcesMessage(t *testing.T) {
	ui := NewOperationsUI()

	t.Run("shows no resources message without panicking", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ShowNoResourcesMessage panicked: %v", r)
			}
		}()

		ui.ShowNoResourcesMessage("clusters", "cleanup")
	})

	t.Run("handles empty parameters without panicking", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ShowNoResourcesMessage panicked: %v", r)
			}
		}()

		ui.ShowNoResourcesMessage("", "")
	})
}
