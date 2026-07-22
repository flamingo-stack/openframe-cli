package eks

import (
	_ "embed"
	"fmt"
	"regexp"
	"strings"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
)

// mainTF is the generated root module. It is static — all per-cluster values
// travel through terraform.tfvars.json, so there is no HCL templating (and no
// HCL-escaping bugs).
//
//go:embed templates/main.tf
var mainTF []byte

// tfvars mirrors the variables block of templates/main.tf.
type tfvars struct {
	ClusterName       string `json:"cluster_name"`
	Region            string `json:"region"`
	Profile           string `json:"profile,omitempty"`
	KubernetesVersion string `json:"kubernetes_version,omitempty"`
	InstanceType      string `json:"instance_type,omitempty"`
	MinNodes          int    `json:"min_nodes,omitempty"`
	MaxNodes          int    `json:"max_nodes,omitempty"`
	DesiredNodes      int    `json:"desired_nodes,omitempty"`
	Spot              bool   `json:"spot,omitempty"`
}

// eksVersionRE matches the <major>.<minor> form EKS expects (e.g. "1.33").
var eksVersionRE = regexp.MustCompile(`^\d+\.\d+$`)

// tfvarsFor maps a validated ClusterConfig onto the template variables.
func tfvarsFor(config models.ClusterConfig) (tfvars, error) {
	cloud := config.Cloud

	version := strings.TrimPrefix(config.K8sVersion, "v")
	if version == "latest" {
		version = "" // template maps empty to the EKS default (its latest)
	}
	if version != "" && !eksVersionRE.MatchString(version) {
		return tfvars{}, models.NewInvalidConfigError("version", config.K8sVersion,
			"EKS expects <major>.<minor> (e.g. 1.33)")
	}

	vars := tfvars{
		ClusterName:       config.Name,
		Region:            cloud.Region,
		Profile:           cloud.Profile,
		KubernetesVersion: version,
		InstanceType:      cloud.MachineType,
		MinNodes:          cloud.MinNodes,
		MaxNodes:          cloud.MaxNodes,
		DesiredNodes:      config.NodeCount,
		Spot:              cloud.Spot,
	}
	return vars, nil
}

// validate enforces the EKS-specific config invariants at the domain boundary.
func validate(config models.ClusterConfig) error {
	if err := models.ValidateClusterName(config.Name); err != nil {
		return models.NewInvalidConfigError("name", config.Name, err.Error())
	}
	if config.Type != models.ClusterTypeEKS {
		return models.NewProviderNotFoundError(config.Type)
	}
	if config.Cloud == nil || config.Cloud.Region == "" {
		return models.NewInvalidConfigError("region", "", "a region is required for EKS clusters (--region)")
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
