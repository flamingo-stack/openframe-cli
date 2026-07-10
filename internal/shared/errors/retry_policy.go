package errors

import (
	"context"
	stderrors "errors"
	"math"
	"math/rand"
	"strings"
	"time"
)

// RetryPolicy defines retry behavior for recoverable errors
type RetryPolicy interface {
	ShouldRetry(err error, attempt int) bool
	GetDelay(attempt int) time.Duration
	GetMaxAttempts() int
}

// RecoverableError interface for errors that can be retried
type RecoverableError interface {
	IsRecoverable() bool
	GetRetryAfter() time.Duration
}

// ExponentialBackoffPolicy implements exponential backoff retry policy
type ExponentialBackoffPolicy struct {
	MaxAttempts   int
	BaseDelay     time.Duration
	MaxDelay      time.Duration
	Multiplier    float64
	Jitter        bool
	RetryableErrs map[string]bool
}

// NewExponentialBackoffPolicy creates a new exponential backoff policy
func NewExponentialBackoffPolicy(maxAttempts int, baseDelay time.Duration) *ExponentialBackoffPolicy {
	return &ExponentialBackoffPolicy{
		MaxAttempts: maxAttempts,
		BaseDelay:   baseDelay,
		MaxDelay:    5 * time.Minute,
		Multiplier:  2.0,
		Jitter:      true,
		RetryableErrs: map[string]bool{
			"network timeout":     true,
			"connection refused":  true,
			"temporary failure":   true,
			"resource not ready":  true,
			"cluster not ready":   true,
			"service unavailable": true,
		},
	}
}

// ShouldRetry determines if an error should be retried
func (p *ExponentialBackoffPolicy) ShouldRetry(err error, attempt int) bool {
	if attempt >= p.MaxAttempts {
		return false
	}

	// Check if it's a recoverable error. errors.As unwraps %w chains, so a
	// recoverable error stays recognized after being wrapped.
	var recoverableErr RecoverableError
	if stderrors.As(err, &recoverableErr) {
		return recoverableErr.IsRecoverable()
	}

	// Check if error message indicates it's retryable
	errMsg := err.Error()
	for retryablePattern := range p.RetryableErrs {
		if contains(errMsg, retryablePattern) {
			return true
		}
	}

	return false
}

// GetDelay calculates the delay for the next retry attempt
func (p *ExponentialBackoffPolicy) GetDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return p.BaseDelay
	}

	delay := time.Duration(float64(p.BaseDelay) * math.Pow(p.Multiplier, float64(attempt-1)))

	if delay > p.MaxDelay {
		delay = p.MaxDelay
	}

	// Add non-negative jitter to prevent thundering herd. Jitter is additive in
	// [0, 10%] so it never shortens the intended backoff (the previous
	// [-10%, +10%] form could reduce the delay below the computed value).
	if p.Jitter {
		jitter := time.Duration(float64(delay) * 0.1 * rand.Float64()) //nolint:gosec // jitter, not security-sensitive
		delay += jitter
		if delay > p.MaxDelay {
			delay = p.MaxDelay
		}
	}

	return delay
}

// GetMaxAttempts returns the maximum number of retry attempts
func (p *ExponentialBackoffPolicy) GetMaxAttempts() int {
	return p.MaxAttempts
}

// RetryExecutor handles retry logic with policies
type RetryExecutor struct {
	policy  RetryPolicy
	onRetry func(err error, attempt int, delay time.Duration)
}

// NewRetryExecutor creates a new retry executor with the given policy
func NewRetryExecutor(policy RetryPolicy) *RetryExecutor {
	return &RetryExecutor{
		policy: policy,
	}
}

// Execute executes a function with retry logic
func (r *RetryExecutor) Execute(ctx context.Context, operation func() error) error {
	var lastErr error

	for attempt := 0; attempt < r.policy.GetMaxAttempts(); attempt++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute the operation
		err := operation()
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if we should retry
		if !r.policy.ShouldRetry(err, attempt) {
			break
		}

		// This is our last attempt
		if attempt == r.policy.GetMaxAttempts()-1 {
			break
		}

		// Calculate delay
		delay := r.policy.GetDelay(attempt + 1)

		// Call retry callback if set
		if r.onRetry != nil {
			r.onRetry(err, attempt+1, delay)
		}

		// Wait for the delay period or context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return lastErr
}

// Predefined retry policies for common scenarios

// InstallationRetryPolicy for installation operations
func InstallationRetryPolicy() RetryPolicy {
	policy := NewExponentialBackoffPolicy(3, 10*time.Second)
	policy.MaxDelay = 5 * time.Minute
	policy.RetryableErrs = map[string]bool{
		"helm not ready":    true,
		"tiller not ready":  true,
		"resource conflict": true,
		"temporary failure": true,
		"rate limited":      true,
	}
	return policy
}

// contains reports whether str contains substr, case-insensitively. (The prior
// hand-rolled implementation was case-sensitive despite its name and duplicated
// strings.Contains with extra, redundant branches.)
func contains(str, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}
