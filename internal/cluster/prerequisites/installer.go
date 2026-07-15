package prerequisites

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/docker"
	"github.com/flamingo-stack/openframe-cli/internal/shared/errors"
	"github.com/flamingo-stack/openframe-cli/internal/shared/ui"
	"github.com/flamingo-stack/openframe-cli/internal/shared/ui/spinner"
	"github.com/pterm/pterm"
)

type Installer struct {
	checker *PrerequisiteChecker
}

func NewInstaller() *Installer {
	return NewInstallerWithChecker(NewPrerequisiteChecker())
}

// NewInstallerWithChecker builds an installer around a specific requirement
// set (k3d local vs EKS cloud); the install/verify flow is data-driven from
// the checker's requirements.
func NewInstallerWithChecker(checker *PrerequisiteChecker) *Installer {
	return &Installer{checker: checker}
}

// requirement returns the checker requirement matching name (case-insensitive).
func (i *Installer) requirement(name string) *Requirement {
	for idx := range i.checker.requirements {
		if strings.EqualFold(i.checker.requirements[idx].Name, name) {
			return &i.checker.requirements[idx]
		}
	}
	return nil
}

func (i *Installer) installSpecificTools(tools []string) error {
	pterm.Info.Printf("Starting installation of %d tool(s): %s\n", len(tools), strings.Join(tools, ", "))

	for idx, tool := range tools {
		// Create a spinner for the installation process
		sp := spinner.New()
		sp.Start(fmt.Sprintf("[%d/%d] Installing %s...", idx+1, len(tools), tool))

		if err := i.installTool(tool); err != nil {
			sp.Fail(fmt.Sprintf("Failed to install %s: %v", tool, err))
			return fmt.Errorf("failed to install %s: %w", tool, err)
		}

		sp.Success(fmt.Sprintf("%s installed successfully", tool))
	}

	// Verify the installed tools are actually usable. Docker is special-cased:
	// its requirement's IsInstalled checks the daemon is RUNNING, which the
	// start/wait phase handles separately — here only binary presence matters.
	var stillMissing []string
	for _, tool := range tools {
		if strings.EqualFold(tool, "docker") {
			if !docker.NewDockerInstaller().IsInstalled() {
				stillMissing = append(stillMissing, "Docker")
			}
			continue
		}
		if req := i.requirement(tool); req != nil && !req.IsInstalled() {
			stillMissing = append(stillMissing, req.Name)
		}
	}

	if len(stillMissing) > 0 {
		pterm.Warning.Printf("Some tools failed to install: %s\n", strings.Join(stillMissing, ", "))
		return fmt.Errorf("installation failed for: %s", strings.Join(stillMissing, ", "))
	}

	// Don't show success here, let the main flow handle it
	return nil
}

// containsTool reports whether tools contains name (case-insensitive).
func containsTool(tools []string, name string) bool {
	for _, t := range tools {
		if strings.EqualFold(t, name) {
			return true
		}
	}
	return false
}

func (i *Installer) installTool(tool string) error {
	req := i.requirement(tool)
	if req == nil || req.Install == nil {
		return fmt.Errorf("unknown tool: %s", tool)
	}
	return req.Install()
}

// CheckAndInstallNonInteractive checks and installs prerequisites with optional non-interactive mode
func (i *Installer) CheckAndInstallNonInteractive(nonInteractive bool) error {
	// PHASE 1: Check what's actually missing vs what's not running
	allPresent, missing := i.checker.CheckAll()
	if allPresent {
		return nil
	}

	// Separate into truly missing tools vs Docker not running
	var missingTools []string
	var dockerNotRunning bool

	for _, tool := range missing {
		switch strings.ToLower(tool) {
		case "docker":
			if docker.NewDockerInstaller().IsInstalled() {
				// Docker is installed but not running - handle later
				dockerNotRunning = true
			} else {
				// Docker is not installed - needs installation
				missingTools = append(missingTools, "Docker")
			}
		default:
			// All other tools are truly missing if they show up in missing list
			missingTools = append(missingTools, tool)
		}
	}

	// PHASE 2: Install missing tools FIRST
	if len(missingTools) > 0 {
		pterm.Warning.Printf("Missing Prerequisites: %s\n", strings.Join(missingTools, ", "))

		var confirmed bool
		if nonInteractive {
			// Auto-approve in non-interactive mode
			pterm.Info.Println("Auto-installing prerequisites (non-interactive mode)...")
			confirmed = true
		} else {
			var err error
			confirmed, err = ui.ConfirmActionInteractive("Would you like me to install them automatically?", true)
			if err := errors.WrapConfirmationError(err, "failed to get user confirmation"); err != nil {
				return err
			}
		}

		if confirmed {
			if err := i.installSpecificTools(missingTools); err != nil {
				// Fail fast in BOTH modes. The old non-interactive path logged
				// "Continuing anyway" and proceeded to a guaranteed, confusingly
				// attributed k3d/helm failure minutes later — CI cannot fix
				// anything "later" anyway.
				return err
			}
			pterm.Success.Println("All missing tools installed successfully!")

			// A freshly installed Docker is not usable yet: Docker Desktop on
			// macOS takes tens of seconds to start, and on Linux the daemon may
			// not be running at all. Route it through the start/wait phase below
			// instead of letting the very next `k3d cluster create` fail.
			if containsTool(missingTools, "Docker") && !docker.IsDockerRunning() {
				dockerNotRunning = true
			}
		} else {
			i.showManualInstructions()
			return fmt.Errorf("required prerequisites are missing")
		}
	}

	// PHASE 3: Now check if Docker needs to be started (after all tools are installed)
	if dockerNotRunning {
		if nonInteractive {
			// In non-interactive mode, try to start Docker automatically
			pterm.Warning.Println("Docker is not running.")
			pterm.Info.Println("Attempting to start Docker automatically (non-interactive mode)...")

			if err := docker.StartDocker(); err != nil {
				// Fail fast: "continuing anyway" only moved the failure into the
				// next cluster operation with a misleading error.
				i.showDockerStartInstructions()
				return fmt.Errorf("the Docker daemon is not running and could not be started automatically (if it was just installed on Linux, a re-login may be needed for docker group membership): %w", err)
			}

			sp := spinner.New()
			sp.Start("Waiting for Docker to start...")
			if err := docker.WaitForDocker(); err != nil {
				sp.Fail("Docker failed to start")
				i.showDockerStartInstructions()
				return fmt.Errorf("timed out waiting for Docker to start (if it was just installed on Linux, a re-login may be needed for docker group membership): %w", err)
			}
			sp.Success("Docker started successfully")
		} else {
			// Interactive mode - prompt user
			pterm.Warning.Println("Docker is not running.")
			confirmed, err := ui.ConfirmActionInteractive("Would you like me to start Docker for you?", true)
			if err != nil {
				// A Ctrl-C interruption flows up as-is; other errors get context.
				return errors.WrapConfirmationError(err, "failed to get Docker start confirmation")
			}
			if confirmed {
				if err := docker.StartDocker(); err != nil {
					pterm.Info.Println("Please start Docker Desktop manually and try again.")
					return fmt.Errorf("failed to start Docker: %w", err)
				}
				sp := spinner.New()
				sp.Start("Waiting for Docker to start...")
				if err := docker.WaitForDocker(); err != nil {
					sp.Fail("Docker failed to start")
					pterm.Info.Println("Please start Docker Desktop manually and try again.")
					return fmt.Errorf("timed out waiting for Docker to start: %w", err)
				}
				sp.Success("Docker started successfully")
			} else {
				i.showDockerStartInstructions()
				return fmt.Errorf("the Docker daemon is not running")
			}
		}
	}

	return nil
}

func (i *Installer) showManualInstructions() {
	fmt.Println()
	pterm.Info.Println("Installation skipped. Here are manual installation instructions:")

	// Get instructions for this checker's prerequisites
	var allInstructions []string
	for _, req := range i.checker.requirements {
		allInstructions = append(allInstructions, req.InstallHelp())
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
}

func (i *Installer) showDockerStartInstructions() {
	fmt.Println()
	pterm.Info.Println("Please start Docker manually and try again:")
	switch runtime.GOOS {
	case "darwin":
		pterm.Printf("• Open Docker Desktop from Applications or Launchpad\n")
		pterm.Printf("• Or run: %s\n", pterm.Cyan("open -a Docker"))
		pterm.Printf("• Wait for Docker to fully start (whale icon in menu bar should be steady)\n")
	case "linux":
		pterm.Printf("• Start Docker daemon:\n")
		pterm.Printf("  %s\n", pterm.Cyan("sudo systemctl start docker"))
		pterm.Printf("• Or if using Docker Desktop:\n")
		pterm.Printf("  %s\n", pterm.Cyan("systemctl --user start docker-desktop"))
		pterm.Printf("• Enable Docker to start on boot (optional):\n")
		pterm.Printf("  %s\n", pterm.Cyan("sudo systemctl enable docker"))
	case "windows":
		pterm.Printf("• Start Docker Desktop from Start Menu or Desktop shortcut\n")
		pterm.Printf("• Or run from Command Prompt:\n")
		pterm.Printf("  %s\n", pterm.Cyan(`"C:\Program Files\Docker\Docker\Docker Desktop.exe"`))
		pterm.Printf("• Wait for Docker to fully start (system tray icon should show running)\n")
	default:
		pterm.Printf("• Start Docker Desktop or Docker daemon according to your system\n")
		pterm.Printf("• Verify Docker is running: %s\n", pterm.Cyan("docker ps"))
	}
}
