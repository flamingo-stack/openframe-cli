package ui

import (
	"testing"

	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
)

func TestGetStatusColor(t *testing.T) {
	// Compare against pterm's own colorizers so the expectations hold whether
	// or not color output is enabled in the test environment.
	cases := []struct {
		status string
		want   string
	}{
		// green: healthy states, case-insensitively
		{"running", pterm.Green("x")},
		{"Running", pterm.Green("x")},
		{"READY", pterm.Green("x")},
		// yellow: transitional / stopped states
		{"stopped", pterm.Yellow("x")},
		{"not ready", pterm.Yellow("x")},
		{"Pending", pterm.Yellow("x")},
		// red: failure states
		{"error", pterm.Red("x")},
		{"failed", pterm.Red("x")},
		{"unhealthy", pterm.Red("x")},
		// gray: anything unrecognized
		{"terminating", pterm.Gray("x")},
		{"", pterm.Gray("x")},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, GetStatusColor(c.status)("x"), "status %q", c.status)
	}
}
