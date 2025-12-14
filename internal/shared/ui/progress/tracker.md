<!-- source-hash: ee63ccb2070e3492e1d03b1bdd2e2440 -->
Provides comprehensive progress tracking functionality for long-running operations with step-by-step execution monitoring, visual feedback, and cancellation support.

## Key Components

**Core Types:**
- `Tracker` - Main progress tracking orchestrator with spinner and progress bar integration
- `Step` - Individual operation step with timing, status, and error tracking
- `StepStatus` - Enumeration for step states (Pending, Running, Completed, Failed, Skipped)

**Primary Methods:**
- `NewTracker()` - Creates tracker with operation name and step definitions
- `Start()` - Begins operation tracking with visual spinner
- `StartStep()`, `CompleteStep()`, `FailStep()`, `SkipStep()` - Step lifecycle management
- `UpdateProgress()` - Updates progress bar based on step weights
- `Complete()`, `Fail()`, `Cancel()` - Operation termination methods
- `Context()` - Returns cancellation context for graceful shutdown

## Usage Example

```go
// Define operation steps
steps := []Step{
    {Name: "Initialize", Weight: 1.0},
    {Name: "Process Data", Weight: 3.0},
    {Name: "Generate Output", Weight: 1.0},
}

// Create and start tracker
tracker := NewTracker("Data Processing", steps)
tracker.Start()

// Execute steps
tracker.StartStep(0)
tracker.CompleteStep(0)

tracker.StartStep(1)
tracker.UpdateProgress(50.0) // 50% through current step
tracker.CompleteStep(1)

tracker.StartStep(2)
tracker.CompleteStep(2)

tracker.Complete() // Shows summary with timing statistics
```