package errors

import (
	stderrors "errors"
	"fmt"
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

// Error reads as "<operation> failed for <component> [on cluster <name>]: <cause>",
// e.g. "waiting failed for ArgoCD applications on cluster dev: ...".
//
// The previous form prefixed every message with a bare "chart" — "chart waiting
// failed for ArgoCD applications" — which was ungrammatical and also wrong: the
// operation is not always on a chart (waiting is on applications).
func (e *ChartError) Error() string {
	if e.ClusterName != "" {
		return fmt.Sprintf("%s failed for %s on cluster %s: %v",
			e.Operation, e.Component, e.ClusterName, e.Cause)
	}
	return fmt.Sprintf("%s failed for %s: %v", e.Operation, e.Component, e.Cause)
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
	ErrChartNotFound         = fmt.Errorf("chart not found")
	ErrChartAlreadyInstalled = fmt.Errorf("chart already installed")
	ErrInvalidConfiguration  = fmt.Errorf("invalid configuration")
	ErrClusterNotReady       = fmt.Errorf("cluster not ready")
	ErrHelmNotAvailable      = fmt.Errorf("helm not available")
	ErrInsufficientResources = fmt.Errorf("insufficient cluster resources")
	ErrAuthenticationFailed  = fmt.Errorf("authentication failed")
	ErrPermissionDenied      = fmt.Errorf("permission denied")
)

// ValidationError represents validation-specific errors
type ValidationError struct {
	*ChartError
	Field      string
	Value      string
	Constraint string
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

// WrapAsChartError wraps a generic error as a chart error
func WrapAsChartError(operation, component string, err error) *ChartError {
	var chartErr *ChartError
	if stderrors.As(err, &chartErr) {
		return chartErr
	}
	return NewChartError(operation, component, err)
}
