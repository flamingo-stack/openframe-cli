package cluster

import (
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/utils"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/flamingo-stack/openframe-cli/tests/testutil"
)

// These tests drive runCreateCluster directly to exercise the non-interactive
// branch logic (node-count validation, default name, dry-run early return,
// cluster-name validation) that the generic wiring harness does not cover.

func setupCreate(t *testing.T) {
	t.Helper()
	utils.InitGlobalFlags()
	utils.SetTestExecutor(testutil.NewTestMockExecutor())
	t.Cleanup(utils.ResetGlobalFlags)
}

func TestRunCreateCluster_RejectsExplicitZeroNodes(t *testing.T) {
	setupCreate(t)
	cmd := getCreateCmd()
	utils.GetGlobalFlags().Create.SkipWizard = true
	// Set marks the "nodes" flag as Changed → the explicit-zero guard fires.
	if err := cmd.Flags().Set("nodes", "0"); err != nil {
		t.Fatal(err)
	}

	err := runCreateCluster(cmd, []string{"my-cluster"})
	if err == nil {
		t.Fatal("expected an error for explicit --nodes 0")
	}
}

func TestRunCreateCluster_RejectsInvalidClusterName(t *testing.T) {
	setupCreate(t)
	cmd := getCreateCmd()
	utils.GetGlobalFlags().Create.SkipWizard = true

	// Underscore + uppercase violate RFC1123 → validation must reject before
	// any cluster is created.
	if err := runCreateCluster(cmd, []string{"Bad_Name"}); err == nil {
		t.Fatal("expected a cluster-name validation error")
	}
}

func TestRunCreateCluster_DryRunReturnsWithoutCreating(t *testing.T) {
	setupCreate(t)
	cmd := getCreateCmd()
	gf := utils.GetGlobalFlags()
	gf.Create.SkipWizard = true
	gf.Create.DryRun = true

	// Dry-run with an explicit valid name must short-circuit to nil.
	if err := runCreateCluster(cmd, []string{"valid-name"}); err != nil {
		t.Fatalf("dry-run should return nil, got %v", err)
	}
}

func TestRunCreateCluster_DryRunDefaultsNameWhenNoArgs(t *testing.T) {
	setupCreate(t)
	cmd := getCreateCmd()
	gf := utils.GetGlobalFlags()
	gf.Create.SkipWizard = true
	gf.Create.DryRun = true

	// No args → the default "openframe-dev" name branch, then dry-run nil.
	if err := runCreateCluster(cmd, nil); err != nil {
		t.Fatalf("dry-run with default name should return nil, got %v", err)
	}
}

func TestRunCreateCluster_GKEFailsWithProviderNotFound(t *testing.T) {
	setupCreate(t)
	cmd := getCreateCmd()
	gf := utils.GetGlobalFlags()
	gf.Create.SkipWizard = true
	gf.Create.ClusterType = "gke"

	// gke passes flag validation (recognized) and the prerequisite gate (no
	// backend → no tools), then fails at the provider factory.
	err := runCreateCluster(cmd, []string{"cloud-cluster"})
	if err == nil {
		t.Fatal("expected ErrProviderNotFound for gke")
	}
	if !strings.Contains(err.Error(), "no provider available for cluster type") {
		t.Fatalf("expected provider-not-found error, got: %v", err)
	}
}

func TestRunCreateCluster_EKSDryRunShowsPlanAndExits(t *testing.T) {
	setupCreate(t)
	cmd := getCreateCmd()
	gf := utils.GetGlobalFlags()
	gf.Create.SkipWizard = true
	gf.Create.DryRun = true
	gf.Create.ClusterType = "eks"
	gf.Create.Region = "us-east-1"

	// Dry-run exits after the summary — before the prerequisite gate (which may
	// install tools) and before any terraform runs, so this is hermetic.
	if err := runCreateCluster(cmd, []string{"cloud-cluster"}); err != nil {
		t.Fatalf("eks dry-run should return nil, got %v", err)
	}
}

// setupWithExecutor wires a specific mock executor into the command service.
func setupWithExecutor(t *testing.T, exec *executor.MockCommandExecutor) {
	t.Helper()
	utils.InitGlobalFlags()
	utils.SetTestExecutor(exec)
	t.Cleanup(utils.ResetGlobalFlags)
}

// When cluster discovery fails, delete/status must surface the error rather
// than proceed to selection.
func TestRunDeleteCluster_ListFailureSurfacesError(t *testing.T) {
	mock := executor.NewMockCommandExecutor()
	mock.SetShouldFail(true, "k3d down")
	setupWithExecutor(t, mock)

	if err := runDeleteCluster(getDeleteCmd(), []string{"c1"}); err == nil {
		t.Fatal("expected an error when cluster listing fails")
	}
}

func TestRunClusterStatus_ListFailureSurfacesError(t *testing.T) {
	mock := executor.NewMockCommandExecutor()
	mock.SetShouldFail(true, "k3d down")
	setupWithExecutor(t, mock)

	if err := runClusterStatus(getStatusCmd(), []string{"c1"}); err == nil {
		t.Fatal("expected an error when cluster listing fails")
	}
}
