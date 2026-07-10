package cluster

import (
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/utils"
	"github.com/flamingo-stack/openframe-cli/tests/testutil"
)

func init() {
	testutil.InitializeTestMode()
}

func TestCreateCommand(t *testing.T) {
	setupFunc := func() {
		utils.SetTestExecutor(testutil.NewTestMockExecutor())
	}
	teardownFunc := func() {
		utils.ResetGlobalFlags()
	}

	testutil.TestClusterCommand(t, "create", getCreateCmd, setupFunc, teardownFunc)
}

// TestCreateHelp_DoesNotPromiseRecreation (M2.5): the help text used to say
// existing clusters "will be recreated". CreateCluster does the opposite — it
// warns, reuses the existing cluster, and returns its rest.Config. A user who
// trusted the help would think a stale cluster had been rebuilt.
func TestCreateHelp_DoesNotPromiseRecreation(t *testing.T) {
	long := getCreateCmd().Long

	if strings.Contains(long, "recreated") {
		t.Errorf("create --help must not promise recreation; CreateCluster reuses an existing cluster:\n%s", long)
	}
	if !strings.Contains(long, "reused") {
		t.Errorf("create --help must state that an existing cluster is reused:\n%s", long)
	}
}
