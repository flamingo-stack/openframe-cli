package certificates

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/platform"
	"github.com/flamingo-stack/openframe-cli/internal/shared/download"
	"github.com/pterm/pterm"
)

type CertificateInstaller struct{}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func isMkcertInstalled() bool {
	return commandExists("mkcert")
}

func certificateInstallHelp() string {
	return platform.InstallHint("certificates")
}

func NewCertificateInstaller() *CertificateInstaller {
	return &CertificateInstaller{}
}

func (c *CertificateInstaller) IsInstalled() bool {
	// Only check if mkcert is installed, not if certificates exist
	// Certificates will be regenerated on every install command anyway
	return isMkcertInstalled()
}

func (c *CertificateInstaller) GetInstallHelp() string {
	return certificateInstallHelp()
}

func (c *CertificateInstaller) Install() error {
	// First install mkcert if needed
	if !isMkcertInstalled() {
		if err := c.installMkcert(); err != nil {
			return fmt.Errorf("failed to install mkcert: %w", err)
		}
	}

	// Then generate certificates
	return c.generateCertificates()
}

// ForceRegenerate always regenerates certificates even if they exist
func (c *CertificateInstaller) ForceRegenerate() error {
	// Check if mkcert is installed
	if !isMkcertInstalled() {
		return fmt.Errorf("mkcert is not installed")
	}

	// Always regenerate certificates
	return c.generateCertificates()
}

func (c *CertificateInstaller) installMkcert() error {
	switch runtime.GOOS {
	case "darwin":
		return c.installMkcertMacOS()
	case "linux":
		return c.installMkcertLinux()
	case "windows":
		return fmt.Errorf("automatic mkcert installation on Windows not supported. Please install from https://github.com/FiloSottile/mkcert")
	default:
		return fmt.Errorf("automatic mkcert installation not supported on %s", runtime.GOOS)
	}
}

func (c *CertificateInstaller) installMkcertMacOS() error {
	if !commandExists("brew") {
		return fmt.Errorf("automatic mkcert installation on macOS requires Homebrew. Please install brew first: https://brew.sh")
	}

	cmd := exec.Command("brew", "install", "mkcert")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install mkcert: %w", err)
	}

	return nil
}

func (c *CertificateInstaller) installMkcertLinux() error {
	// Verified, pinned download (SHA256) into ~/.openframe/bin — replacing the
	// unverified `curl dl.filippo.io/mkcert/latest?for=linux/amd64` install, which
	// hardcoded amd64 and ran an unauthenticated binary that injects a root CA
	// (audit T0.3). The bin dir is on this process's PATH (prepended at startup),
	// so the later `mkcert` invocations resolve it.
	binDir, err := download.UserBinDir()
	if err != nil {
		return fmt.Errorf("resolving bin directory: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pterm.Info.Printf("Downloading verified mkcert %s...\n", download.Mkcert.Version)
	path, err := (download.Downloader{}).InstallPinnedTool(ctx, download.Mkcert, binDir)
	if err != nil {
		return fmt.Errorf("installing verified mkcert: %w", err)
	}
	download.PrependToPath(binDir)
	pterm.Success.Printf("Installed verified mkcert %s to %s\n", download.Mkcert.Version, path)
	return nil
}

func (c *CertificateInstaller) generateCertificates() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	certDir := filepath.Join(homeDir, ".config", "openframe", "certs")
	if err := os.MkdirAll(certDir, 0750); err != nil {
		return fmt.Errorf("failed to create certificate directory: %w", err)
	}

	// Get CAROOT location
	caRootCmd := exec.Command("mkcert", "-CAROOT")
	caRootOutput, err := caRootCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get mkcert CAROOT: %w", err)
	}
	caRoot := strings.TrimSpace(string(caRootOutput))

	// Check if CA is already installed (check for CA key)
	rootCAKeyPath := filepath.Join(caRoot, "rootCA-key.pem")
	if _, err := os.Stat(rootCAKeyPath); os.IsNotExist(err) {
		// Install CA with NSS trust stores for browsers - silently
		installCmd := exec.Command("bash", "-c", "TRUST_STORES=nss mkcert -install 2>&1")
		output, err := installCmd.Output()
		if err != nil {
			// Check if it's a password/sudo issue
			if strings.Contains(string(output), "password") || strings.Contains(string(output), "sudo") {
				// Allow interactive password prompt
				installCmd := exec.Command("bash", "-c", "TRUST_STORES=nss mkcert -install")
				installCmd.Stdin = os.Stdin
				installCmd.Stdout = os.Stdout
				installCmd.Stderr = os.Stderr
				if err := installCmd.Run(); err != nil {
					return fmt.Errorf("failed to install mkcert CA: %w", err)
				}
			} else {
				return fmt.Errorf("failed to install mkcert CA: %w", err)
			}
		}
	}

	// Platform-specific trust handling - EXACTLY as in certificates.sh
	switch runtime.GOOS {
	case "darwin":
		if err := trustCADarwin(caRoot, homeDir); err != nil {
			return err
		}

	case "linux":
		// Optional: install certutil for browser NSS
		if !commandExists("certutil") {
			if commandExists("apt-get") {
				updateCmd := exec.Command("sudo", "apt-get", "update", "-y")
				if err := updateCmd.Run(); err != nil {
					pterm.Debug.Printf("apt-get update failed (certutil install is optional): %v\n", err)
				}
				installCmd := exec.Command("sudo", "apt-get", "install", "-y", "libnss3-tools", "ca-certificates")
				if err := installCmd.Run(); err != nil {
					pterm.Debug.Printf("apt-get install of certutil/ca-certificates failed (optional): %v\n", err)
				}
			}
		}

		// Clean old mkcert CA nicknames to avoid duplicates
		if commandExists("certutil") {
			nssDBPaths := []string{
				filepath.Join(homeDir, ".pki/nssdb"),
			}

			// Add Firefox profiles
			firefoxDir := filepath.Join(homeDir, ".mozilla/firefox")
			if entries, err := os.ReadDir(firefoxDir); err == nil {
				for _, entry := range entries {
					if entry.IsDir() {
						nssDBPaths = append(nssDBPaths, filepath.Join(firefoxDir, entry.Name()))
					}
				}
			}

			for _, dbPath := range nssDBPaths {
				certDBPath := filepath.Join(dbPath, "cert9.db")
				if fileExists(certDBPath) {
					listCmd := exec.Command("certutil", "-L", "-d", "sql:"+dbPath) // #nosec G204 -- explicit argv, no shell; command and args are internal, not untrusted input
					output, _ := listCmd.Output()

					lines := strings.Split(string(output), "\n")
					for _, line := range lines {
						if strings.HasPrefix(line, "mkcert ") {
							parts := strings.Fields(line)
							if len(parts) > 0 {
								nick := parts[0]
								deleteCmd := exec.Command("certutil", "-D", "-d", "sql:"+dbPath, "-n", nick) // #nosec G204 -- explicit argv, no shell; command and args are internal, not untrusted input
								if err := deleteCmd.Run(); err != nil {                                      // best effort
									pterm.Debug.Printf("best-effort removal of old mkcert NSS nickname %q failed: %v\n", nick, err)
								}
							}
						}
					}
				}
			}
		}

		// Install CA to system + NSS
		installSystemCmd := exec.Command("bash", "-c", "TRUST_STORES=system,nss mkcert -install")
		installSystemCmd.Stdin = os.Stdin
		installSystemCmd.Stdout = os.Stdout
		installSystemCmd.Stderr = os.Stderr
		if err := installSystemCmd.Run(); err != nil {
			pterm.Debug.Printf("mkcert -install (system,nss trust stores) failed: %v\n", err)
		}

		// Refresh trust stores
		if commandExists("update-ca-certificates") {
			updateCmd := exec.Command("sudo", "update-ca-certificates")
			if err := updateCmd.Run(); err != nil {
				pterm.Debug.Printf("update-ca-certificates failed: %v\n", err)
			}
		}
		if commandExists("update-ca-trust") {
			updateCmd := exec.Command("sudo", "update-ca-trust", "extract")
			if err := updateCmd.Run(); err != nil {
				pterm.Debug.Printf("update-ca-trust extract failed: %v\n", err)
			}
		}
	}

	// Generate localhost certificates (silently)
	generateCmd := exec.Command("bash", "-c", // #nosec G204 -- shell string built from constant/program-derived values, not untrusted input
		fmt.Sprintf("cd '%s' && mkcert -cert-file localhost.pem -key-file localhost-key.pem localhost 127.0.0.1 ::1 >/dev/null 2>&1", certDir))
	if err := generateCmd.Run(); err != nil {
		return fmt.Errorf("failed to generate certificates: %w", err)
	}

	return nil
}

// trustCADarwin adds the mkcert root CA to the macOS login keychain, removing any
// stale mkcert certificates first. It shells out to `security`, but the fragile
// parts — resolving the keychain, parsing the certificate list, and classifying
// the add-trusted-cert result — are pure helpers (below) so they can be tested
// off a real keychain.
func trustCADarwin(caRoot, homeDir string) error {
	kcOut, _ := exec.Command("security", "default-keychain", "-d", "user").Output()
	keychain := resolveLoginKeychain(string(kcOut), homeDir, fileExists)
	if keychain == "" {
		return nil // no login keychain — nothing to trust into
	}

	// Remove old mkcert certificates from the keychain (best-effort).
	findOut, _ := exec.Command("security", "find-certificate", "-a", "-c", "mkcert", "-Z", keychain).Output() // #nosec G204 -- explicit argv, no shell; keychain is program-derived
	for _, sha := range parseMkcertCertSHAs(string(findOut)) {
		if err := exec.Command("security", "delete-certificate", "-Z", sha, keychain).Run(); err != nil { // #nosec G204 -- explicit argv, no shell; values are program-derived
			pterm.Debug.Printf("best-effort removal of old mkcert certificate failed: %v\n", err)
		}
	}

	// Add the mkcert CA (silently; fall back to interactive if macOS asks).
	rootCAPem := filepath.Join(caRoot, "rootCA.pem")
	trustArgs := []string{"add-trusted-cert", "-r", "trustRoot", "-p", "ssl", "-k", keychain, rootCAPem}
	out, err := exec.Command("security", trustArgs...).CombinedOutput() // #nosec G204 -- explicit argv, no shell; values are program-derived
	switch classifyAddTrustedCert(string(out), err) {
	case trustNeedsInteractive:
		ic := exec.Command("security", trustArgs...) // #nosec G204 -- explicit argv, no shell; values are program-derived
		ic.Stdin, ic.Stdout, ic.Stderr = os.Stdin, os.Stdout, os.Stderr
		if err := ic.Run(); err != nil {
			return fmt.Errorf("certificate trust was not established (user cancelled or error occurred)")
		}
	case trustCancelled:
		return fmt.Errorf("certificate trust was not established (user cancelled or error occurred)")
	case trustAdded, trustOtherError:
		// Added, or a non-fatal error — continue with certificate generation.
	}
	return nil
}

// resolveLoginKeychain picks the macOS login keychain: the one reported by
// `security default-keychain` (quotes stripped) when it exists on disk, else
// ~/Library/Keychains/login.keychain-db, else "" when neither is present. Pure —
// exists abstracts the filesystem so it is testable off macOS.
func resolveLoginKeychain(defaultKeychainOut, homeDir string, exists func(string) bool) string {
	kc := strings.Trim(strings.TrimSpace(defaultKeychainOut), `"`)
	if kc != "" && exists(kc) {
		return kc
	}
	login := filepath.Join(homeDir, "Library/Keychains/login.keychain-db")
	if exists(login) {
		return login
	}
	return ""
}

// parseMkcertCertSHAs extracts SHA-1 hashes from `security find-certificate -a -Z`
// output, whose relevant lines look like "    SHA-1 hash: A1B2C3…". Parsing in Go
// (instead of a piped awk) makes it unit-testable.
func parseMkcertCertSHAs(out string) []string {
	var shas []string
	for _, line := range strings.Split(out, "\n") {
		if after, ok := strings.CutPrefix(strings.TrimSpace(line), "SHA-1 hash:"); ok {
			if h := strings.TrimSpace(after); h != "" {
				shas = append(shas, h)
			}
		}
	}
	return shas
}

// addTrustedCertOutcome classifies the result of `security add-trusted-cert`.
type addTrustedCertOutcome int

const (
	trustAdded            addTrustedCertOutcome = iota // succeeded
	trustNeedsInteractive                              // macOS blocked non-interactive auth; retry with a prompt
	trustCancelled                                     // the user cancelled the authorization
	trustOtherError                                    // some other, non-fatal error
)

// classifyAddTrustedCert maps the command's combined output + error to an outcome.
func classifyAddTrustedCert(output string, err error) addTrustedCertOutcome {
	if err == nil {
		return trustAdded
	}
	switch {
	case strings.Contains(output, "User interaction is not allowed"):
		return trustNeedsInteractive
	case strings.Contains(output, "authorization was canceled by the user"):
		return trustCancelled
	default:
		return trustOtherError
	}
}
