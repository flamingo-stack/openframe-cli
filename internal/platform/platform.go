// Package platform centralizes host-OS detection and the per-tool installation
// guidance shown when a prerequisite is missing. It gives the "non-technical
// user" requirements (reqs 21/30) a single home: macOS and Linux get concrete
// install hints, while Windows and any unknown OS get a documentation link
// instead of a hard failure.
package platform

import (
	"fmt"
	"runtime"
)

// OS identifies the host operating system.
type OS string

const (
	Darwin  OS = "darwin"
	Linux   OS = "linux"
	Windows OS = "windows"
)

// Current returns the host OS as reported by the Go runtime.
func Current() OS { return OS(runtime.GOOS) }

// IsWindows reports whether the host OS is Windows.
func IsWindows() bool { return Current() == Windows }

// IsMac reports whether the host OS is macOS.
func IsMac() bool { return Current() == Darwin }

// IsLinux reports whether the host OS is Linux.
func IsLinux() bool { return Current() == Linux }

// InstallDocs holds a tool's installation guidance per OS. Default is used for
// any OS without a specific entry.
type InstallDocs struct {
	Darwin  string
	Linux   string
	Windows string
	Default string
}

// Hint returns the guidance for the current OS, falling back to Default when
// the OS-specific entry is empty.
func (d InstallDocs) Hint() string { return d.hintFor(Current()) }

// hintFor returns the guidance for a specific OS, falling back to Default.
func (d InstallDocs) hintFor(os OS) string {
	switch os {
	case Darwin:
		if d.Darwin != "" {
			return d.Darwin
		}
	case Linux:
		if d.Linux != "" {
			return d.Linux
		}
	case Windows:
		if d.Windows != "" {
			return d.Windows
		}
	}
	return d.Default
}

// toolDocs maps a tool name to its per-OS installation guidance. This is the
// single source of truth for prerequisite install hints.
var toolDocs = map[string]InstallDocs{
	"docker": {
		Darwin:  "Docker: Install Docker Desktop from https://docker.com/products/docker-desktop or run 'brew install --cask docker'",
		Linux:   "Docker: Install using your package manager or from https://docs.docker.com/engine/install/",
		Windows: "Docker: Install Docker Desktop from https://docker.com/products/docker-desktop",
		Default: "Docker: Please install Docker from https://docker.com/",
	},
	"kubectl": {
		Darwin:  "kubectl: Run 'brew install kubectl' or download from https://kubernetes.io/docs/tasks/tools/install-kubectl-macos/",
		Linux:   "kubectl: Install using your package manager or from https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/",
		Windows: "kubectl: Download from https://kubernetes.io/docs/tasks/tools/install-kubectl-windows/",
		Default: "kubectl: Please install kubectl from https://kubernetes.io/docs/tasks/tools/",
	},
	"k3d": {
		Darwin:  "k3d: Run 'brew install k3d' or download from https://k3d.io/v5.4.6/#installation",
		Linux:   "k3d: Run 'curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash' or download from https://k3d.io/v5.4.6/#installation",
		Windows: "k3d: Download from https://github.com/k3d-io/k3d/releases or use chocolatey 'choco install k3d'",
		Default: "k3d: Please install k3d from https://k3d.io/v5.4.6/#installation",
	},
	"helm": {
		Darwin:  "helm: Run 'brew install helm' or download from https://helm.sh/docs/intro/install/",
		Linux:   "helm: Install via your package manager, run 'curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash', or download from https://helm.sh/docs/intro/install/",
		Windows: "helm: Download from https://helm.sh/docs/intro/install/ or install via chocolatey 'choco install kubernetes-helm'",
		Default: "helm: Please install helm from https://helm.sh/docs/intro/install/",
	},
	"certificates": {
		Darwin:  "Certificates: mkcert will be installed via Homebrew and certificates generated automatically",
		Linux:   "Certificates: mkcert will be downloaded and certificates generated automatically",
		Windows: "Certificates: Please install mkcert manually from https://github.com/FiloSottile/mkcert and run 'mkcert localhost 127.0.0.1'",
		Default: "Certificates: Please install mkcert from https://github.com/FiloSottile/mkcert",
	},
}

// InstallHint returns installation guidance for the named tool on the current
// OS. Unknown tools get a generic message so a missing entry never crashes a
// caller.
func InstallHint(tool string) string {
	if d, ok := toolDocs[tool]; ok {
		return d.Hint()
	}
	return fmt.Sprintf("%s: please install %s and ensure it is on your PATH", tool, tool)
}
