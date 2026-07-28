package helm

import "testing"

// apiDialAddress must produce a dialable host:port for every endpoint shape.
// The regression it guards: GKE/EKS endpoints are a bare host with no port, so
// a naive scheme-strip left the dialer with "missing port in address" and the
// API-port wait always timed out even though ArgoCD was already healthy.
func TestAPIDialAddress(t *testing.T) {
	cases := []struct {
		name string
		host string
		want string
	}{
		{"gke bare host defaults to 443", "https://34.9.1.2", "34.9.1.2:443"},
		{"eks dns host defaults to 443", "https://ABCD.gr7.us-east-1.eks.amazonaws.com", "ABCD.gr7.us-east-1.eks.amazonaws.com:443"},
		{"k3d endpoint keeps its explicit port", "https://127.0.0.1:6550", "127.0.0.1:6550"},
		{"http scheme is stripped too", "http://10.0.0.1", "10.0.0.1:443"},
		{"host already host:port is untouched", "10.0.0.1:6443", "10.0.0.1:6443"},
		{"empty host yields empty", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := apiDialAddress(tc.host); got != tc.want {
				t.Fatalf("apiDialAddress(%q) = %q, want %q", tc.host, got, tc.want)
			}
		})
	}
}
