package config

import (
	"github.com/flamingo-stack/openframe-cli/internal/chart/models"
)

// ChartInstallConfig holds configuration for chart installation
type ChartInstallConfig struct {
	ClusterName string
	// KubeContext, when set, is the explicit kube-context every helm CLI call
	// must target (from --context / the interactive target selector). It wins
	// over the ClusterName-derived k3d context so a single install never talks
	// to two clusters (audit F4).
	KubeContext    string
	Force          bool
	DryRun         bool
	Verbose        bool
	Silent         bool
	NonInteractive bool // Suppresses interactive UI elements and spinners
	SkipCRDs       bool // Skip installation of ArgoCD CRDs
	// SyncStragglersOnStall lets the application wait trigger a one-shot sync of
	// OutOfSync-but-healthy stragglers when progress stalls. Set on the upgrade
	// (ref-change) path: children with autoSync disabled never roll a new ref
	// out by themselves, so waiting for them is provably futile (finding N3).
	SyncStragglersOnStall bool
	// App-of-apps specific configuration
	AppOfApps *models.AppOfAppsConfig
}

// HasAppOfApps returns true if app-of-apps configuration is provided
func (c *ChartInstallConfig) HasAppOfApps() bool {
	return c.AppOfApps != nil && c.AppOfApps.GitHubRepo != ""
}
