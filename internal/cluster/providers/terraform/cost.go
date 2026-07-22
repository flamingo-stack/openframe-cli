package terraform

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	sharedexec "github.com/flamingo-stack/openframe-cli/internal/shared/executor"
)

// Cost estimation is strictly optional: when the infracost CLI
// (https://www.infracost.io) is installed and configured, the dry-run plan
// preview shows a monthly estimate; otherwise the CLI prints no figures at
// all — only a pointer to the provider's pricing page. No pricing is ever
// hardcoded.

// InfracostAvailable reports whether the infracost binary is on PATH.
func InfracostAvailable() bool {
	_, err := exec.LookPath("infracost")
	return err == nil
}

// infracostBreakdown is the subset of `infracost breakdown --format json`
// this integration reads.
type infracostBreakdown struct {
	TotalMonthlyCost string `json:"totalMonthlyCost"`
	Currency         string `json:"currency"`
}

// EstimateMonthlyCost runs infracost against a terraform plan JSON and
// returns a human-readable monthly estimate (e.g. "142.53 USD"). Any failure
// (missing API key, network, unparseable output) is returned as an error —
// callers fall back to the abstract pricing hint, never to made-up numbers.
func EstimateMonthlyCost(ctx context.Context, execer sharedexec.CommandExecutor, planJSON []byte) (string, error) {
	if len(planJSON) == 0 {
		return "", fmt.Errorf("no plan JSON to estimate from")
	}
	tmp, err := os.MkdirTemp("", "openframe-infracost-*")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.RemoveAll(tmp) }()
	planPath := filepath.Join(tmp, "plan.json")
	if err := os.WriteFile(planPath, planJSON, 0o600); err != nil {
		return "", err
	}

	result, err := execer.Execute(ctx, "infracost", "breakdown",
		"--path", planPath, "--format", "json", "--no-color")
	if err != nil {
		return "", fmt.Errorf("infracost breakdown failed: %w", err)
	}
	var breakdown infracostBreakdown
	if err := json.Unmarshal([]byte(result.Stdout), &breakdown); err != nil {
		return "", fmt.Errorf("unparseable infracost output: %w", err)
	}
	cost := strings.TrimSpace(breakdown.TotalMonthlyCost)
	if cost == "" {
		return "", fmt.Errorf("infracost reported no total monthly cost")
	}
	currency := breakdown.Currency
	if currency == "" {
		currency = "USD"
	}
	return cost + " " + currency, nil
}
