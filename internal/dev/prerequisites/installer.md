<!-- source-hash: 6556141f21c61dbd49bf083ddd5c6a01 -->
This file provides a comprehensive tool installer system for managing development prerequisites like telepresence, jq, and skaffold. It handles automatic installation, user confirmation, and fallback to manual installation instructions.

## Key Components

- **Installer struct**: Main installer that manages prerequisite checking and installation
- **ToolInstaller interface**: Defines contract for tool-specific installers with `IsInstalled()`, `GetInstallHelp()`, and `Install()` methods
- **CheckAndInstall()**: Primary method that checks all prerequisites and offers automatic installation
- **CheckSpecificTools()**: Validates only specified tools are installed
- **CheckAndInstallSpecificTools()**: Checks specific tools with installation option
- **CheckInterceptPrerequisites()**: Specialized check for telepresence and jq tools
- **CheckScaffoldPrerequisites()**: Specialized check for skaffold tool

## Usage Example

```go
// Basic usage - check and install all prerequisites
installer := NewInstaller()
if err := installer.CheckAndInstall(); err != nil {
    log.Fatal(err)
}

// Check specific tools only
if err := installer.CheckSpecificTools([]string{"telepresence", "jq"}); err != nil {
    log.Fatal(err)
}

// Check and optionally install specific tools
if err := installer.CheckAndInstallSpecificTools([]string{"skaffold"}); err != nil {
    log.Fatal(err)
}

// Specialized prerequisite checks
if err := CheckInterceptPrerequisites(); err != nil {
    log.Fatal(err)
}
```