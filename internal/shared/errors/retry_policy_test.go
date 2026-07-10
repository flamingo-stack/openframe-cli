package errors

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// nonRecoverable implements RecoverableError and reports itself non-retryable.
type nonRecoverable struct{ msg string }

func (e nonRecoverable) Error() string                { return e.msg }
func (e nonRecoverable) IsRecoverable() bool          { return false }
func (e nonRecoverable) GetRetryAfter() time.Duration { return 0 }

// recoverable implements RecoverableError and reports itself retryable.
type recoverable struct{ msg string }

func (e recoverable) Error() string                { return e.msg }
func (e recoverable) IsRecoverable() bool          { return true }
func (e recoverable) GetRetryAfter() time.Duration { return 0 }

// TestShouldRetry_WrappedRecoverableIsUnwrapped proves the fix: a recoverable
// error wrapped with %w (message carrying NO retryable substring) is still
// retried — only errors.As, not a bare type assertion, recognizes it.
func TestShouldRetry_WrappedRecoverableIsUnwrapped(t *testing.T) {
	p := NewExponentialBackoffPolicy(5, time.Millisecond)
	wrapped := fmt.Errorf("install step failed: %w", recoverable{msg: "boom"})
	assert.True(t, p.ShouldRetry(wrapped, 0), "wrapped recoverable error must still retry")
	assert.False(t, p.ShouldRetry(fmt.Errorf("plain: %w", nonRecoverable{msg: "x"}), 0))
}

func TestShouldRetry_RespectsMaxAttempts(t *testing.T) {
	p := NewExponentialBackoffPolicy(3, time.Millisecond)
	err := errors.New("network timeout")
	assert.True(t, p.ShouldRetry(err, 0))
	assert.True(t, p.ShouldRetry(err, 2))
	assert.False(t, p.ShouldRetry(err, 3), "attempt == MaxAttempts must not retry")
	assert.False(t, p.ShouldRetry(err, 4))
}

func TestShouldRetry_NonRecoverableFailsFast(t *testing.T) {
	p := NewExponentialBackoffPolicy(5, time.Millisecond)
	// Even though the message contains a retryable substring, an error that
	// explicitly declares itself non-recoverable must never be retried.
	err := nonRecoverable{msg: "network timeout but do not retry me"}
	assert.False(t, p.ShouldRetry(err, 0))
}

func TestShouldRetry_RetryableSubstringCaseInsensitive(t *testing.T) {
	p := NewExponentialBackoffPolicy(5, time.Millisecond)
	assert.True(t, p.ShouldRetry(errors.New("Cluster Not Ready yet"), 0),
		"matching must be case-insensitive (audit: contains() was case-sensitive)")
	assert.False(t, p.ShouldRetry(errors.New("permission denied"), 0),
		"non-retryable message must not retry")
}

func TestGetDelay_JitterNeverNegativeOrBelowBase(t *testing.T) {
	p := NewExponentialBackoffPolicy(10, time.Second)
	p.MaxDelay = time.Hour
	for i := 0; i < 5000; i++ {
		for attempt := 1; attempt <= 6; attempt++ {
			base := time.Duration(float64(p.BaseDelay) * pow(p.Multiplier, attempt-1))
			if base > p.MaxDelay {
				base = p.MaxDelay
			}
			d := p.GetDelay(attempt)
			require.GreaterOrEqual(t, d, time.Duration(0), "delay must never be negative")
			require.GreaterOrEqual(t, d, base, "additive jitter must not shorten the backoff")
			require.LessOrEqual(t, d, p.MaxDelay+time.Duration(float64(p.MaxDelay)*0.1)+time.Second)
		}
	}
}

func TestGetDelay_CapsAtMaxDelay(t *testing.T) {
	p := NewExponentialBackoffPolicy(20, time.Second)
	p.MaxDelay = 10 * time.Second
	p.Jitter = false
	assert.Equal(t, p.MaxDelay, p.GetDelay(20), "large attempt must cap at MaxDelay")
}

func TestGetDelay_FirstAttemptIsBase(t *testing.T) {
	p := NewExponentialBackoffPolicy(3, 2*time.Second)
	p.Jitter = false
	assert.Equal(t, 2*time.Second, p.GetDelay(0))
	assert.Equal(t, 2*time.Second, p.GetDelay(1))
	assert.Equal(t, 4*time.Second, p.GetDelay(2))
}

func TestRetryExecutor_RetriesThenSucceeds(t *testing.T) {
	p := NewExponentialBackoffPolicy(5, time.Microsecond)
	exec := NewRetryExecutor(p)
	calls := 0
	err := exec.Execute(context.Background(), func() error {
		calls++
		if calls < 3 {
			return errors.New("cluster not ready")
		}
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 3, calls)
}

func TestRetryExecutor_StopsOnNonRetryable(t *testing.T) {
	p := NewExponentialBackoffPolicy(5, time.Microsecond)
	exec := NewRetryExecutor(p)
	calls := 0
	err := exec.Execute(context.Background(), func() error {
		calls++
		return errors.New("permission denied")
	})
	require.Error(t, err)
	assert.Equal(t, 1, calls, "non-retryable error must not be retried")
}

func TestRetryExecutor_StopsOnContextCancel(t *testing.T) {
	p := NewExponentialBackoffPolicy(10, 50*time.Millisecond)
	exec := NewRetryExecutor(p)
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()
	err := exec.Execute(ctx, func() error {
		calls++
		return errors.New("cluster not ready")
	})
	assert.ErrorIs(t, err, context.Canceled)
	assert.Less(t, calls, 10, "must stop early on context cancel")
}

// pow is a tiny float power helper to avoid importing math in the test.
func pow(base float64, exp int) float64 {
	r := 1.0
	for i := 0; i < exp; i++ {
		r *= base
	}
	return r
}
