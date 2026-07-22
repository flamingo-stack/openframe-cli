package download

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func platformKey() string { return runtime.GOOS + "/" + runtime.GOARCH }

func hexSum(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

func TestInstallPinnedTool_Success(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix file modes (0750) aren't honoured on Windows")
	}
	payload := []byte("#!/bin/sh\necho hi\n")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(payload)
	}))
	defer srv.Close()

	tool := PinnedTool{
		Name:    "faketool",
		Version: "v1.2.3",
		Assets:  map[string]PinnedAsset{platformKey(): {URL: srv.URL, SHA256: hexSum(payload)}},
	}
	binDir := t.TempDir()

	d := Downloader{Client: srv.Client()}
	path, err := d.InstallPinnedTool(context.Background(), tool, binDir)
	if err != nil {
		t.Fatalf("InstallPinnedTool: %v", err)
	}
	if want := filepath.Join(binDir, "faketool"); path != want {
		t.Fatalf("path = %q, want %q", path, want)
	}

	got, err := os.ReadFile(path) // #nosec G304 -- test reads a path it just created under t.TempDir()
	if err != nil {
		t.Fatalf("reading installed binary: %v", err)
	}
	if string(got) != string(payload) {
		t.Fatalf("installed content = %q, want %q", got, payload)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0o750 {
		t.Fatalf("mode = %v, want 0750", info.Mode().Perm())
	}
}

func TestInstallPinnedTool_ChecksumMismatch(t *testing.T) {
	payload := []byte("real bytes")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(payload)
	}))
	defer srv.Close()

	tool := PinnedTool{
		Name:    "faketool",
		Version: "v1.2.3",
		Assets:  map[string]PinnedAsset{platformKey(): {URL: srv.URL, SHA256: hexSum([]byte("different bytes"))}},
	}
	binDir := t.TempDir()

	d := Downloader{Client: srv.Client()}
	if _, err := d.InstallPinnedTool(context.Background(), tool, binDir); err == nil {
		t.Fatal("expected checksum-mismatch error, got nil")
	}

	// Nothing must be left behind: no destination binary and no temp file.
	entries, err := os.ReadDir(binDir)
	if err != nil {
		t.Fatalf("reading bin dir: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("bin dir not empty after failed install: %v", entries)
	}
}

func TestInstallPinnedTool_UnsupportedPlatform(t *testing.T) {
	tool := PinnedTool{
		Name:    "faketool",
		Version: "v1.2.3",
		Assets:  map[string]PinnedAsset{"plan9/riscv": {URL: "http://example.invalid", SHA256: "deadbeef"}},
	}
	_, err := (Downloader{}).InstallPinnedTool(context.Background(), tool, t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "no verified") {
		t.Fatalf("expected 'no verified ... asset' error, got %v", err)
	}
}

func TestPrependToPath(t *testing.T) {
	t.Setenv("PATH", "/usr/bin"+string(os.PathListSeparator)+"/bin")
	dir := filepath.Join(t.TempDir(), "of-bin")

	PrependToPath(dir)
	got := os.Getenv("PATH")
	if !strings.HasPrefix(got, dir+string(os.PathListSeparator)) {
		t.Fatalf("PATH = %q, want it to start with %q", got, dir)
	}

	// Idempotent: a second call must not add a duplicate entry.
	PrependToPath(dir)
	if n := strings.Count(os.Getenv("PATH"), dir); n != 1 {
		t.Fatalf("dir appears %d times in PATH, want 1", n)
	}
}

func TestUserBinDir(t *testing.T) {
	dir, err := UserBinDir()
	if err != nil {
		t.Fatalf("UserBinDir: %v", err)
	}
	home, _ := os.UserHomeDir()
	if want := filepath.Join(home, ".openframe", "bin"); dir != want {
		t.Fatalf("UserBinDir = %q, want %q", dir, want)
	}
}

// TestPinnedAssets_RealDownload verifies that every pinned tool actually
// downloads and passes its SHA256 check for the current platform. This is the
// proof that the pinned versions/checksums are correct and the verified
// download works end to end. It hits the network, so it is skipped under -short.
func TestPinnedAssets_RealDownload(t *testing.T) {
	if testing.Short() {
		t.Skip("network test skipped under -short")
	}
	tools := []PinnedTool{Terraform, Infracost}
	if runtime.GOOS != "windows" {
		// k3d/mkcert/helm have no windows pins: on Windows they run inside WSL.
		// Terraform and infracost are pinned for all supported platforms.
		tools = append(tools, K3d, Mkcert, Helm)
	}
	for _, tool := range tools {
		asset, ok := tool.Asset(runtime.GOOS, runtime.GOARCH)
		if !ok {
			t.Errorf("%s: no asset for %s/%s", tool.Name, runtime.GOOS, runtime.GOARCH)
			continue
		}
		if _, err := (Downloader{}).FetchVerified(context.Background(), asset); err != nil {
			t.Errorf("%s %s (%s): %v", tool.Name, tool.Version, asset.URL, err)
		}
	}
}

// TestHelm_Pins locks the helm pin shape: a versioned .tar.gz + non-empty SHA256
// for each supported linux/darwin platform, marked as a tarball.
func TestHelm_Pins(t *testing.T) {
	if !strings.HasPrefix(Helm.Version, "v") {
		t.Fatalf("Helm.Version must be a v-prefixed tag, got %q", Helm.Version)
	}
	if !Helm.Tarball {
		t.Error("Helm assets are .tar.gz — Tarball must be true")
	}
	for _, p := range []struct{ os, arch string }{
		{"linux", "amd64"}, {"linux", "arm64"}, {"darwin", "amd64"}, {"darwin", "arm64"},
	} {
		asset, ok := Helm.Asset(p.os, p.arch)
		if !ok {
			t.Errorf("no helm asset for %s/%s", p.os, p.arch)
			continue
		}
		if len(asset.SHA256) != 64 {
			t.Errorf("%s/%s: SHA256 must be 64 hex chars, got %q", p.os, p.arch, asset.SHA256)
		}
		if !strings.Contains(asset.URL, Helm.Version) || !strings.HasSuffix(asset.URL, p.os+"-"+p.arch+".tar.gz") {
			t.Errorf("%s/%s: URL %q must contain version and end with platform.tar.gz", p.os, p.arch, asset.URL)
		}
	}
}

// TestTerraform_Pins locks the terraform pin shape: a versioned .zip +
// non-empty SHA256 for every supported platform — including Windows, where
// terraform (unlike k3d/helm) runs natively rather than inside WSL.
func TestTerraform_Pins(t *testing.T) {
	if Terraform.Version == "" {
		t.Fatal("Terraform.Version must be set")
	}
	if !Terraform.Zip {
		t.Error("Terraform assets are .zip — Zip must be true")
	}
	for _, p := range []struct{ os, arch string }{
		{"linux", "amd64"}, {"linux", "arm64"},
		{"darwin", "amd64"}, {"darwin", "arm64"},
		{"windows", "amd64"}, {"windows", "arm64"},
	} {
		asset, ok := Terraform.Asset(p.os, p.arch)
		if !ok {
			t.Errorf("no terraform asset for %s/%s", p.os, p.arch)
			continue
		}
		if len(asset.SHA256) != 64 {
			t.Errorf("%s/%s: SHA256 must be 64 hex chars, got %q", p.os, p.arch, asset.SHA256)
		}
		if !strings.Contains(asset.URL, Terraform.Version) || !strings.HasSuffix(asset.URL, p.os+"_"+p.arch+".zip") {
			t.Errorf("%s/%s: URL %q must contain version and end with platform.zip", p.os, p.arch, asset.URL)
		}
	}
}

// TestInstallPinnedTool_ZipTool covers the .zip install path (terraform's
// asset format): the binary is extracted from the archive root by name.
func TestInstallPinnedTool_ZipTool(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix file modes (0750) aren't honoured on Windows")
	}
	payload := []byte("#!/bin/sh\necho tf\n")
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("faketool")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(payload); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	archive := buf.Bytes()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(archive)
	}))
	defer srv.Close()

	tool := PinnedTool{
		Name:    "faketool",
		Version: "v1.2.3",
		Zip:     true,
		Assets:  map[string]PinnedAsset{platformKey(): {URL: srv.URL, SHA256: hexSum(archive)}},
	}
	binDir := t.TempDir()

	path, err := (Downloader{Client: srv.Client()}).InstallPinnedTool(context.Background(), tool, binDir)
	if err != nil {
		t.Fatalf("InstallPinnedTool(zip): %v", err)
	}
	got, err := os.ReadFile(path) // #nosec G304 -- test reads a path it just created under t.TempDir()
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(payload) {
		t.Fatalf("installed content = %q, want %q", got, payload)
	}
}

// TestMkcert_Pins locks the mkcert pin shape: a versioned URL + non-empty SHA256
// for each supported linux/darwin platform.
func TestMkcert_Pins(t *testing.T) {
	if Mkcert.Version == "" {
		t.Fatal("Mkcert.Version must be set")
	}
	for _, p := range []struct{ os, arch string }{
		{"linux", "amd64"}, {"linux", "arm64"}, {"darwin", "amd64"}, {"darwin", "arm64"},
	} {
		asset, ok := Mkcert.Asset(p.os, p.arch)
		if !ok {
			t.Errorf("no mkcert asset for %s/%s", p.os, p.arch)
			continue
		}
		if len(asset.SHA256) != 64 {
			t.Errorf("%s/%s: SHA256 must be 64 hex chars, got %q", p.os, p.arch, asset.SHA256)
		}
		if !strings.Contains(asset.URL, Mkcert.Version) || !strings.Contains(asset.URL, p.os+"-"+p.arch) {
			t.Errorf("%s/%s: URL %q must contain version and platform", p.os, p.arch, asset.URL)
		}
	}
}

// TestInstallPinnedTool_RealK3dExec installs the real pinned k3d into a temp
// dir and runs it, proving the whole verified-install path works on the current
// platform. Network + exec, so skipped under -short and on Windows (WSL path).
func TestInstallPinnedTool_RealK3dExec(t *testing.T) {
	if testing.Short() {
		t.Skip("network test skipped under -short")
	}
	if runtime.GOOS == "windows" {
		t.Skip("Windows installs k3d via WSL, not the verified path")
	}
	binDir := t.TempDir()
	path, err := (Downloader{}).InstallPinnedTool(context.Background(), K3d, binDir)
	if err != nil {
		t.Fatalf("InstallPinnedTool(k3d): %v", err)
	}
	out, err := exec.Command(path, "version").CombinedOutput() // #nosec G204 -- path is the binary this test just installed under t.TempDir()
	if err != nil {
		t.Fatalf("running installed k3d: %v (%s)", err, out)
	}
	if !strings.Contains(string(out), strings.TrimPrefix(K3d.Version, "v")) {
		t.Fatalf("k3d version output %q does not contain %q", out, K3d.Version)
	}
}

// TestInfracost_Pins locks the infracost pin shape: a versioned .tar.gz +
// non-empty SHA256 for every supported platform (windows/arm64 has no
// upstream asset).
func TestInfracost_Pins(t *testing.T) {
	if !strings.HasPrefix(Infracost.Version, "v") {
		t.Fatalf("Infracost.Version must be a v-prefixed tag, got %q", Infracost.Version)
	}
	if !Infracost.Tarball {
		t.Errorf("Infracost.Tarball must be true (assets are .tar.gz)")
	}
	for _, p := range []struct{ os, arch string }{
		{"linux", "amd64"}, {"linux", "arm64"},
		{"darwin", "amd64"}, {"darwin", "arm64"},
		{"windows", "amd64"},
	} {
		asset, ok := Infracost.Asset(p.os, p.arch)
		if !ok {
			t.Errorf("no infracost asset for %s/%s", p.os, p.arch)
			continue
		}
		if len(asset.SHA256) != 64 {
			t.Errorf("%s/%s: SHA256 must be 64 hex chars, got %q", p.os, p.arch, asset.SHA256)
		}
		if !strings.Contains(asset.URL, Infracost.Version) || !strings.HasSuffix(asset.URL, p.os+"-"+p.arch+".tar.gz") {
			t.Errorf("%s/%s: URL %q must contain version and end with platform.tar.gz", p.os, p.arch, asset.URL)
		}
	}
}
