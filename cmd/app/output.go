package app

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// outputFormat reads and validates the --output flag (text|json). Commands
// without the flag report "text".
func outputFormat(cmd *cobra.Command) (string, error) {
	f, _ := cmd.Flags().GetString("output")
	switch f {
	case "", "text":
		return "text", nil
	case "json":
		return "json", nil
	default:
		return "", fmt.Errorf("invalid --output %q (want \"text\" or \"json\")", f)
	}
}

// isJSONOutput reports whether the command was asked for JSON. It is used in
// PersistentPreRunE (before flag validation) to switch to machine mode.
func isJSONOutput(cmd *cobra.Command) bool {
	f, _ := cmd.Flags().GetString("output")
	return f == "json"
}

// addOutputFlag registers the shared --output/-o flag.
func addOutputFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("output", "o", "text", "Output format: text or json")
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
