package ui

import (
	"io"
	"os"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// ApplyGlobalOutputFlags applies the root command's --silent/--verbose
// contract. Cobra runs only the CLOSEST parent's PersistentPreRunE, so any
// command group that defines its own hook shadows the root's — it must call
// this first, or --verbose is silently inert for its whole subtree (the
// executor's failed-command diagnostics, among ~40 other pterm.Debug sites,
// live behind it). --silent wins when both are given.
func ApplyGlobalOutputFlags(cmd *cobra.Command) {
	ApplyColorContract()
	silentFlag, _ := cmd.Flags().GetBool("silent")
	if silentFlag {
		SetSilent()
	}
	if p, _ := cmd.Flags().GetBool("plain"); p {
		SetPlain()
	}
	if v, _ := cmd.Flags().GetBool("verbose"); v && !silentFlag {
		pterm.EnableDebugMessages()
		// Timestamped debug lines: --verbose exists to correlate the CLI's
		// actions with cluster events, which needs a clock on every line.
		pterm.Debug = *pterm.Debug.WithWriter(NewTimestampWriter(os.Stdout))
	}
	// Last: the theme reads IsPlain/IsSilent, which the flags above just set.
	ApplyStatusPrefixTheme()
}

// silent records whether --silent suppressed non-error output. Read by the logo
// renderer so it can honor the flag.
var silent bool

// SetSilent honors the --silent flag's contract ("suppress all output except
// errors"): it routes every non-error pterm printer to io.Discard and marks the
// UI silent so the ASCII logo is skipped. Error and Fatal printers are left
// untouched so failures are still surfaced. It mutates pterm's package-level
// printers, so it must be called once, early — from the root command's
// PersistentPreRun — and is not meant to be reversed within a process.
// IsSilent reports whether --silent was applied. Components that write to
// stdout through their OWN writer — the spinner prints its final line via
// pterm.Success.WithWriter(s.out), which overrides the io.Discard writer
// SetSilent installs on the package-level printers — must consult this, or
// they leak output that --silent promises to suppress.
func IsSilent() bool { return silent }

func SetSilent() {
	silent = true
	pterm.Info = *pterm.Info.WithWriter(io.Discard)
	pterm.Success = *pterm.Success.WithWriter(io.Discard)
	pterm.Warning = *pterm.Warning.WithWriter(io.Discard)
	pterm.Debug = *pterm.Debug.WithWriter(io.Discard)
	pterm.DefaultBasicText = *pterm.DefaultBasicText.WithWriter(io.Discard)
	pterm.DefaultBox = *pterm.DefaultBox.WithWriter(io.Discard)
	pterm.DefaultHeader = *pterm.DefaultHeader.WithWriter(io.Discard)
	pterm.DefaultTable = *pterm.DefaultTable.WithWriter(io.Discard)
	pterm.DefaultSection = *pterm.DefaultSection.WithWriter(io.Discard)
	// Interactive prompt printers (DefaultInteractiveConfirm/TextInput) are left
	// alone on purpose: discarding their writer would hide the prompt text while
	// it still blocks on stdin — a silent hang, worse than a visible prompt.
}
