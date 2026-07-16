package ui

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/pterm/pterm"
	"golang.org/x/term"
)

// IsNonInteractive reports whether the CLI must avoid interactive prompts:
// either a recognized CI environment, or stdin is not a terminal (piped /
// redirected, as in CI). Prompt-driven flows (e.g. the prerequisite gate) should
// take their non-interactive path so they never block waiting for a Y/N that
// no one can type.
func IsNonInteractive() bool {
	for _, v := range []string{"CI", "GITHUB_ACTIONS", "GITLAB_CI", "CIRCLECI"} {
		if os.Getenv(v) != "" {
			return true
		}
	}
	return !term.IsTerminal(int(os.Stdin.Fd()))
}

// confirm shows pterm's styled interactive y/N confirmation with the given
// default. It is the single implementation behind the exported confirm helpers.
func confirm(message string, defaultYes bool) (bool, error) {
	return pterm.DefaultInteractiveConfirm.
		WithDefaultText(message).
		WithDefaultValue(defaultYes).
		Show()
}

// ConfirmActionInteractive prompts the user with a polished interactive
// confirmation (colored styling, clear y/N format) defaulting to defaultValue.
func ConfirmActionInteractive(message string, defaultValue bool) (bool, error) {
	return confirm(message, defaultValue)
}

// RequireConfirmation prompts like ConfirmActionInteractive, but in a
// non-interactive environment (CI, piped stdin) it fails fast with guidance
// instead of blocking on a prompt no one can answer — or worse, silently
// proceeding with a destructive action. flagHint names the flag that skips the
// prompt (e.g. "--yes", "--force"); callers must check that flag BEFORE calling.
func RequireConfirmation(message, flagHint string, defaultValue bool) (bool, error) {
	if IsNonInteractive() {
		return false, fmt.Errorf("confirmation required but the session is non-interactive; re-run with %s", flagHint)
	}
	return confirm(message, defaultValue)
}

// ConfirmDeletion prompts for deletion confirmation (defaults to No).
func ConfirmDeletion(resourceType, resourceName string) (bool, error) {
	return confirm(fmt.Sprintf("Are you sure you want to delete %s '%s'?", resourceType, pterm.Cyan(resourceName)), false)
}

// selectTemplates is the shared styling for the interactive list selectors.
var selectTemplates = &promptui.SelectTemplates{
	Label:    "{{ . }}?",
	Active:   "→ {{ . | cyan }}",
	Inactive: "  {{ . | white }}",
	Selected: "✓ {{ . | green }}",
}

// SelectFromList prompts the user to select from a list of options.
func SelectFromList(label string, items []string) (int, string, error) {
	prompt := promptui.Select{
		Label:     label,
		Items:     items,
		Templates: selectTemplates,
	}
	return prompt.Run()
}

// ValidateNonEmpty validates that input is not empty after trimming
func ValidateNonEmpty(fieldName string) func(string) error {
	return func(input string) error {
		if strings.TrimSpace(input) == "" {
			return fmt.Errorf("%s cannot be empty", fieldName)
		}
		return nil
	}
}

// ValidateIntRange validates that input is an integer within specified range
func ValidateIntRange(min, max int, fieldName string) func(string) error {
	return func(input string) error {
		val, err := strconv.Atoi(input)
		if err != nil {
			return fmt.Errorf("please enter a valid number for %s", fieldName)
		}
		if val < min || val > max {
			return fmt.Errorf("%s must be between %d and %d", fieldName, min, max)
		}
		return nil
	}
}
