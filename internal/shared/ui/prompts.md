<!-- source-hash: 22ebb78a5892c4696fdada3d21fe6b1c -->
Provides interactive user interface utilities for terminal-based prompts, confirmations, and selections. Built on top of promptui and pterm libraries for polished CLI experiences.

## Key Components

**Confirmation Functions:**
- `ConfirmActionInteractive` - Interactive confirmation with pterm styling
- `ConfirmDeletion` - Specialized deletion confirmation prompt
- `ConfirmAction` - Single-character confirmation (Y/n) with terminal raw mode

**Selection Functions:**
- `SelectFromList` - Basic item selection from a list
- `SelectFromListWithSearch` - List selection with search/filter capability
- `SelectFromListWithCustomTemplates` - Customizable selection styling

**Input Functions:**
- `GetInput` - Text input with validation
- `GetMultiChoice` - Multiple item selection
- `HandleResourceSelection` - Resource selection from args or interactive list

**Validation Helpers:**
- `ValidateNonEmpty` - Ensures non-empty input
- `ValidateIntRange` - Validates integer within range

## Usage Example

```go
// Simple confirmation
confirmed, err := ui.ConfirmAction("Deploy to production?")
if err != nil || !confirmed {
    return
}

// Interactive list selection
items := []string{"dev", "staging", "prod"}
_, env, err := ui.SelectFromList("Select environment", items)

// Input with validation
name, err := ui.GetInput("Project name", "", ui.ValidateNonEmpty("Project name"))

// Resource selection (from args or interactive)
resourceName, err := ui.HandleResourceSelection(args, availableResources, "Select resource")
```