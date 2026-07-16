package platform

import "fmt"

// ErrWindowsNeedsWSL is returned when a cluster operation cannot run as a native
// Windows process. OpenFrame talks to Kubernetes via the native Go client
// (client-go); on Windows the cluster (Docker + k3d) runs inside WSL2, which a
// native Windows process cannot reach reliably. The supported path is to run
// OpenFrame itself inside WSL.
var ErrWindowsNeedsWSL = fmt.Errorf("cluster operations require running OpenFrame inside WSL on Windows")

// WSLClusterHint returns a user-facing error explaining how to perform the given
// operation on Windows, where cluster access must happen inside WSL. Returns nil
// on non-Windows hosts, so callers can use it as a guard:
//
//	if err := platform.WSLClusterHint("wait for ArgoCD"); err != nil {
//	    return err
//	}
func WSLClusterHint(operation string) error {
	if !IsWindows() {
		return nil
	}
	return windowsWSLError(operation)
}

// windowsWSLError builds the Windows-only guidance error. Split out so it is
// unit-testable regardless of the host OS.
func windowsWSLError(operation string) error {
	return fmt.Errorf(`%s: %w

On Windows the Kubernetes cluster (Docker + k3d) runs inside WSL2, and OpenFrame
reaches it through the native Kubernetes client — which cannot connect from a
native Windows process. Run OpenFrame inside WSL instead:

    wsl -d Ubuntu
    openframe <your command>

Or perform the step manually with kubectl from inside WSL`, operation, ErrWindowsNeedsWSL)
}
