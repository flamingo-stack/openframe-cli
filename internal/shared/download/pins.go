package download

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Pinned tool definitions replace the unverified "curl | bash" / "curl -o
// /tmp/tool && sudo mv" installs flagged by the audit (I5/M1). Versions and
// checksums are pinned deliberately: bumping a tool means updating both the
// Version and every asset's SHA256 (taken from the upstream release's
// published checksums). That is the security tradeoff — a fixed, verified
// version instead of an unverified "latest".

// K3d is the pinned k3d CLI. Upstream: https://github.com/k3d-io/k3d/releases
// Checksums: the release's checksums.txt.
var K3d = PinnedTool{
	Name:    "k3d",
	Version: "v5.9.0",
	Assets: map[string]PinnedAsset{
		"linux/amd64":  {URL: "https://github.com/k3d-io/k3d/releases/download/v5.9.0/k3d-linux-amd64", SHA256: "06d8f25bc3a971c4eb29e0ff08429b180402db0f4dec838c9eac427e296800a0"},
		"linux/arm64":  {URL: "https://github.com/k3d-io/k3d/releases/download/v5.9.0/k3d-linux-arm64", SHA256: "03cde5cf23e6e8e67de5a039ecf26e5b85aca82fba3e5d13dadf904cd218a250"},
		"darwin/amd64": {URL: "https://github.com/k3d-io/k3d/releases/download/v5.9.0/k3d-darwin-amd64", SHA256: "b4aabc37534f95b9c764e7823f2df923f50d57600837aa60a06266cce47db732"},
		"darwin/arm64": {URL: "https://github.com/k3d-io/k3d/releases/download/v5.9.0/k3d-darwin-arm64", SHA256: "fe106541d5d0a3f18debcd4d432a16f8c0ce3e6ddc06f8fbb6f696a122313e00"},
	},
}

// UserBinDir returns the CLI-managed bin directory (~/.openframe/bin) where
// verified tool binaries are installed. It does not create the directory.
func UserBinDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, ".openframe", "bin"), nil
}

// InstallPinnedTool installs the pinned tool for the current platform into
// binDir (created if missing), verifying its checksum, and returns the
// installed binary path. It returns an error if no asset is pinned for the
// current GOOS/GOARCH.
func (d Downloader) InstallPinnedTool(ctx context.Context, tool PinnedTool, binDir string) (string, error) {
	asset, ok := tool.Asset(runtime.GOOS, runtime.GOARCH)
	if !ok {
		return "", fmt.Errorf("no verified %s %s asset for %s/%s", tool.Name, tool.Version, runtime.GOOS, runtime.GOARCH)
	}
	if err := os.MkdirAll(binDir, 0o750); err != nil {
		return "", fmt.Errorf("creating %s: %w", binDir, err)
	}
	dest := filepath.Join(binDir, tool.Name)
	if err := d.InstallVerified(ctx, asset, dest, 0o750); err != nil {
		return "", err
	}
	return dest, nil
}

// PrependToPath puts dir at the front of the current process PATH when it is
// not already present, so tools installed there are found by later exec calls
// in this process. It only affects this process's environment, never the
// user's shell configuration.
func PrependToPath(dir string) {
	path := os.Getenv("PATH")
	for _, p := range filepath.SplitList(path) {
		if p == dir {
			return
		}
	}
	if path == "" {
		_ = os.Setenv("PATH", dir)
		return
	}
	_ = os.Setenv("PATH", dir+string(os.PathListSeparator)+path)
}
