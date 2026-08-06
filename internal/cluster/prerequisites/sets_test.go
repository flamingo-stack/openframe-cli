package prerequisites

import (
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	fw "github.com/flamingo-stack/openframe-cli/internal/prerequisites"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertSetStructure verifies a prerequisite set is assembled correctly. It
// does not invoke IsSatisfied/Install (those touch the host), only checks the
// set's shape so a mis-wired adapter is caught.
func assertSetStructure(t *testing.T, set fw.Set, wantName string, wantItems []string) {
	t.Helper()
	assert.Equal(t, wantName, set.Name)

	names := make([]string, 0, len(set.Items))
	for _, it := range set.Items {
		names = append(names, it.Name)
		assert.NotNilf(t, it.IsSatisfied, "%s must have a check", it.Name)
		assert.NotNilf(t, it.Install, "%s must have an installer", it.Name)
		assert.NotEmptyf(t, it.DocsURL, "%s must carry manual setup guidance", it.Name)
	}

	require.ElementsMatch(t, wantItems, names)
}

func TestClusterSet_Structure(t *testing.T) {
	assertSetStructure(t, ClusterSet(), "cluster", []string{"Docker", "k3d", "helm"})
}

func TestEKSSet_Structure(t *testing.T) {
	assertSetStructure(t, EKSSet(), "eks", []string{"terraform", "AWS CLI"})
}

func TestGKESet_Structure(t *testing.T) {
	assertSetStructure(t, GKESet(), "gke", []string{"terraform", "gcloud", "gke-gcloud-auth-plugin"})
}

// TestSetForClusterType verifies the type→set mapping without invoking any
// checks or installs: k3d (and the empty default) map to the local set, the
// cloud types to their cloud sets, and unknown types are an error.
func TestSetForClusterType(t *testing.T) {
	cases := []struct {
		clusterType models.ClusterType
		wantName    string
	}{
		{models.ClusterTypeK3d, "cluster"},
		{"", "cluster"},
		{models.ClusterTypeEKS, "eks"},
		{models.ClusterTypeGKE, "gke"},
	}
	for _, tc := range cases {
		t.Run(string(tc.clusterType), func(t *testing.T) {
			set, err := SetForClusterType(tc.clusterType)
			require.NoError(t, err)
			assert.Equal(t, tc.wantName, set.Name)
			assert.NotEmpty(t, set.Items)
		})
	}

	t.Run("unknown", func(t *testing.T) {
		_, err := SetForClusterType("minikube")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "minikube")
		assert.Contains(t, err.Error(), "k3d, eks, or gke")
	})
}
