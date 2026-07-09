// Package prerequisites wires the OS-aware prerequisites framework into the
// `openframe prerequisites` command, so users can check and install the tools
// OpenFrame needs as an explicit, first-class step (req 20).
package prerequisites

import (
	"fmt"

	clusterprereq "github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites"
	fw "github.com/flamingo-stack/openframe-cli/internal/prerequisites"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// GetPrerequisitesCmd returns the prerequisites command and its subcommands.
func GetPrerequisitesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "prerequisites",
		Aliases: []string{"prereq", "prereqs"},
		Short:   "Check and install the tools OpenFrame needs",
		Long: `Prerequisites - check and install the tools OpenFrame needs

Verifies that Docker, kubectl, k3d, and helm are available (and Docker running).

  • check   - report what is installed, without changing anything
  • install - install anything missing (macOS/Linux); on Windows, print the docs
              links to install them manually

Examples:
  openframe prerequisites check
  openframe prerequisites install`,
		RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}
	cmd.AddCommand(checkCmd(), installCmd())
	return cmd
}

func checkCmd() *cobra.Command {
	return &cobra.Command{
		Use:           "check",
		Short:         "Report which prerequisites are installed (no changes)",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			set := clusterprereq.ClusterSet()
			res := fw.NewRunner().Check(set)
			printResult(res)
			if !res.OK() {
				return fmt.Errorf("%d prerequisite(s) missing — run 'openframe prerequisites install'", len(res.Missing))
			}
			return nil
		},
	}
}

func installCmd() *cobra.Command {
	return &cobra.Command{
		Use:           "install",
		Short:         "Install any missing prerequisites (macOS/Linux)",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			set := clusterprereq.ClusterSet()
			runner := fw.NewRunner()
			if !runner.AutoInstalls() {
				pterm.Warning.Println("Automatic install isn't supported on this OS — please install the tools below manually.")
			}
			res := runner.Run(cmd.Context(), set)
			printResult(res)
			if !res.OK() {
				return fmt.Errorf("%d prerequisite(s) still missing", len(res.Missing))
			}
			return nil
		},
	}
}

// printResult renders a friendly, plain-language summary for non-technical users.
func printResult(res fw.Result) {
	for _, name := range res.Satisfied {
		pterm.Success.Printf("✓ %s\n", name)
	}
	for _, name := range res.Installed {
		pterm.Success.Printf("✓ %s (installed)\n", name)
	}
	for _, m := range res.Missing {
		pterm.Error.Printf("✗ %s is not installed\n", m.Name)
		if m.DocsURL != "" {
			pterm.Info.Printf("   How to install: %s\n", m.DocsURL)
		}
		if m.Err != nil {
			pterm.Debug.Printf("   (%v)\n", m.Err)
		}
	}
	if res.OK() {
		pterm.Success.Println("All prerequisites are satisfied.")
	}
}
