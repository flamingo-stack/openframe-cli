package ui

import (
	"fmt"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/manifoldco/promptui"
	"github.com/pterm/pterm"
)

// ClusterConfig holds cluster configuration for wizard
type ClusterConfig struct {
	Name       string
	Type       models.ClusterType
	NodeCount  int
	K8sVersion string
	// Cloud-only answers (EKS/GKE)
	Region      string
	Project     string
	MachineType string
}

// ToDomain converts the wizard answers into the domain config, attaching the
// cloud block only for cloud types.
func (c ClusterConfig) ToDomain() models.ClusterConfig {
	domain := models.ClusterConfig{
		Name:       c.Name,
		Type:       c.Type,
		NodeCount:  c.NodeCount,
		K8sVersion: c.K8sVersion,
	}
	if c.Type == models.ClusterTypeEKS || c.Type == models.ClusterTypeGKE {
		domain.Cloud = &models.CloudConfig{
			Region:      c.Region,
			Project:     c.Project,
			MachineType: c.MachineType,
		}
	}
	return domain
}

// ConfigWizard provides interactive configuration for cluster creation
type ConfigWizard struct {
	config ClusterConfig
}

// NewConfigWizard creates a new configuration wizard
func NewConfigWizard() *ConfigWizard {
	return &ConfigWizard{
		config: ClusterConfig{
			Name:       "openframe-dev",
			Type:       models.ClusterTypeK3d,
			NodeCount:  3,
			K8sVersion: "latest",
		},
	}
}

// SetDefaults sets the default values for the wizard
func (w *ConfigWizard) SetDefaults(name string, clusterType models.ClusterType, nodeCount int, k8sVersion string) {
	w.config.Name = name
	w.config.Type = clusterType
	w.config.NodeCount = nodeCount
	w.config.K8sVersion = k8sVersion
}

// Run starts the interactive configuration wizard
func (w *ConfigWizard) Run() (ClusterConfig, error) {
	pterm.Info.Println("Cluster Configuration Wizard")
	pterm.Info.Println("Configure your new Kubernetes cluster step by step")
	pterm.Println()

	steps := NewWizardSteps()

	// Step 1: Cluster name
	name, err := steps.PromptClusterName(w.config.Name)
	if err != nil {
		return ClusterConfig{}, err
	}
	w.config.Name = name

	// Step 2: Cluster type
	clusterType, err := steps.PromptClusterType()
	if err != nil {
		return ClusterConfig{}, err
	}
	w.config.Type = clusterType

	// Step 3 (cloud only): project/region + instance type. The k3s version
	// list below is meaningless for cloud clusters, whose version comes from
	// the module default.
	if clusterType == models.ClusterTypeEKS || clusterType == models.ClusterTypeGKE {
		defaultRegion, defaultMachine := "us-east-1", "m6i.large"
		regionLabel := "AWS Region"
		if clusterType == models.ClusterTypeGKE {
			defaultRegion, defaultMachine = "us-central1", "e2-standard-4"
			regionLabel = "GCP Region"

			project, err := steps.PromptProject()
			if err != nil {
				return ClusterConfig{}, err
			}
			w.config.Project = project
		}

		region, err := steps.PromptRegion(regionLabel, defaultRegion)
		if err != nil {
			return ClusterConfig{}, err
		}
		w.config.Region = region

		machineType, err := steps.PromptMachineType(defaultMachine)
		if err != nil {
			return ClusterConfig{}, err
		}
		w.config.MachineType = machineType
		w.config.K8sVersion = ""
	}

	// Step 4: Node count
	nodeCount, err := steps.PromptNodeCount(w.config.NodeCount)
	if err != nil {
		return ClusterConfig{}, err
	}
	w.config.NodeCount = nodeCount

	// Step 5 (k3d only): Kubernetes version
	if clusterType == models.ClusterTypeK3d {
		k8sVersion, err := steps.PromptK8sVersion()
		if err != nil {
			return ClusterConfig{}, err
		}
		w.config.K8sVersion = k8sVersion
	}

	// Step 6: Confirmation
	domainConfig := w.config.ToDomain()
	confirmed, err := steps.ConfirmConfiguration(domainConfig)
	if err != nil {
		return ClusterConfig{}, err
	}

	if !confirmed {
		// User wants to modify - restart wizard
		return w.Run()
	}

	return w.config, nil
}

// ConfigurationHandler handles cluster configuration flows
type ConfigurationHandler struct{}

// NewConfigurationHandler creates a new configuration handler
func NewConfigurationHandler() *ConfigurationHandler {
	return &ConfigurationHandler{}
}

// GetClusterConfig handles the complete cluster configuration flow
func (h *ConfigurationHandler) GetClusterConfig(clusterName string) (models.ClusterConfig, error) {
	// Show creation mode selection
	modeChoice, err := h.showCreationModeSelection()
	if err != nil {
		return models.ClusterConfig{}, err
	}

	if modeChoice == "quick" {
		return h.getQuickConfig(clusterName), nil
	}

	return h.getWizardConfig(clusterName)
}

// showCreationModeSelection shows the initial creation mode selection
func (h *ConfigurationHandler) showCreationModeSelection() (string, error) {
	pterm.Info.Printf("How would you like to create your cluster?\n")
	fmt.Println()

	prompt := promptui.Select{
		Label: "Creation Mode",
		Items: []string{
			"Default configuration",
			"Interactive configuration",
		},
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}:",
			Active:   "→ {{ . | cyan }}",
			Inactive: "  {{ . }}",
			Selected: "{{ . | green }}",
		},
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return "", err
	}

	if idx == 0 {
		return "quick", nil
	}
	return "wizard", nil
}

// getQuickConfig creates a quick default configuration
func (h *ConfigurationHandler) getQuickConfig(clusterName string) models.ClusterConfig {
	if clusterName == "" {
		clusterName = "openframe-dev"
	}

	return models.ClusterConfig{
		Name:       clusterName,
		Type:       models.ClusterTypeK3d,
		K8sVersion: "latest",
		NodeCount:  3,
	}
}

// getWizardConfig runs the interactive configuration wizard
func (h *ConfigurationHandler) getWizardConfig(clusterName string) (models.ClusterConfig, error) {
	wizard := NewConfigWizard()

	// Set defaults if cluster name provided
	if clusterName != "" {
		wizard.SetDefaults(clusterName, models.ClusterTypeK3d, 3, "latest")
	}

	wizardConfig, err := wizard.Run()
	if err != nil {
		return models.ClusterConfig{}, err
	}

	return wizardConfig.ToDomain(), nil
}
