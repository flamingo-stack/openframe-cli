package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flamingo-stack/openframe-cli/internal/chart/models"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// Repository handles git operations for chart repositories using go-git — no
// external `git` binary, and private-repo tokens are supplied in memory only
// (never in the URL, argv, or an on-disk credentials file).
type Repository struct{}

// NewRepository creates a new git repository handler.
func NewRepository() *Repository {
	return &Repository{}
}

// CloneChartRepository clones a GitHub repository to a temporary directory with
// a shallow, single-branch checkout.
func (r *Repository) CloneChartRepository(ctx context.Context, config *models.AppOfAppsConfig) (*CloneResult, error) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "openframe-chart-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	// Separate any embedded credential from the URL so the token is passed only
	// via the in-memory auth method (audit I1) — never in the URL, argv, or a
	// credentials file on disk.
	auth := extractGitAuth(config.GitHubRepo)

	_, err = gogit.PlainCloneContext(ctx, tempDir, false, &gogit.CloneOptions{
		URL:           auth.cleanURL,
		Auth:          auth.buildAuth(),
		ReferenceName: plumbing.NewBranchReferenceName(config.GitHubBranch),
		SingleBranch:  true,
		Depth:         1,
		Tags:          gogit.NoTags,
	})
	if err != nil {
		r.Cleanup(tempDir)
		if isBranchNotFound(err) {
			return nil, fmt.Errorf("branch '%s' does not exist in repository. Please check if the branch name is correct or use 'main' branch", config.GitHubBranch)
		}
		// Mask the token in any surfaced output as defense-in-depth.
		return nil, fmt.Errorf("failed to clone repository: %s", maskToken(err.Error(), auth.token))
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

// isBranchNotFound reports whether err is go-git's "requested branch does not
// exist on the remote" condition, so callers can surface a friendly message.
func isBranchNotFound(err error) bool {
	var noRef gogit.NoMatchingRefSpecError
	if errors.As(err, &noRef) || errors.Is(err, plumbing.ErrReferenceNotFound) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "couldn't find remote ref") ||
		strings.Contains(msg, "reference not found")
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
