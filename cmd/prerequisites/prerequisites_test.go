package prerequisites

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPrerequisitesCmd_Structure(t *testing.T) {
	cmd := GetPrerequisitesCmd()
	require.NotNil(t, cmd)
	assert.Equal(t, "prerequisites", cmd.Name())
	assert.Contains(t, cmd.Aliases, "prereq")
	assert.Contains(t, cmd.Aliases, "prereqs")
	assert.NotEmpty(t, cmd.Short)

	sub := map[string]bool{}
	for _, c := range cmd.Commands() {
		sub[c.Name()] = true
	}
	assert.True(t, sub["check"], "must have a check subcommand")
	assert.True(t, sub["install"], "must have an install subcommand")
}
