package ui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/pterm/pterm"
)

const (
	// Logo configuration constants
	logoTitle         = "OpenFrame Platform Bootstrapper"
	borderChar        = "━"
	topLeftCorner     = "┏"
	topRightCorner    = "┓"
	bottomLeftCorner  = "┗"
	bottomRightCorner = "┛"
	verticalChar      = "┃"
)

var (
	// TestMode suppresses logo output during testing
	TestMode bool

	// logoArt contains the beautiful Unicode logo for OpenFrame
	logoArt = []string{
		" ██████╗ ██████╗ ███████╗███╗   ██╗███████╗██████╗  █████╗ ███╗   ███╗███████╗",
		"██╔═══██╗██╔══██╗██╔════╝████╗  ██║██╔════╝██╔══██╗██╔══██╗████╗ ████║██╔════╝",
		"██║   ██║██████╔╝█████╗  ██╔██╗ ██║█████╗  ██████╔╝███████║██╔████╔██║█████╗  ",
		"██║   ██║██╔═══╝ ██╔══╝  ██║╚██╗██║██╔══╝  ██╔══██╗██╔══██║██║╚██╔╝██║██╔══╝  ",
		"╚██████╔╝██║     ███████╗██║ ╚████║██║     ██║  ██║██║  ██║██║ ╚═╝ ██║███████╗",
		" ╚═════╝ ╚═╝     ╚══════╝╚═╝  ╚═══╝╚═╝     ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝╚══════╝",
	}
)

// ShowLogo displays the OpenFrame ASCII logo
func ShowLogo() {
	ShowLogoConditional(false)
}

// ShowLogoConditional displays the OpenFrame ASCII logo with optional suppression
func ShowLogoConditional(suppress bool) {
	if TestMode || suppress || silent {
		return
	}

	// Check if we should use fancy formatting
	// Use plain logo by default to avoid escape sequence issues
	// Only use fancy logo when explicitly requested or in known good environments
	useFancy := false

	// Check for explicit preference
	if os.Getenv("OPENFRAME_FANCY_LOGO") == "true" {
		useFancy = true
	} else if os.Getenv("OPENFRAME_FANCY_LOGO") == "false" {
		useFancy = false
	} else {
		// Auto-detect: use fancy only for truly interactive terminals with color support
		useFancy = isTerminalEnvironment() && pterm.PrintColor && os.Getenv("TERM") != "" && os.Getenv("NO_COLOR") == ""
	}

	if useFancy {
		showFancyLogo()
	} else {
		showPlainLogo()
	}
}

// contextKey is used for context values
type contextKey string

const suppressLogoKey contextKey = "suppressLogo"

// ShowLogoWithContext displays the logo unless suppressed via context
func ShowLogoWithContext(ctx context.Context) {
	if TestMode {
		return
	}

	// Check if logo should be suppressed via context
	if suppress, ok := ctx.Value(suppressLogoKey).(bool); ok && suppress {
		return
	}

	ShowLogoConditional(false)
}

// isTerminalEnvironment checks if we're running in a proper terminal
func isTerminalEnvironment() bool {
	// Check if stdout is a terminal
	if stat, err := os.Stdout.Stat(); err == nil {
		return (stat.Mode() & os.ModeCharDevice) != 0
	}
	return false
}

// showFancyLogo displays the logo using pterm for enhanced terminals
func showFancyLogo() {
	// Create custom box style with gradient colors
	boxStyle := pterm.NewStyle(pterm.FgCyan, pterm.Bold)
	titleStyle := pterm.NewStyle(pterm.FgLightCyan, pterm.Bold)

	// Use pterm default box with custom styling
	pterm.DefaultBox.BoxStyle = boxStyle

	// Create padded logo lines with a vertical cyan→purple gradient. pterm
	// downsamples RGB on terminals without truecolor, so this is safe
	// everywhere colors are safe at all.
	gradStart := pterm.NewRGB(0, 200, 255)
	gradEnd := pterm.NewRGB(168, 108, 255)
	paddedLines := make([]string, 0, len(logoArt)+3)
	paddedLines = append(paddedLines, "") // Top padding
	for i, line := range logoArt {
		c := gradStart.Fade(0, float32(len(logoArt)-1), float32(i), gradEnd)
		paddedLines = append(paddedLines, " "+c.Sprint(line)+" ")
	}
	paddedLines = append(paddedLines, "") // Bottom padding

	logo := strings.Join(paddedLines, "\n")

	// Create styled title
	styledTitle := titleStyle.Sprint(" " + logoTitle + " ")

	pterm.DefaultBox.WithTitle(styledTitle).
		WithTitleTopCenter().
		WithBoxStyle(boxStyle).
		Println(logo)

	// Add a subtle separator after the logo
	fmt.Println()
}

// plainLogoLines builds the plain-text logo box. Every width is derived from
// the art itself — hardcoded border lengths previously disagreed with the art
// width (88 vs 84 vs 82 runes), rendering the box broken.
func plainLogoLines() []string {
	// All art lines share one width; pad one space each side, plus the borders.
	inner := utf8.RuneCountInString(logoArt[0]) + 2

	// Center the title in the top border.
	titleSegment := "┫ " + logoTitle + " ┣"
	fill := inner - utf8.RuneCountInString(titleSegment)
	topBorder := topLeftCorner + strings.Repeat(borderChar, fill/2) + titleSegment + strings.Repeat(borderChar, fill-fill/2) + topRightCorner

	middleSeparator := verticalChar + strings.Repeat("─", inner) + verticalChar
	bottomBorder := bottomLeftCorner + strings.Repeat(borderChar, inner) + bottomRightCorner

	lines := make([]string, 0, len(logoArt)+4)
	lines = append(lines, topBorder, middleSeparator)
	for _, line := range logoArt {
		lines = append(lines, verticalChar+" "+line+" "+verticalChar)
	}
	lines = append(lines, middleSeparator, bottomBorder)
	return lines
}

// showPlainLogo displays a simple plain text logo for non-terminal environments
func showPlainLogo() {
	for _, line := range plainLogoLines() {
		fmt.Println(line)
	}
	fmt.Println()
}
