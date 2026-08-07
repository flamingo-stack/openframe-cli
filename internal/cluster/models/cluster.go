package models

import (
	"fmt"
	"strings"
	"time"
)

// ClusterType represents different types of Kubernetes clusters
type ClusterType string

const (
	ClusterTypeK3d ClusterType = "k3d"
	ClusterTypeGKE ClusterType = "gke"
	ClusterTypeEKS ClusterType = "eks"
)

// ParseClusterType normalizes a user-supplied cluster type: case-insensitive,
// accepting the provider names as aliases (aws → eks, gcp → gke) — the mental
// model "my AWS cluster" is as common as the product name. Every --type flag
// must parse through here so the aliases work identically across commands.
// Empty stays empty (the caller's default applies); unknown values are an
// error naming the accepted set.
func ParseClusterType(s string) (ClusterType, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "":
		return "", nil
	case "k3d":
		return ClusterTypeK3d, nil
	case "eks", "aws":
		return ClusterTypeEKS, nil
	case "gke", "gcp":
		return ClusterTypeGKE, nil
	default:
		return "", fmt.Errorf("unknown cluster type '%s' (supported: k3d, eks, gke)", s)
	}
}

// ClusterConfig holds cluster configuration
type ClusterConfig struct {
	Name       string      `json:"name"`
	Type       ClusterType `json:"type"`
	NodeCount  int         `json:"node_count"`
	K8sVersion string      `json:"k8s_version"`
	// Cloud carries the settings that only make sense for managed cloud
	// clusters (GKE/EKS). Nil for local clusters; the k3d backend rejects a
	// config that sets it.
	Cloud *CloudConfig `json:"cloud,omitempty"`
}

// GKERegionalZones is how many zones GKE spreads a regional cluster across by
// default (the module passes no explicit node locations). A regional node
// pool's node count is PER ZONE, so every display of an HA cluster's nodes
// must show the ×zones math — "Nodes: 3" that silently provisions 9 was a
// verification-report finding (S2).
const GKERegionalZones = 3

// CloudConfig holds the provider-agnostic knobs for a managed cloud cluster.
type CloudConfig struct {
	Region      string `json:"region"`
	Project     string `json:"project,omitempty"` // GCP project
	Profile     string `json:"profile,omitempty"` // AWS profile
	MachineType string `json:"machine_type,omitempty"`
	MinNodes    int    `json:"min_nodes,omitempty"`
	MaxNodes    int    `json:"max_nodes,omitempty"`
	Spot        bool   `json:"spot,omitempty"`
	// HA requests a regional (multi-zone) control plane and node pool. Default
	// (false) is a single-zone cluster, so the node count is exact — N means N,
	// not N-per-zone. GKE only.
	HA bool `json:"ha,omitempty"`
	// Zone is the single zone a zonal (non-HA) GKE cluster lives in, e.g.
	// "us-central1-a". Derived from Region by the provider; empty for HA.
	Zone string `json:"zone,omitempty"`
	// BackendConfig is an optional remote-state location
	// (s3://bucket/prefix for EKS, gcs://bucket/prefix for GKE);
	// empty means local state in the cluster workspace.
	BackendConfig string `json:"backend_config,omitempty"`
}

// ClusterSource says who owns a cluster's lifecycle.
type ClusterSource string

const (
	// SourceOpenframe: created by this CLI, has a workspace — full lifecycle.
	SourceOpenframe ClusterSource = "openframe"
	// SourceExternal: discovered in the cloud without a workspace — strictly
	// read-only; every mutating command must refuse it.
	SourceExternal ClusterSource = "external"
)

// ClusterInfo represents information about a cluster
type ClusterInfo struct {
	Name string      `json:"name"`
	Type ClusterType `json:"type"`
	// Source is empty for local (k3d) clusters, where ownership is implicit.
	Source ClusterSource `json:"source,omitempty"`
	// Context is the kubeconfig context that reaches this cluster, when known.
	Context string `json:"context,omitempty"`
	Project string `json:"project,omitempty"`
	// Profile is the AWS profile this cluster is reached with (EKS only).
	Profile string `json:"profile,omitempty"`
	Region  string `json:"region,omitempty"`
	// Status is a human-readable server fraction ("1/1"). Machine consumers
	// should prefer ReadyServers/TotalServers (verification report: a string
	// fraction forces JSON consumers to parse it).
	Status       string     `json:"status"`
	ReadyServers int        `json:"ready_servers"`
	TotalServers int        `json:"total_servers"`
	NodeCount    int        `json:"node_count"`
	K8sVersion   string     `json:"k8s_version,omitempty"`
	CreatedAt    time.Time  `json:"created_at,omitempty"`
	Nodes        []NodeInfo `json:"nodes,omitempty"`
}

// NodeInfo represents information about a node in the cluster
type NodeInfo struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Role   string `json:"role"`
}
