package selfupdate

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// prevBinaryPath is where the previously-installed binary is kept so `openframe
// update --rollback` can restore it without any download. Its version is not
// stored separately — the binary self-reports it via `--version`.
func prevBinaryPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".openframe", "state", "openframe.prev"), nil
}

// savePrevious copies the just-replaced binary (currently at backupPath) to the
// rollback slot.
func savePrevious(backupPath string) error {
	dst, err := prevBinaryPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
		return err
	}
	return copyFile(backupPath, dst, 0o755)
}

// PreviousVersion reports whether a rollback point exists and, best-effort, the
// version it holds (read from the binary itself; "" if it cannot report one).
func PreviousVersion() (string, bool) {
	p, err := prevBinaryPath()
	if err != nil {
		return "", false
	}
	if _, err := os.Stat(p); err != nil {
		return "", false
	}
	return binaryVersion(p), true
}

// binaryVersion runs "<path> --version" and returns the leading version token
// (the CLI prints "<version> (<commit>) built on <date>").
func binaryVersion(path string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, path, "--version").Output()
	if err != nil {
		return ""
	}
	if fields := strings.Fields(string(out)); len(fields) > 0 {
		return fields[0]
	}
	return ""
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
