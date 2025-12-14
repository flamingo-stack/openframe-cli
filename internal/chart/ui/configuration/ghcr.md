<!-- source-hash: 4fffd00c3c6e5cb0cdd902adc1895742 -->
Provides interactive configuration functionality for GitHub Container Registry (GHCR) credentials within a chart configuration wizard. The function handles both new credential setup and updating existing credentials with a user-friendly selection interface.

## Key Components

- **configureGHCRCredentials()** - Main method that orchestrates GHCR credential configuration
  - Detects existing credentials from chart configuration
  - Presents interactive options to keep or update credentials
  - Collects username, password/token, and email through masked input prompts
  - Returns validated credential strings or error

## Usage Example

```go
// Initialize configuration wizard
wizard := &ConfigurationWizard{}
config := &types.ChartConfiguration{
    ExistingValues: map[string]interface{}{
        "registry": map[string]interface{}{
            "ghcr": map[string]interface{}{
                "username": "existing-user",
                "email": "user@example.com",
            },
        },
    },
}

// Configure GHCR credentials
username, password, email, err := wizard.configureGHCRCredentials(config)
if err != nil {
    log.Fatal("Failed to configure GHCR credentials:", err)
}

fmt.Printf("Configured GHCR user: %s, email: %s\n", username, email)
```

The function automatically detects existing credentials and offers options to keep them (requiring only password re-entry) or update all credential fields interactively.