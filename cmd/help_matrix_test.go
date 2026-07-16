package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// These tests sweep the ENTIRE command tree, so every current and future
// command is covered automatically — no per-command test to forget.
//
// Contract per command:
//   - `<path> --help` succeeds and prints a Usage section (help must never be
//     blocked by prerequisite gates, cluster access, or the network);
//   - it has a non-empty Short description (shown in the parent's command list);
//   - every locally-declared flag has a non-empty usage string.

// collectCommandPaths returns the arg-path of every command in the tree.
func collectCommandPaths(c *cobra.Command, prefix []string) [][]string {
	var out [][]string
	for _, sub := range c.Commands() {
		if sub.Name() == "help" {
			continue // cobra's built-in
		}
		path := append(append([]string{}, prefix...), sub.Name())
		out = append(out, path)
		out = append(out, collectCommandPaths(sub, path)...)
	}
	return out
}

func TestEveryCommand_HelpWorks(t *testing.T) {
	paths := collectCommandPaths(GetRootCmd(DefaultVersionInfo), nil)
	if len(paths) < 15 {
		t.Fatalf("expected a substantial command tree, found only %d commands", len(paths))
	}

	for _, path := range paths {
		t.Run(strings.Join(path, "_"), func(t *testing.T) {
			// Fresh tree per execution: cobra commands carry parsed-flag state.
			root := GetRootCmd(DefaultVersionInfo)
			var buf bytes.Buffer
			root.SetOut(&buf)
			root.SetErr(&buf)
			root.SetArgs(append(append([]string{}, path...), "--help"))

			if err := root.Execute(); err != nil {
				t.Fatalf("%s --help failed: %v", strings.Join(path, " "), err)
			}
			out := buf.String()
			if !strings.Contains(out, "Usage:") {
				t.Errorf("%s --help printed no Usage section:\n%s", strings.Join(path, " "), out)
			}
		})
	}
}

func TestEveryCommand_HasShortAndFlagUsages(t *testing.T) {
	root := GetRootCmd(DefaultVersionInfo)
	for _, path := range collectCommandPaths(root, nil) {
		cmd, _, err := root.Find(path)
		if err != nil {
			t.Fatalf("find %v: %v", path, err)
		}
		name := strings.Join(path, " ")
		if strings.TrimSpace(cmd.Short) == "" {
			t.Errorf("%s: empty Short description", name)
		}
		cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
			if strings.TrimSpace(f.Usage) == "" {
				t.Errorf("%s: flag --%s has no usage text", name, f.Name)
			}
		})
	}
}
