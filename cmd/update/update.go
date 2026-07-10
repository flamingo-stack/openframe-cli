// Package update implements the `openframe update` command: check for a newer
// release and replace the running binary in place (checksum-verified).
package update

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/flamingo-stack/openframe-cli/internal/shared/selfupdate"
	"github.com/flamingo-stack/openframe-cli/internal/shared/ui"
	"github.com/flamingo-stack/openframe-cli/internal/shared/ui/spinner"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

// GetUpdateCmd returns the `openframe update` command tree. currentVersion is
// the running CLI version (from the root command's VersionInfo).
//
// Modes are subcommands rather than flags:
//
//	openframe update            update to the latest release
//	openframe update v1.4.0     switch to a specific release (up or down)
//	openframe update check      report whether an update is available
//	openframe update rollback   revert to the previous version, offline
func GetUpdateCmd(currentVersion string) *cobra.Command {
	var (
		assumeYes bool
		force     bool
	)
	cmd := &cobra.Command{
		Use:   "update [version]",
		Short: "Update the OpenFrame CLI to the latest (or a specific) release",
		Long: `Download the verified OpenFrame CLI binary and replace the running
executable in place. With no argument it updates to the latest release; pass a
version (e.g. v1.4.0) to switch to a specific one, up or down.

Every download is checksum-verified before it touches disk. A backup of the
current binary is kept and automatically restored if the new one fails to run,
and the previous binary is retained so 'openframe update rollback' can revert
instantly, offline.

Opt into automatic updates by setting OPENFRAME_AUTO_UPDATE=1 (checked once a
day, skips major versions, never runs in CI/non-interactive shells).`,
		Example: `  openframe update             # update to the latest release
  openframe update v1.4.0      # switch to a specific release (up or down)
  openframe update check       # only report whether an update is available
  openframe update rollback    # revert to the previous version, no download`,
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			target := ""
			if len(args) == 1 {
				target = args[0]
			}
			return run(cmd.Context(), currentVersion, target, assumeYes, force)
		},
	}
	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "Skip the confirmation prompt")
	cmd.Flags().BoolVar(&force, "force", false, "Reinstall even if already up to date")
	cmd.AddCommand(newCheckCmd(currentVersion))
	cmd.AddCommand(newRollbackCmd(currentVersion))
	return cmd
}

// newCheckCmd is `openframe update check`: report availability, change nothing.
func newCheckCmd(current string) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "check",
		Short:        "Report whether an update is available, without changing anything",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			u := selfupdate.Updater{Current: current, Client: selfupdate.Client{Token: os.Getenv("GITHUB_TOKEN")}}

			// Spinner only in human (text) mode — json/yaml must keep stdout clean.
			var sp *spinner.Spinner
			if format, _ := cmd.Flags().GetString("output"); format == "" || format == "text" {
				sp = spinner.Start("Checking for updates...")
			}
			st, _, err := u.Check(cmd.Context(), "")
			if err != nil {
				if sp != nil {
					sp.Fail("Update check failed")
				}
				return fmt.Errorf("checking for updates: %w", err)
			}
			if sp != nil {
				sp.Stop()
			}
			return reportStatus(cmd, st)
		},
	}
	cmd.Flags().StringP("output", "o", "text", "Output format: text, json, or yaml")
	return cmd
}

// newRollbackCmd is `openframe update rollback`: revert to the previous version.
func newRollbackCmd(current string) *cobra.Command {
	var assumeYes bool
	cmd := &cobra.Command{
		Use:          "rollback",
		Short:        "Revert to the previously-installed version (offline, no download)",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRollback(cmd.Context(), current, assumeYes)
		},
	}
	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "Skip the confirmation prompt")
	return cmd
}

func run(ctx context.Context, current, target string, assumeYes, force bool) error {
	u := selfupdate.Updater{
		Current: current,
		Client:  selfupdate.Client{Token: os.Getenv("GITHUB_TOKEN")},
	}
	sp := spinner.Start("Checking for updates...")
	st, rel, err := u.Check(ctx, target)
	if err != nil {
		sp.Fail("Update check failed")
		return fmt.Errorf("checking for updates: %w", err)
	}

	if st.DevBuild {
		sp.Warning("This is a development build; self-update is disabled. Install a released version to enable updates.")
		return nil
	}
	if !st.Available && !force && target == "" {
		sp.Success(fmt.Sprintf("Already up to date (%s).", st.Current))
		return nil
	}
	sp.Stop() // stop before the interactive confirm prompt

	// "Update", "Downgrade", or "Reinstall" depending on the target direction.
	// Replacing the running binary needs explicit consent: interactively via the
	// prompt, non-interactively via --yes. The old `!assumeYes && !IsNonInteractive()`
	// guard silently AUTO-CONFIRMED in CI/piped sessions — inverted polarity.
	// (Opt-in unattended updates go through OPENFRAME_AUTO_UPDATE, not this path.)
	verb := selfupdate.ChangeVerb(st.Current, rel.TagName)
	if !assumeYes {
		ok, err := ui.RequireConfirmation(fmt.Sprintf("%s from %s to %s?", verb, st.Current, rel.TagName), "--yes", true)
		if err != nil {
			return err
		}
		if !ok {
			pterm.Info.Println("Cancelled.")
			return nil
		}
	}

	apply := spinner.Start(fmt.Sprintf("%s to %s...", verb, rel.TagName))
	if err := u.Apply(ctx, rel, func(msg string) { apply.UpdateText(msg) }); err != nil {
		apply.Fail(fmt.Sprintf("%s failed", verb))
		return err
	}
	apply.Success(fmt.Sprintf("OpenFrame CLI is now %s.", rel.TagName))
	return nil
}

// runRollback reverts to the binary retained by the last successful update.
func runRollback(ctx context.Context, current string, assumeYes bool) error {
	u := selfupdate.Updater{Current: current}
	prev, ok := selfupdate.PreviousVersion()
	if !ok {
		pterm.Warning.Println("No previous version to roll back to (nothing was saved by a prior update).")
		return nil
	}
	label := prev
	if label == "" {
		label = "the previous version" // binary exists but couldn't report its version
	}
	// Same consent rule as run(): never replace the binary in a non-interactive
	// session without an explicit --yes.
	if !assumeYes {
		confirmed, err := ui.RequireConfirmation(fmt.Sprintf("Roll back from %s to %s?", current, label), "--yes", true)
		if err != nil {
			return err
		}
		if !confirmed {
			pterm.Info.Println("Cancelled.")
			return nil
		}
	}
	sp := spinner.Start(fmt.Sprintf("Rolling back to %s...", label))
	if err := u.Rollback(ctx, func(msg string) { sp.UpdateText(msg) }); err != nil {
		sp.Fail("Rollback failed")
		return err
	}
	sp.Success(fmt.Sprintf("Rolled back to %s.", label))
	return nil
}

// reportStatus renders a --check result as text (default), json, or yaml.
func reportStatus(cmd *cobra.Command, st selfupdate.Status) error {
	switch format, _ := cmd.Flags().GetString("output"); format {
	case "json":
		b, err := json.MarshalIndent(st, "", "  ")
		if err != nil {
			return fmt.Errorf("encoding JSON: %w", err)
		}
		fmt.Println(string(b))
		return nil
	case "yaml":
		b, err := yaml.Marshal(st)
		if err != nil {
			return fmt.Errorf("encoding YAML: %w", err)
		}
		fmt.Print(string(b))
		return nil
	case "", "text":
		switch {
		case st.DevBuild:
			pterm.Info.Println("Development build — no release to compare against.")
		case st.Available:
			pterm.Warning.Printfln("Update available: %s → %s", st.Current, st.Latest)
			if st.ReleaseURL != "" {
				pterm.Info.Println(st.ReleaseURL)
			}
		default:
			pterm.Success.Printfln("Up to date (%s).", st.Current)
		}
		return nil
	default:
		return fmt.Errorf("invalid --output %q (want \"text\", \"json\", or \"yaml\")", format)
	}
}
