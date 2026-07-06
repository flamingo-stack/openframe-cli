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
const (
	k3dVersion = "v5.9.0"
	k3dBaseURL = "https://github.com/k3d-io/k3d/releases/download/" + k3dVersion + "/k3d-"

	k3dSHA256LinuxAMD64  = "06d8f25bc3a971c4eb29e0ff08429b180402db0f4dec838c9eac427e296800a0"
	k3dSHA256LinuxARM64  = "03cde5cf23e6e8e67de5a039ecf26e5b85aca82fba3e5d13dadf904cd218a250"
	k3dSHA256DarwinAMD64 = "b4aabc37534f95b9c764e7823f2df923f50d57600837aa60a06266cce47db732"
	k3dSHA256DarwinARM64 = "fe106541d5d0a3f18debcd4d432a16f8c0ce3e6ddc06f8fbb6f696a122313e00"
)

var K3d = PinnedTool{
	Name:    "k3d",
	Version: k3dVersion,
	Assets: map[string]PinnedAsset{
		"linux/amd64":  {URL: k3dBaseURL + "linux-amd64", SHA256: k3dSHA256LinuxAMD64},
		"linux/arm64":  {URL: k3dBaseURL + "linux-arm64", SHA256: k3dSHA256LinuxARM64},
		"darwin/amd64": {URL: k3dBaseURL + "darwin-amd64", SHA256: k3dSHA256DarwinAMD64},
		"darwin/arm64": {URL: k3dBaseURL + "darwin-arm64", SHA256: k3dSHA256DarwinARM64},
	},
}

// Mkcert is the pinned mkcert CLI. Upstream:
// https://github.com/FiloSottile/mkcert/releases. The release ships bare
// per-platform binaries (no checksums.txt), so these SHA256 are computed from
// the published v1.4.4 assets. Replacing the unverified
// `curl dl.filippo.io/mkcert/latest` install (audit T0.3) — critical because
// mkcert injects a root CA into the system/NSS trust stores.
const (
	mkcertVersion = "v1.4.4"
	mkcertBaseURL = "https://github.com/FiloSottile/mkcert/releases/download/" + mkcertVersion + "/mkcert-" + mkcertVersion + "-"

	mkcertSHA256LinuxAMD64  = "6d31c65b03972c6dc4a14ab429f2928300518b26503f58723e532d1b0a3bbb52"
	mkcertSHA256LinuxARM64  = "b98f2cc69fd9147fe4d405d859c57504571adec0d3611c3eefd04107c7ac00d0"
	mkcertSHA256DarwinAMD64 = "a32dfab51f1845d51e810db8e47dcf0e6b51ae3422426514bf5a2b8302e97d4e"
	mkcertSHA256DarwinARM64 = "c8af0df44bce04359794dad8ea28d750437411d632748049d08644ffb66a60c6"
)

var Mkcert = PinnedTool{
	Name:    "mkcert",
	Version: mkcertVersion,
	Assets: map[string]PinnedAsset{
		"linux/amd64":  {URL: mkcertBaseURL + "linux-amd64", SHA256: mkcertSHA256LinuxAMD64},
		"linux/arm64":  {URL: mkcertBaseURL + "linux-arm64", SHA256: mkcertSHA256LinuxARM64},
		"darwin/amd64": {URL: mkcertBaseURL + "darwin-amd64", SHA256: mkcertSHA256DarwinAMD64},
		"darwin/arm64": {URL: mkcertBaseURL + "darwin-arm64", SHA256: mkcertSHA256DarwinARM64},
	},
}

// Helm is the pinned Helm CLI. Upstream: https://github.com/helm/helm/releases
// (binaries served from get.helm.sh). Pinned to the latest Helm 3.x — the CLI is
// built and tested against Helm 3 (get-helm-3), so this is a deliberate v3 pin,
// not "latest". The assets are .tar.gz (Tarball), so the helm binary is extracted
// from "<os>-<arch>/helm". SHA256 from helm's published *.tar.gz.sha256sum.
// Replaces the unverified `curl get-helm-3 | bash` install (audit T0.3).
const (
	helmVersion = "v3.21.2"
	helmBaseURL = "https://get.helm.sh/helm-" + helmVersion + "-"

	helmSHA256LinuxAMD64  = "0a745198de24545d0055cd8414bc8d2ba10363ef5f5d38369ea1b399671cc083"
	helmSHA256LinuxARM64  = "bbd559fc0547f1d96ccbc68fe4f1cb98f01808f36538139e669369066b781267"
	helmSHA256DarwinAMD64 = "82ac9105e657267cb029b5bf27ed28e35db104777328a036a84d345046f9f329"
	helmSHA256DarwinARM64 = "aea537342b4c03cf58e089cb8dc99468087bb1a0218531df40462faca3f6c5d3"
)

var Helm = PinnedTool{
	Name:    "helm",
	Version: helmVersion,
	Tarball: true,
	Assets: map[string]PinnedAsset{
		"linux/amd64":  {URL: helmBaseURL + "linux-amd64.tar.gz", SHA256: helmSHA256LinuxAMD64},
		"linux/arm64":  {URL: helmBaseURL + "linux-arm64.tar.gz", SHA256: helmSHA256LinuxARM64},
		"darwin/amd64": {URL: helmBaseURL + "darwin-amd64.tar.gz", SHA256: helmSHA256DarwinAMD64},
		"darwin/arm64": {URL: helmBaseURL + "darwin-arm64.tar.gz", SHA256: helmSHA256DarwinARM64},
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
	if tool.Tarball {
		member := fmt.Sprintf("%s-%s/%s", runtime.GOOS, runtime.GOARCH, tool.Name)
		if err := d.InstallVerifiedTarGz(ctx, asset, member, dest, 0o750); err != nil {
			return "", err
		}
		return dest, nil
	}
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
