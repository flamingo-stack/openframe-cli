package errors

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewChartError(t *testing.T) {
	cause := errors.New("test error")
	chartErr := NewChartError("installation", "ArgoCD", cause)

	assert.NotNil(t, chartErr)
	assert.Equal(t, "installation", chartErr.Operation)
	assert.Equal(t, "ArgoCD", chartErr.Component)
	assert.Equal(t, cause, chartErr.Cause)
	assert.False(t, chartErr.Recoverable)
	assert.Equal(t, time.Duration(0), chartErr.RetryAfter)
	assert.Empty(t, chartErr.ClusterName)
}

func TestNewRecoverableChartError(t *testing.T) {
	cause := errors.New("recoverable error")
	retryAfter := 30 * time.Second
	chartErr := NewRecoverableChartError("installation", "Helm", cause, retryAfter)

	assert.NotNil(t, chartErr)
	assert.Equal(t, "installation", chartErr.Operation)
	assert.Equal(t, "Helm", chartErr.Component)
	assert.Equal(t, cause, chartErr.Cause)
	assert.True(t, chartErr.Recoverable)
	assert.Equal(t, retryAfter, chartErr.RetryAfter)
	assert.True(t, chartErr.IsRecoverable())
	assert.Equal(t, retryAfter, chartErr.GetRetryAfter())
}

func TestChartError_Error(t *testing.T) {
	cause := errors.New("test error")
	chartErr := NewChartError("installation", "ArgoCD", cause)

	errorMsg := chartErr.Error()
	assert.Contains(t, errorMsg, "installation")
	assert.Contains(t, errorMsg, "ArgoCD")
	assert.Contains(t, errorMsg, "test error")
}

func TestChartError_ErrorWithCluster(t *testing.T) {
	cause := errors.New("test error")
	chartErr := NewChartError("installation", "ArgoCD", cause).WithCluster("test-cluster")

	errorMsg := chartErr.Error()
	assert.Contains(t, errorMsg, "installation")
	assert.Contains(t, errorMsg, "ArgoCD")
	assert.Contains(t, errorMsg, "test-cluster")
	assert.Contains(t, errorMsg, "test error")
}

func TestChartError_WithCluster(t *testing.T) {
	cause := errors.New("test error")
	chartErr := NewChartError("installation", "ArgoCD", cause)

	result := chartErr.WithCluster("test-cluster")

	assert.Equal(t, chartErr, result) // Should return same instance
	assert.Equal(t, "test-cluster", chartErr.ClusterName)
}

func TestChartError_WithRecovery(t *testing.T) {
	cause := errors.New("test error")
	chartErr := NewChartError("installation", "ArgoCD", cause)
	retryAfter := 45 * time.Second

	result := chartErr.WithRecovery(retryAfter)

	assert.Equal(t, chartErr, result) // Should return same instance
	assert.True(t, chartErr.Recoverable)
	assert.Equal(t, retryAfter, chartErr.RetryAfter)
}

func TestChartError_Unwrap(t *testing.T) {
	cause := errors.New("original error")
	chartErr := NewChartError("installation", "ArgoCD", cause)

	unwrapped := chartErr.Unwrap()
	assert.Equal(t, cause, unwrapped)
}

func TestNewValidationError(t *testing.T) {
	valErr := NewValidationError("github-repo", "", "URL is required")

	assert.NotNil(t, valErr)
	assert.Equal(t, "validation", valErr.Operation)
	assert.Equal(t, "configuration", valErr.Component)
	assert.Equal(t, "github-repo", valErr.Field)
	assert.Equal(t, "", valErr.Value)
	assert.Equal(t, "URL is required", valErr.Constraint)
}

func TestValidationError_Error(t *testing.T) {
	valErr := NewValidationError("github-repo", "invalid-url", "must be valid URL")

	errorMsg := valErr.Error()
	assert.Contains(t, errorMsg, "validation failed")
	assert.Contains(t, errorMsg, "github-repo")
	assert.Contains(t, errorMsg, "invalid-url")
	assert.Contains(t, errorMsg, "must be valid URL")
}

func TestWrapAsChartError(t *testing.T) {
	// Test wrapping regular error
	normalErr := errors.New("test error")
	wrapped := WrapAsChartError("installation", "helm", normalErr)

	assert.IsType(t, &ChartError{}, wrapped)
	assert.Equal(t, "installation", wrapped.Operation)
	assert.Equal(t, "helm", wrapped.Component)
	assert.Equal(t, normalErr, wrapped.Cause)

	// Test wrapping existing ChartError
	chartErr := NewChartError("existing", "component", normalErr)
	wrapped2 := WrapAsChartError("new", "new-component", chartErr)

	assert.Equal(t, chartErr, wrapped2) // Should return same instance
}
