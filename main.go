package main

import (
	stderrors "errors"
	"fmt"
	"os"

	"github.com/flamingo-stack/openframe-cli/cmd"
	sharederrors "github.com/flamingo-stack/openframe-cli/internal/shared/errors"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
)

func main() {
	if err := cmd.Execute(); err != nil {
		// Errors already shown to the user (via HandleGlobalError / the command
		// error handler) carry the AlreadyHandledError sentinel — exit non-zero
		// without re-printing. Everything else is printed here.
		var handled *sharederrors.AlreadyHandledError
		if !stderrors.As(err, &handled) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(exitCode(err))
	}
}

// exitCode preserves a failed external command's exit code (exit-code fidelity
// for automation) when it is a valid Unix code; otherwise it is a generic 1.
func exitCode(err error) int {
	var ce *executor.CommandError
	if stderrors.As(err, &ce) && ce.ExitCode > 0 && ce.ExitCode < 256 {
		return ce.ExitCode
	}
	return 1
}
