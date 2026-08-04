package update

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/shared/selfupdate"
	"github.com/flamingo-stack/openframe-cli/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// An invalid --output must fail BEFORE the network round-trip. The context is
// pre-cancelled so a regression (validating only in reportStatus, after
// u.Check) surfaces as a "checking for updates" error instead of the format
// error — deterministically, with no network involved.
func TestCheck_InvalidOutputFailsBeforeNetwork(t *testing.T) {
	cmd := GetUpdateCmd("v1.0.0")
	cmd.SetArgs([]string{"check", "-o", "bogus"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := cmd.ExecuteContext(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), `invalid --output "bogus"`)
	assert.NotContains(t, err.Error(), "checking for updates", "must fail before u.Check runs")
}

// captureStdout redirects os.Stdout around fn. Only fmt-based output (the
// json/yaml renderers) is reliably captured — pterm's text printers may hold
// the original stdout — so text cases assert behavior, not bytes.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	defer func() { os.Stdout = old }()
	fn()
	require.NoError(t, w.Close())
	out, err := io.ReadAll(r)
	require.NoError(t, err)
	return string(out)
}

func TestReportStatus_RendersEachFormat(t *testing.T) {
	st := selfupdate.Status{
		Current:    "v1.0.0",
		Latest:     "v1.2.0",
		Available:  true,
		ReleaseURL: "https://example.com/rel",
	}

	tests := []struct {
		name    string
		format  string
		wantErr string
		check   func(t *testing.T, out string)
	}{
		{
			name:   "json round-trips the status",
			format: "json",
			check: func(t *testing.T, out string) {
				var got selfupdate.Status
				require.NoError(t, json.Unmarshal([]byte(out), &got))
				assert.Equal(t, st, got)
			},
		},
		{
			name:   "yaml carries the fields",
			format: "yaml",
			check: func(t *testing.T, out string) {
				assert.Contains(t, out, "current: v1.0.0")
				assert.Contains(t, out, "latest: v1.2.0")
			},
		},
		{name: "text succeeds", format: "text"},
		{name: "empty format falls back to text", format: ""},
		{name: "invalid format errors", format: "bogus", wantErr: `invalid --output "bogus"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := testutil.FindSubcommand(t, GetUpdateCmd("v1.0.0"), "check")
			require.NoError(t, cmd.Flags().Set("output", tt.format))

			var err error
			out := captureStdout(t, func() { err = reportStatus(cmd, st) })

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Empty(t, strings.TrimSpace(out), "an invalid format must print nothing")
				return
			}
			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, out)
			}
		})
	}
}
