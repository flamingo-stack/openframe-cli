package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/shared/redact"
)

// TestLoadExistingValues_RegistersSecrets is the B5 guard for the redaction
// chokepoint: loading a values file that carries credentials must register
// them, so no later code path (verbose logs, command errors) can echo them.
// Before this, redact.RegisterSecret had ZERO call sites — the exact-match
// half of the redaction system was inert.
func TestLoadExistingValues_RegistersSecrets(t *testing.T) {
	redact.ClearSecrets()
	t.Cleanup(redact.ClearSecrets)

	content := `
registry:
  docker:
    username: user
    password: docker-pass-secret-1
deployment:
  ingress:
    ngrok:
      credentials:
        apiKey: ngrok-api-key-secret-2
        authtoken: ngrok-authtoken-secret-3
`
	p := filepath.Join(t.TempDir(), "openframe-helm-values.yaml")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := NewHelmValuesModifier().LoadExistingValues(p); err != nil {
		t.Fatalf("LoadExistingValues: %v", err)
	}

	line := "helm upgrade --set registry.docker.password=docker-pass-secret-1 --set ngrok.apiKey=ngrok-api-key-secret-2 --set ngrok.authtoken=ngrok-authtoken-secret-3"
	got := redact.Redact(line)
	for _, secret := range []string{"docker-pass-secret-1", "ngrok-api-key-secret-2", "ngrok-authtoken-secret-3"} {
		if strings.Contains(got, secret) {
			t.Errorf("secret %q survived redaction after loading the values file:\n%s", secret, got)
		}
	}
}
