<!-- source-hash: 6e56d694689e1f4e2c6acae195a113c4 -->
Provides interactive credential prompting functionality for various authentication scenarios including GitHub, Docker, and generic services with comprehensive validation and user-friendly UI feedback.

## Key Components

- **CredentialsPrompter** - Main service for handling credential collection
- **GitHubCredentials** - Username and token structure for GitHub authentication
- **DockerCredentials** - Registry, username, and password for Docker registries
- **GenericCredentials** - Basic username/password credentials
- **APIKeyCredentials** - API key and secret pair for API-based authentication
- **CredentialsOptions** - Configuration options for customizing prompt behavior

## Usage Example

```go
// Create credentials prompter
prompter := NewCredentialsPrompter()

// Prompt for GitHub credentials
githubCreds, err := prompter.PromptForGitHubCredentials("https://github.com/private/repo")
if err != nil {
    log.Fatal(err)
}

// Prompt for Docker registry credentials
dockerCreds, err := prompter.PromptForDockerCredentials("registry.example.com")
if err != nil {
    log.Fatal(err)
}

// Prompt with custom options
opts := CredentialsOptions{
    AllowEmpty:      false,
    DefaultUsername: "admin",
    CustomMessage:   "Please enter your service credentials",
}
genericCreds, err := prompter.PromptWithOptions("MyService", opts)
if err != nil {
    log.Fatal(err)
}

// Validate existing credentials
err = prompter.ValidateCredentials(username, password)
if err != nil {
    log.Printf("Invalid credentials: %v", err)
}
```