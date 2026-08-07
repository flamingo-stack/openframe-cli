package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/pterm/pterm"

	"github.com/flamingo-stack/openframe-cli/internal/shared/config"
	"github.com/flamingo-stack/openframe-cli/internal/shared/ui"
	"github.com/flamingo-stack/openframe-cli/tests/testutil"
)

func init() {
	// Suppress logo output during tests
	ui.TestMode = true
	testutil.InitializeTestMode()
}

func TestRootCommand(t *testing.T) {
	// Test basic command structure using testutil
	cmd := GetRootCmd(DefaultVersionInfo)

	// Note: Root command doesn't have RunE function, so we use custom validation
	if cmd.Use != "openframe" {
		t.Errorf("expected Use to be 'openframe', got %q", cmd.Use)
	}

	expectedShort := "OpenFrame CLI - provision Kubernetes clusters and deploy the OpenFrame platform"
	if cmd.Short != expectedShort {
		t.Errorf("expected Short to be %q, got %q", expectedShort, cmd.Short)
	}

	if cmd.Long == "" {
		t.Error("Command should have long description")
	}
}

func TestRootCommandHelp(t *testing.T) {
	// Test help command using testutil
	cmd := GetRootCmd(DefaultVersionInfo)
	testutil.TestCLICommand(t, cmd, []string{"--help"}, false, "OpenFrame CLI", "Available Commands")
}

func TestRootCommandVersion(t *testing.T) {
	// Test version flag using testutil. Only "dev" is asserted: commit and date
	// are backfilled from the VCS stamp on any build inside a git checkout, so
	// "none"/"unknown" only appear when no VCS info is embedded.
	cmd := GetRootCmd(DefaultVersionInfo)
	testutil.TestCLICommand(t, cmd, []string{"--version"}, false, "dev")
}

func TestBackfillFromVCS(t *testing.T) {
	dev := VersionInfo{Version: "dev", Commit: "none", Date: "unknown"}

	t.Run("fills commit and date from VCS, truncating the sha", func(t *testing.T) {
		got := backfillFromVCS(dev, []debug.BuildSetting{
			{Key: "vcs.revision", Value: "85c7c15f11b9cf079785edace4882b196008d6ee"},
			{Key: "vcs.time", Value: "2026-07-29T09:21:35Z"},
			{Key: "vcs.modified", Value: "false"},
		})
		if got.Commit != "85c7c15f11b9" {
			t.Errorf("commit = %q, want truncated sha 85c7c15f11b9", got.Commit)
		}
		if got.Date != "2026-07-29T09:21:35Z" {
			t.Errorf("date = %q, want the vcs.time", got.Date)
		}
	})

	t.Run("marks a dirty working tree", func(t *testing.T) {
		got := backfillFromVCS(dev, []debug.BuildSetting{
			{Key: "vcs.revision", Value: "abcdef1234567890"},
			{Key: "vcs.modified", Value: "true"},
		})
		if got.Commit != "abcdef123456-dirty" {
			t.Errorf("commit = %q, want abcdef123456-dirty", got.Commit)
		}
	})

	t.Run("release values are never overwritten", func(t *testing.T) {
		release := VersionInfo{Version: "1.2.3", Commit: "realsha", Date: "2026-01-01"}
		got := backfillFromVCS(release, []debug.BuildSetting{
			{Key: "vcs.revision", Value: "shouldNotWin"},
			{Key: "vcs.time", Value: "9999-01-01"},
		})
		if got.Commit != "realsha" || got.Date != "2026-01-01" {
			t.Errorf("release metadata must not be overwritten, got %+v", got)
		}
	})
}

func TestGetRootCmd(t *testing.T) {
	versionInfo := VersionInfo{
		Version: "test-version",
		Commit:  "test-commit",
		Date:    "test-date",
	}

	cmd := GetRootCmd(versionInfo)

	if cmd.Use != "openframe" {
		t.Errorf("expected Use to be 'openframe', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	// The version must stay the FIRST whitespace token — selfupdate's rollback
	// labels the saved binary by parsing `--version` output that way — followed
	// by the commit/date and the toolchain/platform suffix.
	expectedPrefix := "test-version (test-commit) built on test-date — "
	if !strings.HasPrefix(cmd.Version, expectedPrefix) {
		t.Errorf("expected version to start with %q, got %q", expectedPrefix, cmd.Version)
	}
	if !strings.Contains(cmd.Version, runtime.GOOS+"/"+runtime.GOARCH) {
		t.Errorf("expected version to name the platform, got %q", cmd.Version)
	}
}

func TestSystemService(t *testing.T) {
	// Test system service
	service := config.NewSystemService()

	err := service.Initialize()
	if err != nil {
		t.Errorf("Initialize() should not error: %v", err)
	}

	// Check that the default log directory exists
	logDir := filepath.Join(os.TempDir(), "openframe-deployment-logs")
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		t.Error("Service should create log directory")
	}
}

func TestVersionInfo(t *testing.T) {
	// Test default version info
	if DefaultVersionInfo.Version == "" {
		t.Error("DefaultVersionInfo.Version should be initialized")
	}
	if DefaultVersionInfo.Commit == "" {
		t.Error("DefaultVersionInfo.Commit should be initialized")
	}
	if DefaultVersionInfo.Date == "" {
		t.Error("DefaultVersionInfo.Date should be initialized")
	}
}

// TestVerboseEnablesDebugOutput (M1.2) locks the wiring that makes the ~36
// pterm.Debug call sites reachable at all. pterm suppresses Debug unless
// PrintDebugMessages is set, and nothing in the CLI ever set it: every debug
// diagnostic in the codebase was written but unreachable. The assertion is on
// real output, not on the flag, so removing EnableDebugMessages fails here.
func TestVerboseEnablesDebugOutput(t *testing.T) {
	restore := pterm.PrintDebugMessages
	// The --silent probe below calls ui.SetSilent(), which permanently rewires
	// pterm's package-level printers to io.Discard. Snapshot and restore them,
	// or every later test in this binary silently loses its output.
	info, success, warning, debug := pterm.Info, pterm.Success, pterm.Warning, pterm.Debug
	basic, box := pterm.DefaultBasicText, pterm.DefaultBox
	header, table := pterm.DefaultHeader, pterm.DefaultTable
	t.Cleanup(func() {
		pterm.PrintDebugMessages = restore
		pterm.Info, pterm.Success, pterm.Warning, pterm.Debug = info, success, warning, debug
		pterm.DefaultBasicText, pterm.DefaultBox = basic, box
		pterm.DefaultHeader, pterm.DefaultTable = header, table
	})

	probe := func(args ...string) string {
		pterm.DisableDebugMessages()

		root := GetRootCmd(DefaultVersionInfo)
		root.SetOut(io.Discard)
		root.SetErr(io.Discard)
		root.SetArgs(args)
		if err := root.Execute(); err != nil {
			t.Fatalf("Execute(%v): %v", args, err)
		}

		var buf bytes.Buffer
		old := pterm.Debug
		pterm.Debug = *pterm.Debug.WithWriter(&buf)
		defer func() { pterm.Debug = old }()
		pterm.Debug.Println("diagnostic line")
		return buf.String()
	}

	if out := probe("--verbose"); !strings.Contains(out, "diagnostic line") {
		t.Errorf("--verbose must make pterm.Debug print; got %q", out)
	}
	if out := probe(); strings.Contains(out, "diagnostic line") {
		t.Errorf("debug output must stay off by default; got %q", out)
	}
	// --silent means "nothing but errors"; it must win over --verbose.
	if out := probe("--verbose", "--silent"); strings.Contains(out, "diagnostic line") {
		t.Errorf("--silent must suppress debug output even with --verbose; got %q", out)
	}
}
