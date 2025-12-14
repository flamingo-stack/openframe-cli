<!-- source-hash: cc1feeb642f8578eaa85e6c6b37561fe -->
Provides comprehensive retry logic with configurable policies for handling transient errors in distributed systems. Implements exponential and linear backoff strategies with jitter, context cancellation support, and predefined policies for common scenarios.

## Key Components

- **RetryPolicy Interface**: Defines contract for retry behavior with delay calculation and attempt limits
- **ExponentialBackoffPolicy**: Implements exponential backoff with jitter and configurable error patterns
- **LinearBackoffPolicy**: Simple linear delay increment strategy
- **RetryExecutor**: Orchestrates retry logic with context cancellation and callback support
- **RecoverableError Interface**: Allows errors to specify their retry characteristics
- **Predefined Policies**: Ready-to-use configurations for network, resource, and installation operations
- **Callback Functions**: Built-in logging callbacks for different verbosity levels

## Usage Example

```go
// Basic retry with exponential backoff
executor := NewRetryExecutor(NetworkRetryPolicy()).
    WithRetryCallback(DefaultRetryCallback("API call"))

err := executor.Execute(ctx, func() error {
    return makeAPICall()
})

// Retry operation with result
result, err := executor.ExecuteWithResult(ctx, func() (interface{}, error) {
    data, err := fetchData()
    return data, err
})

// Custom retry policy
policy := NewExponentialBackoffPolicy(5, time.Second)
policy.MaxDelay = 30 * time.Second

executor = NewRetryExecutor(policy)
```