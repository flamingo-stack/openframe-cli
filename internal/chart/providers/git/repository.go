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
	"github.com/pterm/pterm"
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
// a shallow, single-ref checkout. GitHubBranch carries a general git ref: it is
// tried first as a branch and, if no such branch exists, as a tag — so a release
// tag (e.g. "v1.2.3") works as well as a branch name.
func (r *Repository) CloneChartRepository(ctx context.Context, config *models.AppOfAppsConfig) (*CloneResult, error) {
	// Separate any embedded credential from the URL so the token is passed only
	// via the in-memory auth method (audit I1) — never in the URL, argv, or a
	// credentials file on disk.
	auth := extractGitAuth(config.GitHubRepo)

	// Try the ref as a branch first, then as a tag. A branch that is present
	// succeeds on the first attempt; a tag falls through the branch-not-found
	// path to the second. Any other error (auth, network) aborts immediately.
	var lastErr error
	for _, refName := range []plumbing.ReferenceName{
		plumbing.NewBranchReferenceName(config.GitHubBranch),
		plumbing.NewTagReferenceName(config.GitHubBranch),
	} {
		tempDir, err := os.MkdirTemp("", "openframe-chart-*")
		if err != nil {
			return nil, fmt.Errorf("failed to create temporary directory: %w", err)
		}

		_, err = gogit.PlainCloneContext(ctx, tempDir, false, &gogit.CloneOptions{
			URL:           auth.cleanURL,
			Auth:          auth.buildAuth(),
			ReferenceName: refName,
			SingleBranch:  true,
			Depth:         1,
			Tags:          gogit.NoTags,
		})
		if err == nil {
			return r.chartResult(tempDir, config.ChartPath)
		}

		r.Cleanup(tempDir)
		lastErr = err
		// Only a missing ref is worth retrying as the other ref type; a real
		// failure (auth/network) must surface as-is.
		if !isBranchNotFound(err) {
			break
		}
	}

	if isBranchNotFound(lastErr) {
		return nil, fmt.Errorf("ref '%s' does not exist in repository (tried as both branch and tag). Please check the branch/tag name", config.GitHubBranch)
	}
	// Mask the token in any surfaced output as defense-in-depth.
	return nil, fmt.Errorf("failed to clone repository: %s", maskToken(lastErr.Error(), auth.token))
}

// chartResult validates that chartPath exists inside the freshly cloned tempDir
// and returns the CloneResult, cleaning up on failure.
func (r *Repository) chartResult(tempDir, chartSubPath string) (*CloneResult, error) {
	chartPath := filepath.Join(tempDir, chartSubPath)
	if _, err := os.Stat(chartPath); os.IsNotExist(err) {
		r.Cleanup(tempDir)
		return nil, fmt.Errorf("chart path '%s' does not exist in repository", chartSubPath)
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
			// Log the error but don't fail the operation: this is cleanup, and
			// aborting the main flow over a leftover temp dir is worse than the
			// leak. pterm.Warning, not fmt.Printf — the latter writes straight to
			// stdout and ignores --silent.
			pterm.Warning.Printfln("Failed to clean up the temporary directory %s: %v", tempDir, err)
		}
	}
}
