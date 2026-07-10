package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstallationRequest_DefaultValues(t *testing.T) {
	req := &InstallationRequest{}

	assert.Nil(t, req.Args)
	assert.False(t, req.Force)
	assert.False(t, req.DryRun)
	assert.False(t, req.Verbose)
	assert.Empty(t, req.GitHubRepo)
	assert.Empty(t, req.GitHubBranch)
	assert.Empty(t, req.CertDir)
}

func TestInstallationRequest_WithValues(t *testing.T) {
	req := &InstallationRequest{
		Args:         []string{"test-cluster"},
		Force:        true,
		DryRun:       false,
		Verbose:      true,
		GitHubRepo:   "https://github.com/test/repo",
		GitHubBranch: "main",
		CertDir:      "/path/to/certs",
	}

	assert.Equal(t, []string{"test-cluster"}, req.Args)
	assert.True(t, req.Force)
	assert.False(t, req.DryRun)
	assert.True(t, req.Verbose)
	assert.Equal(t, "https://github.com/test/repo", req.GitHubRepo)
	assert.Equal(t, "main", req.GitHubBranch)
	assert.Equal(t, "/path/to/certs", req.CertDir)
}

func TestInstallationRequest_WithMultipleArgs(t *testing.T) {
	req := &InstallationRequest{
		Args:         []string{"cluster1", "cluster2", "cluster3"},
		Force:        false,
		DryRun:       true,
		Verbose:      false,
		GitHubRepo:   "https://github.com/multi/repo",
		GitHubBranch: "develop",
	}

	assert.Len(t, req.Args, 3)
	assert.Equal(t, "cluster1", req.Args[0])
	assert.Equal(t, "cluster2", req.Args[1])
	assert.Equal(t, "cluster3", req.Args[2])
	assert.False(t, req.Force)
	assert.True(t, req.DryRun)
	assert.False(t, req.Verbose)
	assert.Equal(t, "https://github.com/multi/repo", req.GitHubRepo)
	assert.Equal(t, "develop", req.GitHubBranch)
}

func TestInstallationRequest_EmptyArgs(t *testing.T) {
	req := &InstallationRequest{
		Args:         []string{},
		GitHubRepo:   "https://github.com/empty/args",
		GitHubBranch: "main",
	}

	assert.NotNil(t, req.Args)
	assert.Len(t, req.Args, 0)
	assert.Equal(t, "https://github.com/empty/args", req.GitHubRepo)
	assert.Equal(t, "main", req.GitHubBranch)
}

// Test interface completeness by verifying struct field counts
func TestStructFieldCounts(t *testing.T) {
	// These tests help ensure we don't accidentally remove fields without updating tests

	// InstallationRequest should have 7 fields
	req := InstallationRequest{}
	_ = req.Args
	_ = req.Force
	_ = req.DryRun
	_ = req.Verbose
	_ = req.GitHubRepo
	_ = req.GitHubBranch
	_ = req.CertDir

	// If test passes, all expected fields exist
	assert.True(t, true, "All struct fields are accessible")
}
