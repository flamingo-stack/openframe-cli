package config

import (
	"net"
	"net/url"
	"strings"

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

	// Only bypass TLS verification for LOCAL clusters (k3d/kind on the
	// loopback/host interface). For any other server — a cluster reached via
	// --context, a remote/production cluster — honor the kubeconfig's TLS
	// settings instead of silently disabling verification.
	if !isLocalAPIServer(config.Host) {
		return config
	}

	// 1. Force TLS bypass - this tells client-go to skip server certificate verification
	config.Insecure = true

	// 2. Clear CA data to prevent certificate validation conflicts
	// The CA data would otherwise be used to verify the server certificate
	config.CAData = nil
	config.CAFile = ""

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

// isLocalAPIServer reports whether serverURL points at a cluster running on
// this host — loopback (127.0.0.0/8, ::1), the unspecified address 0.0.0.0
// (used by k3d), localhost, or host.docker.internal (Docker Desktop's alias
// for this host). Anything else — including *.local names — is treated as
// remote: mDNS/legacy-AD `.local` domains are common for REAL corporate
// clusters, and a suffix match here silently disabled TLS verification for
// them, enabling MITM with zero warning (audit B5/T1-10).
func isLocalAPIServer(serverURL string) bool {
	host := serverURL
	if u, err := url.Parse(serverURL); err == nil && u.Host != "" {
		host = u.Hostname()
	} else if h, _, err := net.SplitHostPort(serverURL); err == nil {
		host = h
	}
	host = strings.ToLower(host)

	switch host {
	case "localhost", "0.0.0.0", "host.docker.internal", "::1":
		return true
	}
	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		return true
	}
	return false
}

// ApplyInsecureTransport is an alias for ApplyInsecureTLSConfig for backward compatibility.
// Deprecated: Use ApplyInsecureTLSConfig instead.
func ApplyInsecureTransport(config *rest.Config) *rest.Config {
	return ApplyInsecureTLSConfig(config)
}
