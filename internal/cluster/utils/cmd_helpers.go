package utils

// NOTE: this package must never import tests/... — the shipped binary used to
// compile testutil (and testify) in via the integration-test helpers that
// lived here (audit B7).

import (
	stderrors "errors"
	"sync"

	"github.com/flamingo-stack/openframe-cli/internal/cluster"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/shared/errors"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/spf13/cobra"
)

// Global flag container for all cluster commands
var globalFlags *cluster.FlagContainer
var globalFlagsMutex sync.Mutex

// InitGlobalFlags initializes the global flag container if not already set
func InitGlobalFlags() {
	globalFlagsMutex.Lock()
	defer globalFlagsMutex.Unlock()

	if globalFlags == nil {
		globalFlags = cluster.NewFlagContainer()
	}
}

// GetCommandService creates a command service for business logic operations
func GetCommandService() *cluster.ClusterService {
	// Use injected executor if available (for testing)
	if globalFlags != nil && globalFlags.Executor != nil {
		return cluster.NewClusterService(globalFlags.Executor)
	}

	// Create real executor with current flags
	dryRun := globalFlags != nil && globalFlags.Global != nil && globalFlags.Global.DryRun
	verbose := globalFlags != nil && globalFlags.Global != nil && globalFlags.Global.Verbose
	exec := executor.NewRealCommandExecutor(dryRun, verbose)
	return cluster.NewClusterService(exec)
}

// WrapCommandWithCommonSetup wraps a command function with common CLI setup and error handling
func WrapCommandWithCommonSetup(runFunc func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Logo is now shown in PersistentPreRunE, not here

		// Execute the command
		err := runFunc(cmd, args)
		if err == nil {
			return nil
		}

		// Already displayed inside runFunc (e.g. via HandleGlobalError): keep the
		// sentinel so main.go exits non-zero without re-printing.
		var handled *errors.AlreadyHandledError
		if stderrors.As(err, &handled) {
			return err
		}

		// Otherwise display it once here and mark it handled. Every command error
		// now yields a non-zero exit code (previously only string-matched errors
		// did, so genuine failures could exit 0) — the root has SilenceErrors, so
		// cobra will not re-print, and main.go skips the sentinel.
		verbose := globalFlags != nil && globalFlags.Global != nil && globalFlags.Global.Verbose
		errors.NewErrorHandler(verbose).HandleError(err)
		return &errors.AlreadyHandledError{OriginalError: err}
	}
}

// SyncGlobalFlags synchronizes global flags to all command flags
func SyncGlobalFlags() {
	if globalFlags != nil && globalFlags.Global != nil {
		globalFlags.SyncGlobalFlags()
	}
}

// ValidateGlobalFlags validates global flags
func ValidateGlobalFlags() error {
	if globalFlags != nil && globalFlags.Global != nil {
		return models.ValidateGlobalFlags(globalFlags.Global)
	}
	return nil
}

// GetGlobalFlags returns the current global flags instance
func GetGlobalFlags() *cluster.FlagContainer {
	InitGlobalFlags()
	return globalFlags
}

// SetTestExecutor injects a mock executor into the global flag container. It
// exists ONLY for cmd-layer tests (paired with ResetGlobalFlags in cleanup);
// production never calls it.
func SetTestExecutor(exec executor.CommandExecutor) {
	InitGlobalFlags()
	globalFlags.Executor = exec
}

// ResetGlobalFlags clears the global flag container (test cleanup).
func ResetGlobalFlags() {
	globalFlagsMutex.Lock()
	defer globalFlagsMutex.Unlock()
	globalFlags = nil
}
