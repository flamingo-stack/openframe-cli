// Package download provides verified binary downloads for prerequisite tools.
//
// It replaces the "curl … | bash" / unverified "curl -o /usr/local/bin/tool"
// installs flagged in the audit (I5/M1): every asset is pinned to a version and
// a SHA256 digest, downloaded to a temp file, checksum-verified, and only then
// moved into place. A mismatch aborts the install and leaves nothing behind.
package download

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// PinnedAsset is a single platform's download, pinned to a content digest.
type PinnedAsset struct {
	URL    string
	SHA256 string // lowercase hex
}

// PinnedTool describes a tool pinned to a version, with one asset per platform
// keyed by "<GOOS>/<GOARCH>" (e.g. "linux/amd64").
type PinnedTool struct {
	Name    string
	Version string
	Assets  map[string]PinnedAsset
}

// Asset returns the pinned asset for the given platform.
func (t PinnedTool) Asset(goos, goarch string) (PinnedAsset, bool) {
	a, ok := t.Assets[goos+"/"+goarch]
	return a, ok
}

// VerifyChecksum returns an error unless sha256(data) equals wantHex.
func VerifyChecksum(data []byte, wantHex string) error {
	if wantHex == "" {
		return fmt.Errorf("no expected checksum provided")
	}
	sum := sha256.Sum256(data)
	got := hex.EncodeToString(sum[:])
	if !strings.EqualFold(got, wantHex) {
		return fmt.Errorf("checksum mismatch: got %s, want %s", got, wantHex)
	}
	return nil
}

// Downloader fetches and verifies pinned assets. The zero value is usable and
// uses http.DefaultClient; tests can inject a client pointed at httptest.
type Downloader struct {
	Client *http.Client
}

func (d Downloader) client() *http.Client {
	if d.Client != nil {
		return d.Client
	}
	return http.DefaultClient
}

// FetchVerified downloads asset.URL, verifies its SHA256, and returns the bytes.
// It never returns unverified data.
func (d Downloader) FetchVerified(ctx context.Context, asset PinnedAsset) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.URL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := d.client().Do(req)
	if err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading download body: %w", err)
	}
	if err := VerifyChecksum(body, asset.SHA256); err != nil {
		return nil, err
	}
	return body, nil
}

// InstallVerified downloads and verifies the asset, then writes it to destPath
// with mode perm via a temp file + atomic rename. On any failure the
// destination is left untouched and no partial file remains.
func (d Downloader) InstallVerified(ctx context.Context, asset PinnedAsset, destPath string, perm os.FileMode) error {
	body, err := d.FetchVerified(ctx, asset)
	if err != nil {
		return err
	}
	dir := filepath.Dir(destPath)
	tmp, err := os.CreateTemp(dir, ".of-download-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpName) }
	if _, err := tmp.Write(body); err != nil {
		_ = tmp.Close()
		cleanup()
		return err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return err
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		cleanup()
		return err
	}
	if err := os.Rename(tmpName, destPath); err != nil {
		cleanup()
		return fmt.Errorf("installing to %s: %w", destPath, err)
	}
	return nil
}
