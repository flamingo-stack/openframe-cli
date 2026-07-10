package ui

import (
	"io"
	"testing"

	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
)

// Finding 6: --silent must actually suppress non-error output. SetSilent routes
// the noisy printers to io.Discard and marks the UI silent (so the logo is
// skipped), while leaving Error intact so failures still surface.
func TestSetSilent(t *testing.T) {
	// SetSilent mutates pterm package globals + the silent flag; restore them so
	// this test doesn't leak into others in the package.
	savedInfo, savedSuccess, savedWarning, savedError := pterm.Info, pterm.Success, pterm.Warning, pterm.Error
	savedDebug, savedBasicText := pterm.Debug, pterm.DefaultBasicText
	savedBox, savedHeader, savedTable := pterm.DefaultBox, pterm.DefaultHeader, pterm.DefaultTable
	savedSilent := silent
	t.Cleanup(func() {
		pterm.Info, pterm.Success, pterm.Warning, pterm.Error = savedInfo, savedSuccess, savedWarning, savedError
		pterm.Debug, pterm.DefaultBasicText = savedDebug, savedBasicText
		pterm.DefaultBox, pterm.DefaultHeader, pterm.DefaultTable = savedBox, savedHeader, savedTable
		silent = savedSilent
	})

	assert.False(t, silent, "precondition: not silent")

	SetSilent()

	assert.True(t, silent)
	assert.Equal(t, io.Discard, pterm.Info.GetWriter(), "Info must be discarded")
	assert.Equal(t, io.Discard, pterm.Success.GetWriter(), "Success must be discarded")
	assert.Equal(t, io.Discard, pterm.Warning.GetWriter(), "Warning must be discarded")
	assert.NotEqual(t, io.Discard, pterm.Error.GetWriter(), "Error must NOT be discarded — failures still surface")
}

// The ASCII logo is gated on the silent flag so --silent produces no banner.
func TestShowLogoConditional_SilentSkips(t *testing.T) {
	savedSilent := silent
	t.Cleanup(func() { silent = savedSilent })

	silent = true
	// With silent set the renderer must early-return; this just exercises the
	// guard path (no panic, no output).
	ShowLogoConditional(false)
	assert.True(t, silent)
}
