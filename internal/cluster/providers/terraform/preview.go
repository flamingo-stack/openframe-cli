package terraform

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// CopyPlanInputs copies the inputs a preview plan of an EXISTING workspace
// needs into a throwaway directory: the local state (the diff base) and the
// backend block (a remote-state workspace must preview against its real
// state). The generated module is deliberately NOT copied — the preview
// regenerates it from the current template and flags, exactly like a resume
// (CreateCluster) does, so the plan shows what the resume would actually
// apply. Missing files are fine: a workspace that failed before the first
// apply has no state yet.
func CopyPlanInputs(srcDir, dstDir string) error {
	for _, name := range []string{"terraform.tfstate", "backend.tf"} {
		data, err := os.ReadFile(filepath.Join(srcDir, name)) // #nosec G304 -- path is CLI-managed under ~/.openframe
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dstDir, name), data, 0o600); err != nil { // #nosec G703 -- dstDir is our own MkdirTemp dir; srcDir is CLI-managed under ~/.openframe
			return err
		}
	}
	return nil
}
