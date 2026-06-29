package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedact_RegisteredSecret(t *testing.T) {
	ClearSecrets()
	defer ClearSecrets()

	RegisterSecret("ghp_supersecrettoken")
	out := Redact("helm upgrade --set token=ghp_supersecrettoken --namespace argocd")
	assert.NotContains(t, out, "ghp_supersecrettoken")
	assert.Contains(t, out, "***")
	assert.Contains(t, out, "--namespace argocd", "non-secret content must survive")
}

func TestRedact_URLCredentials(t *testing.T) {
	ClearSecrets()
	out := Redact("cloning https://x-access-token:ghp_abc123def456@github.com/org/repo")
	assert.NotContains(t, out, "ghp_abc123def456")
	assert.Contains(t, out, "x-access-token:***@github.com", "username kept, password masked")
}

func TestRedact_URLCredentialsWithoutRegistration(t *testing.T) {
	ClearSecrets()
	// Token was never registered, but the URL structure still reveals it.
	out := Redact("https://user:p4ssw0rd-unregistered@example.com/x")
	assert.NotContains(t, out, "p4ssw0rd-unregistered")
}

func TestRegisterSecret_IgnoresShortValues(t *testing.T) {
	ClearSecrets()
	defer ClearSecrets()
	RegisterSecret("ab") // too short — would over-redact common text
	assert.Equal(t, "ab cat tab", Redact("ab cat tab"))
}

func TestRedact_LongerSecretFirst(t *testing.T) {
	ClearSecrets()
	defer ClearSecrets()
	RegisterSecret("secret")
	RegisterSecret("secretLONGER")
	out := Redact("value=secretLONGER")
	assert.Equal(t, "value=***", out, "longer secret must be redacted as a whole")
}

func TestRedact_NoSecretsIsIdentity(t *testing.T) {
	ClearSecrets()
	in := "kubectl get pods -n argocd"
	assert.Equal(t, in, Redact(in))
}
