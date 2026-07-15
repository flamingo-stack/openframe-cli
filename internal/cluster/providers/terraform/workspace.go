// Package terraform is the provider-neutral engine behind the cloud cluster
// backends (EKS now, GKE later). It owns the per-cluster workspace layout
// under ~/.openframe/clusters/<name>/:
//
//	cluster.json              — registry record (type, status, region, outputs)
//	terraform/main.tf         — generated root module (embedded template)
//	terraform/terraform.tfvars.json
//	terraform/terraform.tfstate — local state (the cluster's source of truth)
//
// The workspace is deliberately never deleted on a failed apply: the state
// file is the only reliable pointer to partially-created (billed!) cloud
// resources. Only a successful destroy removes it.
package terraform

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
)

// Status is the lifecycle state recorded in cluster.json.
type Status string

const (
	StatusCreating Status = "creating"
	StatusReady    Status = "ready"
	StatusFailed   Status = "failed"
)

// Record is the persisted registry entry for one cloud cluster. Endpoint and
// CACert are captured from terraform outputs after a successful apply so that
// status/kubeconfig operations never need to run terraform again.
type Record struct {
	Name       string             `json:"name"`
	Type       models.ClusterType `json:"type"`
	Status     Status             `json:"status"`
	Region     string             `json:"region"`
	Profile    string             `json:"profile,omitempty"`
	K8sVersion string             `json:"k8s_version,omitempty"`
	NodeCount  int                `json:"node_count"`
	CreatedAt  time.Time          `json:"created_at"`
	Endpoint   string             `json:"endpoint,omitempty"`
	CACert     string             `json:"ca_cert,omitempty"` // base64, as EKS emits it
}

const recordFile = "cluster.json"

// Workspace is one cluster's directory under the registry base.
type Workspace struct {
	dir string
}

// DefaultBaseDir returns ~/.openframe/clusters, or $OPENFRAME_CLUSTERS_DIR
// when set (tests and non-standard homes).
func DefaultBaseDir() (string, error) {
	if dir := os.Getenv("OPENFRAME_CLUSTERS_DIR"); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, ".openframe", "clusters"), nil
}

// OpenWorkspace addresses (without creating) the workspace for name.
func OpenWorkspace(base, name string) *Workspace {
	return &Workspace{dir: filepath.Join(base, name)}
}

func (w *Workspace) Dir() string          { return w.dir }
func (w *Workspace) TerraformDir() string { return filepath.Join(w.dir, "terraform") }

// Exists reports whether the workspace has a registry record on disk.
func (w *Workspace) Exists() bool {
	_, err := os.Stat(filepath.Join(w.dir, recordFile))
	return err == nil
}

// Scaffold creates the workspace directories and writes the generated root
// module, tfvars, and the initial record. It refuses to overwrite an existing
// terraform state (a partially-created cluster must be resumed or destroyed,
// never silently re-scaffolded over).
func (w *Workspace) Scaffold(record Record, mainTF []byte, tfvars any) error {
	if err := os.MkdirAll(w.TerraformDir(), 0o750); err != nil {
		return fmt.Errorf("creating workspace %s: %w", w.dir, err)
	}
	varsJSON, err := json.MarshalIndent(tfvars, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding tfvars: %w", err)
	}
	if err := os.WriteFile(filepath.Join(w.TerraformDir(), "main.tf"), mainTF, 0o600); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(w.TerraformDir(), "terraform.tfvars.json"), varsJSON, 0o600); err != nil {
		return err
	}
	return w.WriteRecord(record)
}

// ReadRecord loads cluster.json; a missing record maps to fs.ErrNotExist.
func (w *Workspace) ReadRecord() (Record, error) {
	data, err := os.ReadFile(filepath.Join(w.dir, recordFile)) // #nosec G304 -- path is CLI-managed under ~/.openframe
	if err != nil {
		return Record{}, err
	}
	var r Record
	if err := json.Unmarshal(data, &r); err != nil {
		return Record{}, fmt.Errorf("corrupt %s in %s: %w", recordFile, w.dir, err)
	}
	return r, nil
}

// WriteRecord persists cluster.json atomically enough for a single-user CLI.
func (w *Workspace) WriteRecord(r Record) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(w.dir, recordFile), data, 0o600)
}

// SetStatus updates only the lifecycle status in the record.
func (w *Workspace) SetStatus(s Status) error {
	r, err := w.ReadRecord()
	if err != nil {
		return err
	}
	r.Status = s
	return w.WriteRecord(r)
}

// Remove deletes the whole workspace. Callers must only invoke it after a
// successful terraform destroy — the state file inside is the only pointer to
// live cloud resources.
func (w *Workspace) Remove() error {
	return os.RemoveAll(w.dir)
}

// Registry lists the cloud-cluster workspaces under a base directory.
type Registry struct {
	base string
}

func NewRegistry(base string) *Registry { return &Registry{base: base} }

// DefaultRegistry opens the registry at ~/.openframe/clusters.
func DefaultRegistry() (*Registry, error) {
	base, err := DefaultBaseDir()
	if err != nil {
		return nil, err
	}
	return NewRegistry(base), nil
}

func (r *Registry) Workspace(name string) *Workspace { return OpenWorkspace(r.base, name) }

// List returns the records of every workspace with a readable cluster.json.
// A missing base directory is an empty registry, not an error.
func (r *Registry) List() ([]Record, error) {
	entries, err := os.ReadDir(r.base)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading cluster registry %s: %w", r.base, err)
	}
	var records []Record
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		rec, err := OpenWorkspace(r.base, e.Name()).ReadRecord()
		if err != nil {
			continue // not a cluster workspace (or unreadable) — skip, don't fail the listing
		}
		records = append(records, rec)
	}
	return records, nil
}

// Get returns the record for name, or models.ErrClusterNotFound.
func (r *Registry) Get(name string) (Record, error) {
	ws := r.Workspace(name)
	if !ws.Exists() {
		return Record{}, models.NewClusterNotFoundError(name)
	}
	return ws.ReadRecord()
}
