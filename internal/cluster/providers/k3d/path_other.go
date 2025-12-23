//go:build !windows

package k3d

// expandShortPath is a no-op on non-Windows platforms.
// Windows short filenames (8.3 format) are only relevant on Windows.
func expandShortPath(path string) (string, error) {
	return path, nil
}
