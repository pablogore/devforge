package ports

// GitClient performs git operations in a repository directory.
type GitClient interface {
	GetCurrentBranch(dir string) (string, error)
	GetLatestTag(dir string) (string, error)
	GetCommitsSince(dir, sinceTag string) ([]string, error)
	CreateTag(dir, version string) error
	IsWorkingTreeClean(dir string) (bool, error)
	HasFullHistory(dir string) (bool, error)
	GetHeadHash(dir string) (string, error)
	GetTagHash(dir, tag string) (string, error)
	DiffExitCode(dir string, files ...string) error
	GetLatestCommitMessage(dir string) (string, error)
}
