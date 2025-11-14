package executor

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// MockCommandExecutor implements CommandExecutor for testing
// It simulates command execution without actually running external commands
type MockCommandExecutor struct {
	mu             sync.RWMutex              // Protects all fields for concurrent access
	commands       []string                  // Log of executed commands
	responses      map[string]*CommandResult // Predefined responses for specific commands
	defaultResult  *CommandResult            // Default response when no specific response is set
	shouldFail     bool                      // Whether to simulate failures
	failMessage    string                    // Error message for simulated failures
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

	// Build full command string
	fullCommand := options.Command
	if len(options.Args) > 0 {
		fullCommand += " " + strings.Join(options.Args, " ")
	}

	// Lock for writing to shared state
	m.mu.Lock()
	// Log the command
	m.commands = append(m.commands, fullCommand)

	// Check if we should simulate failure
	shouldFail := m.shouldFail
	failMessage := m.failMessage

	// Look for specific response for this command
	var matchedResponse *CommandResult
	for pattern, response := range m.responses {
		if strings.Contains(fullCommand, pattern) {
			matchedResponse = response
			break
		}
	}

	// Get default result
	defaultResult := m.defaultResult
	m.mu.Unlock()

	// Process result outside of lock
	if shouldFail {
		result := &CommandResult{
			ExitCode: 1,
			Stderr:   failMessage,
			Duration: time.Since(start),
		}
		return result, fmt.Errorf("mock command failure: %s", failMessage)
	}

	if matchedResponse != nil {
		result := *matchedResponse // Copy the response
		result.Duration = time.Since(start)
		if result.ExitCode != 0 {
			return &result, fmt.Errorf("mock command failed with exit code %d", result.ExitCode)
		}
		return &result, nil
	}

	// Return default result
	result := *defaultResult // Copy the default result
	result.Duration = time.Since(start)

	return &result, nil
}

// GetExecutedCommands returns the list of commands that were executed
func (m *MockCommandExecutor) GetExecutedCommands() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]string(nil), m.commands...) // Return a copy
}

// GetCommandCount returns the number of commands executed
func (m *MockCommandExecutor) GetCommandCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.commands)
}

// Reset clears all executed commands and responses
func (m *MockCommandExecutor) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.commands = make([]string, 0)
	m.responses = make(map[string]*CommandResult)
	m.shouldFail = false
	m.failMessage = ""
}

// WasCommandExecuted checks if a command containing the given pattern was executed
func (m *MockCommandExecutor) WasCommandExecuted(pattern string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, cmd := range m.commands {
		if strings.Contains(cmd, pattern) {
			return true
		}
	}
	return false
}

// GetLastCommand returns the last executed command, or empty string if none
func (m *MockCommandExecutor) GetLastCommand() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.commands) == 0 {
		return ""
	}
	return m.commands[len(m.commands)-1]
}