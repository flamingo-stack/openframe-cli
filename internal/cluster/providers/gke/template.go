package gke

import (
	_ "embed"
	"fmt"
	"regexp"
	"strings"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
)

// mainTF is the generated root module. It is static — all per-cluster values
// travel through terraform.tfvars.json, so there is no HCL templating.
//
//go:embed templates/main.tf
var mainTF []byte

// tfvars mirrors the variables block of templates/main.tf.
type tfvars struct {
	ClusterName       string `json:"cluster_name"`
	Project           string `json:"project"`
	Region            string `json:"region"`
	KubernetesVersion string `json:"kubernetes_version,omitempty"`
	InstanceType      string `json:"instance_type,omitempty"`
	MinNodes          int    `json:"min_nodes,omitempty"`
	MaxNodes          int    `json:"max_nodes,omitempty"`
	DesiredNodes      int    `json:"desired_nodes,omitempty"`
	Spot              bool   `json:"spot,omitempty"`
}

// gkeVersionRE matches the <major>.<minor> form GKE expects (e.g. "1.33").
var gkeVersionRE = regexp.MustCompile(`^\d+\.\d+$`)

// tfvarsFor maps a validated ClusterConfig onto the template variables.
func tfvarsFor(config models.ClusterConfig) (tfvars, error) {
	cloud := config.Cloud

	version := strings.TrimPrefix(config.K8sVersion, "v")
	if version == "latest" {
		version = "" // template maps empty to the GKE default
	}
	if version != "" && !gkeVersionRE.MatchString(version) {
		return tfvars{}, models.NewInvalidConfigError("version", config.K8sVersion,
			"GKE expects <major>.<minor> (e.g. 1.33)")
	}

	return tfvars{
		ClusterName:       config.Name,
		Project:           cloud.Project,
		Region:            cloud.Region,
		KubernetesVersion: version,
		InstanceType:      cloud.MachineType,
		MinNodes:          cloud.MinNodes,
		MaxNodes:          cloud.MaxNodes,
		DesiredNodes:      config.NodeCount,
		Spot:              cloud.Spot,
	}, nil
}

// validate enforces the GKE-specific config invariants at the domain boundary.
func validate(config models.ClusterConfig) error {
	if err := models.ValidateClusterName(config.Name); err != nil {
		return models.NewInvalidConfigError("name", config.Name, err.Error())
	}
	if config.Type != models.ClusterTypeGKE {
		return models.NewProviderNotFoundError(config.Type)
	}
	if config.Cloud == nil || config.Cloud.Region == "" {
		return models.NewInvalidConfigError("region", "", "a region is required for GKE clusters (--region)")
	}
	if config.Cloud.Project == "" {
		return models.NewInvalidConfigError("project", "", "a GCP project is required for GKE clusters (--project)")
	}
	if config.NodeCount < 1 {
		return models.NewInvalidConfigError("nodeCount", config.NodeCount, "node count must be at least 1")
	}
	c := config.Cloud
	if c.MinNodes < 0 || c.MaxNodes < 0 {
		return models.NewInvalidConfigError("nodes", fmt.Sprintf("min=%d max=%d", c.MinNodes, c.MaxNodes), "node bounds must not be negative")
	}
	if c.MinNodes > 0 && c.MaxNodes > 0 && c.MinNodes > c.MaxNodes {
		return models.NewInvalidConfigError("nodes", fmt.Sprintf("min=%d max=%d", c.MinNodes, c.MaxNodes), "min nodes must not exceed max nodes")
	}
	return nil
}
