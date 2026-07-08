package models

// Platform repository URL. The CLI supports only the OSS (oss-tenant)
// deployment; this public repository is the single source of truth for the
// app-of-apps source, so a repo rename is a one-line change here.
const (
	RepoOSSTenant = "https://github.com/flamingo-stack/openframe-oss-tenant"

	// DefaultGitBranch is the default branch of the platform app-of-apps repo.
	DefaultGitBranch = "main"
)
