package selfupdate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// prevInfo records what the retained rollback binary is.
type prevInfo struct {
	Version string `json:"version"`
}

// prevBinaryPath is where the previously-installed binary is kept so `openframe
// update --rollback` can restore it without any download. prevInfoPath holds its
// version metadata.
func prevBinaryPath() (string, error) { return statePath("openframe.prev") }
func prevInfoPath() (string, error)   { return statePath("openframe.prev.json") }

func statePath(name string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".openframe", "state", name), nil
}

// savePrevious copies the just-replaced binary (currently at backupPath) to the
// rollback slot and records the version it holds.
func savePrevious(backupPath, version string) error {
	dst, err := prevBinaryPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
		return err
	}
	if err := copyFile(backupPath, dst, 0o755); err != nil {
		return err
	}
	info, err := prevInfoPath()
	if err != nil {
		return err
	}
	b, err := json.Marshal(prevInfo{Version: version})
	if err != nil {
		return err
	}
	return os.WriteFile(info, b, 0o600)
}

// PreviousVersion returns the version of the retained rollback binary, if any.
func PreviousVersion() (string, bool) {
	p, err := prevInfoPath()
	if err != nil {
		return "", false
	}
	b, err := os.ReadFile(p) //nolint:gosec // G304: fixed CLI-owned path
	if err != nil {
		return "", false
	}
	var info prevInfo
	if json.Unmarshal(b, &info) != nil || info.Version == "" {
		return "", false
	}
	return info.Version, true
}

// Rollback restores the binary saved by the most recent successful update,
// reverting to the previous version with no download. The rollback point is
// consumed (one level deep). progress, if non-nil, receives step messages.
func (u Updater) Rollback(ctx context.Context, progress func(string)) error {
	if u.goos() == "windows" {
		return fmt.Errorf("self-update is not supported for the native Windows launcher")
	}
	log := func(s string) {
		if progress != nil {
			progress(s)
		}
	}

	prev, err := prevBinaryPath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(prev); err != nil {
		return fmt.Errorf("no previous version to roll back to (nothing was saved by a prior update)")
	}
	exePath, err := u.resolveExe()
	if err != nil {
		return err
	}
	if err := dirWritable(filepath.Dir(exePath)); err != nil {
		return fmt.Errorf("cannot roll back %s: %w", exePath, err)
	}

	// Stage the saved binary next to the running one (same filesystem → atomic
	// rename), then swap it in.
	staged := exePath + ".rollback"
	defer func() { _ = os.Remove(staged) }()
	if err := copyFile(prev, staged, 0o755); err != nil {
		return fmt.Errorf("staging the previous binary: %w", err)
	}
	backup, err := swapExecutable(ctx, exePath, staged, log)
	if err != nil {
		return err
	}
	_ = os.Remove(backup)

	// The rollback point has been consumed.
	_ = os.Remove(prev)
	if info, e := prevInfoPath(); e == nil {
		_ = os.Remove(info)
	}
	log("Rolled back to the previous version.")
	return nil
}

// copyFile copies src to dst with mode perm via a temp file + atomic rename.
func copyFile(src, dst string, perm os.FileMode) error {
	in, err := os.Open(src) //nolint:gosec // G304: CLI-managed binary paths
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	tmp, err := os.CreateTemp(filepath.Dir(dst), ".of-copy-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := io.Copy(tmp, in); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, dst); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return nil
}
