<!-- source-hash: 660f17eca83ef5007a3ce05df1adb779 -->
This file provides namespace management functionality for Telepresence, allowing services to query the current namespace and switch between namespaces while maintaining connections.

## Key Components

- **getCurrentNamespace()** - Retrieves the current Telepresence namespace using `telepresence status` and `jq` parsing, defaults to "default" on failure
- **switchNamespace()** - Switches Telepresence from current to target namespace by disconnecting and reconnecting with the new namespace parameter

## Usage Example

```go
// Get current namespace
ctx := context.Background()
service := &Service{executor: executor, verbose: true}

currentNS, err := service.getCurrentNamespace(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Current namespace: %s\n", currentNS)

// Switch to a different namespace
targetNS := "production"
err = service.switchNamespace(ctx, currentNS, targetNS)
if err != nil {
    log.Fatalf("Failed to switch namespace: %v", err)
}
```

The namespace switching preserves the traffic manager connection by using disconnect/reconnect rather than quit/connect, making it more efficient for namespace changes during intercept operations.