package config

import (
	"k8s.io/client-go/rest"
)

// ApplyInsecureTLSConfig configures the rest.Config to skip TLS certificate verification
// while preserving client certificate authentication data.
//
// This approach lets client-go build the TLS config correctly (ingesting client certs/keys
// from the kubeconfig), then overrides only the certificate verification part.
//
// This is necessary for k3d/WSL2 environments where:
// - The API server's certificate is issued to a hostname that doesn't match 127.0.0.1
// - Self-signed certificates are used for local development clusters
//
// WARNING: Only use this for local development clusters. Never in production.
func ApplyInsecureTLSConfig(config *rest.Config) *rest.Config {
	if config == nil {
		return nil
	}

	// 1. Force TLS bypass - this tells client-go to skip server certificate verification
	config.Insecure = true

	// 2. Clear CA data to prevent certificate validation conflicts
	// The CA data would otherwise be used to verify the server certificate
	config.TLSClientConfig.CAData = nil
	config.TLSClientConfig.CAFile = ""

	// 3. Clear any custom transport that might conflict with client-go's handling
	// This ensures client-go builds the transport itself with the auth data intact
	config.Transport = nil
	config.WrapTransport = nil

	// NOTE: We intentionally preserve these authentication fields:
	// - config.TLSClientConfig.CertData (client certificate)
	// - config.TLSClientConfig.KeyData (client private key)
	// - config.TLSClientConfig.CertFile
	// - config.TLSClientConfig.KeyFile
	// These are used by client-go to authenticate to the API server

	return config
}

// ApplyInsecureTransport is an alias for ApplyInsecureTLSConfig for backward compatibility.
// Deprecated: Use ApplyInsecureTLSConfig instead.
func ApplyInsecureTransport(config *rest.Config) *rest.Config {
	return ApplyInsecureTLSConfig(config)
}
