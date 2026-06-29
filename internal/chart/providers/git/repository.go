package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flamingo-stack/openframe-cli/internal/chart/models"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
)

// Repository handles git operations for chart repositories
type Repository struct {
	executor executor.CommandExecutor
}

// NewRepository creates a new git repository handler
func NewRepository(exec executor.CommandExecutor) *Repository {
	return &Repository{
		executor: exec,
	}
}

// CloneChartRepository clones a GitHub repository to a temporary directory with depth 1
func (r *Repository) CloneChartRepository(ctx context.Context, config *models.AppOfAppsConfig) (*CloneResult, error) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "openframe-chart-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	// Separate any embedded credential from the URL so the token never reaches
	// the git command line (audit I1). For private repos it is supplied via a
	// 0600 credentials file and the `store` helper instead.
	auth := extractGitAuth(config.GitHubRepo)

	opts := executor.ExecuteOptions{Command: "git"}
	var args []string
	if line, ok := auth.credentialLine(); ok {
		credFile, cleanup, cerr := writeGitCredentials(line)
		if cerr != nil {
			r.Cleanup(tempDir)
			return nil, fmt.Errorf("failed to set up git credentials: %w", cerr)
		}
		defer cleanup()
		// Reset inherited helpers (-c credential.helper=) then use only our
		// file-based store helper. The token is in the file, not in argv.
		args = append(args,
			"-c", "credential.helper=",
			"-c", "credential.helper=store --file="+credFile,
		)
		opts.Env = map[string]string{"GIT_TERMINAL_PROMPT": "0"}
	}

	// Clone with depth 1 and optimizations for speed
	args = append(args,
		"clone",
		"--depth", "1",
		"--single-branch",
		"--no-tags",
		"--branch", config.GitHubBranch,
		auth.cleanURL,
		tempDir,
	)
	opts.Args = args

	result, err := r.executor.ExecuteWithOptions(ctx, opts)
	if err != nil {
		r.Cleanup(tempDir)
		// Mask the token in any surfaced output as defense-in-depth.
		errMsg := maskToken(err.Error(), auth.token)
		// Check for branch not found error
		if result != nil && result.Stderr != "" {
			stderr := maskToken(result.Stderr, auth.token)
			if strings.Contains(stderr, "Remote branch") && strings.Contains(stderr, "not found") {
				return nil, fmt.Errorf("branch '%s' does not exist in repository. Please check if the branch name is correct or use 'main' branch", config.GitHubBranch)
			}
			return nil, fmt.Errorf("failed to clone repository: %s\nGit output: %s", errMsg, stderr)
		}
		return nil, fmt.Errorf("failed to clone repository: %s", errMsg)
	}

	// Build the path to the chart within the cloned repository
	chartPath := filepath.Join(tempDir, config.ChartPath)

	// Verify the chart directory exists
	if _, err := os.Stat(chartPath); os.IsNotExist(err) {
		r.Cleanup(tempDir)
		return nil, fmt.Errorf("chart path '%s' does not exist in repository", config.ChartPath)
	}

	return &CloneResult{
		TempDir:   tempDir,
		ChartPath: chartPath,
	}, nil
}

// Cleanup removes the temporary directory
func (r *Repository) Cleanup(tempDir string) {
	if tempDir != "" {
		if err := os.RemoveAll(tempDir); err != nil {
			// Log the error but don't fail the operation
			// This is cleanup so we don't want to break the main flow
			fmt.Printf("Warning: failed to cleanup temporary directory %s: %v\n", tempDir, err)
		}
	}
}
