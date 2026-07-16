package services

import (
	"fmt"

	"github.com/flamingo-stack/openframe-cli/internal/chart/providers/argocd"
	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
)

// ValidateHelmValuesFile pre-flights the default openframe-helm-values.yaml in
// the current directory. Bootstrap runs it BEFORE cluster creation: without it
// a malformed `argocd:` override costs a full k3d cluster create (minutes)
// before the chart install rejects the same file (0.4.9 verification
// observation). The install workflow re-validates the temp values file it
// actually feeds to helm; this is the earliest, cheapest gate.
func ValidateHelmValuesFile() error {
	path := config.NewPathResolver().GetHelmValuesFile()
	if err := argocd.ValidateUserValuesFile(path); err != nil {
		return fmt.Errorf("helm values pre-flight failed: %w", err)
	}
	return nil
}
