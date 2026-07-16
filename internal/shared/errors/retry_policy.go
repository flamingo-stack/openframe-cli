package errors

import (
	"context"
	stderrors "errors"
	"io"
	"math"
	"math/rand"
	"net"
	"strings"
	"syscall"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
		// Substrings for tools we shell out to, whose Go error is only an exit
		// code. These are the strings the tools ACTUALLY print. "network timeout"
		// used to be listed here and is a phrase Go never emits: the standard
		// library says "i/o timeout" (verified), so that entry matched nothing.
		RetryableErrs: map[string]bool{
			"i/o timeout":           true,
			"connection refused":    true,
			"connection reset":      true,
			"tls handshake timeout": true,
			"unexpected eof":        true,
			"temporary failure":     true,
			"resource not ready":    true,
			"cluster not ready":     true,
			"service unavailable":   true,
			"too many requests":     true,
		},
	}
}

// ShouldRetry determines if an error should be retried
func (p *ExponentialBackoffPolicy) ShouldRetry(err error, attempt int) bool {
	if attempt >= p.MaxAttempts {
		return false
	}
	if err == nil {
		return false
	}

	// Check if it's a recoverable error. errors.As unwraps %w chains, so a
	// recoverable error stays recognized after being wrapped.
	var recoverableErr RecoverableError
	if stderrors.As(err, &recoverableErr) {
		return recoverableErr.IsRecoverable()
	}

	// Structural classification first — it does not depend on the wording of
	// somebody else's error string.
	if retry, decided := classifyTransient(err); decided {
		return retry
	}

	// Fallback for shelled-out tools (helm, k3d, docker): their failure is an
	// exit code, and the reason only exists as text on stderr. Since
	// executor.CommandError carries that stderr in its Error() string, matching
	// substrings here actually reaches the tool's own message.
	errMsg := err.Error()
	for retryablePattern := range p.RetryableErrs {
		if contains(errMsg, retryablePattern) {
			return true
		}
	}

	return false
}

// classifyTransient answers "is this error transient?" structurally, returning
// decided=false when it has no opinion.
//
// Order matters, and was verified against the standard library: an
// http.Client timeout satisfies BOTH net.Error.Timeout() and
// errors.Is(err, context.DeadlineExceeded), while a cancelled request is a
// net.Error whose Timeout() is false. So the timeout check must come first, or
// every network timeout would be misfiled as "the operation is over".
func classifyTransient(err error) (retry, decided bool) {
	// A timed-out network operation is the canonical retryable failure.
	var netErr net.Error
	if stderrors.As(err, &netErr) && netErr.Timeout() {
		return true, true
	}

	// The user pressed Ctrl-C, or the overall budget is spent. Retrying is
	// pointless and, for cancellation, wrong.
	if stderrors.Is(err, context.Canceled) || stderrors.Is(err, context.DeadlineExceeded) {
		return false, true
	}

	// Connection-level failures against an API server that is still starting.
	for _, errno := range []error{
		syscall.ECONNREFUSED, syscall.ECONNRESET, syscall.EPIPE,
		syscall.EHOSTUNREACH, syscall.ENETUNREACH,
	} {
		if stderrors.Is(err, errno) {
			return true, true
		}
	}
	if stderrors.Is(err, io.ErrUnexpectedEOF) {
		return true, true
	}

	// Kubernetes API server backpressure and optimistic-concurrency conflicts.
	if apierrors.IsConflict(err) || apierrors.IsTooManyRequests(err) ||
		apierrors.IsServerTimeout(err) || apierrors.IsServiceUnavailable(err) ||
		apierrors.IsTimeout(err) {
		return true, true
	}

	return false, false
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

// InstallationRetryPolicy for installation operations.
//
// The substrings are the fallback for helm/kubectl failures, which reach us as
// "exit status 1" plus the tool's stderr. Structural classification
// (classifyTransient) handles everything the Go type system can see.
//
// "tiller not ready" used to be listed here: Tiller was removed in Helm 3
// (2019) and this CLI drives Helm 3/4, so it could never match. Likewise
// "rate limited" — GitHub and the Kubernetes API server both say "too many
// requests".
func InstallationRetryPolicy() RetryPolicy {
	policy := NewExponentialBackoffPolicy(3, 10*time.Second)
	policy.MaxDelay = 5 * time.Minute
	policy.RetryableErrs = map[string]bool{
		// helm's own transient conditions
		"another operation (install/upgrade/rollback) is in progress": true,
		"the server is currently unable to handle the request":        true,
		"etcdserver: request timed out":                               true,
		// generic transport failures visible only as text from a child process
		"i/o timeout":           true,
		"connection refused":    true,
		"connection reset":      true,
		"tls handshake timeout": true,
		"unexpected eof":        true,
		"too many requests":     true,
		"temporary failure":     true,
	}
	return policy
}

// contains reports whether str contains substr, case-insensitively. (The prior
// hand-rolled implementation was case-sensitive despite its name and duplicated
// strings.Contains with extra, redundant branches.)
func contains(str, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}
