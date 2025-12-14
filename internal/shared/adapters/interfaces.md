<!-- source-hash: 09c486a699d1bdb6139be0bd21e39ed2 -->
Defines core interfaces and base implementations for building Cobra command adapters with standardized flag handling and validation.

## Key Components

**Interfaces:**
- `CommandHandler` - Defines command execution interface
- `FlagsProvider` - Provides flag definitions and setup
- `CommandAdapter` - Combines command handling and flag management

**Base Implementation:**
- `BaseCommandAdapter` - Common functionality for command adapters with flag extraction and validation

**Support Types:**
- `FlagDefinition` - Command flag specification
- `CommandMetadata` - Command metadata container
- `ExampleBuilder` - Builds formatted command examples
- `PreRunEChain` - Chains multiple PreRunE functions
- `RequiredFlagError`/`InvalidFlagError` - Validation error types

## Usage Example

```go
// Create a custom command adapter
type MyCommandAdapter struct {
    *BaseCommandAdapter
}

func (a *MyCommandAdapter) Execute(cmd *cobra.Command, args []string) error {
    flags := a.ExtractFlags(cmd)
    if result := a.ValidateRequired(flags.GetAll()); !result.IsValid() {
        return result.Error()
    }
    // Command logic here
    return nil
}

func (a *MyCommandAdapter) GetFlagDefinitions() []FlagDefinition {
    return []FlagDefinition{
        {Name: "output", Shorthand: "o", DefaultValue: "json", Usage: "Output format"},
    }
}

// Build command examples
examples := NewExampleBuilder().
    Add("Output as JSON", "--output json").
    Add("Use shorthand", "-o yaml").
    Build("mycommand")
```