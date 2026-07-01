package k8s

import "fmt"

// Prompter abstracts the interactive prompts the context-selection flow needs,
// so the flow is fully testable without a real terminal. The production
// implementation (wired by the command layer) maps these onto the shared UI
// helpers: Confirm → ui.ConfirmActionInteractive, Choose → ui.SelectFromList.
type Prompter interface {
	// Confirm asks a yes/no question with a default answer.
	Confirm(message string, defaultYes bool) (bool, error)
	// Choose presents a labelled list and returns the selected index.
	Choose(label string, options []string) (int, error)
}

// SelectContext runs the context-selection flow for non-technical users (req 27):
//
//  1. If there is a current context, ask "Use the current context 'X'?"
//     (default yes) — the safe, one-keystroke path.
//  2. If the user declines, or there is no current context, show every context
//     and let them pick one.
//
// It returns the chosen context name.
func SelectContext(contexts []ContextInfo, current string, p Prompter) (string, error) {
	if len(contexts) == 0 {
		return "", fmt.Errorf("no kubeconfig contexts found — is a cluster configured?")
	}

	// Step 1: offer the current context as the default.
	if current != "" {
		use, err := p.Confirm(fmt.Sprintf("Use the current context %q?", current), true)
		if err != nil {
			return "", err
		}
		if use {
			return current, nil
		}
	}

	// Step 2: let the user choose from the full list.
	labels := make([]string, len(contexts))
	for i, c := range contexts {
		if c.Cluster != "" {
			labels[i] = fmt.Sprintf("%s  (cluster: %s)", c.Name, c.Cluster)
		} else {
			labels[i] = c.Name
		}
	}

	idx, err := p.Choose("Select a context", labels)
	if err != nil {
		return "", err
	}
	if idx < 0 || idx >= len(contexts) {
		return "", fmt.Errorf("invalid context selection")
	}
	return contexts[idx].Name, nil
}
