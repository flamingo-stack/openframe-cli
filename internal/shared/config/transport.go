package config

import (
	"crypto/tls"
	"crypto/x509"
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

	// Build the TLS config for the custom transport
	tlsConfig := &tls.Config{
		// Skip server certificate verification (the main goal)
		InsecureSkipVerify: true,
	}

	// CRITICAL: Preserve client certificate authentication data
	// The client cert/key are used to authenticate to the Kubernetes API server
	if len(config.TLSClientConfig.CertData) > 0 && len(config.TLSClientConfig.KeyData) > 0 {
		cert, err := tls.X509KeyPair(config.TLSClientConfig.CertData, config.TLSClientConfig.KeyData)
		if err == nil {
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
	}

	// Also preserve CA data if present (for client cert verification by server)
	if len(config.TLSClientConfig.CAData) > 0 {
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(config.TLSClientConfig.CAData)
		tlsConfig.RootCAs = certPool
	}

	// Create a custom transport that explicitly skips TLS verification at the HTTP level
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	// Set the custom RoundTripper on the rest.Config
	// This overrides any default transport behavior
	config.Transport = transport

	// CRITICAL: Clear the internal TLS fields to prevent conflict with the custom transport.
	// The client-go library throws "not allowed" error when both Insecure=true and custom Transport are set.
	// Since the custom transport is now handling ALL TLS (including client certs), we can safely clear these.
	config.Insecure = false // Must be false to allow custom Transport!
	config.TLSClientConfig = rest.TLSClientConfig{}

	return config
}
