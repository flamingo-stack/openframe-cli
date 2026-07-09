package testutil

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// FlagSpec describes the frozen contract of a single command-line flag: its
// name, optional shorthand, type, and default (stringified) value. Contract
// tests use it so any silent change to the CLI surface — a renamed flag, a
// dropped shorthand, a changed default — fails loudly.
type FlagSpec struct {
	Name      string
	Shorthand string
	Type      string // cobra's flag Value.Type(): "bool", "string", "int", ...
	Default   string // flag.DefValue
}

// AssertFlag verifies a single flag on cmd matches the spec exactly.
func AssertFlag(t *testing.T, cmd *cobra.Command, spec FlagSpec) {
	t.Helper()
	f := cmd.Flags().Lookup(spec.Name)
	if !assert.NotNilf(t, f, "%s: flag --%s must exist", cmd.Name(), spec.Name) {
		return
	}
	assert.Equalf(t, spec.Shorthand, f.Shorthand, "%s: --%s shorthand", cmd.Name(), spec.Name)
	assert.Equalf(t, spec.Type, f.Value.Type(), "%s: --%s type", cmd.Name(), spec.Name)
	assert.Equalf(t, spec.Default, f.DefValue, "%s: --%s default", cmd.Name(), spec.Name)
}

// AssertFlags verifies every spec in specs against cmd.
func AssertFlags(t *testing.T, cmd *cobra.Command, specs []FlagSpec) {
	t.Helper()
	for _, s := range specs {
		AssertFlag(t, cmd, s)
	}
}

// FindSubcommand returns the immediate subcommand of parent with the given
// name (as it appears in `Use`), or fails the test.
func FindSubcommand(t *testing.T, parent *cobra.Command, name string) *cobra.Command {
	t.Helper()
	for _, c := range parent.Commands() {
		if c.Name() == name {
			return c
		}
	}
	t.Fatalf("%s: subcommand %q not found", parent.Name(), name)
	return nil
}

// AssertSubcommands verifies parent exposes exactly the named subcommands
// (order-independent). Extra or missing subcommands fail the test.
func AssertSubcommands(t *testing.T, parent *cobra.Command, names ...string) {
	t.Helper()
	got := map[string]bool{}
	for _, c := range parent.Commands() {
		got[c.Name()] = true
	}
	for _, n := range names {
		assert.Truef(t, got[n], "%s: expected subcommand %q", parent.Name(), n)
	}
	assert.Lenf(t, parent.Commands(), len(names), "%s: subcommand count (got %v, want %v)", parent.Name(), keys(got), names)
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
