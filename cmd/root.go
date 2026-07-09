package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/flamingo-stack/openframe-cli/cmd/app"
	"github.com/flamingo-stack/openframe-cli/cmd/bootstrap"
	"github.com/flamingo-stack/openframe-cli/cmd/cluster"
	"github.com/flamingo-stack/openframe-cli/cmd/prerequisites"
	"github.com/flamingo-stack/openframe-cli/cmd/update"
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
var DefaultVersionInfo = VersionInfo{
	Version: version,
	Commit:  commit,
	Date:    date,
}

// GetRootCmd returns the root command following cluster command pattern
func GetRootCmd(versionInfo VersionInfo) *cobra.Command {
	return buildRootCommand(versionInfo)
}

// buildRootCommand constructs the root command with given version info
func buildRootCommand(versionInfo VersionInfo) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "openframe",
		Short: "OpenFrame CLI - Kubernetes cluster bootstrapping and chart deployment",
		Long: `OpenFrame CLI - Interactive Kubernetes Platform Bootstrapper

OpenFrame CLI replaces the shell scripts with a modern, interactive terminal UI
for managing OpenFrame Kubernetes deployments. Built following best practices
for CLI design with wizard-style interactive prompts.

Key Features:
  - Interactive Wizard - Step-by-step guided setup
  - Cluster Management - K3d, Kind, and cloud provider support
  - Helm Integration - App-of-Apps pattern with ArgoCD
  - Prerequisite Checking - Validates tools before running

The CLI provides both interactive modes for new users and flag-based
operation for automation and power users.`,
		Version: fmt.Sprintf("%s (%s) built on %s", versionInfo.Version, versionInfo.Commit, versionInfo.Date),
		// Silence errors and usage globally - we handle our own error display
		SilenceErrors: true,
		SilenceUsage:  true,
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

// getAppCmd returns the app command (formerly "chart"; "chart" remains an alias)
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
