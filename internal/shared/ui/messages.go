package ui

import (
	"fmt"

	"github.com/pterm/pterm"
)

// ShowNoResourcesMessage displays a friendly message when no resources are available
func ShowNoResourcesMessage(resourceType, operation, createCommand, listCommand string) {
	pterm.Warning.Printf("No %s found for %s operation\n", resourceType, operation)
	pterm.DefaultBasicText.Println()

	boxContent := fmt.Sprintf(
		"No %s are currently available.\n\n"+
			"To get started:\n"+
			"  • Create a new %s: %s\n"+
			"  • List existing %s: %s\n\n"+
			"Need help? Try: %s",
		resourceType,
		resourceType,
		pterm.Green(createCommand),
		resourceType,
		pterm.Cyan(listCommand),
		pterm.Gray("--help"),
	)

	pterm.DefaultBox.
		WithTitle(fmt.Sprintf(" No %s Available ", resourceType)).
		WithTitleTopCenter().
		Println(boxContent)
	pterm.DefaultBasicText.Println()
}

// ShowOperationError displays a friendly error message with troubleshooting tips
func ShowOperationError(operation, resourceName string, err error, troubleshootingTips []TroubleshootingTip) {
	pterm.Error.Printf("Operation '%s' failed for %s\n", operation, pterm.Cyan(resourceName))
	pterm.DefaultBasicText.Printf("Error details: %s\n\n", pterm.Red(err.Error()))

	if len(troubleshootingTips) > 0 {
		// Show helpful suggestions
		tableData := pterm.TableData{}

		for i, tip := range troubleshootingTips {
			tableData = append(tableData, []string{
				fmt.Sprintf("%d.", i+1),
				pterm.Gray(tip.Description) + " " + pterm.Cyan(tip.Command),
			})
		}

		pterm.Info.Println("Troubleshooting Tips:")
		if err := pterm.DefaultTable.WithData(tableData).Render(); err != nil {
			pterm.DefaultBasicText.Printf("Troubleshooting:\n")
			for i, tip := range troubleshootingTips {
				pterm.DefaultBasicText.Printf("  %d. %s: %s\n", i+1, tip.Description, pterm.Cyan(tip.Command))
			}
		}
	}
	pterm.DefaultBasicText.Println()
}

// TroubleshootingTip represents a troubleshooting suggestion
type TroubleshootingTip struct {
	Description string
	Command     string
}
