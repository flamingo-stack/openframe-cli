package models

import "time"

// ClusterType represents different types of Kubernetes clusters
type ClusterType string

const (
	ClusterTypeK3d ClusterType = "k3d"
	ClusterTypeGKE ClusterType = "gke"
	ClusterTypeEKS ClusterType = "eks"
)

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

// CloudConfig holds the provider-agnostic knobs for a managed cloud cluster.
type CloudConfig struct {
	Region      string `json:"region"`
	Project     string `json:"project,omitempty"` // GCP project
	Profile     string `json:"profile,omitempty"` // AWS profile
	MachineType string `json:"machine_type,omitempty"`
	MinNodes    int    `json:"min_nodes,omitempty"`
	MaxNodes    int    `json:"max_nodes,omitempty"`
	Spot        bool   `json:"spot,omitempty"`
	// BackendConfig is an optional remote-state location
	// (s3://bucket/prefix for EKS, gcs://bucket/prefix for GKE);
	// empty means local state in the cluster workspace.
	BackendConfig string `json:"backend_config,omitempty"`
}

// ClusterInfo represents information about a cluster
type ClusterInfo struct {
	Name string      `json:"name"`
	Type ClusterType `json:"type"`
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
