package selfupdate

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// releaseFixture serves a fake GitHub release (tag 9.9.9) with a linux/amd64
// archive and checksums, but NO signature bundle — like a hypothetical
// tampered/unsigned release.
func releaseFixture(t *testing.T, binary []byte, withBundle bool) *httptest.Server {
	t.Helper()
	tgz := makeTarGz(t, binaryName, binary, 0o755)
	sum := sha256.Sum256(tgz)
	archive := archiveName("linux", "amd64")

	mux := http.NewServeMux()
	mux.HandleFunc("/archive.tgz", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(tgz) })
	mux.HandleFunc("/checksums.txt", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, "%s  %s\n", hex.EncodeToString(sum[:]), archive)
	})
	var srv *httptest.Server
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/flamingo-stack/openframe-cli/releases/tags/9.9.9" {
			assets := fmt.Sprintf(`{"name":%q,"browser_download_url":%q},{"name":"checksums.txt","browser_download_url":%q}`,
				archive, srv.URL+"/archive.tgz", srv.URL+"/checksums.txt")
			if withBundle {
				assets += fmt.Sprintf(`,{"name":"checksums.txt.bundle","browser_download_url":%q}`, srv.URL+"/bundle")
			}
			fmt.Fprintf(w, `{"tag_name":"9.9.9","assets":[%s]}`, assets)
			return
		}
		mux.ServeHTTP(w, r)
	})
	srv = httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

// TestFetchVerifiedLinuxBinary_InsecureOptOut exercises the WSL-install fetch
// chain offline: with the (strictly parsed) insecure opt-in the unsigned
// release is accepted on checksum alone, and the extracted binary matches.
// It also covers ForTag's v-prefix fallback (version passed as "v9.9.9",
// release tagged "9.9.9").
func TestFetchVerifiedLinuxBinary_InsecureOptOut(t *testing.T) {
	t.Setenv(insecureSkipEnv, "1")
	want := []byte("#!/bin/sh\necho linux binary\n")
	srv := releaseFixture(t, want, false)

	// Route the package-level client at the fixture via the Client in the
	// helper — FetchVerifiedLinuxBinary builds its own Client, so inject the
	// API base through the environment-independent seam: none exists, so call
	// the internals the same way it does.
	client := Client{APIBase: srv.URL}
	rel, err := client.ForTag(context.Background(), "v9.9.9")
	if err != nil {
		t.Fatalf("ForTag: %v", err)
	}
	got, err := fetchVerifiedLinuxBinary(context.Background(), client, rel, "amd64", nil)
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("binary mismatch: got %q", got)
	}
}

// TestFetchVerifiedLinuxBinary_UnsignedRefused: without the insecure opt-in an
// unsigned release must be REFUSED — and the opt-in is strictly parsed, so
// "0" keeps verification ON (the old any-non-empty check disabled it).
func TestFetchVerifiedLinuxBinary_UnsignedRefused(t *testing.T) {
	t.Setenv(insecureSkipEnv, "0") // strict parsing: NOT an opt-in
	srv := releaseFixture(t, []byte("#!/bin/sh\n"), false)

	client := Client{APIBase: srv.URL}
	rel, err := client.ForTag(context.Background(), "9.9.9")
	if err != nil {
		t.Fatalf("ForTag: %v", err)
	}
	_, err = fetchVerifiedLinuxBinary(context.Background(), client, rel, "amd64", nil)
	if err == nil {
		t.Fatal("an unsigned release must be refused when the insecure opt-in is not set")
	}
	if !strings.Contains(err.Error(), "not signed") {
		t.Errorf("error should say the release is not signed, got: %v", err)
	}
}
