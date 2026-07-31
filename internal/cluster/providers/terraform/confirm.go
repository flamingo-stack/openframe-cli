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
	RenderPlanDiff(summary)
	pterm.DefaultBasicText.Println()

	confirmed, err := sharedUI.ConfirmActionInteractive(
		fmt.Sprintf("Apply this plan (%d to add, %d to change, %d to destroy)?",
			summary.Add, summary.Change, summary.Destroy), true)
	return err == nil && confirmed
}

// RenderPlanDiff prints the plan as a git-style colored diff, grouped by
// action — creates, then updates, then destroys/replaces — with a colored
// counts line and an explicit destructive-changes warning. The flat
// plan-order list with bare "+/~/-" prefixes hid a destroy in the middle of
// forty creates.
func RenderPlanDiff(summary PlanSummary) {
	styles := map[string]func(string) string{
		"+":   func(s string) string { return pterm.FgGreen.Sprint(s) },
		"~":   func(s string) string { return pterm.FgYellow.Sprint(s) },
		"-":   func(s string) string { return pterm.FgRed.Sprint(s) },
		"-/+": func(s string) string { return pterm.FgMagenta.Sprint(s) },
	}
	for _, action := range []string{"+", "~", "-", "-/+"} {
		style := styles[action]
		for _, change := range summary.Changes {
			if change.Action != action {
				continue
			}
			pterm.DefaultBasicText.Printf("  %s\n", style(fmt.Sprintf("%-3s %s", change.Action, change.Address)))
		}
	}
	pterm.DefaultBasicText.Printf("\nPlan: %s, %s, %s\n",
		pterm.FgGreen.Sprintf("%d to add", summary.Add),
		pterm.FgYellow.Sprintf("%d to change", summary.Change),
		pterm.FgRed.Sprintf("%d to destroy", summary.Destroy))
	if summary.Destroy > 0 {
		pterm.Warning.Println("This plan DESTROYS resources — review the red lines above before applying")
	}
}
