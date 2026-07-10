package ui

import (
	"bufio"
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

// ConfirmAction prompts the user to confirm an action with friendly UX:
// - Enter = yes (default)
// - y = yes (immediate, no Enter needed)
// - n = no (immediate, no Enter needed)
func ConfirmAction(message string) (bool, error) {
	fmt.Printf("%s (Y/n): ", pterm.Bold.Sprint(message))

	// Get the file descriptor for stdin
	fd := int(os.Stdin.Fd())

	// Check if stdin is a terminal
	if !term.IsTerminal(fd) {
		// Fallback for non-terminal input (like pipes/tests)
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return false, err
		}
		input = strings.ToLower(strings.TrimSpace(input))
		return input == "" || input == "y" || input == "yes", nil
	}

	// Save the current terminal state
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return false, err
	}

	// restore returns the terminal to its cooked state. It is called before any
	// output on every exit path (so prints get normal newline handling) rather
	// than deferred. A failure would leave the terminal in raw mode — surface it
	// (Debug) instead of silently dropping it; the user's captured choice is the
	// function's real result, so we don't return the restore error.
	restore := func() {
		if err := term.Restore(fd, oldState); err != nil {
			pterm.Debug.Printf("failed to restore terminal state (it may be left in raw mode): %v\n", err)
		}
	}

	// Read single character
	buf := make([]byte, 1)
	for {
		_, err := os.Stdin.Read(buf)
		if err != nil {
			restore()
			return false, err
		}

		char := buf[0]

		switch char {
		case '\r', '\n': // Enter key
			restore()
			fmt.Println()
			return true, nil // Default to yes
		case 'y', 'Y':
			restore()
			fmt.Println("y")
			return true, nil
		case 'n', 'N':
			restore()
			fmt.Println("n")
			return false, nil
		case 3: // Ctrl+C
			restore()
			fmt.Println()
			return false, fmt.Errorf("interrupted")
			// Ignore other characters and continue reading
		}
	}
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

// GetInput prompts the user for text input
func GetInput(label, defaultValue string, validate func(string) error) (string, error) {
	prompt := promptui.Prompt{
		Label:    label,
		Default:  defaultValue,
		Validate: validate,
	}

	return prompt.Run()
}

// GetMultiChoice prompts the user to select multiple items from a list
func GetMultiChoice(label string, items []string, defaults []bool) ([]bool, error) {
	if len(items) != len(defaults) {
		return nil, fmt.Errorf("items and defaults must have the same length")
	}

	results := make([]bool, len(items))
	copy(results, defaults)

	for i, item := range items {
		confirmed, err := ConfirmAction(fmt.Sprintf("%s - %s", label, item))
		if err != nil {
			return nil, err
		}
		results[i] = confirmed
	}

	return results, nil
}

// HandleResourceSelection resolves a resource name from args or interactive
// selection: it returns the first non-empty arg if given, otherwise prompts the
// user to pick from items via SelectFromList.
func HandleResourceSelection(args []string, items []string, prompt string) (string, error) {
	if len(args) > 0 {
		resourceName := strings.TrimSpace(args[0])
		if resourceName == "" {
			return "", fmt.Errorf("resource name cannot be empty")
		}
		return resourceName, nil
	}
	if len(items) == 0 {
		return "", fmt.Errorf("no items available for selection")
	}
	_, selected, err := SelectFromList(prompt, items)
	if err != nil {
		return "", fmt.Errorf("selection failed: %w", err)
	}
	return selected, nil
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
