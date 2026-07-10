//go:build integration

package bootstrap_integration

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/flamingo-stack/openframe-cli/tests/integration/common"
)

// TestMain builds the CLI binary once for the package and tears it down after.
func TestMain(m *testing.M) {
	if err := common.InitializeCLI(); err != nil {
		panic("Failed to build CLI binary: " + err.Error())
	}
	code := m.Run()
	common.CleanupCLI()
	os.Exit(code)
}

// TestBootstrapOSSTenantHappyPath is the end-to-end anchor for the OSS-tenant
// contract (reqs 7/22/24): `openframe bootstrap <name> --non-interactive` must
// create a local k3d cluster and install ArgoCD + the app-of-apps from the
// public OSS repo with no credentials, and exit 0 only after ArgoCD reports the
// apps synced.
//
// It is heavy (real cluster + image pulls + ArgoCD sync) so it is gated:
//   - skipped under `-short`,
//   - skipped when Docker or k3d are unavailable,
//   - the test cluster is always torn down via t.Cleanup.
//
// Run it explicitly with a timeout that EXCEEDS bootstrap's internal app sync
// wait (40m), otherwise the test is killed before the platform converges:
//
//	go test -run TestBootstrapOSSTenantHappyPath -timeout 40m ./tests/integration/bootstrap/
func TestBootstrapOSSTenantHappyPath(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live bootstrap e2e in short mode")
	}
	if !common.Docker.IsAvailable() {
		t.Skip("Docker required for the bootstrap e2e")
	}
	if !common.K3d.IsAvailable() {
		t.Skip("k3d required for the bootstrap e2e")
	}

	// Start from a clean slate and guarantee teardown even on failure/panic.
	common.CleanupAllTestClusters()
	clusterName := fmt.Sprintf("of-e2e-%d", time.Now().Unix()%100000)
	t.Cleanup(func() { common.CleanupTestCluster(clusterName) })

	t.Logf("Running: openframe bootstrap %s --non-interactive", clusterName)
	start := time.Now()
	result := common.RunCLI("bootstrap", clusterName, "--non-interactive")
	t.Logf("bootstrap finished in %v (exit=%d)", time.Since(start), result.ExitCode)

	if !result.Success() {
		t.Fatalf("bootstrap oss-tenant failed (exit=%d)\n--- stdout ---\n%s\n--- stderr ---\n%s",
			result.ExitCode, result.Output(), result.ErrorMessage())
	}

	// Exit 0 means ArgoCD synced; also confirm the cluster is really there.
	exists, err := common.ClusterExists(clusterName)
	if err != nil {
		t.Fatalf("checking cluster existence: %v", err)
	}
	if !exists {
		t.Fatalf("expected cluster %q to exist after a successful bootstrap", clusterName)
	}
}
