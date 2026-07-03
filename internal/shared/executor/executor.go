package executor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/shared/redact"
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

// WakeUpWSL sends a simple command to WSL to ensure it's responsive
// This is useful before critical operations as WSL can become unresponsive when idle
// Returns nil if WSL is responsive, error otherwise
func WakeUpWSL() error {
	if runtime.GOOS != "windows" {
		return nil
	}

	// Quick ping to WSL - just echo something
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "wsl", "-d", "Ubuntu", "echo", "ping")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("WSL wake-up failed: %w", err)
	}

	if strings.TrimSpace(string(output)) != "ping" {
		return fmt.Errorf("WSL wake-up returned unexpected output: %s", string(output))
	}

	return nil
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

	// Wrap command for Windows if needed (kubectl/helm via WSL)
	command, args := e.wrapCommandForWindows(options.Command, options.Args)

	// Build full command string for logging (use original command for readability)
	fullCommand := options.Command
	if len(options.Args) > 0 {
		fullCommand += " " + strings.Join(options.Args, " ")
	}

	result := &CommandResult{
		Stdout: "",
		Stderr: "",
	}

	// Handle dry-run mode
	if e.dryRun {
		if e.verbose {
			fmt.Printf("Would run: %s\n", redact.Redact(fullCommand))
		}
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
			result.Stderr = string(exitError.Stderr)
		} else {
			result.ExitCode = -1
		}

		// Log error in verbose mode
		if e.verbose {
			fmt.Printf("Command failed: %s (exit code: %d)\n", redact.Redact(fullCommand), result.ExitCode)
			if result.Stderr != "" {
				fmt.Printf("Stderr: %s\n", redact.Redact(result.Stderr))
			}
		}

		// Check for WSL-specific errors on Windows
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
					Stderr:     errorOutput,
					Suggestion: GetWSLErrorSuggestion(result.ExitCode, fullCommand),
				}
				return result, wslErr
			}
			// Detect other WSL errors
			if command == "wsl" && result.ExitCode != 0 {
				wslErr := &WSLError{
					Operation:  fmt.Sprintf("executing %s via WSL", options.Command),
					ExitCode:   result.ExitCode,
					Stderr:     errorOutput,
					Suggestion: GetWSLErrorSuggestion(result.ExitCode, fullCommand),
				}
				return result, wslErr
			}
		}

		return result, fmt.Errorf("command failed: %s (exit code: %d): %w", fullCommand, result.ExitCode, err)
	}

	result.ExitCode = 0

	// Log success in verbose mode
	if e.verbose {
		fmt.Printf("Command completed successfully: %s (took %v)\n", redact.Redact(fullCommand), result.Duration)
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

// wslUserPattern allows only safe POSIX-ish usernames. Anything else falls back
// to the default, so a hostile WSL_USER value cannot influence command construction.
var wslUserPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]*$`)

// sanitizeWSLUser returns a safe WSL username, defaulting to "runner".
func sanitizeWSLUser(user string) string {
	if user == "" || !wslUserPattern.MatchString(user) {
		return "runner"
	}
	return user
}

// filterFlag removes "--flag value" and "--flag=value" occurrences from args.
func filterFlag(args []string, flag string) []string {
	out := make([]string, 0, len(args))
	skip := false
	for _, a := range args {
		if skip {
			skip = false
			continue
		}
		if a == flag {
			skip = true
			continue
		}
		if strings.HasPrefix(a, flag+"=") {
			continue
		}
		out = append(out, a)
	}
	return out
}

// The WSL wrapper scripts below are CONSTANT: no argument is ever interpolated
// into them. The target tool and its arguments are passed as positional
// parameters after the script (… bash -c <script> bash <HOME> <tool> <args…>)
// and reach the program via "$@"/exec, so the shell never re-parses them.
//
// This is the structural fix for the previous design, which concatenated
// (incompletely escaped) arguments into the script string and allowed command
// injection via $(...) / ; / | embedded in arguments. $1 carries HOME (also a
// positional, so even an exotic home path is inert).
const helmWSLScript = `set -e
export HOME="$1"; shift
mkdir -p /tmp/helm/cache /tmp/helm/config /tmp/helm/data
export HELM_CACHE_HOME=/tmp/helm/cache
export HELM_CONFIG_HOME=/tmp/helm/config
export HELM_DATA_HOME=/tmp/helm/data
if [ -f "$HOME/.kube/config" ]; then cp "$HOME/.kube/config" "$HOME/.kube/config.openframe.bak" 2>/dev/null || true; sed "s|server: https://0\.0\.0\.0:|server: https://127.0.0.1:|g" "$HOME/.kube/config" > "$HOME/.kube/config.openframe.tmp" 2>/dev/null && mv "$HOME/.kube/config.openframe.tmp" "$HOME/.kube/config" || true; fi
"$@" 2>&1`

// buildWSLCommand wraps helm and k3d so they run inside WSL2 Ubuntu.
// It is a pure function (no OS calls) so it can be unit-tested on any platform.
//
// Tool arguments are ALWAYS passed as discrete argv elements — never spliced
// into a shell string — so values containing shell metacharacters are inert.
func buildWSLCommand(command string, args []string, wslUser string) (string, []string) {
	user := sanitizeWSLUser(wslUser)
	home := "/home/" + user
	base := []string{"-d", "Ubuntu", "-u", user}

	switch command {
	case "k3d":
		// k3d needs Docker access; run via sudo -E (preserves env like KUBECONFIG).
		out := append([]string{}, base...)
		out = append(out, "sudo", "-E", "k3d")
		return "wsl", append(out, args...)
	case "helm":
		filtered := filterFlag(args, "--kube-context")
		out := append([]string{}, base...)
		out = append(out, "bash", "-c", helmWSLScript, "bash", home, "helm")
		return "wsl", append(out, filtered...)
	default:
		return command, args
	}
}

// wrapCommandForWindows wraps helm and k3d commands to run directly in WSL2.
// On non-Windows platforms it is a no-op. The actual command construction lives in
// buildWSLCommand (a pure, testable seam).
func (e *RealCommandExecutor) wrapCommandForWindows(command string, args []string) (string, []string) {
	// Only wrap on Windows
	if runtime.GOOS != "windows" {
		return command, args
	}

	// Only wrap helm and k3d commands
	if command != "helm" && command != "k3d" {
		return command, args
	}

	return buildWSLCommand(command, args, os.Getenv("WSL_USER"))
}
