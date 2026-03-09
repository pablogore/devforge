//nolint:revive // package name matches adapter purpose; stdlib conflict accepted
package runtime

import "os"

// Environment describes where DevForge is running (CI, GitHub Actions, GitLab CI, or local).
type Environment struct {
	IsCI     bool
	IsGitHub bool
	IsGitLab bool
	IsLocal  bool
}

// DetectEnvironment returns the current runtime environment based on standard env vars.
// CI is set by most CI systems; GITHUB_ACTIONS by GitHub Actions; GITLAB_CI by GitLab CI.
func DetectEnvironment() Environment {
	_, ci := os.LookupEnv("CI")
	_, gh := os.LookupEnv("GITHUB_ACTIONS")
	_, gl := os.LookupEnv("GITLAB_CI")

	return Environment{
		IsCI:     ci,
		IsGitHub: gh,
		IsGitLab: gl,
		IsLocal:  !ci,
	}
}
