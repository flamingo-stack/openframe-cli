package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strings"
	"syscall"

	"github.com/flamingo-stack/openframe-cli/cmd/app"
	"github.com/flamingo-stack/openframe-cli/cmd/bootstrap"
	"github.com/flamingo-stack/openframe-cli/cmd/cluster"
	"github.com/flamingo-stack/openframe-cli/cmd/prerequisites"
	"github.com/flamingo-stack/openframe-cli/cmd/update"
	"github.com/flamingo-stack/openframe-cli/internal/chart/providers/argocd"
	"github.com/flamingo-stack/openframe-cli/internal/shared/config"
	"github.com/flamingo-stack/openframe-cli/internal/shared/download"
	"github.com/flamingo-stack/openframe-cli/internal/shared/selfupdate"
	"github.com/flamingo-stack/openframe-cli/internal/shared/ui"
	"github.com/flamingo-stack/openframe-cli/internal/shared/wsllauncher"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// VersionInfo holds version information for the CLI
type VersionInfo struct {
	Version string
	Commit  string
	Date    string
}

// Build-time version metadata. These MUST be package-level string vars: the
// linker's -X flag can only set package-level strings, not struct fields, so a
// prior `-X ...DefaultVersionInfo.Version=...` silently no-op'd and every
// release reported "dev" (which also disabled self-update). See .goreleaser.yml.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// DefaultVersionInfo provides default version information, populated from the
// build-time vars above (overridden via -ldflags -X at release time).
var DefaultVersionInfo = resolveVersionInfo(version, commit, date)

// resolveVersionInfo returns the build-time version metadata. On a dev build —
// where the release-time -ldflags were not applied, so commit is "none" and
// date "unknown" — it backfills them from the VCS stamp Go embeds in every
// `go build`/`go install`/`make build` binary, so the build is identifiable
// ("dev (85c7c15f11b9) built on <time>") instead of "dev (none) built on
// unknown". The version string stays "dev", which is what keeps self-update
// disabled for non-release builds.
func resolveVersionInfo(version, commit, date string) VersionInfo {
	info := VersionInfo{Version: version, Commit: commit, Date: date}
	if commit != "none" && date != "unknown" {
		return info // release build — ldflags were applied
	}
	if bi, ok := debug.ReadBuildInfo(); ok {
		info = backfillFromVCS(info, bi.Settings)
	}
	return info
}

// backfillFromVCS fills commit/date from Go's embedded VCS settings, but only
// where they are still at their dev defaults.
func backfillFromVCS(info VersionInfo, settings []debug.BuildSetting) VersionInfo {
	var rev, vcsTime string
	var modified bool
	for _, s := range settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.time":
			vcsTime = s.Value
		case "vcs.modified":
			modified = s.Value == "true"
		}
	}
	if info.Commit == "none" && rev != "" {
		if len(rev) > 12 {
			rev = rev[:12]
		}
		if modified {
			rev += "-dirty"
		}
		info.Commit = rev
	}
	if info.Date == "unknown" && vcsTime != "" {
		info.Date = vcsTime
	}
	return info
}

// GetRootCmd returns the root command following cluster command pattern
func GetRootCmd(versionInfo VersionInfo) *cobra.Command {
	return buildRootCommand(versionInfo)
}

// pinnedDependencies renders the versions this build installs (verified,
// checksum-pinned downloads) and deploys — so `--version` answers not just
// "which CLI" but "which terraform/helm/argocd comes with it". Sources: the
// PinnedTool definitions in internal/shared/download and the ArgoCD chart pin
// in internal/chart/providers/argocd.
func pinnedDependencies() string {
	var b strings.Builder
	b.WriteString("Pinned dependencies (installed verified at exactly these versions):\n")
	for _, dep := range []struct{ name, version string }{
		{"terraform", download.Terraform.Version},
		{"helm", download.Helm.Version},
		{"k3d", download.K3d.Version},
		{"mkcert", download.Mkcert.Version},
		{"infracost", download.Infracost.Version + " (optional, cost estimates)"},
		{"argo-cd", "chart " + argocd.ArgoCDChartVersion},
	} {
		fmt.Fprintf(&b, "  %-10s %s\n", dep.name, dep.version)
	}
	return strings.TrimRight(b.String(), "\n")
}

// buildRootCommand constructs the root command with given version info
func buildRootCommand(versionInfo VersionInfo) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "openframe",
		Short: "OpenFrame CLI - provision Kubernetes clusters and deploy the OpenFrame platform",
		Long: `OpenFrame CLI - Kubernetes Platform Bootstrapper

Provision a Kubernetes cluster — local k3d for development, or cloud GKE/EKS
via Terraform — install the OpenFrame platform onto it (ArgoCD app-of-apps),
and manage the full lifecycle: prerequisites, status, upgrades, teardown.

Typical flows:
  openframe bootstrap                    # local: k3d cluster + platform in one step
  openframe cluster create --type gke    # cloud: plan, confirm, provision...
  openframe app install                  # ...then install the platform onto it

Command groups:
  cluster        create, delete, list, status, use, cleanup (prune node images)
  app            install, upgrade, status, access, uninstall
  bootstrap      cluster create + app install in one step
  prerequisites  check and install required tools (--type k3d|eks|gke)
  update         update this CLI to a newer release

Every command runs interactively by default (wizards, confirmations) and
non-interactively with flags for CI and automation. Cloud creates show a full
terraform plan (and an infracost estimate, when installed) before anything is
applied; deletes require typed confirmation and clean up after themselves.`,
		// The version MUST stay the first whitespace token: selfupdate's
		// rollback labels the saved binary by parsing `--version` output that
		// way (binaryVersion in internal/shared/selfupdate). The toolchain and
		// platform ride along because they are the first questions of any bug
		// report about a downloaded release; the pinned-dependency block below
		// them answers the second ("which terraform/helm/argocd does this build
		// install?") without digging through the source.
		Version: fmt.Sprintf("%s (%s) built on %s — %s %s/%s\n\n%s",
			versionInfo.Version, versionInfo.Commit, versionInfo.Date,
			runtime.Version(), runtime.GOOS, runtime.GOARCH,
			pinnedDependencies()),
		// Silence errors and usage globally - we handle our own error display
		SilenceErrors: true,
		SilenceUsage:  true,
		// Apply --silent before any command runs so it honors its contract
		// ("suppress all output except errors") across every subcommand.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// --verbose enables pterm's Debug printer. Without this the ~40
			// pterm.Debug call sites across the codebase (executed helm/k3d
			// command lines, ArgoCD wait internals, prerequisite decisions)
			// print NOTHING, ever. NOTE: cobra runs only the CLOSEST parent's
			// PersistentPreRunE — command groups with their own hook (app,
			// cluster) shadow this one and must call the same helper.
			ui.ApplyGlobalOutputFlags(cmd)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Show logo when no subcommand is provided
			ui.ShowLogo()
			return cmd.Help()
		},
	}

	// Add subcommands
	rootCmd.AddCommand(getClusterCmd())
	rootCmd.AddCommand(getAppCmd())
	rootCmd.AddCommand(getBootstrapCmd())
	rootCmd.AddCommand(getPrerequisitesCmd())
	rootCmd.AddCommand(getUpdateCmd(versionInfo.Version))

	// Add global flags following cluster pattern
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().Bool("silent", false, "Suppress all output except errors")
	rootCmd.PersistentFlags().Bool("plain", false, "Sequential output without spinners or in-place redraws (for scripts, tmux logging, watch)")

	// Version template
	rootCmd.SetVersionTemplate(`{{printf "%s\n" .Version}}`)

	// Custom usage template with better formatting
	rootCmd.SetUsageTemplate(`Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`)

	return rootCmd
}

// Execute runs the root command with default version info
func Execute() error {
	return ExecuteWithVersion(DefaultVersionInfo)
}

// ExecuteWithVersion runs the root command with specified version info
func ExecuteWithVersion(versionInfo VersionInfo) error {
	// On Windows, re-run the whole CLI inside WSL — the cluster and the native
	// Kubernetes client live there (Option 1). The Linux build inside WSL does
	// not forward, so this happens at most once.
	if wsllauncher.ShouldForward() {
		code, err := wsllauncher.Forward(versionInfo.Version, os.Args[1:])
		if err != nil {
			return err
		}
		os.Exit(code)
	}
	rootCmd := GetRootCmd(versionInfo)

	// Initialize configuration using service layer
	service := config.NewSystemService()
	if err := service.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: initialization failed: %v\n", err)
	}

	// Ensure the CLI-managed bin dir (where verified tool binaries are
	// installed) is on this process's PATH, so tools installed by earlier
	// runs are found without editing the user's shell configuration.
	if binDir, err := download.UserBinDir(); err == nil {
		download.PrependToPath(binDir)
	}

	// Run with a signal-cancelled context so Ctrl-C / SIGTERM cancels every
	// command via cmd.Context(). This replaces the per-operation signal handlers
	// that individual services used to install by hand.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err := rootCmd.ExecuteContext(ctx)

	// Restore default signal handling before the post-command window: the
	// NotifyContext handler keeps capturing SIGINT after the context has been
	// consumed, so during an opted-in auto-update (up to 10 minutes) Ctrl-C was
	// swallowed with no way to interrupt. After stop(), Ctrl-C terminates the
	// process again.
	stop()

	// Post-command self-update handling, best-effort and printed to stderr so it
	// never blocks the command, changes its exit code, or corrupts machine output
	// on stdout. All paths are suppressed in CI / non-TTY / dev builds and by
	// OPENFRAME_NO_UPDATE_CHECK. When OPENFRAME_AUTO_UPDATE is opted in, apply the
	// update in place (skipping major bumps); otherwise just show a notice.
	interactive := !ui.IsNonInteractive()
	stderr := func(s string) { pterm.Info.WithWriter(os.Stderr).Println(s) }
	if selfupdate.AutoUpdateEnabled() {
		if msg := selfupdate.MaybeAutoUpdate(context.Background(), versionInfo.Version, interactive, stderr); msg != "" {
			stderr(msg)
		}
	} else if msg := selfupdate.MaybeNotify(context.Background(), versionInfo.Version, interactive); msg != "" {
		stderr(msg)
	}
	return err
}

// getClusterCmd returns the cluster command
func getClusterCmd() *cobra.Command {
	return cluster.GetClusterCmd()
}

// getAppCmd returns the app command (formerly "chart")
func getAppCmd() *cobra.Command {
	return app.GetAppCmd()
}

// getBootstrapCmd returns the bootstrap command
func getBootstrapCmd() *cobra.Command {
	return bootstrap.GetBootstrapCmd()
}

// getPrerequisitesCmd returns the prerequisites command
func getPrerequisitesCmd() *cobra.Command {
	return prerequisites.GetPrerequisitesCmd()
}

// getUpdateCmd returns the self-update command, bound to the running version.
func getUpdateCmd(currentVersion string) *cobra.Command {
	return update.GetUpdateCmd(currentVersion)
}
