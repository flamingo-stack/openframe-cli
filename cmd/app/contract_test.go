package app

import (
	"testing"

	"github.com/flamingo-stack/openframe-cli/tests/testutil"
	"github.com/stretchr/testify/assert"
)

// These tests freeze the public CLI contract of the `app` command tree — names,
// aliases, subcommands, flags (name/shorthand/type/default), and the readonly
// annotations. Any accidental drift in the user-facing surface fails here.

func TestAppContract_RootShape(t *testing.T) {
	app := GetAppCmd()

	assert.Equal(t, "app", app.Name())
	assert.ElementsMatch(t, []string{"chart", "c"}, app.Aliases, "chart/c aliases are part of the contract")
	assert.NotEmpty(t, app.Short)

	testutil.AssertSubcommands(t, app, "install", "status", "access", "uninstall")
}

func TestAppContract_InstallFlags(t *testing.T) {
	install := testutil.FindSubcommand(t, GetAppCmd(), "install")

	testutil.AssertFlags(t, install, []testutil.FlagSpec{
		{Name: "force", Shorthand: "f", Type: "bool", Default: "false"},
		{Name: "dry-run", Type: "bool", Default: "false"},
		{Name: "github-repo", Type: "string", Default: "https://github.com/flamingo-stack/openframe-oss-tenant"},
		{Name: "github-branch", Type: "string", Default: "main"},
		{Name: "cert-dir", Type: "string", Default: ""},
		{Name: "deployment-mode", Type: "string", Default: ""},
		{Name: "non-interactive", Type: "bool", Default: "false"},
		{Name: "context", Type: "string", Default: ""},
	})
}

func TestAppContract_StatusAndAccessAreReadonly(t *testing.T) {
	app := GetAppCmd()
	for _, name := range []string{"status", "access"} {
		cmd := testutil.FindSubcommand(t, app, name)
		assert.Equalf(t, "true", cmd.Annotations["readonly"], "%s must be annotated readonly (skips the prereq gate)", name)
		testutil.AssertFlags(t, cmd, []testutil.FlagSpec{
			{Name: "context", Type: "string", Default: ""},
			{Name: "output", Shorthand: "o", Type: "string", Default: "text"},
		})
	}
}

func TestAppContract_UninstallFlags(t *testing.T) {
	uninstall := testutil.FindSubcommand(t, GetAppCmd(), "uninstall")

	// Uninstall mutates the cluster → must NOT be readonly.
	assert.NotEqual(t, "true", uninstall.Annotations["readonly"], "uninstall must run the prereq gate")
	testutil.AssertFlags(t, uninstall, []testutil.FlagSpec{
		{Name: "context", Type: "string", Default: ""},
		{Name: "yes", Shorthand: "y", Type: "bool", Default: "false"},
		{Name: "delete-namespace", Type: "bool", Default: "false"},
	})
}
