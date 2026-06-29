package download

import (
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
