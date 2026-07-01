package prerequisites

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClusterSet_Structure verifies the cluster prerequisite set is assembled
// correctly. It does not invoke IsSatisfied/Install (those touch the host), only
// checks the set's shape so a mis-wired adapter is caught.
func TestClusterSet_Structure(t *testing.T) {
	set := ClusterSet()
	assert.Equal(t, "cluster", set.Name)

	names := make([]string, 0, len(set.Items))
	for _, it := range set.Items {
		names = append(names, it.Name)
		assert.NotNilf(t, it.IsSatisfied, "%s must have a check", it.Name)
		assert.NotNilf(t, it.Install, "%s must have an installer", it.Name)
		assert.NotEmptyf(t, it.DocsURL, "%s must carry manual setup guidance", it.Name)
	}

	require.ElementsMatch(t, []string{"Docker", "kubectl", "k3d", "helm"}, names)
}
