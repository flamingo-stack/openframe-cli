package app

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

// outputFormat reads and validates the --output flag (text|json|yaml). Commands
// without the flag report "text".
func outputFormat(cmd *cobra.Command) (string, error) {
	f, _ := cmd.Flags().GetString("output")
	switch f {
	case "", "text":
		return "text", nil
	case "json":
		return "json", nil
	case "yaml":
		return "yaml", nil
	default:
		return "", fmt.Errorf("invalid --output %q (want \"text\", \"json\", or \"yaml\")", f)
	}
}

// isMachineOutput reports whether the command was asked for a machine-readable
// format (json or yaml). It is used in PersistentPreRunE (before flag
// validation) to switch to machine mode: no logo, no prerequisite gate, clean
// stdout for scripts.
func isMachineOutput(cmd *cobra.Command) bool {
	switch f, _ := cmd.Flags().GetString("output"); f {
	case "json", "yaml":
		return true
	default:
		return false
	}
}

// addOutputFlag registers the shared --output/-o flag.
func addOutputFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("output", "o", "text", "Output format: text, json, or yaml")
}

// printJSON writes v to stdout as indented JSON.
func printJSON(v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}
	fmt.Println(string(b))
	return nil
}

// printYAML writes v to stdout as YAML. sigs.k8s.io/yaml round-trips through
// JSON, so it reuses the same `json:` struct tags — json and yaml output carry
// identical field names.
func printYAML(v any) error {
	b, err := yaml.Marshal(v)
	if err != nil {
		return fmt.Errorf("encoding YAML: %w", err)
	}
	fmt.Print(string(b)) // yaml.Marshal already terminates with a newline
	return nil
}

// renderMachine writes v in the requested machine-readable format. Callers gate
// on format != "text" (the human-readable path); json is the default machine
// format, yaml the alternative.
func renderMachine(format string, v any) error {
	if format == "yaml" {
		return printYAML(v)
	}
	return printJSON(v)
}
