package selfupdate

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
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

// TestForTag_TogglesVPrefixOnMiss is the T0-3 regression guard: releases in
// this repo are tagged with the bare semver ("0.4.7"), but the help text and
// user habit say "v0.4.7". ForTag must find the release either way.
func TestForTag_TogglesVPrefixOnMiss(t *testing.T) {
	const body = `{"tag_name":"0.4.7","html_url":"https://example/rel"}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/flamingo-stack/openframe-cli/releases/tags/0.4.7" {
			_, _ = w.Write([]byte(body))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()
	c := Client{APIBase: srv.URL}

	// User typed the v-prefixed form; the bare tag must be found via fallback.
	rel, err := c.ForTag(context.Background(), "v0.4.7")
	if err != nil || rel.TagName != "0.4.7" {
		t.Fatalf("ForTag(v0.4.7) = (%+v, %v), want the bare-tagged release", rel, err)
	}
	// The exact form still works without a fallback.
	if rel, err = c.ForTag(context.Background(), "0.4.7"); err != nil || rel.TagName != "0.4.7" {
		t.Fatalf("ForTag(0.4.7) = (%+v, %v)", rel, err)
	}
}

// TestForTag_BareMissFindsVPrefixed covers the inverse convention (tags with
// "v"), so the fallback is symmetric and survives a future tag-format change.
func TestForTag_BareMissFindsVPrefixed(t *testing.T) {
	const body = `{"tag_name":"v1.4.0","html_url":"https://example/rel"}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/flamingo-stack/openframe-cli/releases/tags/v1.4.0" {
			_, _ = w.Write([]byte(body))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	rel, err := (Client{APIBase: srv.URL}).ForTag(context.Background(), "1.4.0")
	if err != nil || rel.TagName != "v1.4.0" {
		t.Fatalf("ForTag(1.4.0) = (%+v, %v), want the v-tagged release", rel, err)
	}
}

// TestForTag_BothMissesError: neither spelling exists -> a clear error naming
// both tried tags, and no infinite toggling.
func TestForTag_BothMissesError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(http.NotFound))
	defer srv.Close()

	_, err := (Client{APIBase: srv.URL}).ForTag(context.Background(), "9.9.9")
	if err == nil {
		t.Fatal("expected an error when both tag spellings 404")
	}
	for _, want := range []string{"9.9.9", "v9.9.9"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q should mention tried tag %q", err, want)
		}
	}
}

func TestAlternateTag(t *testing.T) {
	if got := alternateTag("0.4.7"); got != "v0.4.7" {
		t.Errorf("alternateTag(0.4.7) = %q", got)
	}
	if got := alternateTag("v0.4.7"); got != "0.4.7" {
		t.Errorf("alternateTag(v0.4.7) = %q", got)
	}
}
