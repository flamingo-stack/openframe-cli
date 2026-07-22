package terraform

import (
	"fmt"

	sharedUI "github.com/flamingo-stack/openframe-cli/internal/shared/ui"
	"github.com/pterm/pterm"
)

// ConfirmApplyInteractive is the default pre-apply gate for cluster creation
// — the interactive `terraform apply` shape: show the full plan, then ask.
// Non-interactive sessions auto-approve (the previous behavior of every
// scripted/CI create), and a plan with no changes needs no question.
func ConfirmApplyInteractive(summary PlanSummary) bool {
	if !summary.HasChanges() {
		return true
	}
	if sharedUI.IsNonInteractive() {
		return true
	}

	pterm.DefaultBasicText.Println()
	for _, change := range summary.Changes {
		pterm.DefaultBasicText.Printf("  %-3s %s\n", change.Action, change.Address)
	}
	pterm.DefaultBasicText.Printf("\nPlan: %d to add, %d to change, %d to destroy\n\n", summary.Add, summary.Change, summary.Destroy)

	total := summary.Add + summary.Change + summary.Destroy
	confirmed, err := sharedUI.ConfirmActionInteractive(
		fmt.Sprintf("Apply this plan (%d resources)?", total), true)
	return err == nil && confirmed
}
