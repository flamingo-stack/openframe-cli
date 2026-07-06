package selfupdate

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestNormalizeVersion(t *testing.T) {
	cases := map[string]string{
		"v1.2.3":  "v1.2.3",
		"1.2.3":   "v1.2.3", // GoReleaser's .Version has no leading v
		"v1.2.0":  "v1.2.0",
		"dev":     "",
		"":        "",
		"garbage": "",
	}
	for in, want := range cases {
		if got := normalizeVersion(in); got != want {
			t.Errorf("normalizeVersion(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestIsNewer(t *testing.T) {
	cases := []struct {
		current, latest string
		want            bool
	}{
		{"v1.0.0", "v1.0.1", true},
		{"1.0.0", "v1.1.0", true}, // mixed prefixes still compare
		{"v1.2.3", "v1.2.3", false},
		{"v1.2.3", "v1.2.2", false},
		{"dev", "v1.0.0", false},   // dev builds never self-update
		{"v1.0.0", "dev", false},   // unparseable latest → no update
		{"v1.0.0", "garbage", false},
	}
	for _, c := range cases {
		if got := IsNewer(c.current, c.latest); got != c.want {
			t.Errorf("IsNewer(%q, %q) = %v, want %v", c.current, c.latest, got, c.want)
		}
	}
}

func TestArchiveName(t *testing.T) {
	if got := archiveName("linux", "amd64"); got != "openframe-cli_linux_amd64.tar.gz" {
		t.Fatalf("archiveName = %q", got)
	}
}

func TestApplyRefusesNativeWindows(t *testing.T) {
	u := Updater{Current: "v1.0.0", GOOS: "windows"}
	if err := u.Apply(context.Background(), Release{TagName: "v1.1.0"}, nil); err == nil {
		t.Fatal("expected Apply to refuse on native windows")
	}
}

// TestApplyEndToEnd exercises the full happy path: fetch release metadata,
// verify the archive against checksums.txt, extract the binary, atomically
// replace the current executable, and smoke-test it.
func TestApplyEndToEnd(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("smoke test runs a /bin/sh stub binary; unix-only")
	}

	// A stand-in "release binary": a shell script that satisfies `--version`.
	stub := []byte("#!/bin/sh\necho v9.9.9\n")
	tgz := makeTarGz(t, binaryName, stub, 0o755)
	sum := sha256.Sum256(tgz)
	archive := archiveName("linux", "amd64")

	mux := http.NewServeMux()
	mux.HandleFunc("/archive.tgz", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(tgz) })
	mux.HandleFunc("/checksums.txt", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, "%s  %s\n", hex.EncodeToString(sum[:]), archive)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	rel := Release{
		TagName: "v9.9.9",
		Assets: []Asset{
			{Name: checksumsFile, URL: srv.URL + "/checksums.txt"},
			{Name: archive, URL: srv.URL + "/archive.tgz"},
		},
	}

	// The "currently installed" binary we expect to be replaced.
	dir := t.TempDir()
	exe := filepath.Join(dir, "openframe")
	if err := os.WriteFile(exe, []byte("OLD BINARY"), 0o755); err != nil {
		t.Fatal(err)
	}

	u := Updater{Current: "v1.0.0", GOOS: "linux", GOARCH: "amd64", Client: Client{APIBase: srv.URL}, exePath: exe}
	if err := u.Apply(context.Background(), rel, nil); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	got, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, stub) {
		t.Fatalf("binary not replaced: got %q", got)
	}
	// No backup / staging files should be left behind.
	for _, leftover := range []string{exe + ".bak", exe + ".new"} {
		if _, err := os.Stat(leftover); !os.IsNotExist(err) {
			t.Errorf("leftover file not cleaned up: %s", leftover)
		}
	}
}

// TestApplyChecksumMismatch verifies a tampered archive is rejected and the
// current binary is left untouched.
func TestApplyChecksumMismatch(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix-only")
	}
	tgz := makeTarGz(t, binaryName, []byte("#!/bin/sh\necho hi\n"), 0o755)
	archive := archiveName("linux", "amd64")

	mux := http.NewServeMux()
	mux.HandleFunc("/archive.tgz", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(tgz) })
	mux.HandleFunc("/checksums.txt", func(w http.ResponseWriter, _ *http.Request) {
		// Deliberately wrong digest.
		fmt.Fprintf(w, "%s  %s\n", "0000000000000000000000000000000000000000000000000000000000000000", archive)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	rel := Release{TagName: "v9.9.9", Assets: []Asset{
		{Name: checksumsFile, URL: srv.URL + "/checksums.txt"},
		{Name: archive, URL: srv.URL + "/archive.tgz"},
	}}

	dir := t.TempDir()
	exe := filepath.Join(dir, "openframe")
	_ = os.WriteFile(exe, []byte("OLD BINARY"), 0o755)

	u := Updater{Current: "v1.0.0", GOOS: "linux", GOARCH: "amd64", Client: Client{APIBase: srv.URL}, exePath: exe}
	if err := u.Apply(context.Background(), rel, nil); err == nil {
		t.Fatal("expected a checksum-mismatch error")
	}
	if got, _ := os.ReadFile(exe); string(got) != "OLD BINARY" {
		t.Fatalf("current binary was modified on a failed update: %q", got)
	}
}

func makeTarGz(t *testing.T, member string, content []byte, mode int64) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{
		Name: member, Mode: mode, Size: int64(len(content)), Typeflag: tar.TypeReg,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
