package main

import (
	stderrors "errors"
	"fmt"
	"os"

	"github.com/flamingo-stack/openframe-cli/cmd"
	sharederrors "github.com/flamingo-stack/openframe-cli/internal/shared/errors"
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
		os.Exit(1)
	}
}
