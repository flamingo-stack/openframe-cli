package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExitCode proves the top level propagates a failed command's exit code when
// it is a valid Unix code, and otherwise falls back to a generic 1.
func TestExitCode(t *testing.T) {
	assert.Equal(t, 1, exitCode(nil))
	assert.Equal(t, 1, exitCode(errors.New("plain error")))

	// A CommandError's code is preserved, bare and wrapped.
	assert.Equal(t, 125, exitCode(&executor.CommandError{ExitCode: 125}))
	assert.Equal(t, 125, exitCode(fmt.Errorf("cluster create failed: %w", &executor.CommandError{ExitCode: 125})))

	// Out-of-range / non-failure codes fall back to 1.
	assert.Equal(t, 1, exitCode(&executor.CommandError{ExitCode: 0}))
	assert.Equal(t, 1, exitCode(&executor.CommandError{ExitCode: -1}))
	assert.Equal(t, 1, exitCode(&executor.CommandError{ExitCode: 4294967295}))
}

func TestMainIntegration(t *testing.T) {
	// Build test binary
	testBinary := "openframe-test-main"
	buildCmd := exec.Command("go", "build", "-o", testBinary, ".")
	require.NoError(t, buildCmd.Run())
	defer os.Remove(testBinary)

	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains string
	}{
		{
			name:     "help",
			args:     []string{"--help"},
			wantErr:  false,
			contains: "OpenFrame CLI",
		},
		{
			name:     "version",
			args:     []string{"--version"},
			wantErr:  false,
			contains: "dev",
		},
		{
			name:     "invalid flag",
			args:     []string{"--invalid"},
			wantErr:  true,
			contains: "unknown flag",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command("./"+testBinary, tc.args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			output := stdout.String() + stderr.String()

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Contains(t, output, tc.contains)
		})
	}
}
