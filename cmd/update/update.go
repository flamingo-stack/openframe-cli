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
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

// GetUpdateCmd returns the `openframe update` command. currentVersion is the
// running CLI version (from the root command's VersionInfo).
func GetUpdateCmd(currentVersion string) *cobra.Command {
	var (
		checkOnly bool
		targetVer string
		assumeYes bool
		force     bool
		rollback  bool
	)
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update the OpenFrame CLI to the latest release",
		Long: `Check for a newer OpenFrame CLI release and, unless --check is given,
download the verified binary and replace the running executable in place.

Every download is checksum-verified before it touches disk. A backup of the
current binary is kept and automatically restored if the new one fails to run.
The previous binary is retained so --rollback can revert instantly, offline.

Opt into automatic updates by setting OPENFRAME_AUTO_UPDATE=1 (checked once a
day, skips major versions, never runs in CI/non-interactive shells).`,
		Example: `  openframe update             # update to the latest release
  openframe update --check     # only report whether an update is available
  openframe update --version v1.4.0   # switch to a specific release (up or down)
  openframe update --rollback  # revert to the previous version, no download`,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if rollback && (checkOnly || targetVer != "" || force) {
				return fmt.Errorf("--rollback cannot be combined with --check, --version, or --force")
			}
			if rollback {
				return runRollback(cmd.Context(), currentVersion, assumeYes)
			}
			return run(cmd.Context(), cmd, currentVersion, checkOnly, targetVer, assumeYes, force)
		},
	}
	cmd.Flags().BoolVar(&checkOnly, "check", false, "Only report whether an update is available; make no changes")
	cmd.Flags().StringVar(&targetVer, "version", "", "Target a specific release (e.g. v1.4.0) instead of the latest")
	cmd.Flags().BoolVar(&rollback, "rollback", false, "Revert to the previously-installed version (offline, no download)")
	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "Skip the confirmation prompt")
	cmd.Flags().BoolVar(&force, "force", false, "Reinstall even if already up to date")
	cmd.Flags().StringP("output", "o", "text", "Output format for --check: text, json, or yaml")
	return cmd
}

func run(ctx context.Context, cmd *cobra.Command, current string, checkOnly bool, target string, assumeYes, force bool) error {
	u := selfupdate.Updater{
		Current: current,
		Client:  selfupdate.Client{Token: os.Getenv("GITHUB_TOKEN")},
	}
	st, rel, err := u.Check(ctx, target)
	if err != nil {
		return fmt.Errorf("checking for updates: %w", err)
	}

	if checkOnly {
		return reportStatus(cmd, st)
	}

	if st.DevBuild {
		pterm.Warning.Println("This is a development build; self-update is disabled. Install a released version to enable updates.")
		return nil
	}
	if !st.Available && !force && target == "" {
		pterm.Success.Printfln("Already up to date (%s).", st.Current)
		return nil
	}

	// "Update", "Downgrade", or "Reinstall" depending on the target direction.
	verb := selfupdate.ChangeVerb(st.Current, rel.TagName)
	if !assumeYes && !ui.IsNonInteractive() {
		ok, err := ui.ConfirmActionInteractive(fmt.Sprintf("%s from %s to %s?", verb, st.Current, rel.TagName), true)
		if err != nil {
			return err
		}
		if !ok {
			pterm.Info.Println("Cancelled.")
			return nil
		}
	}

	if err := u.Apply(ctx, rel, func(msg string) { pterm.Info.Println(msg) }); err != nil {
		return err
	}
	pterm.Success.Printfln("OpenFrame CLI is now %s.", rel.TagName)
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
	if !assumeYes && !ui.IsNonInteractive() {
		confirmed, err := ui.ConfirmActionInteractive(fmt.Sprintf("Roll back from %s to %s?", current, prev), true)
		if err != nil {
			return err
		}
		if !confirmed {
			pterm.Info.Println("Cancelled.")
			return nil
		}
	}
	if err := u.Rollback(ctx, func(msg string) { pterm.Info.Println(msg) }); err != nil {
		return err
	}
	pterm.Success.Printfln("Rolled back to %s.", prev)
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
