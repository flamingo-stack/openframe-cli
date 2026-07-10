package prerequisites

import (
	"fmt"
	"strings"

	"github.com/flamingo-stack/openframe-cli/internal/chart/prerequisites/certificates"
	"github.com/flamingo-stack/openframe-cli/internal/chart/prerequisites/helm"
	"github.com/flamingo-stack/openframe-cli/internal/chart/prerequisites/memory"
	"github.com/flamingo-stack/openframe-cli/internal/shared/errors"
	"github.com/flamingo-stack/openframe-cli/internal/shared/ui"
	"github.com/flamingo-stack/openframe-cli/internal/shared/ui/spinner"
	"github.com/pterm/pterm"
)

type Installer struct {
	checker *PrerequisiteChecker
}

func NewInstaller() *Installer {
	return &Installer{
		checker: NewPrerequisiteChecker(),
	}
}

// installMissingToolsNonInteractive installs missing tools with optional non-interactive mode
func (i *Installer) installMissingToolsNonInteractive(tools []string, nonInteractive bool) error {
	if len(tools) == 0 {
		pterm.Success.Println("All prerequisites are already installed.")
		return nil
	}

	pterm.Info.Printf("Starting installation of %d prerequisite(s): %s\n", len(tools), strings.Join(tools, ", "))

	certsSkipped := false
	for idx, tool := range tools {
		// Skip memory as it can't be installed
		if strings.ToLower(tool) == "memory" {
			continue
		}
		// Certificate setup can't run non-interactively: generateCertificates runs
		// `mkcert -install`, which installs a local root CA into the system/browser
		// trust store and may need an interactive sudo password. Skip it here (the
		// consequence is that localhost HTTPS is served with an untrusted cert)
		// rather than run a no-op that reports "installed successfully" while the
		// re-check still finds it missing.
		if nonInteractive && strings.ToLower(tool) == "certificates" {
			pterm.Info.Println("Skipping certificates: mkcert -install needs interactive trust-store/sudo access — localhost HTTPS will be untrusted")
			certsSkipped = true
			continue
		}

		// Create a spinner for the installation process
		sp := spinner.New()
		sp.Start(fmt.Sprintf("[%d/%d] Installing %s...", idx+1, len(tools), tool))

		if err := i.installToolNonInteractive(tool, nonInteractive); err != nil {
			// In non-interactive mode, log error but continue with next tool
			if nonInteractive {
				sp.Warning(fmt.Sprintf("Skipped %s: %v", tool, err))
				continue
			}
			sp.Fail(fmt.Sprintf("Failed to install %s: %v", tool, err))
			return fmt.Errorf("failed to install %s: %w", tool, err)
		}

		sp.Success(fmt.Sprintf("%s installed successfully", tool))
	}

	// Verify all tools are now installed
	_, stillMissing := i.checker.CheckAll()

	// Filter out memory (not installable) and, in non-interactive mode,
	// certificates (intentionally skipped above) from verification.
	stillMissingInstallable := []string{}
	for _, tool := range stillMissing {
		lc := strings.ToLower(tool)
		if lc == "memory" || (nonInteractive && lc == "certificates") {
			continue
		}
		stillMissingInstallable = append(stillMissingInstallable, tool)
	}

	if len(stillMissingInstallable) > 0 {
		// Fail fast in BOTH modes: "continuing with available tools" just moved
		// the failure into the helm install minutes later with a misleading
		// error, which is worse in CI, not better.
		pterm.Warning.Printf("Some tools are still missing: %s\n", strings.Join(stillMissingInstallable, ", "))
		return fmt.Errorf("installation completed but some tools are still missing: %s", strings.Join(stillMissingInstallable, ", "))
	}

	// Don't claim "all installed" when certificates were deliberately skipped —
	// the re-check will still find them missing, so be honest about it.
	if certsSkipped {
		pterm.Success.Println("Installable prerequisites ready — certificates skipped (localhost HTTPS will be untrusted).")
	} else {
		pterm.Success.Println("All prerequisites installed successfully!")
	}
	return nil
}

// installToolNonInteractive installs a single tool with optional non-interactive mode
func (i *Installer) installToolNonInteractive(tool string, nonInteractive bool) error {
	switch strings.ToLower(tool) {
	case "helm":
		installer := helm.NewHelmInstaller()
		return installer.Install()
	case "memory":
		// Memory cannot be automatically installed
		return fmt.Errorf("memory cannot be automatically increased. Please add more physical RAM or increase virtual memory allocation")
	case "certificates":
		// Non-interactive callers skip certificates before reaching here (see
		// installMissingToolsNonInteractive); this path installs them interactively.
		installer := certificates.NewCertificateInstaller()
		return installer.Install()
	default:
		return fmt.Errorf("unknown tool: %s", tool)
	}
}

// CheckAndInstallNonInteractive checks and installs prerequisites with optional non-interactive mode
func (i *Installer) CheckAndInstallNonInteractive(nonInteractive bool) error {
	_, missing := i.checker.CheckAll()

	// Check memory separately for warning
	memChecker := memory.NewMemoryChecker()
	current, recommended, sufficient := memChecker.GetMemoryInfo()

	// Show memory warning if insufficient (but don't block)
	if !sufficient {
		pterm.Warning.Printfln("Insufficient memory: %d MB available, %d MB recommended", current, recommended)
		pterm.Info.Println("Charts may not deploy successfully with insufficient memory. Consider adding more RAM.")
		fmt.Println()
	}

	// Filter out memory from missing tools (we handle it as warning only)
	installableMissing := []string{}
	for _, tool := range missing {
		if strings.ToLower(tool) != "memory" {
			installableMissing = append(installableMissing, tool)
		}
	}

	if len(installableMissing) > 0 {
		// Show missing prerequisites with nice formatting
		pterm.Warning.Printf("Missing Prerequisites: %s\n", strings.Join(installableMissing, ", "))

		var confirmed bool
		if nonInteractive {
			// Auto-approve in non-interactive mode
			pterm.Info.Println("Auto-installing prerequisites (non-interactive mode)...")
			confirmed = true
		} else {
			// Single confirmation using shared UI
			var err error
			confirmed, err = ui.ConfirmActionInteractive("Would you like me to install them automatically?", true)
			if err := errors.WrapConfirmationError(err, "failed to get user confirmation"); err != nil {
				return err
			}
		}

		if confirmed {
			if err := i.installMissingToolsNonInteractive(installableMissing, nonInteractive); err != nil {
				// Fail fast in BOTH modes: the old non-interactive "continuing
				// anyway" deferred the failure to a guaranteed helm error with
				// the real cause buried in a scrolled-past warning.
				return err
			}
		} else {
			// Show manual installation instructions and exit
			fmt.Println()
			pterm.Info.Println("Installation skipped. Here are manual installation instructions:")

			// Get instructions for all prerequisites
			allInstructions := []string{
				helm.NewHelmInstaller().GetInstallHelp(),
				memory.NewMemoryChecker().GetInstallHelp(),
				certificates.NewCertificateInstaller().GetInstallHelp(),
			}

			tableData := pterm.TableData{{"Tool", "Installation Instructions"}}
			for _, instruction := range allInstructions {
				parts := strings.SplitN(instruction, ": ", 2)
				if len(parts) == 2 {
					tableData = append(tableData, []string{pterm.Cyan(parts[0]), parts[1]})
				} else {
					tableData = append(tableData, []string{"", instruction})
				}
			}

			_ = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
			return fmt.Errorf("required prerequisites are missing")
		}
	}

	return nil
}

// RegenerateCertificatesOnly just regenerates certificates without checking other prerequisites
// This should be used for the install command only
func (i *Installer) RegenerateCertificatesOnly() error {
	certInstaller := certificates.NewCertificateInstaller()
	sp := spinner.New()
	sp.Start("Refreshing certificates...")
	if err := certInstaller.ForceRegenerate(); err != nil {
		if strings.Contains(err.Error(), "user cancelled") {
			sp.Warning("Certificate trust skipped (deployment would be unsecure)")
		} else {
			sp.Warning(fmt.Sprintf("Could not refresh certificates: %v", err))
		}
		// Non-fatal - continue anyway
	} else {
		sp.Info("Certificates refreshed")
	}

	return nil
}
