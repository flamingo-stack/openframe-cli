package models

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/flamingo-stack/openframe-cli/internal/shared/flags"
	"github.com/spf13/cobra"
)

// Use CommonFlags from internal/common as the single source of truth
type GlobalFlags = flags.CommonFlags

// CreateFlags contains flags specific to create command
type CreateFlags struct {
	GlobalFlags
	ClusterType string
	NodeCount   int
	K8sVersion  string
	SkipWizard  bool

	// Cloud-only flags (EKS/GKE)
	Region        string
	Profile       string // AWS
	Project       string // GCP
	MachineType   string
	MinNodes      int
	MaxNodes      int
	Spot          bool
	BackendConfig string
}

// ListFlags contains flags specific to list command
type ListFlags struct {
	GlobalFlags
	Quiet bool
	// All additionally discovers external cloud clusters (read-only).
	All bool
}

// StatusFlags contains flags specific to status command
type StatusFlags struct {
	GlobalFlags
	Detailed bool
	NoApps   bool
}

// DeleteFlags contains flags specific to delete command
type DeleteFlags struct {
	GlobalFlags
	Force bool // Delete-specific force flag
}

// CleanupFlags contains flags specific to cleanup command
type CleanupFlags struct {
	GlobalFlags
	Force bool // Cleanup-specific force flag
}

// Flag setup functions

// AddGlobalFlags adds global flags to a cluster command
func AddGlobalFlags(cmd *cobra.Command, global *GlobalFlags) {
	flagManager := flags.NewFlagManager(global)
	flagManager.AddCommonFlags(cmd)
}

// AddCreateFlags adds create-specific flags to a command
func AddCreateFlags(cmd *cobra.Command, flags *CreateFlags) {
	cmd.Flags().StringVarP(&flags.ClusterType, "type", "t", "", "Cluster type (k3d, eks, gke)")
	cmd.Flags().IntVarP(&flags.NodeCount, "nodes", "n", 3, "Number of nodes (default 3)")
	cmd.Flags().StringVar(&flags.K8sVersion, "version", "", "Kubernetes version")
	cmd.Flags().BoolVar(&flags.SkipWizard, "skip-wizard", false, "Skip interactive wizard")

	cmd.Flags().StringVar(&flags.Region, "region", "", "Cloud region (required for cloud types)")
	cmd.Flags().StringVar(&flags.Profile, "profile", "", "AWS credentials profile (eks only)")
	cmd.Flags().StringVar(&flags.Project, "project", "", "GCP project (required for --type gke)")
	cmd.Flags().StringVar(&flags.MachineType, "machine-type", "", "Node instance type (cloud only; defaults: m6i.large on eks, e2-standard-4 on gke)")
	cmd.Flags().IntVar(&flags.MinNodes, "min-nodes", 0, "Node group minimum size (cloud only)")
	cmd.Flags().IntVar(&flags.MaxNodes, "max-nodes", 0, "Node group maximum size (cloud only)")
	cmd.Flags().BoolVar(&flags.Spot, "spot", false, "Use spot capacity for nodes (cloud only)")
	cmd.Flags().StringVar(&flags.BackendConfig, "backend-config", "", "Remote terraform state: s3://bucket/prefix (eks) or gcs://bucket/prefix (gke); default is local state")
}

// AddListFlags adds list-specific flags to a command
func AddListFlags(cmd *cobra.Command, flags *ListFlags) {
	cmd.Flags().BoolVarP(&flags.Quiet, "quiet", "q", false, "Only show cluster names")
	cmd.Flags().BoolVarP(&flags.All, "all", "a", false, "Also discover external cloud clusters (read-only; needs provider CLI auth)")
}

// AddStatusFlags adds status-specific flags to a command
func AddStatusFlags(cmd *cobra.Command, flags *StatusFlags) {
	cmd.Flags().BoolVarP(&flags.Detailed, "detailed", "d", false, "Show detailed resource information")
	cmd.Flags().BoolVar(&flags.NoApps, "no-apps", false, "Skip application status checking")
}

// AddDeleteFlags adds delete-specific flags to a command
func AddDeleteFlags(cmd *cobra.Command, flags *DeleteFlags) {
	cmd.Flags().BoolVarP(&flags.Force, "force", "f", false, "Skip confirmation prompt")
}

// AddCleanupFlags adds cleanup-specific flags to a command
func AddCleanupFlags(cmd *cobra.Command, flags *CleanupFlags) {
	cmd.Flags().BoolVarP(&flags.Force, "force", "f", false, "Skip confirmation prompt and enable aggressive cleanup (remove all images, volumes, networks)")
}

// ValidateClusterName validates cluster name according to Kubernetes naming conventions
func ValidateClusterName(name string) error {
	// Trim whitespace and check if empty after trimming
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("cluster name cannot be empty or contain only whitespace")
	}

	// Check length (DNS-1123 subdomain: max 253 characters, but k3d has stricter limits)
	if len(trimmed) > 63 {
		return fmt.Errorf("cluster name is too long: %d characters (max 63)", len(trimmed))
	}

	// Check minimum length
	if len(trimmed) < 1 {
		return fmt.Errorf("cluster name must be at least 1 character")
	}

	// Check for invalid characters (DNS-1123 subdomain rules, but allow uppercase)
	// Must contain only alphanumeric characters or '-'
	// Must start and end with an alphanumeric character
	// Single character names are allowed if they are alphanumeric
	if len(trimmed) == 1 {
		if !regexp.MustCompile(`^[a-zA-Z0-9]$`).MatchString(trimmed) {
			return fmt.Errorf("cluster name '%s' is invalid: must be an alphanumeric character", trimmed)
		}
	} else {
		if !regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9]$`).MatchString(trimmed) {
			return fmt.Errorf("cluster name '%s' is invalid: must contain only letters, numbers, and hyphens, and must start and end with an alphanumeric character", trimmed)
		}
	}

	return nil
}

// Flag validation functions

// ValidateGlobalFlags validates global flag combinations
func ValidateGlobalFlags(globalFlags *GlobalFlags) error {
	return flags.ValidateCommonFlags(globalFlags)
}

// ValidateCreateFlags validates create flag combinations
func ValidateCreateFlags(flags *CreateFlags) error {
	if err := ValidateGlobalFlags(&flags.GlobalFlags); err != nil {
		return err
	}

	// Reject unknown --type values up front.
	clusterType := ClusterType(flags.ClusterType)
	switch clusterType {
	case "", ClusterTypeK3d, ClusterTypeGKE, ClusterTypeEKS:
		// known
	default:
		return fmt.Errorf("unknown cluster type '%s' (supported: k3d, eks, gke)", flags.ClusterType)
	}

	// The wizard prompts for these; in skip-wizard mode they must come from
	// flags. EKS is exempt while its creation is gated behind the coming-soon
	// banner — the banner must win over a missing-flag error.
	isCloud := clusterType == ClusterTypeEKS || clusterType == ClusterTypeGKE
	if clusterType == ClusterTypeGKE && flags.SkipWizard {
		if flags.Region == "" {
			return fmt.Errorf("--region is required for --type gke with --skip-wizard")
		}
		if flags.Project == "" {
			return fmt.Errorf("--project is required for --type gke with --skip-wizard")
		}
	}
	if flags.BackendConfig != "" && !isCloud {
		return fmt.Errorf("--backend-config only applies to cloud cluster types (eks, gke)")
	}
	if flags.MinNodes < 0 || flags.MaxNodes < 0 {
		return fmt.Errorf("node bounds must not be negative: min=%d max=%d", flags.MinNodes, flags.MaxNodes)
	}
	if flags.MinNodes > 0 && flags.MaxNodes > 0 && flags.MinNodes > flags.MaxNodes {
		return fmt.Errorf("--min-nodes (%d) must not exceed --max-nodes (%d)", flags.MinNodes, flags.MaxNodes)
	}

	// Validate node count - this validation is now handled at command level
	// to distinguish between explicitly set values and defaults
	if flags.NodeCount <= 0 {
		return fmt.Errorf("node count must be at least 1: %d", flags.NodeCount)
	}

	return nil
}

// ValidateListFlags validates list flag combinations
func ValidateListFlags(flags *ListFlags) error {
	return ValidateGlobalFlags(&flags.GlobalFlags)
}

// ValidateStatusFlags validates status flag combinations
func ValidateStatusFlags(flags *StatusFlags) error {
	return ValidateGlobalFlags(&flags.GlobalFlags)
}

// ValidateDeleteFlags validates delete flag combinations
func ValidateDeleteFlags(flags *DeleteFlags) error {
	return ValidateGlobalFlags(&flags.GlobalFlags)
}

// ValidateCleanupFlags validates cleanup flag combinations
func ValidateCleanupFlags(flags *CleanupFlags) error {
	return ValidateGlobalFlags(&flags.GlobalFlags)
}
