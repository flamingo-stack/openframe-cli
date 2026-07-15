// Package download provides verified binary downloads for prerequisite tools.
//
// It replaces the "curl … | bash" / unverified "curl -o /usr/local/bin/tool"
// installs flagged in the audit (I5/M1): every asset is pinned to a version and
// a SHA256 digest, downloaded to a temp file, checksum-verified, and only then
// moved into place. A mismatch aborts the install and leaves nothing behind.
package download

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// maxAssetBytes bounds FetchVerified's in-memory read so a misbehaving or
// oversized response cannot exhaust memory before the checksum runs.
const maxAssetBytes = 512 << 20 // 512 MiB

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
	// Tarball marks the assets as .tar.gz archives (e.g. helm). The binary is
	// extracted from the archive member "<GOOS>-<GOARCH>/<Name>" (the layout
	// helm and many Go tools ship). Bare-binary tools leave this false.
	Tarball bool
	// Zip marks the assets as .zip archives with the bare binary named <Name>
	// at the archive root (the layout HashiCorp releases ship, e.g. terraform).
	Zip bool
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
// uses a client with a generous overall timeout; tests can inject a client
// pointed at httptest.
type Downloader struct {
	Client *http.Client
}

// defaultClient bounds every download: http.DefaultClient has NO timeout, so a
// stalled GitHub connection hung the spinner forever (worse for MaybeAutoUpdate,
// which runs before the user's actual command). Generous because release
// archives are tens of MB on slow links; healthy downloads finish long before.
var defaultClient = &http.Client{Timeout: 5 * time.Minute}

func (d Downloader) client() *http.Client {
	if d.Client != nil {
		return d.Client
	}
	return defaultClient
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
	// Cap the read so a misbehaving/oversized response can't exhaust memory. The
	// pinned assets (tool binaries / archives) are tens of MB; 512 MiB is a
	// generous ceiling. Read one byte past to detect an over-cap body.
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxAssetBytes+1))
	if err != nil {
		return nil, fmt.Errorf("reading download body: %w", err)
	}
	if int64(len(body)) > maxAssetBytes {
		return nil, fmt.Errorf("download exceeds the %d-byte cap", maxAssetBytes)
	}
	if err := VerifyChecksum(body, asset.SHA256); err != nil {
		return nil, err
	}
	return body, nil
}

// InstallVerified downloads and verifies the asset (a bare binary), then writes
// it to destPath with mode perm via a temp file + atomic rename. On any failure
// the destination is left untouched and no partial file remains.
func (d Downloader) InstallVerified(ctx context.Context, asset PinnedAsset, destPath string, perm os.FileMode) error {
	body, err := d.FetchVerified(ctx, asset)
	if err != nil {
		return err
	}
	return writeFileAtomic(body, destPath, perm)
}

// InstallVerifiedTarGz downloads and verifies a .tar.gz asset, extracts the
// regular file named member (a slash path within the archive, e.g.
// "linux-amd64/helm"), and installs it to destPath with mode perm (atomic).
func (d Downloader) InstallVerifiedTarGz(ctx context.Context, asset PinnedAsset, member, destPath string, perm os.FileMode) error {
	body, err := d.FetchVerified(ctx, asset)
	if err != nil {
		return err
	}
	extracted, err := extractTarGzMember(body, member)
	if err != nil {
		return err
	}
	return writeFileAtomic(extracted, destPath, perm)
}

// InstallVerifiedZipMember downloads and verifies a .zip asset, extracts the
// regular file named member (a slash path within the archive, e.g.
// "terraform"), and installs it to destPath with mode perm (atomic).
func (d Downloader) InstallVerifiedZipMember(ctx context.Context, asset PinnedAsset, member, destPath string, perm os.FileMode) error {
	body, err := d.FetchVerified(ctx, asset)
	if err != nil {
		return err
	}
	extracted, err := extractZipMember(body, member)
	if err != nil {
		return err
	}
	return writeFileAtomic(extracted, destPath, perm)
}

// FetchVerifiedTarGzMember downloads and verifies a .tar.gz asset and returns
// the bytes of the regular file named member — for callers that stream the
// binary elsewhere (e.g. into WSL via stdin) instead of installing it locally.
func (d Downloader) FetchVerifiedTarGzMember(ctx context.Context, asset PinnedAsset, member string) ([]byte, error) {
	body, err := d.FetchVerified(ctx, asset)
	if err != nil {
		return nil, err
	}
	return extractTarGzMember(body, member)
}

// extractTarGzMember returns the bytes of the regular file named member inside a
// gzip-compressed tar archive. The member is matched by its cleaned path.
func extractTarGzMember(data []byte, member string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("opening gzip: %w", err)
	}
	defer func() { _ = gz.Close() }()
	tr := tar.NewReader(gz)
	want := path.Clean(member)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("member %q not found in archive", member)
		}
		if err != nil {
			return nil, fmt.Errorf("reading archive: %w", err)
		}
		if hdr.Typeflag != tar.TypeReg || path.Clean(hdr.Name) != want {
			continue
		}
		// Cap extraction to guard against a decompression bomb.
		b, err := io.ReadAll(io.LimitReader(tr, 200<<20))
		if err != nil {
			return nil, fmt.Errorf("extracting %q: %w", member, err)
		}
		return b, nil
	}
}

// extractZipMember returns the bytes of the regular file named member inside a
// zip archive. The member is matched by its cleaned path.
func extractZipMember(data []byte, member string) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("opening zip: %w", err)
	}
	want := path.Clean(member)
	for _, f := range zr.File {
		if f.Mode().IsRegular() && path.Clean(f.Name) == want {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("extracting %q: %w", member, err)
			}
			defer func() { _ = rc.Close() }()
			// Cap extraction to guard against a decompression bomb.
			b, err := io.ReadAll(io.LimitReader(rc, 200<<20))
			if err != nil {
				return nil, fmt.Errorf("extracting %q: %w", member, err)
			}
			return b, nil
		}
	}
	return nil, fmt.Errorf("member %q not found in archive", member)
}

// writeFileAtomic writes body to destPath with mode perm via a temp file in the
// same directory + atomic rename. On any failure nothing partial remains.
func writeFileAtomic(body []byte, destPath string, perm os.FileMode) error {
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
