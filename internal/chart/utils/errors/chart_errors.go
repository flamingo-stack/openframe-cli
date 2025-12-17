package errors

import (
	"fmt"
	"strings"
	"time"
)

// ChartError represents chart-specific errors with enhanced context
type ChartError struct {
	Operation   string
	Component   string
	ClusterName string
	Cause       error
	Timestamp   time.Time
	Recoverable bool
	RetryAfter  time.Duration
}

// Error implements the error interface
func (e *ChartError) Error() string {
	if e.ClusterName != "" {
		return fmt.Sprintf("chart %s failed for %s on cluster %s: %v", 
			e.Operation, e.Component, e.ClusterName, e.Cause)
	}
	return fmt.Sprintf("chart %s failed for %s: %v", e.Operation, e.Component, e.Cause)
}

// Unwrap returns the underlying error
func (e *ChartError) Unwrap() error {
	return e.Cause
}

// IsRecoverable returns whether this error can be retried
func (e *ChartError) IsRecoverable() bool {
	return e.Recoverable
}

// GetRetryAfter returns the suggested retry delay
func (e *ChartError) GetRetryAfter() time.Duration {
	return e.RetryAfter
}

// NewChartError creates a new chart error
func NewChartError(operation, component string, cause error) *ChartError {
	return &ChartError{
		Operation: operation,
		Component: component,
		Cause:     cause,
		Timestamp: time.Now(),
	}
}

// NewRecoverableChartError creates a recoverable chart error
func NewRecoverableChartError(operation, component string, cause error, retryAfter time.Duration) *ChartError {
	return &ChartError{
		Operation:   operation,
		Component:   component,
		Cause:       cause,
		Timestamp:   time.Now(),
		Recoverable: true,
		RetryAfter:  retryAfter,
	}
}

// WithCluster adds cluster context to the error
func (e *ChartError) WithCluster(clusterName string) *ChartError {
	e.ClusterName = clusterName
	return e
}

// WithRecovery marks the error as recoverable with retry delay
func (e *ChartError) WithRecovery(retryAfter time.Duration) *ChartError {
	e.Recoverable = true
	e.RetryAfter = retryAfter
	return e
}

// Chart Error Types
var (
	ErrChartNotFound        = fmt.Errorf("chart not found")
	ErrChartAlreadyInstalled = fmt.Errorf("chart already installed")
	ErrInvalidConfiguration = fmt.Errorf("invalid configuration")
	ErrClusterNotReady      = fmt.Errorf("cluster not ready")
	ErrHelmNotAvailable     = fmt.Errorf("helm not available")
	ErrKubectlNotAvailable  = fmt.Errorf("kubectl not available")
	ErrInsufficientResources = fmt.Errorf("insufficient cluster resources")
	ErrNetworkTimeout       = fmt.Errorf("network timeout")
	ErrAuthenticationFailed = fmt.Errorf("authentication failed")
	ErrPermissionDenied     = fmt.Errorf("permission denied")
)

// InstallationError represents installation-specific errors
type InstallationError struct {
	*ChartError
	Phase       string
	StepsFailed []string
	Suggestions []string
}

// Error implements error interface for InstallationError
func (e *InstallationError) Error() string {
	baseError := e.ChartError.Error()
	if e.Phase != "" {
		return fmt.Sprintf("%s during phase '%s'", baseError, e.Phase)
	}
	return baseError
}

// GetTroubleshootingSteps returns suggested troubleshooting steps
func (e *InstallationError) GetTroubleshootingSteps() []string {
	steps := []string{
		"Check cluster connectivity: kubectl cluster-info",
		"Verify cluster resources: kubectl top nodes",
		"Check helm installation: helm version",
	}
	
	// Add error-specific steps
	steps = append(steps, e.Suggestions...)
	
	return steps
}

// NewInstallationError creates a new installation error
func NewInstallationError(component, phase string, cause error) *InstallationError {
	return &InstallationError{
		ChartError: NewChartError("installation", component, cause),
		Phase:      phase,
	}
}

// WithSuggestions adds troubleshooting suggestions
func (e *InstallationError) WithSuggestions(suggestions []string) *InstallationError {
	e.Suggestions = suggestions
	return e
}

// ValidationError represents validation-specific errors  
type ValidationError struct {
	*ChartError
	Field       string
	Value       string
	Constraint  string
}

// Error implements error interface for ValidationError
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation failed for field '%s': %s (value: '%s')", 
			e.Field, e.Constraint, e.Value)
	}
	return fmt.Sprintf("validation failed: %v", e.Cause)
}

// NewValidationError creates a new validation error
func NewValidationError(field, value, constraint string) *ValidationError {
	cause := fmt.Errorf("constraint violation: %s", constraint)
	return &ValidationError{
		ChartError: NewChartError("validation", "configuration", cause),
		Field:      field,
		Value:      value,
		Constraint: constraint,
	}
}

// ConfigurationError represents configuration-specific errors
type ConfigurationError struct {
	*ChartError
	ConfigFile  string
	Section     string
	MissingKeys []string
}

// Error implements error interface for ConfigurationError  
func (e *ConfigurationError) Error() string {
	if e.ConfigFile != "" {
		return fmt.Sprintf("configuration error in file '%s': %v", e.ConfigFile, e.Cause)
	}
	return fmt.Sprintf("configuration error: %v", e.Cause)
}

// GetMissingKeys returns list of missing configuration keys
func (e *ConfigurationError) GetMissingKeys() []string {
	return e.MissingKeys
}

// NewConfigurationError creates a new configuration error
func NewConfigurationError(configFile, section string, cause error) *ConfigurationError {
	return &ConfigurationError{
		ChartError: NewChartError("configuration", "validation", cause),
		ConfigFile: configFile,
		Section:    section,
	}
}

// WithMissingKeys adds missing keys information
func (e *ConfigurationError) WithMissingKeys(keys []string) *ConfigurationError {
	e.MissingKeys = keys
	return e
}

// Helper functions for common error patterns

// IsTimeout checks if an error is timeout-related
func IsTimeout(err error) bool {
	if chartErr, ok := err.(*ChartError); ok {
		return chartErr.Cause == ErrNetworkTimeout
	}
	return false
}

// IsRecoverable checks if an error is recoverable
func IsRecoverable(err error) bool {
	if chartErr, ok := err.(*ChartError); ok {
		return chartErr.IsRecoverable()
	}
	return false
}

// GetRetryDelay gets the retry delay for recoverable errors
func GetRetryDelay(err error) time.Duration {
	if chartErr, ok := err.(*ChartError); ok && chartErr.IsRecoverable() {
		return chartErr.GetRetryAfter()
	}
	return 0
}

// WrapAsChartError wraps a generic error as a chart error
func WrapAsChartError(operation, component string, err error) *ChartError {
	if chartErr, ok := err.(*ChartError); ok {
		return chartErr
	}
	return NewChartError(operation, component, err)
}

// SkippedInstallationError represents when installation is skipped (not an actual error)
type SkippedInstallationError struct {
	Component string
	Reason    string
}

// Error implements the error interface
func (e *SkippedInstallationError) Error() string {
	return fmt.Sprintf("%s installation skipped: %s", e.Component, e.Reason)
}

// IsSkipped returns true, indicating this is a skipped installation
func (e *SkippedInstallationError) IsSkipped() bool {
	return true
}

// CombineErrors combines multiple errors into a single error message
func CombineErrors(errors []error) error {
	if len(errors) == 0 {
		return nil
	}
	
	if len(errors) == 1 {
		return errors[0]
	}
	
	var messages []string
	for _, err := range errors {
		if err != nil {
			messages = append(messages, err.Error())
		}
	}
	
	return fmt.Errorf("multiple errors occurred: %v", messages)
}

// IsSkippedInstallation checks if an error is a skipped installation
func IsSkippedInstallation(err error) bool {
	_, ok := err.(*SkippedInstallationError)
	return ok
}

// RegistryDNSError represents errors where container registry DNS resolution fails
// This is common on Windows/WSL2 where Docker Hub DNS can be flaky
type RegistryDNSError struct {
	*InstallationError
	Registry string
}

// Error implements error interface
func (e *RegistryDNSError) Error() string {
	return fmt.Sprintf("registry DNS resolution failed for %s: %s", e.Registry, e.ChartError.Error())
}

// NewRegistryDNSError creates a new registry DNS error (always recoverable)
func NewRegistryDNSError(component, registry string, cause error) *RegistryDNSError {
	instErr := NewInstallationError(component, "helm-install", cause)
	instErr.ChartError.Recoverable = true
	instErr.ChartError.RetryAfter = 2 * time.Minute
	instErr.Suggestions = []string{
		"Check WSL2 DNS configuration in /etc/resolv.conf",
		"Ensure registry-1.docker.io is reachable from WSL: curl -I https://registry-1.docker.io/v2/",
		"Restart Docker daemon: sudo systemctl restart docker",
		"Retry 'openframe bootstrap' after the network stabilizes",
	}
	return &RegistryDNSError{
		InstallationError: instErr,
		Registry:          registry,
	}
}

// IsRegistryDNSError checks if an error is a registry DNS error
func IsRegistryDNSError(err error) bool {
	_, ok := err.(*RegistryDNSError)
	return ok
}

// IsHelmTimeoutWithRegistryDNS checks if an error indicates Helm timed out due to registry DNS issues
// This pattern is common on Windows/WSL2 where Docker Hub DNS can be flaky
func IsHelmTimeoutWithRegistryDNS(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()

	// Check for Helm timeout patterns
	isHelmTimeout := (strings.Contains(msg, "failed pre-install") ||
		strings.Contains(msg, "failed post-install") ||
		strings.Contains(msg, "timed out waiting")) &&
		strings.Contains(msg, "timed out waiting for the condition")

	// Check for registry/DNS patterns in the error message
	isRegistryDNS := strings.Contains(msg, "lookup registry-1.docker.io") ||
		strings.Contains(msg, "failed to pull image") ||
		strings.Contains(msg, "failed to resolve reference") ||
		strings.Contains(msg, "ErrImagePull") ||
		strings.Contains(msg, "ImagePullBackOff") ||
		(strings.Contains(msg, "dial tcp") && strings.Contains(msg, "i/o timeout"))

	return isHelmTimeout && isRegistryDNS
}

// IsHelmPreInstallTimeout checks if an error indicates Helm timed out during pre-install
// This is a broader check than IsHelmTimeoutWithRegistryDNS - it catches any pre-install timeout
// On Windows/WSL2, these timeouts are almost always caused by registry DNS issues
func IsHelmPreInstallTimeout(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()

	// Check for Helm pre-install timeout pattern
	// This catches cases where helm returns just the timeout without DNS details
	return strings.Contains(msg, "failed pre-install") &&
		strings.Contains(msg, "timed out waiting for the condition")
}

// ClassifyInstallError examines an error and returns a more specific error type if possible
// This is useful for detecting infrastructure issues that should be treated as recoverable
func ClassifyInstallError(component, clusterName string, err error) error {
	if err == nil {
		return nil
	}

	// Check for registry DNS issues (common on Windows/WSL2)
	if IsHelmTimeoutWithRegistryDNS(err) {
		regErr := NewRegistryDNSError(component, "registry-1.docker.io", err)
		regErr.ChartError.ClusterName = clusterName
		return regErr
	}

	// Default: return as a standard chart error
	return WrapAsChartError("installation", component, err).WithCluster(clusterName)
}