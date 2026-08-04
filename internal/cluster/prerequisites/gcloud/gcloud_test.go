package gcloud

import (
	"errors"
	"strings"
	"testing"
)

// stubLookPath makes only the named commands resolvable, restoring the real
// lookup after the test.
func stubLookPath(t *testing.T, available ...string) {
	t.Helper()
	orig := lookPath
	t.Cleanup(func() { lookPath = orig })
	lookPath = func(cmd string) (string, error) {
		for _, name := range available {
			if cmd == name {
				return "/usr/bin/" + cmd, nil
			}
		}
		return "", errors.New("not found: " + cmd)
	}
}

// stubRunQuiet records every invocation and never executes anything.
func stubRunQuiet(t *testing.T, err error) *[][]string {
	t.Helper()
	orig := runQuiet
	t.Cleanup(func() { runQuiet = orig })
	var calls [][]string
	runQuiet = func(name string, args ...string) error {
		calls = append(calls, append([]string{name}, args...))
		return err
	}
	return &calls
}

// The auth plugin is installed THROUGH gcloud's component manager: without
// gcloud the install must fail up front, not shell out into the void.
func TestAuthPluginInstall_RequiresGcloud(t *testing.T) {
	stubLookPath(t) // gcloud missing
	calls := stubRunQuiet(t, nil)

	err := NewAuthPluginInstaller().Install()
	if err == nil {
		t.Fatal("expected an error when gcloud is missing")
	}
	if !strings.Contains(err.Error(), "gcloud") {
		t.Fatalf("error must name gcloud as the missing piece, got: %v", err)
	}
	if len(*calls) != 0 {
		t.Fatalf("no command may run without gcloud, got %v", *calls)
	}
}

func TestAuthPluginInstall_UsesGcloudComponents(t *testing.T) {
	stubLookPath(t, "gcloud")
	calls := stubRunQuiet(t, nil)

	if err := NewAuthPluginInstaller().Install(); err != nil {
		t.Fatalf("install must succeed when gcloud components succeeds: %v", err)
	}
	if len(*calls) != 1 {
		t.Fatalf("want one gcloud invocation, got %v", *calls)
	}
	got := strings.Join((*calls)[0], " ")
	if want := "gcloud components install gke-gcloud-auth-plugin --quiet"; got != want {
		t.Fatalf("command = %q, want %q", got, want)
	}
}

// Package-manager gcloud installs lack the component manager; the failure must
// point at the OS-package alternative instead of a bare exit code.
func TestAuthPluginInstall_ComponentFailureCarriesHint(t *testing.T) {
	stubLookPath(t, "gcloud")
	stubRunQuiet(t, errors.New("components manager disabled"))

	err := NewAuthPluginInstaller().Install()
	if err == nil {
		t.Fatal("expected the component-manager failure to surface")
	}
	if !strings.Contains(err.Error(), "cloud.google.com") {
		t.Fatalf("error must carry the docs pointer for package-manager installs, got: %v", err)
	}
}

func TestGetInstallHelp_NonEmpty(t *testing.T) {
	if NewGcloudInstaller().GetInstallHelp() == "" {
		t.Fatal("gcloud install help must not be empty")
	}
	if NewAuthPluginInstaller().GetInstallHelp() == "" {
		t.Fatal("auth plugin install help must not be empty")
	}
}
