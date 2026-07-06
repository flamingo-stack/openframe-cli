package download

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeTarGz builds a gzip-compressed tar with a single regular file.
func makeTarGz(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	require.NoError(t, tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(content)), Typeflag: tar.TypeReg}))
	_, err := tw.Write(content)
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())
	return buf.Bytes()
}

func TestInstallVerifiedTarGz_ExtractsMember(t *testing.T) {
	want := []byte("#!/bin/sh\necho helm\n")
	archive := makeTarGz(t, "linux-amd64/helm", want)
	srv := serve(t, archive)

	dest := filepath.Join(t.TempDir(), "helm")
	d := Downloader{}
	require.NoError(t, d.InstallVerifiedTarGz(context.Background(),
		PinnedAsset{URL: srv.URL, SHA256: sha256hex(archive)}, "linux-amd64/helm", dest, 0o755))

	got, err := os.ReadFile(dest)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestExtractTarGzMember_NotFound(t *testing.T) {
	archive := makeTarGz(t, "linux-amd64/helm", []byte("x"))
	_, err := extractTarGzMember(archive, "linux-amd64/nope")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestInstallVerifiedTarGz_RejectsChecksumMismatch(t *testing.T) {
	archive := makeTarGz(t, "linux-amd64/helm", []byte("x"))
	srv := serve(t, archive)
	dest := filepath.Join(t.TempDir(), "helm")
	err := Downloader{}.InstallVerifiedTarGz(context.Background(),
		PinnedAsset{URL: srv.URL, SHA256: "00"}, "linux-amd64/helm", dest, 0o755)
	require.Error(t, err)
	_, statErr := os.Stat(dest)
	assert.True(t, os.IsNotExist(statErr), "nothing must be written on checksum mismatch")
}

func sha256hex(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

func serve(t *testing.T, body []byte) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestVerifyChecksum(t *testing.T) {
	data := []byte("hello tool binary")
	assert.NoError(t, VerifyChecksum(data, sha256hex(data)))
	assert.Error(t, VerifyChecksum(data, "deadbeef"), "mismatch must error")
	assert.Error(t, VerifyChecksum(data, ""), "empty expected checksum must error")
}

func TestFetchVerified_RejectsChecksumMismatch(t *testing.T) {
	body := []byte("the real binary")
	srv := serve(t, body)
	d := Downloader{Client: srv.Client()}

	// Wrong digest must be rejected.
	_, err := d.FetchVerified(context.Background(), PinnedAsset{URL: srv.URL, SHA256: sha256hex([]byte("something else"))})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "checksum mismatch")

	// Correct digest returns the bytes.
	got, err := d.FetchVerified(context.Background(), PinnedAsset{URL: srv.URL, SHA256: sha256hex(body)})
	require.NoError(t, err)
	assert.Equal(t, body, got)
}

func TestInstallVerified_WritesExecutableOnlyWhenValid(t *testing.T) {
	body := []byte("#!/bin/sh\necho tool\n")
	srv := serve(t, body)
	d := Downloader{Client: srv.Client()}
	dest := filepath.Join(t.TempDir(), "tool")

	require.NoError(t, d.InstallVerified(context.Background(),
		PinnedAsset{URL: srv.URL, SHA256: sha256hex(body)}, dest, 0o755))

	info, err := os.Stat(dest)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o755), info.Mode().Perm())
}

func TestInstallVerified_LeavesNothingOnMismatch(t *testing.T) {
	body := []byte("tampered")
	srv := serve(t, body)
	d := Downloader{Client: srv.Client()}
	dir := t.TempDir()
	dest := filepath.Join(dir, "tool")

	err := d.InstallVerified(context.Background(),
		PinnedAsset{URL: srv.URL, SHA256: sha256hex([]byte("expected different"))}, dest, 0o755)
	require.Error(t, err)

	_, statErr := os.Stat(dest)
	assert.True(t, os.IsNotExist(statErr), "destination must not exist after a failed install")

	// No leftover temp files either.
	entries, _ := os.ReadDir(dir)
	assert.Empty(t, entries, "no partial/temp files should remain")
}

func TestPinnedTool_Asset(t *testing.T) {
	tool := PinnedTool{
		Name:    "telepresence",
		Version: "2.22.4",
		Assets: map[string]PinnedAsset{
			"linux/amd64":  {URL: "https://example/telepresence-linux-amd64", SHA256: "abc"},
			"darwin/arm64": {URL: "https://example/telepresence-darwin-arm64", SHA256: "def"},
		},
	}
	a, ok := tool.Asset("linux", "amd64")
	require.True(t, ok)
	assert.Equal(t, "abc", a.SHA256)

	_, ok = tool.Asset("windows", "arm64")
	assert.False(t, ok, "unsupported platform must report missing")
}
