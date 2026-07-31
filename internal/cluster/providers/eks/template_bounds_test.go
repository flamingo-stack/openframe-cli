package eks

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --nodes outside the scaling bounds must fail validation up front — not
// mid-apply, after the VPC and control plane are already created and billed.
func TestValidate_NodeCountWithinScalingBounds(t *testing.T) {
	cfg := eksConfig("demo")
	cfg.NodeCount = 6
	cfg.Cloud.MaxNodes = 4
	err := validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must not exceed max nodes")

	cfg = eksConfig("demo")
	cfg.NodeCount = 2
	cfg.Cloud.MinNodes = 3
	err = validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must not be below min nodes")

	cfg = eksConfig("demo")
	cfg.NodeCount = 3
	cfg.Cloud.MinNodes = 1
	cfg.Cloud.MaxNodes = 5
	require.NoError(t, validate(cfg))
}

// With --max-nodes omitted, the template default (4) must not silently cap a
// larger --nodes: the derived max grows to cover the requested count.
func TestTfvarsFor_DerivesMaxNodesFromNodeCount(t *testing.T) {
	cfg := eksConfig("demo")
	cfg.NodeCount = 6
	vars, err := tfvarsFor(cfg)
	require.NoError(t, err)
	assert.Equal(t, 6, vars.MaxNodes, "unset max must grow to cover --nodes")

	cfg.NodeCount = 3
	vars, err = tfvarsFor(cfg)
	require.NoError(t, err)
	assert.Equal(t, 0, vars.MaxNodes)

	cfg.Cloud.MaxNodes = 8
	vars, err = tfvarsFor(cfg)
	require.NoError(t, err)
	assert.Equal(t, 8, vars.MaxNodes)
}
