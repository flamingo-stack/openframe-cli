package executor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/shared/redact"
	"github.com/pterm/pterm"
)

// WSL error exit codes
const (
	// WSLExitCodeDistroNotFound indicates the WSL distribution was not found or not accessible
	// This is 0xFFFFFFFF (-1 as unsigned 32-bit) which Windows returns when WSL can't reach the distro
	WSLExitCodeDistroNotFound = 4294967295
	// WSLExitCodeGenericError is a generic WSL error
	WSLExitCodeGenericError = 1
)

// WSLError represents an error specific to WSL operations
type WSLError struct {
	Operation  string
	ExitCode   int
	Stderr     string
	Suggestion string
}

func (e *WSLError) Error() string {
	msg := fmt.Sprintf("WSL error during %s (exit code: %d)", e.Operation, e.ExitCode)
	if e.Stderr != "" {
		msg += fmt.Sprintf(": %s", e.Stderr)
	}
	if e.Suggestion != "" {
		msg += fmt.Sprintf("\nSuggestion: %s", e.Suggestion)
	}
	return msg
}

// CommandError is returned when an external command exits non-zero. It carries
// the child's exit code so the top level can propagate it (exit-code fidelity
// for automation) AND the child's stderr — without it the message degrades to
// the useless "exit status 1" (`*exec.ExitError`'s own string), and the actual
// reason ("port 6550 already allocated", "no space left on device") was only
// ever printed under --verbose. Modelled on WSLError above.
//
// Stderr arrives already redacted from the executor (secrets can be echoed back
// by child processes).
type CommandError struct {
	Command  string
	ExitCode int
	Stderr   string
	cause    error
}

// maxStderrInError bounds how much of a chatty child's stderr lands in the
// error string; the full text is still available via the Stderr field.
const maxStderrInError = 2000

func (e *CommandError) Error() string {
	msg := fmt.Sprintf("command failed: %s (exit code: %d)", e.Command, e.ExitCode)
	if reason := strings.TrimSpace(e.Stderr); reason != "" {
		if len(reason) > maxStderrInError {
			reason = "..." + reason[len(reason)-maxStderrInError:]
		}
		return msg + ": " + reason
	}
	// No stderr (e.g. the child only wrote to stdout): fall back to the exec
	// error, which at least carries the signal/exit description.
	return fmt.Sprintf("%s: %v", msg, e.cause)
}

// Unwrap exposes the underlying exec error so errors.As/Is still reach it.
func (e *CommandError) Unwrap() error { return e.cause }

// wslAvailabilityCache caches the WSL availability check result
var (
	wslAvailable     bool
	wslChecked       bool
	wslCheckMutex    sync.Mutex
	wslUbuntuChecked bool
	wslUbuntuAvail   bool
)

// IsWSLAvailable checks if WSL is available on the system
func IsWSLAvailable() bool {
	if runtime.GOOS != "windows" {
		return false
	}

	wslCheckMutex.Lock()
	defer wslCheckMutex.Unlock()

	if wslChecked {
		return wslAvailable
	}

	// Try to run wsl --status
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "wsl", "--status")
	err := cmd.Run()
	wslAvailable = err == nil
	wslChecked = true

	return wslAvailable
}

// IsWSLUbuntuAvailable checks if the Ubuntu distribution is available and accessible in WSL
func IsWSLUbuntuAvailable() bool {
	if runtime.GOOS != "windows" {
		return false
	}

	wslCheckMutex.Lock()
	defer wslCheckMutex.Unlock()

	if wslUbuntuChecked {
		return wslUbuntuAvail
	}

	// Try to run a simple command in Ubuntu
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "wsl", "-d", "Ubuntu", "echo", "ok")
	output, err := cmd.Output()
	wslUbuntuAvail = err == nil && strings.TrimSpace(string(output)) == "ok"
	wslUbuntuChecked = true

	return wslUbuntuAvail
}

// ResetWSLCache resets the WSL availability cache (useful for testing or after WSL restart)
func ResetWSLCache() {
	wslCheckMutex.Lock()
	defer wslCheckMutex.Unlock()
	wslChecked = false
	wslUbuntuChecked = false
}

// TryRecoverWSL attempts to recover WSL connectivity by terminating and restarting the distribution
// This is a last-resort operation when WSL becomes completely unresponsive
// Returns nil if recovery was successful, error otherwise
func TryRecoverWSL() error {
	if runtime.GOOS != "windows" {
		return nil
	}

	// First, try to terminate the Ubuntu distribution
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	terminateCmd := exec.CommandContext(ctx, "wsl", "--terminate", "Ubuntu")
	_ = terminateCmd.Run() // Ignore error - distribution might not be running

	// Wait a moment for WSL to fully terminate
	time.Sleep(2 * time.Second)

	// Now try to start Ubuntu with a simple command
	startCtx, startCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer startCancel()

	startCmd := exec.CommandContext(startCtx, "wsl", "-d", "Ubuntu", "echo", "recovered")
	output, err := startCmd.Output()
	if err != nil {
		return fmt.Errorf("WSL recovery failed - could not restart Ubuntu: %w", err)
	}

	if strings.TrimSpace(string(output)) != "recovered" {
		return fmt.Errorf("WSL recovery returned unexpected output: %s", string(output))
	}

	// Reset the cache since we just restarted WSL
	ResetWSLCache()

	// After WSL restart, Docker daemon needs to be restarted too
	// Docker CE runs as a background process in WSL, not as a systemd service
	if err := RestartDockerInWSL(); err != nil {
		// Log warning but don't fail - Docker might already be running or not installed
		return fmt.Errorf("WSL recovered but Docker restart failed: %w", err)
	}

	return nil
}

// RestartDockerInWSL starts the Docker daemon inside WSL2 Ubuntu
// This is needed after WSL restart since Docker CE runs as a background process
func RestartDockerInWSL() error {
	if runtime.GOOS != "windows" {
		return nil
	}

	// Start Docker daemon in WSL using the start-docker.sh script we created during installation
	// If the script doesn't exist, fall back to starting dockerd directly
	startScript := `
if [ -x /usr/local/bin/start-docker.sh ]; then
    sudo /usr/local/bin/start-docker.sh
else
    # Fallback: start dockerd directly if script doesn't exist
    if ! pgrep -x dockerd > /dev/null; then
        sudo dockerd > /dev/null 2>&1 &
    fi
fi

# Wait for Docker to be ready (up to 30 seconds)
for i in $(seq 1 30); do
    if sudo docker ps > /dev/null 2>&1; then
        echo "docker_ready"
        exit 0
    fi
    sleep 1
done
echo "docker_timeout"
exit 1
`

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "wsl", "-d", "Ubuntu", "-u", "root", "bash", "-c", startScript)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to start Docker in WSL: %w", err)
	}

	result := strings.TrimSpace(string(output))
	if result == "docker_timeout" {
		return fmt.Errorf("timeout waiting for Docker to start in WSL")
	}

	return nil
}

// GetWSLErrorSuggestion returns a helpful suggestion based on the WSL error
func GetWSLErrorSuggestion(exitCode int, command string) string {
	switch exitCode {
	case WSLExitCodeDistroNotFound, -1:
		return "The Ubuntu WSL distribution is not accessible. Try:\n" +
			"  1. Run 'wsl --list --verbose' to check distribution status\n" +
			"  2. Run 'wsl --terminate Ubuntu' followed by 'wsl -d Ubuntu' to restart it\n" +
			"  3. If Ubuntu is not installed, run 'wsl --install -d Ubuntu'"
	case WSLExitCodeGenericError:
		if strings.Contains(command, "wslpath") {
			return "The wslpath command failed. The path may not exist or may not be accessible from WSL."
		}
		return "WSL command failed. Check that the Ubuntu distribution is properly configured."
	default:
		return "WSL command failed unexpectedly. Check WSL status with 'wsl --status'"
	}
}

// CommandExecutor provides an abstraction layer for executing external commands
// This interface allows for dependency injection and testing without running real commands
type CommandExecutor interface {
	Execute(ctx context.Context, name string, args ...string) (*CommandResult, error)
	ExecuteWithOptions(ctx context.Context, options ExecuteOptions) (*CommandResult, error)
}

// CommandResult represents the result of a command execution
type CommandResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
}

// Output returns combined stdout and stderr for backward compatibility
func (r *CommandResult) Output() string {
	if r.Stderr != "" {
		return r.Stdout + "\n" + r.Stderr
	}
	return r.Stdout
}

// ExecuteOptions provides fine-grained control over command execution
type ExecuteOptions struct {
	Command string
	Args    []string
	Dir     string            // Working directory
	Env     map[string]string // Environment variables
	Timeout time.Duration     // Execution timeout
	Stdin   []byte            // Data piped to the process stdin (e.g. `helm -f -`); nil = no stdin
}

// RealCommandExecutor implements CommandExecutor using actual system commands
type RealCommandExecutor struct {
	dryRun  bool
	verbose bool
}

// NewRealCommandExecutor creates a new real command executor
func NewRealCommandExecutor(dryRun, verbose bool) CommandExecutor {
	return &RealCommandExecutor{
		dryRun:  dryRun,
		verbose: verbose,
	}
}

// Execute implements CommandExecutor.Execute
func (e *RealCommandExecutor) Execute(ctx context.Context, name string, args ...string) (*CommandResult, error) {
	options := ExecuteOptions{
		Command: name,
		Args:    args,
		Dir:     "",
		Env:     nil,
		Timeout: 0,
	}
	return e.ExecuteWithOptions(ctx, options)
}

// ExecuteWithOptions implements CommandExecutor.ExecuteWithOptions
func (e *RealCommandExecutor) ExecuteWithOptions(ctx context.Context, options ExecuteOptions) (*CommandResult, error) {
	start := time.Now()

	command, args := options.Command, options.Args

	// Build full command string for logging (use original command for readability)
	fullCommand := options.Command
	if len(options.Args) > 0 {
		fullCommand += " " + strings.Join(options.Args, " ")
	}

	result := &CommandResult{
		Stdout: "",
		Stderr: "",
	}

	// Handle dry-run mode. The "Would run:" line prints UNCONDITIONALLY (not
	// only under --verbose): showing what would execute is dry-run's entire
	// purpose — without it a dry-run was indistinguishable from a real
	// successful run (audit B6/T2-9). pterm.Info honors --silent.
	if e.dryRun {
		pterm.Info.Printf("Would run: %s\n", redact.Redact(fullCommand))
		result.Duration = time.Since(start)
		return result, nil
	}

	// Create the command with wrapped command/args
	cmd := exec.CommandContext(ctx, command, args...) // #nosec G204 -- central executor: explicit argv (no shell); callers pass internal tool names + controlled args

	// Set working directory if specified
	if options.Dir != "" {
		cmd.Dir = options.Dir
	}

	// Set environment variables if specified
	if len(options.Env) > 0 {
		// Start with current environment and add custom variables
		cmd.Env = append(os.Environ(), e.buildEnvStrings(options.Env)...)
	}

	// Apply timeout if specified
	if options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
		cmd = exec.CommandContext(ctx, command, args...) // #nosec G204 -- central executor: explicit argv (no shell); callers pass internal tool names + controlled args

		// Reapply directory and env since we recreated the command
		if options.Dir != "" {
			cmd.Dir = options.Dir
		}
		if len(options.Env) > 0 {
			// Start with current environment and add custom variables
			cmd.Env = append(os.Environ(), e.buildEnvStrings(options.Env)...)
		}
	}

	// Pipe stdin data if provided (e.g. helm reading values from `-f -`).
	// Set once here so it survives the timeout-driven command recreation above.
	if len(options.Stdin) > 0 {
		cmd.Stdin = bytes.NewReader(options.Stdin)
	}

	// Execute the command
	stdout, err := cmd.Output()
	result.Duration = time.Since(start)
	result.Stdout = string(stdout)

	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			result.ExitCode = exitError.ExitCode()
			// Redact at the population chokepoint: callers embed Stderr in
			// user-facing errors even in non-verbose mode (e.g. the helm
			// manager's "Helm output: %s"), and a child process can echo a
			// token back. Control-flow substring checks downstream match
			// generic phrases, never secret values, so redaction is safe here.
			result.Stderr = redact.Redact(string(exitError.Stderr))
		} else {
			result.ExitCode = -1
		}

		// Log error in verbose mode. pterm.Debug, not fmt.Printf: the latter
		// writes straight to stdout, so these diagnostics survived --silent and
		// corrupted machine-readable output (`cluster list -o json`).
		if e.verbose {
			pterm.Debug.Printfln("Command failed: %s (exit code: %d)", redact.Redact(fullCommand), result.ExitCode)
			if result.Stderr != "" {
				pterm.Debug.Printfln("Stderr: %s", redact.Redact(result.Stderr))
			}
		}

		// Check for WSL-specific errors on Windows.
		// Error fields are REDACTED at construction: unlike the verbose prints
		// above, these errors reach user-facing output through the error handler
		// even in non-verbose mode, so a secret in argv or echoed back on stderr
		// (e.g. a URL-embedded token) must never survive into them (audit B5).
		if runtime.GOOS == "windows" && (command == "wsl" || options.Command == "helm" || options.Command == "k3d") {
			// For WSL commands, stderr is often redirected to stdout via 2>&1
			// Use stdout as error output if stderr is empty
			errorOutput := result.Stderr
			if errorOutput == "" && result.Stdout != "" {
				errorOutput = result.Stdout
			}

			// Detect WSL distribution not found error
			if result.ExitCode == WSLExitCodeDistroNotFound || result.ExitCode == -1 {
				wslErr := &WSLError{
					Operation:  fmt.Sprintf("executing %s", options.Command),
					ExitCode:   result.ExitCode,
					Stderr:     redact.Redact(errorOutput),
					Suggestion: GetWSLErrorSuggestion(result.ExitCode, redact.Redact(fullCommand)),
				}
				return result, wslErr
			}
			// Detect other WSL errors
			if command == "wsl" && result.ExitCode != 0 {
				wslErr := &WSLError{
					Operation:  fmt.Sprintf("executing %s via WSL", options.Command),
					ExitCode:   result.ExitCode,
					Stderr:     redact.Redact(errorOutput),
					Suggestion: GetWSLErrorSuggestion(result.ExitCode, redact.Redact(fullCommand)),
				}
				return result, wslErr
			}
		}

		// result.Stderr was already redacted where it was populated.
		return result, &CommandError{
			Command:  redact.Redact(fullCommand),
			ExitCode: result.ExitCode,
			Stderr:   result.Stderr,
			cause:    err,
		}
	}

	result.ExitCode = 0

	// Log success in verbose mode (see above: pterm.Debug, not fmt.Printf).
	if e.verbose {
		pterm.Debug.Printfln("Command completed successfully: %s (took %v)", redact.Redact(fullCommand), result.Duration)
	}

	return result, nil
}

// buildEnvStrings converts environment map to string slice
func (e *RealCommandExecutor) buildEnvStrings(env map[string]string) []string {
	var envStrings []string
	for key, value := range env {
		envStrings = append(envStrings, fmt.Sprintf("%s=%s", key, value))
	}
	return envStrings
}
