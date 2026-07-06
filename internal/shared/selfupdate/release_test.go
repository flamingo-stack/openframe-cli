package selfupdate

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseChecksum(t *testing.T) {
	listing := "abc123  openframe-cli_linux_amd64.tar.gz\n" +
		"def456 *openframe-cli_darwin_arm64.tar.gz\n" +
		"# a comment line that must be ignored\n"

	got, err := parseChecksum(listing, "openframe-cli_linux_amd64.tar.gz")
	if err != nil || got != "abc123" {
		t.Fatalf("linux: got (%q, %v), want (abc123, nil)", got, err)
	}
	// The '*' binary-mode prefix must be tolerated.
	got, err = parseChecksum(listing, "openframe-cli_darwin_arm64.tar.gz")
	if err != nil || got != "def456" {
		t.Fatalf("darwin: got (%q, %v), want (def456, nil)", got, err)
	}
	if _, err := parseChecksum(listing, "missing.tar.gz"); err == nil {
		t.Fatal("expected an error for a filename not in the listing")
	}
}

func TestClientLatestAndForTag(t *testing.T) {
	const body = `{"tag_name":"v2.0.0","html_url":"https://example/rel","assets":[` +
		`{"name":"checksums.txt","browser_download_url":"https://example/checksums.txt"}]}`
	var latestHit, tagHit bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/flamingo-stack/openframe-cli/releases/latest":
			latestHit = true
		case "/repos/flamingo-stack/openframe-cli/releases/tags/v2.0.0":
			tagHit = true
		default:
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	c := Client{APIBase: srv.URL}
	rel, err := c.Latest(context.Background())
	if err != nil || rel.TagName != "v2.0.0" || rel.HTMLURL != "https://example/rel" {
		t.Fatalf("Latest = (%+v, %v)", rel, err)
	}
	if _, err := c.ForTag(context.Background(), "v2.0.0"); err != nil {
		t.Fatalf("ForTag error: %v", err)
	}
	if !latestHit || !tagHit {
		t.Fatalf("endpoints not both hit: latest=%v tag=%v", latestHit, tagHit)
	}
}

func TestClientNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.NotFound(w, req)
	}))
	defer srv.Close()
	if _, err := (Client{APIBase: srv.URL}).ForTag(context.Background(), "v9.9.9"); err == nil {
		t.Fatal("expected an error for a 404 release")
	}
}
