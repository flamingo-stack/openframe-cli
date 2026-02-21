package bootstrap

import (
	"fmt"
	"runtime"
	"strings"

	chartCerts "github.com/flamingo-stack/openframe-cli/internal/chart/prerequisites/certificates"
	chartGit "github.com/flamingo-stack/openframe-cli/internal/chart/prerequisites/git"
	chartHelm "github.com/flamingo-stack/openframe-cli/internal/chart/prerequisites/helm"
	chartMemory "github.com/flamingo-stack/openframe-cli/internal/chart/prerequisites/memory"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/docker"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/k3d"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/kubectl"
	sharedErrors "github.com/flamingo-stack/openframe-cli/internal/shared/errors"
	"github.com/flamingo-stack/openframe-cli/internal/shared/ui"
	"github.com/pterm/pterm"
)

// PreflightChecker runs all prerequisite checks upfront before any work begins.
// This unifies the cluster and chart prerequisite gates so users don't create a
// cluster only to fail on chart prerequisites.
type PreflightChecker struct {
	nonInteractive bool
	force          bool
	verbose        bool
}

// NewPreflightChecker creates a new unified preflight checker.
func NewPreflightChecker(nonInteractive, force, verbose bool) *PreflightChecker {
	return &PreflightChecker{
		nonInteractive: nonInteractive,
		force:          force,
		verbose:        verbose,
	}
}

// preflightTool represents a tool to check during preflight.
type preflightTool struct {
	Name        string
	Category    string // "cluster" or "chart"
	IsInstalled func() bool
	InstallHelp func() string
	Installable bool // false for things like memory
}

// CheckAll runs all prerequisite checks and installs missing tools.
// It checks cluster prerequisites (Docker, kubectl, k3d, helm) and chart
// prerequisites (git, helm, memory, certificates) in a single pass.
func (p *PreflightChecker) CheckAll() error {
	// Phase 1: Check memory upfront — fail fast if insufficient
	if err := p.checkMemory(); err != nil {
		return err
	}

	// Phase 2: Check all tools
	tools := p.getAllTools()
	var missing []preflightTool

	for _, tool := range tools {
		if !tool.IsInstalled() {
			missing = append(missing, tool)
		}
	}

	if len(missing) == 0 {
		return nil
	}

	// Separate truly missing (installable) from Docker-not-running
	var installable []preflightTool
	var dockerNotRunning bool

	for _, tool := range missing {
		if tool.Name == "Docker" {
			if docker.NewDockerInstaller().IsInstalled() {
				dockerNotRunning = true
			} else {
				installable = append(installable, tool)
			}
		} else {
			installable = append(installable, tool)
		}
	}

	// Phase 3: Install missing tools
	if len(installable) > 0 {
		names := make([]string, len(installable))
		for i, t := range installable {
			names[i] = t.Name
		}
		pterm.Warning.Printf("Missing Prerequisites: %s\n", strings.Join(names, ", "))

		var confirmed bool
		if p.nonInteractive {
			pterm.Info.Println("Auto-installing prerequisites (non-interactive mode)...")
			confirmed = true
		} else {
			var err error
			confirmed, err = ui.ConfirmActionInteractive("Would you like me to install them automatically?", true)
			if err := sharedErrors.WrapConfirmationError(err, "failed to get user confirmation"); err != nil {
				return err
			}
		}

		if confirmed {
			if err := p.installTools(installable); err != nil {
				if p.nonInteractive {
					pterm.Warning.Printf("Failed to install some prerequisites: %v\n", err)
					pterm.Info.Println("Continuing anyway (non-interactive mode)...")
				} else {
					return err
				}
			}
		} else {
			p.showManualInstructions(installable)
			return fmt.Errorf("prerequisites not installed")
		}
	}

	// Phase 4: Start Docker if needed
	if dockerNotRunning {
		if err := p.startDocker(); err != nil {
			return err
		}
	}

	return nil
}

// checkMemory validates system memory against the recommended minimum.
func (p *PreflightChecker) checkMemory() error {
	memChecker := chartMemory.NewMemoryChecker()
	current, recommended, sufficient := memChecker.GetMemoryInfo()

	if !sufficient {
		pterm.Warning.Printf("Memory Warning: %d MB available, %d MB recommended\n", current, recommended)
		if p.force {
			pterm.Info.Println("Continuing anyway (--force specified).")
			return nil
		}
		if p.nonInteractive {
			pterm.Info.Println("Continuing anyway (non-interactive mode).")
			return nil
		}
		pterm.Info.Println("Charts may not deploy successfully with insufficient memory.")
		confirmed, err := ui.ConfirmActionInteractive("Continue anyway?", false)
		if err != nil {
			return fmt.Errorf("failed to get memory confirmation: %w", err)
		}
		if !confirmed {
			return fmt.Errorf("insufficient memory: %d MB available, %d MB recommended. Use --force to override", current, recommended)
		}
	}
	return nil
}

// getAllTools returns all prerequisite tools for both cluster and chart phases.
func (p *PreflightChecker) getAllTools() []preflightTool {
	return []preflightTool{
		// Cluster prerequisites
		{
			Name:        "Docker",
			Category:    "cluster",
			IsInstalled: func() bool { return docker.IsDockerRunning() },
			InstallHelp: func() string { return docker.NewDockerInstaller().GetInstallHelp() },
			Installable: true,
		},
		{
			Name:        "kubectl",
			Category:    "cluster",
			IsInstalled: func() bool { return kubectl.NewKubectlInstaller().IsInstalled() },
			InstallHelp: func() string { return kubectl.NewKubectlInstaller().GetInstallHelp() },
			Installable: true,
		},
		{
			Name:        "k3d",
			Category:    "cluster",
			IsInstalled: func() bool { return k3d.NewK3dInstaller().IsInstalled() },
			InstallHelp: func() string { return k3d.NewK3dInstaller().GetInstallHelp() },
			Installable: true,
		},
		{
			Name:        "Helm",
			Category:    "cluster",
			IsInstalled: func() bool { return chartHelm.NewHelmInstaller().IsInstalled() },
			InstallHelp: func() string { return chartHelm.NewHelmInstaller().GetInstallHelp() },
			Installable: true,
		},
		// Chart prerequisites (excluding memory — handled separately)
		{
			Name:        "Git",
			Category:    "chart",
			IsInstalled: func() bool { return chartGit.NewGitChecker().IsInstalled() },
			InstallHelp: func() string { return chartGit.NewGitChecker().GetInstallInstructions() },
			Installable: false, // git install is manual
		},
		{
			Name:        "Certificates",
			Category:    "chart",
			IsInstalled: func() bool { return chartCerts.NewCertificateInstaller().IsInstalled() },
			InstallHelp: func() string { return chartCerts.NewCertificateInstaller().GetInstallHelp() },
			Installable: true,
		},
	}
}

// installTools installs the given list of missing tools.
func (p *PreflightChecker) installTools(tools []preflightTool) error {
	for idx, tool := range tools {
		spinner, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("[%d/%d] Installing %s...", idx+1, len(tools), tool.Name))

		if err := p.installTool(tool); err != nil {
			if p.nonInteractive {
				spinner.Warning(fmt.Sprintf("Skipped %s: %v", tool.Name, err))
				continue
			}
			spinner.Fail(fmt.Sprintf("Failed to install %s: %v", tool.Name, err))
			return fmt.Errorf("failed to install %s: %w", tool.Name, err)
		}

		spinner.Success(fmt.Sprintf("%s installed successfully", tool.Name))
	}
	return nil
}

// installTool installs a single tool by name.
func (p *PreflightChecker) installTool(tool preflightTool) error {
	switch tool.Name {
	case "Docker":
		return docker.NewDockerInstaller().Install()
	case "kubectl":
		return kubectl.NewKubectlInstaller().Install()
	case "k3d":
		return k3d.NewK3dInstaller().Install()
	case "Helm":
		return chartHelm.NewHelmInstaller().Install()
	case "Certificates":
		if p.nonInteractive {
			pterm.Info.Println("Skipping certificate generation in non-interactive mode")
			return nil
		}
		return chartCerts.NewCertificateInstaller().Install()
	case "Git":
		return fmt.Errorf("git is not installed. %s", chartGit.NewGitChecker().GetInstallInstructions())
	default:
		return fmt.Errorf("unknown tool: %s", tool.Name)
	}
}

// startDocker attempts to start Docker.
func (p *PreflightChecker) startDocker() error {
	if p.nonInteractive {
		pterm.Warning.Println("Docker is not running.")
		pterm.Info.Println("Attempting to start Docker automatically (non-interactive mode)...")

		if err := docker.StartDocker(); err != nil {
			pterm.Warning.Printf("Could not start Docker automatically: %v\n", err)
			pterm.Info.Println("Docker must be started manually.")
			return nil
		}

		spinner, _ := pterm.DefaultSpinner.Start("Waiting for Docker to start...")
		if err := docker.WaitForDocker(); err != nil {
			spinner.Warning("Docker failed to start automatically")
			return nil
		}
		spinner.Success("Docker started successfully")
		return nil
	}

	pterm.Warning.Println("Docker is not running.")
	confirmed, err := ui.ConfirmActionInteractive("Would you like me to start Docker for you?", true)
	if err != nil {
		return fmt.Errorf("failed to get Docker start confirmation: %w", err)
	}

	if confirmed {
		if err := docker.StartDocker(); err != nil {
			pterm.Error.Printf("Failed to start Docker: %v\n", err)
			pterm.Info.Println("Please start Docker Desktop manually and try again.")
			return fmt.Errorf("failed to start Docker: %w", err)
		}
		spinner, _ := pterm.DefaultSpinner.Start("Waiting for Docker to start...")
		if err := docker.WaitForDocker(); err != nil {
			spinner.Fail("Docker failed to start")
			return fmt.Errorf("Docker failed to start: %w", err)
		}
		spinner.Success("Docker started successfully")
	} else {
		p.showDockerStartInstructions()
		return fmt.Errorf("Docker is not running")
	}

	return nil
}

// showManualInstructions displays manual installation instructions for missing tools.
func (p *PreflightChecker) showManualInstructions(tools []preflightTool) {
	fmt.Println()
	pterm.Info.Println("Installation skipped. Here are manual installation instructions:")

	tableData := pterm.TableData{{"Tool", "Installation Instructions"}}
	for _, tool := range tools {
		help := tool.InstallHelp()
		parts := strings.SplitN(help, ": ", 2)
		if len(parts) == 2 {
			tableData = append(tableData, []string{pterm.Cyan(parts[0]), parts[1]})
		} else {
			tableData = append(tableData, []string{pterm.Cyan(tool.Name), help})
		}
	}

	pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
}

// showDockerStartInstructions displays platform-specific Docker start instructions.
func (p *PreflightChecker) showDockerStartInstructions() {
	fmt.Println()
	pterm.Info.Println("Please start Docker manually and try again:")
	switch runtime.GOOS {
	case "darwin":
		pterm.Printf("  Open Docker Desktop or run: %s\n", pterm.Cyan("open -a Docker"))
	case "linux":
		pterm.Printf("  %s\n", pterm.Cyan("sudo systemctl start docker"))
	case "windows":
		pterm.Printf("  Start Docker Desktop from Start Menu\n")
	default:
		pterm.Printf("  Verify Docker is running: %s\n", pterm.Cyan("docker ps"))
	}
}
