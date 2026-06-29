package executor

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// RecordedCommand captures a single execution with its argv kept as discrete
// elements (not joined into a string). Security spec tests rely on this to
// distinguish "passed $(x) as one literal arg" from "built a shell string",
// which a flattened log cannot tell apart.
type RecordedCommand struct {
	Name string
	Args []string
	Env  map[string]string
}

// String renders the command as a single line for human-readable assertions.
func (rc RecordedCommand) String() string {
	if len(rc.Args) == 0 {
		return rc.Name
	}
	return rc.Name + " " + strings.Join(rc.Args, " ")
}

// MockCommandExecutor implements CommandExecutor for testing
// It simulates command execution without actually running external commands.
// All shared state is guarded by mu so the mock is safe under `go test -race`.
type MockCommandExecutor struct {
	mu            sync.Mutex
	commands      []string                  // Log of executed commands (flattened, legacy)
	recorded      []RecordedCommand         // Structured log (name + discrete args + env)
	responses     map[string]*CommandResult // Predefined responses for specific commands
	defaultResult *CommandResult            // Default response when no specific response is set
	shouldFail    bool                      // Whether to simulate failures
	failMessage   string                    // Error message for simulated failures
}

// NewMockCommandExecutor creates a new mock command executor
func NewMockCommandExecutor() *MockCommandExecutor {
	return &MockCommandExecutor{
		commands:  make([]string, 0),
		responses: make(map[string]*CommandResult),
		defaultResult: &CommandResult{
			ExitCode: 0,
			Stdout:   "mock output",
			Duration: 100 * time.Millisecond,
		},
	}
}

// SetShouldFail configures the mock to simulate command failures
func (m *MockCommandExecutor) SetShouldFail(fail bool, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail = fail
	m.failMessage = message
}

// SetResponse sets a specific response for a command pattern
func (m *MockCommandExecutor) SetResponse(commandPattern string, result *CommandResult) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses[commandPattern] = result
}

// SetDefaultResult sets the default response for unmatched commands
func (m *MockCommandExecutor) SetDefaultResult(result *CommandResult) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultResult = result
}

// Execute implements CommandExecutor.Execute
func (m *MockCommandExecutor) Execute(ctx context.Context, name string, args ...string) (*CommandResult, error) {
	options := ExecuteOptions{
		Command: name,
		Args:    args,
	}
	return m.ExecuteWithOptions(ctx, options)
}

// ExecuteWithOptions implements CommandExecutor.ExecuteWithOptions
func (m *MockCommandExecutor) ExecuteWithOptions(ctx context.Context, options ExecuteOptions) (*CommandResult, error) {
	start := time.Now()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Build full command string (flattened, legacy log)
	fullCommand := options.Command
	if len(options.Args) > 0 {
		fullCommand += " " + strings.Join(options.Args, " ")
	}

	// Log the command both ways: flattened (legacy) and structured.
	m.commands = append(m.commands, fullCommand)
	m.recorded = append(m.recorded, RecordedCommand{
		Name: options.Command,
		Args: append([]string(nil), options.Args...), // defensive copy
		Env:  copyEnv(options.Env),
	})

	// Check if we should simulate failure
	if m.shouldFail {
		result := &CommandResult{
			ExitCode: 1,
			Stderr:   m.failMessage,
			Duration: time.Since(start),
		}
		return result, fmt.Errorf("mock command failure: %s", m.failMessage)
	}

	// Look for specific response for this command
	for pattern, response := range m.responses {
		if strings.Contains(fullCommand, pattern) {
			result := *response // Copy the response
			result.Duration = time.Since(start)
			if result.ExitCode != 0 {
				return &result, fmt.Errorf("mock command failed with exit code %d", result.ExitCode)
			}
			return &result, nil
		}
	}

	// Return default result
	result := *m.defaultResult // Copy the default result
	result.Duration = time.Since(start)

	return &result, nil
}

// copyEnv returns a defensive copy of an env map (nil-safe).
func copyEnv(env map[string]string) map[string]string {
	if env == nil {
		return nil
	}
	out := make(map[string]string, len(env))
	for k, v := range env {
		out[k] = v
	}
	return out
}

// GetExecutedCommands returns the list of commands that were executed
func (m *MockCommandExecutor) GetExecutedCommands() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string(nil), m.commands...) // Return a copy
}

// Commands returns a copy of the structured command log (name + discrete args + env).
// Prefer this over GetExecutedCommands for security assertions that must inspect
// individual argv elements rather than a flattened string.
func (m *MockCommandExecutor) Commands() []RecordedCommand {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]RecordedCommand, len(m.recorded))
	for i, rc := range m.recorded {
		out[i] = RecordedCommand{
			Name: rc.Name,
			Args: append([]string(nil), rc.Args...),
			Env:  copyEnv(rc.Env),
		}
	}
	return out
}

// GetCommandCount returns the number of commands executed
func (m *MockCommandExecutor) GetCommandCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.commands)
}

// Reset clears all executed commands and responses
func (m *MockCommandExecutor) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.commands = make([]string, 0)
	m.recorded = nil
	m.responses = make(map[string]*CommandResult)
	m.shouldFail = false
	m.failMessage = ""
}

// WasCommandExecuted checks if a command containing the given pattern was executed
func (m *MockCommandExecutor) WasCommandExecuted(pattern string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, cmd := range m.commands {
		if strings.Contains(cmd, pattern) {
			return true
		}
	}
	return false
}

// GetLastCommand returns the last executed command, or empty string if none
func (m *MockCommandExecutor) GetLastCommand() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.commands) == 0 {
		return ""
	}
	return m.commands[len(m.commands)-1]
}
