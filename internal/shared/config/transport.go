package config

import (
	"crypto/tls"
	"net/http"

	"k8s.io/client-go/rest"
)

// ApplyInsecureTransport forces the rest.Config to use a transport that truly skips TLS verification.
// This is the most aggressive way to ensure certificate checks are skipped, bypassing
// Go's default http.Client behavior which can cache or fail TLS handshakes at a deeper level.
//
// This is necessary for k3d/WSL2 environments where:
// - The API server's certificate is issued to a hostname that doesn't match 127.0.0.1
// - Go's net/http may cache TLS state or fail handshakes despite rest.Config.Insecure=true
// - Dynamically mapped ports on WSL may have connection timing issues
//
// WARNING: Only use this for local development clusters. Never in production.
func ApplyInsecureTransport(config *rest.Config) *rest.Config {
	if config == nil {
		return nil
	}

	// Create a custom transport that explicitly skips TLS verification at the HTTP level
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			// The magic setting: skips all server certificate verification checks
			InsecureSkipVerify: true,
		},
	}

	// Set the custom RoundTripper on the rest.Config
	// This overrides any default transport behavior
	config.Transport = transport

	// Clear the old TLS config fields to prevent any conflicts
	config.TLSClientConfig = rest.TLSClientConfig{}

	// Also explicitly set Insecure = true as a secondary check
	// Some code paths may check this flag independently
	config.Insecure = true

	return config
}
