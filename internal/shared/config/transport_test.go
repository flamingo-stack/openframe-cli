package config

import (
	"testing"

	"k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestIsLocalAPIServer(t *testing.T) {
	local := []string{
		"https://0.0.0.0:63625",   // k3d
		"https://127.0.0.1:26443", // orbstack / kind
		"https://localhost:6443",  // docker-desktop style
		"https://host.docker.internal:6550",
		"https://[::1]:6443",
		"127.0.0.1:6443", // bare host:port (no scheme)
	}
	for _, s := range local {
		if !isLocalAPIServer(s) {
			t.Errorf("isLocalAPIServer(%q) = false, want true (local)", s)
		}
	}

	remote := []string{
		"https://api.prod.example.com:6443",
		"https://10.0.5.20:6443",    // private but routable → treat as remote
		"https://192.168.1.50:6443", // LAN → remote
		"https://34.120.0.1:443",    // public IP
		"https://eks.amazonaws.com",
		// B5 regression guard: mDNS/legacy-AD `.local` domains are used by REAL
		// corporate clusters — a suffix match here disabled TLS verification for
		// them with no warning. `.local` must be treated as remote.
		"https://my-cluster.local:6443",
		"https://k8s.corp.local:6443",
	}
	for _, s := range remote {
		if isLocalAPIServer(s) {
			t.Errorf("isLocalAPIServer(%q) = true, want false (remote)", s)
		}
	}
}

func TestApplyInsecureTLSConfig_BypassesLocalOnly(t *testing.T) {
	// Local (k3d) → bypass applied.
	loc := ApplyInsecureTLSConfig(&rest.Config{Host: "https://0.0.0.0:63625", TLSClientConfig: rest.TLSClientConfig{CAData: []byte("ca")}})
	if !loc.Insecure {
		t.Error("local cluster: expected Insecure=true")
	}
	if loc.CAData != nil {
		t.Error("local cluster: expected CA cleared")
	}

	// Remote → TLS untouched (verification preserved).
	rem := ApplyInsecureTLSConfig(&rest.Config{Host: "https://api.prod.example.com:6443", TLSClientConfig: rest.TLSClientConfig{CAData: []byte("ca")}})
	if rem.Insecure {
		t.Error("remote cluster: TLS must NOT be bypassed")
	}
	if string(rem.CAData) != "ca" {
		t.Error("remote cluster: CA must be preserved")
	}
}

// TestApplyInsecureTLSConfig_CloudExecConfigsUntouched locks the cloud-cluster
// guarantee: EKS/GKE rest.Configs (public endpoint + CA + exec-plugin auth)
// pass through the chart subsystem's "defense-in-depth" insecure wraps without
// any TLS downgrade — the local-only guard must keep covering them.
func TestApplyInsecureTLSConfig_CloudExecConfigsUntouched(t *testing.T) {
	hosts := []string{
		"https://ABCDEF123.gr7.us-east-1.eks.amazonaws.com", // EKS
		"https://34.10.20.30",                               // GKE (bare public IP)
	}
	for _, host := range hosts {
		cfg := ApplyInsecureTLSConfig(&rest.Config{
			Host:            host,
			TLSClientConfig: rest.TLSClientConfig{CAData: []byte("ca")},
			ExecProvider:    &clientcmdapi.ExecConfig{Command: "aws"},
		})
		if cfg.Insecure {
			t.Errorf("%s: TLS must NOT be bypassed for a cloud endpoint", host)
		}
		if string(cfg.CAData) != "ca" {
			t.Errorf("%s: CA must be preserved for a cloud endpoint", host)
		}
		if cfg.ExecProvider == nil {
			t.Errorf("%s: exec auth must be preserved", host)
		}
	}
}

func TestApplyInsecureTLSConfig_NilSafe(t *testing.T) {
	if ApplyInsecureTLSConfig(nil) != nil {
		t.Error("nil config must return nil")
	}
}
