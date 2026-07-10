package ui

import (
	"strings"

	"github.com/pterm/pterm"
)

// GetStatusColor returns a color function appropriate for a status string
// (green for running/ready, yellow for pending, red for failures).
func GetStatusColor(status string) func(string) string {
	switch strings.ToLower(status) {
	case "running", "ready":
		return func(s string) string { return pterm.Green(s) }
	case "stopped", "not ready", "pending":
		return func(s string) string { return pterm.Yellow(s) }
	case "error", "failed", "unhealthy":
		return func(s string) string { return pterm.Red(s) }
	default:
		return func(s string) string { return pterm.Gray(s) }
	}
}
