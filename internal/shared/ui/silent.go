package ui

import (
	"io"

	"github.com/pterm/pterm"
)

// silent records whether --silent suppressed non-error output. Read by the logo
// renderer so it can honor the flag.
var silent bool

// SetSilent honors the --silent flag's contract ("suppress all output except
// errors"): it routes every non-error pterm printer to io.Discard and marks the
// UI silent so the ASCII logo is skipped. Error and Fatal printers are left
// untouched so failures are still surfaced. It mutates pterm's package-level
// printers, so it must be called once, early — from the root command's
// PersistentPreRun — and is not meant to be reversed within a process.
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
	// Interactive prompt printers (DefaultInteractiveConfirm/TextInput) are left
	// alone on purpose: discarding their writer would hide the prompt text while
	// it still blocks on stdin — a silent hang, worse than a visible prompt.
}
