package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateDeploymentMode(t *testing.T) {
	// empty is allowed (interactive / defaulting)
	assert.NoError(t, ValidateDeploymentMode(""))

	for _, mode := range ValidDeploymentModes {
		assert.NoErrorf(t, ValidateDeploymentMode(mode), "%s should be valid", mode)
	}

	err := ValidateDeploymentMode("bogus")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid deployment mode: bogus")
	assert.Contains(t, err.Error(), "oss-tenant") // lists the valid options
}
