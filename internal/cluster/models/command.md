<!-- source-hash: a277c0f4ffd444fe898e67c7193a4888 -->
This file defines core interfaces for implementing a command-line interface structure using the Cobra framework, providing contracts for command flags, execution, and cluster operations.

## Key Components

- **CommandFlags**: Interface for accessing global command flags
- **CommandExecutor**: Interface providing access to command execution context
- **ClusterCommand**: Main interface defining the contract for all cluster commands with methods for:
  - Command creation and configuration
  - Execution logic
  - Flag validation and setup

## Usage Example

```go
// Implementing a cluster command
type MyClusterCommand struct{}

func (c *MyClusterCommand) GetCommand(flags CommandFlags) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "deploy",
        Short: "Deploy to cluster",
    }
    c.SetupFlags(cmd, flags)
    return cmd
}

func (c *MyClusterCommand) Execute(cmd *cobra.Command, args []string, flags CommandFlags, executor CommandExecutor) error {
    // Access global flags
    globalFlags := flags.GetGlobal()
    
    // Use executor for command execution
    cmdExec := executor.GetExecutor()
    return cmdExec.Run("kubectl", "apply", "-f", "config.yaml")
}

func (c *MyClusterCommand) ValidateFlags(flags CommandFlags) error {
    // Validate command-specific flags
    return nil
}

func (c *MyClusterCommand) SetupFlags(cmd *cobra.Command, flags CommandFlags) {
    cmd.Flags().StringP("namespace", "n", "default", "Kubernetes namespace")
}
```