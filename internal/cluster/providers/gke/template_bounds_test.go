package gke

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --nodes outside the autoscaler bounds must fail validation up front — not
// mid-apply, after the VPC/NAT/control plane are already created and billed.
func TestValidate_NodeCountWithinAutoscalerBounds(t *testing.T) {
	cfg := gkeConfig("demo")
	cfg.NodeCount = 6
	cfg.Cloud.MaxNodes = 4
	err := validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must not exceed max nodes")

	cfg = gkeConfig("demo")
	cfg.NodeCount = 2
	cfg.Cloud.MinNodes = 3
	err = validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must not be below min nodes")

	// Inside explicit bounds — fine.
	cfg = gkeConfig("demo")
	cfg.NodeCount = 3
	cfg.Cloud.MinNodes = 1
	cfg.Cloud.MaxNodes = 5
	require.NoError(t, validate(cfg))
}

// With --max-nodes omitted, the template default (4) must not silently cap a
// larger --nodes: the derived max grows to cover the requested count.
func TestTfvarsFor_DerivesMaxNodesFromNodeCount(t *testing.T) {
	cfg := gkeConfig("demo")
	cfg.NodeCount = 6
	vars, err := tfvarsFor(cfg)
	require.NoError(t, err)
	assert.Equal(t, 6, vars.MaxNodes, "unset max must grow to cover --nodes")
	assert.Equal(t, 6, vars.DesiredNodes)

	// At or under the template default the max stays unset (template default applies).
	cfg.NodeCount = 3
	vars, err = tfvarsFor(cfg)
	require.NoError(t, err)
	assert.Equal(t, 0, vars.MaxNodes)

	// An explicit max is never overridden.
	cfg.NodeCount = 3
	cfg.Cloud.MaxNodes = 8
	vars, err = tfvarsFor(cfg)
	require.NoError(t, err)
	assert.Equal(t, 8, vars.MaxNodes)
}
