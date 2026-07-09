package services

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

// TestAppSubsystemIsolatedFromClusterCreation enforces the architecture rule
// (reqs 18/19): the app (chart) subsystem must NOT import the cluster-CREATION
// subsystem. The concrete cluster service is injected by the composition root
// (cmd / bootstrap) through the types.ClusterAccess interface instead.
//
// Allowed cluster imports from the app subsystem:
//   - internal/cluster/models — shared domain types (ClusterInfo, …)
//   - internal/cluster/ui     — a reused cluster-selection presentation widget
//
// Forbidden: the root internal/cluster service package (ClusterService) and any
// internal/cluster/provider… package (k3d/gke creation).
func TestAppSubsystemIsolatedFromClusterCreation(t *testing.T) {
	const modulePrefix = "github.com/flamingo-stack/openframe-cli/internal/cluster"

	forbidden := func(importPath string) bool {
		if importPath == modulePrefix {
			return true // the root cluster service package (creation/orchestration)
		}
		return strings.HasPrefix(importPath, modulePrefix+"/provider")
	}

	// The test runs with CWD = this package dir; ".." is internal/chart.
	root := ".."
	fset := token.NewFileSet()
	var violations []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, perr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if perr != nil {
			return perr
		}
		for _, imp := range f.Imports {
			p := strings.Trim(imp.Path.Value, `"`)
			if forbidden(p) {
				violations = append(violations, path+" imports "+p)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking internal/chart: %v", err)
	}

	if len(violations) > 0 {
		t.Errorf("app subsystem must not import cluster-creation code (reqs 18/19):\n  %s",
			strings.Join(violations, "\n  "))
	}
}
