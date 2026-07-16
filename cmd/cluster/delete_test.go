package cluster

import (
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/utils"
	"github.com/flamingo-stack/openframe-cli/tests/testutil"
)

func init() {
	testutil.InitializeTestMode()
}

func TestDeleteCommand(t *testing.T) {
	setupFunc := func() {
		utils.SetTestExecutor(testutil.NewTestMockExecutor())
	}
	teardownFunc := func() {
		utils.ResetGlobalFlags()
	}

	testutil.TestClusterCommand(t, "delete", getDeleteCmd, setupFunc, teardownFunc)
}

// TestConfirmCloudDeletion locks the stronger destroy gate: cloud clusters
// delete billed infrastructure, so a non-interactive delete without --force
// must refuse instead of proceeding; --force and local clusters pass through.
func TestConfirmCloudDeletion(t *testing.T) {
	t.Setenv("CI", "true") // force ui.IsNonInteractive()

	t.Run("k3d passes without confirmation", func(t *testing.T) {
		proceed, err := confirmCloudDeletion(models.ClusterTypeK3d, "local", false)
		if err != nil || !proceed {
			t.Fatalf("k3d must pass through, got proceed=%v err=%v", proceed, err)
		}
	})

	t.Run("cloud with force passes", func(t *testing.T) {
		for _, clusterType := range []models.ClusterType{models.ClusterTypeEKS, models.ClusterTypeGKE} {
			proceed, err := confirmCloudDeletion(clusterType, "cloudy", true)
			if err != nil || !proceed {
				t.Fatalf("%s with --force must pass, got proceed=%v err=%v", clusterType, proceed, err)
			}
		}
	})

	t.Run("cloud non-interactive without force refuses", func(t *testing.T) {
		for _, clusterType := range []models.ClusterType{models.ClusterTypeEKS, models.ClusterTypeGKE} {
			proceed, err := confirmCloudDeletion(clusterType, "cloudy", false)
			if proceed || err == nil {
				t.Fatalf("%s must refuse, got proceed=%v err=%v", clusterType, proceed, err)
			}
			if !strings.Contains(err.Error(), "refusing to destroy cloud cluster") {
				t.Fatalf("expected the refusal message, got: %v", err)
			}
		}
	})
}
