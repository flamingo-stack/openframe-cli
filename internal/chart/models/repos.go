package models

// Platform repository URLs by deployment mode. Single source of truth: the
// mode→URL mapping in types.GetRepositoryURL returns these, and the install
// defaults reference them, so a repo rename is a one-line change here.
const (
	RepoOSSTenant  = "https://github.com/flamingo-stack/openframe-oss-tenant"
	RepoSaaSTenant = "https://github.com/flamingo-stack/openframe-saas-tenant"
	RepoSaaSShared = "https://github.com/flamingo-stack/openframe-saas-shared"

	// DefaultGitBranch is the default branch of the platform app-of-apps repo.
	DefaultGitBranch = "main"
)
