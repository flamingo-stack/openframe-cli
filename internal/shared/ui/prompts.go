package ui

import (
	stderrors "errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
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

// ErrPromptInterrupted is returned when the user aborts an interactive prompt
// with Ctrl+C (huh's quit binding; Esc only manages the list filter). Its text
// is exactly "interrupted": the shared error handler matches it structurally
// via errors.Is and by that string (see errors.isInterruption) and prints a
// friendly "cancelled" notice instead of a failure panel.
var ErrPromptInterrupted = stderrors.New("interrupted")

// normalizePromptError maps huh's abort sentinel onto ErrPromptInterrupted so
// every prompt in the CLI reports user cancellation the same way.
func normalizePromptError(err error) error {
	if stderrors.Is(err, huh.ErrUserAborted) {
		return ErrPromptInterrupted
	}
	return err
}

// selectPageSize caps how many rows a long select shows at once; longer lists
// scroll (and are two filter keystrokes away from any entry).
const selectPageSize = 10

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

// SelectFromList prompts the user to select from a list of options. Pressing
// "/" filters the list (fuzzy, case-insensitive), so a 30-cluster list is a
// few keystrokes away from the right entry; arrow keys navigate as before.
// The label is rendered with a trailing "?", matching the CLI's historical
// picker wording.
func SelectFromList(label string, items []string) (int, string, error) {
	return runSelect(label+"?", items)
}

// SelectOption prompts the user to pick from a short fixed list (wizard steps,
// yes/no style choices). Same interaction as SelectFromList; the separate name
// keeps call sites explicit about intent.
func SelectOption(label string, items []string) (int, string, error) {
	return runSelect(label, items)
}

// requireInteractive fails fast when no one can answer a prompt (CI, piped
// stdin). Unix CI already fails fast — bubbletea cannot open /dev/tty there —
// but Windows runners DO have a console, where a prompt blocks reading it until
// the job times out (a wizard test hung a Windows runner for its full 10
// minutes exactly this way). Same contract as RequireConfirmation.
func requireInteractive(label string) error {
	if IsNonInteractive() {
		return fmt.Errorf("prompt %q requires an interactive terminal", label)
	}
	return nil
}

func runSelect(label string, items []string) (int, string, error) {
	if err := requireInteractive(label); err != nil {
		return 0, "", err
	}
	options := make([]huh.Option[int], len(items))
	for i, item := range items {
		options[i] = huh.NewOption(item, i)
	}
	var idx int
	sel := huh.NewSelect[int]().
		Title(label).
		Options(options...).
		Value(&idx)
	if len(items) > selectPageSize {
		// Long lists scroll; tell the user the filter exists. huh's Height
		// includes the title and description rows (viewport = height - chrome),
		// so +2 shows exactly selectPageSize option rows; while filtering the
		// title row becomes the filter input — still one row, so this holds.
		sel = sel.
			Description("type / to filter").
			Height(selectPageSize + 2)
	}
	if err := sel.Run(); err != nil {
		return 0, "", normalizePromptError(err)
	}
	return idx, items[idx], nil
}

// PromptInput shows a single-line text prompt. defaultVal pre-fills the field
// (editable in place, so Enter accepts it as-is); validate, when non-nil, runs
// on each submission attempt and blocks until it passes. The result is
// whitespace-trimmed.
func PromptInput(label, defaultVal string, validate func(string) error) (string, error) {
	if err := requireInteractive(label); err != nil {
		return "", err
	}
	value := defaultVal
	in := huh.NewInput().
		Title(label).
		Value(&value)
	if validate != nil {
		in = in.Validate(validate)
	}
	if err := in.Run(); err != nil {
		return "", normalizePromptError(err)
	}
	return strings.TrimSpace(value), nil
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
