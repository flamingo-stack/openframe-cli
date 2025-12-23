//go:build windows

package k3d

import (
	"syscall"
	"unsafe"
)

var (
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	procGetLongPathNameW = kernel32.NewProc("GetLongPathNameW")
)

// expandShortPath expands Windows 8.3 short filenames to their full long path names.
// For example: C:\Users\RUNNER~1\... -> C:\Users\runneradmin\...
// This is necessary because WSL doesn't understand Windows short filenames.
func expandShortPath(path string) (string, error) {
	if path == "" {
		return path, nil
	}

	// Convert path to UTF-16 for Windows API
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return path, err
	}

	// First call to get required buffer size
	n, _, _ := procGetLongPathNameW.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		0,
		0,
	)

	if n == 0 {
		// GetLongPathNameW failed - path might not exist or other error
		// Return original path as fallback
		return path, nil
	}

	// Allocate buffer and get the long path
	buf := make([]uint16, n)
	n, _, _ = procGetLongPathNameW.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(n),
	)

	if n == 0 {
		// Failed to get long path, return original
		return path, nil
	}

	return syscall.UTF16ToString(buf[:n]), nil
}
