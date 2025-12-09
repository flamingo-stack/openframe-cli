package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
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
	Operation   string
	ExitCode    int
	Stderr      string
	Suggestion  string
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
			fmt.Printf("Would run: %s\n", fullCommand)
		}
		result.Duration = time.Since(start)
		return result, nil
	}

	// Create the command with wrapped command/args
	cmd := exec.CommandContext(ctx, command, args...)
	
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
		cmd = exec.CommandContext(ctx, command, args...)

		// Reapply directory and env since we recreated the command
		if options.Dir != "" {
			cmd.Dir = options.Dir
		}
		if len(options.Env) > 0 {
			// Start with current environment and add custom variables
			cmd.Env = append(os.Environ(), e.buildEnvStrings(options.Env)...)
		}
	}
	
	// Log command execution in verbose mode
	if e.verbose {
		fmt.Printf("Executing: %s\n", fullCommand)
	}
	
	// Execute the command
	stdout, err := cmd.Output()
	result.Duration = time.Since(start)
	result.Stdout = string(stdout)
	
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
			result.Stderr = string(exitError.Stderr)
		} else {
			result.ExitCode = -1
		}

		// Log error in verbose mode
		if e.verbose {
			fmt.Printf("Command failed: %s (exit code: %d)\n", fullCommand, result.ExitCode)
			if result.Stderr != "" {
				fmt.Printf("Stderr: %s\n", result.Stderr)
			}
		}

		// Check for WSL-specific errors on Windows
		if runtime.GOOS == "windows" && (command == "wsl" || options.Command == "kubectl" || options.Command == "helm" || options.Command == "k3d") {
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
		fmt.Printf("Command completed successfully: %s (took %v)\n", fullCommand, result.Duration)
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

// shellEscape escapes an argument for safe use when passing to WSL
// WSL passes arguments directly to the target command, so we only need to handle
// characters that could confuse the WSL argument parser itself (spaces, quotes, backslashes)
// We should NOT escape characters like {}, $, etc. that are part of command syntax (e.g., jsonpath)
func shellEscape(arg string) string {
	// Only escape if the argument contains spaces, quotes, or backslashes
	// These are the characters that WSL argument parsing cares about
	needsEscape := false
	for _, ch := range arg {
		if ch == ' ' || ch == '"' || ch == '\'' || ch == '\\' {
			needsEscape = true
			break
		}
	}

	if !needsEscape {
		return arg
	}

	// For arguments with spaces or quotes, wrap in double quotes
	// and escape internal double quotes and backslashes
	escaped := strings.ReplaceAll(arg, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
	return "\"" + escaped + "\""
}

// wrapCommandForWindows wraps kubectl, helm, and k3d commands to run directly in WSL2
// This avoids issues with batch file wrappers not preserving special characters
// and ensures all Kubernetes tools run in the same environment
func (e *RealCommandExecutor) wrapCommandForWindows(command string, args []string) (string, []string) {
	// Only wrap on Windows
	if runtime.GOOS != "windows" {
		return command, args
	}

	// Only wrap kubectl, helm, and k3d commands
	if command != "kubectl" && command != "helm" && command != "k3d" {
		return command, args
	}

	// Determine WSL user - try to detect from environment or use default
	wslUser := os.Getenv("WSL_USER")
	if wslUser == "" {
		// Default to "runner" for CI environments, but could be configured
		wslUser = "runner"
	}

	// Escape arguments that contain special characters for shell interpretation
	escapedArgs := make([]string, len(args))
	for i, arg := range args {
		escapedArgs[i] = shellEscape(arg)
	}

	// For k3d, we need Docker access which requires elevated permissions
	// Use 'sudo -E' to run k3d with necessary permissions while preserving environment
	// The -E flag preserves environment variables like KUBECONFIG
	if command == "k3d" {
		// Build command with sudo -E prefix
		newArgs := make([]string, 0, len(escapedArgs)+6)
		newArgs = append(newArgs, "-d", "Ubuntu", "-u", wslUser, "sudo", "-E", command)
		newArgs = append(newArgs, escapedArgs...)
		return "wsl", newArgs
	}

	// For helm, run directly via WSL with environment variables set using 'env'
	// This is more reliable than using a wrapper script, as passing arguments through
	// 'bash /script.sh arg1 arg2' can fail in some WSL/CI environments
	// The 'env' command sets environment variables and then runs helm with all arguments
	// Windows path conversion is already done in the calling code (manager.go)
	// We use bash -c to first create the helm directories, then run helm with proper env vars
	if command == "helm" {
		// Build the helm command with all arguments properly quoted
		helmCmd := "helm"
		for _, arg := range escapedArgs {
			helmCmd += " " + arg
		}

		// Create directories and run helm in a single bash command
		// This ensures directories exist before helm tries to use them
		// We use 2>&1 to redirect stderr to stdout so error messages are captured
		// through the WSL/bash chain (otherwise stderr from helm gets lost)
		//
		// CRITICAL FIX: Dynamically rewrite the kubeconfig server address before running helm
		// The kubeconfig may have 127.0.0.1 or 0.0.0.0 as the server address, but from inside
		// WSL Ubuntu, we need to use the WSL internal IP (eth0) to reach the k3d cluster
		// running in Docker Desktop. We do this inline to ensure it's always correct.
		//
		// The sed command rewrites server addresses like:
		//   server: https://127.0.0.1:6550 -> server: https://172.x.x.x:6550
		//   server: https://0.0.0.0:6550   -> server: https://172.x.x.x:6550
		bashScript := "mkdir -p /tmp/helm/cache /tmp/helm/config /tmp/helm/data && " +
			"export HELM_CACHE_HOME=/tmp/helm/cache && " +
			"export HELM_CONFIG_HOME=/tmp/helm/config && " +
			"export HELM_DATA_HOME=/tmp/helm/data && " +
			"export HOME=/home/" + wslUser + " && " +
			// Get WSL internal IP and rewrite kubeconfig if needed
			"WSL_IP=$(ip -4 addr show eth0 2>/dev/null | grep -oP 'inet \\K[0-9.]+' | head -1) && " +
			"if [ -n \"$WSL_IP\" ] && [ -f ~/.kube/config ]; then " +
			"sed -i \"s|server: https://127\\.0\\.0\\.1:|server: https://$WSL_IP:|g; " +
			"s|server: https://0\\.0\\.0\\.0:|server: https://$WSL_IP:|g\" ~/.kube/config 2>/dev/null || true; " +
			"fi && " +
			helmCmd + " 2>&1"

		newArgs := []string{"-d", "Ubuntu", "-u", wslUser, "bash", "-c", bashScript}
		return "wsl", newArgs
	}

	// For kubectl, run via bash to ensure kubeconfig has correct WSL IP
	// Same fix as helm - dynamically rewrite 127.0.0.1/0.0.0.0 to WSL internal IP
	kubectlCmd := "kubectl"
	for _, arg := range escapedArgs {
		kubectlCmd += " " + arg
	}

	bashScript := "export HOME=/home/" + wslUser + " && " +
		// Get WSL internal IP and rewrite kubeconfig if needed
		"WSL_IP=$(ip -4 addr show eth0 2>/dev/null | grep -oP 'inet \\K[0-9.]+' | head -1) && " +
		"if [ -n \"$WSL_IP\" ] && [ -f ~/.kube/config ]; then " +
		"sed -i \"s|server: https://127\\.0\\.0\\.1:|server: https://$WSL_IP:|g; " +
		"s|server: https://0\\.0\\.0\\.0:|server: https://$WSL_IP:|g\" ~/.kube/config 2>/dev/null || true; " +
		"fi && " +
		kubectlCmd

	newArgs := []string{"-d", "Ubuntu", "-u", wslUser, "bash", "-c", bashScript}
	return "wsl", newArgs
}