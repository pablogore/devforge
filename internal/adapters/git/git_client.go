// Package git provides a ports.GitClient implementation using os/exec.
package git

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/pablogore/devforge/internal/ports"
)

// Client runs git commands in a given directory (implements ports.GitClient).
type Client struct{}

// NewGitClient returns a new Client.
func NewGitClient() ports.GitClient {
	return &Client{}
}

// GetCurrentBranch returns the current branch name in dir.
func (g *Client) GetCurrentBranch(dir string) (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// GetLatestTag returns the most recent tag in dir, or empty string if none.
func (g *Client) GetLatestTag(dir string) (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		if strings.Contains(string(out), "no tags") || err.Error() == "exit status 128" {
			return "", nil
		}
		return "", fmt.Errorf("failed to get latest tag: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// GetCommitsSince returns commit subjects between sinceTag and HEAD.
func (g *Client) GetCommitsSince(dir, sinceTag string) ([]string, error) {
	var cmd *exec.Cmd
	if sinceTag == "" {
		cmd = exec.Command("git", "log", "--format=%s", "-n", "1000")
	} else {
		rangeArg := sinceTag + "..HEAD"
		//nolint:gosec // sinceTag is from GetLatestTag or caller; range is internal git ref range
		cmd = exec.Command("git", "log", rangeArg, "--format=%s")
	}
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var commits []string
	for _, line := range lines {
		if line != "" {
			commits = append(commits, line)
		}
	}
	return commits, nil
}

// CreateTag creates an annotated tag at HEAD in dir.
func (g *Client) CreateTag(dir, version string) error {
	tagMsg := "Release " + version
	//nolint:gosec // version is derived from commit history; tag message is internal
	cmd := exec.Command("git", "tag", "-a", version, "-m", tagMsg)
	cmd.Dir = dir
	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}
	return nil
}

// IsWorkingTreeClean reports whether there are uncommitted changes in dir.
func (g *Client) IsWorkingTreeClean(dir string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check working tree: %w", err)
	}
	return len(strings.TrimSpace(string(out))) == 0, nil
}

// HasFullHistory reports whether the repo in dir is not a shallow clone.
func (g *Client) HasFullHistory(dir string) (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--is-shallow-repository")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check if shallow: %w", err)
	}
	isShallow := strings.TrimSpace(string(out)) == "true"
	return !isShallow, nil
}

// GetHeadHash returns the commit hash of HEAD in dir.
func (g *Client) GetHeadHash(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD hash: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// GetTagHash returns the commit hash the given tag points to in dir.
func (g *Client) GetTagHash(dir, tag string) (string, error) {
	ref := tag + "^{commit}"
	//nolint:gosec // tag is from GetLatestTag or version derivation; ref is internal
	cmd := exec.Command("git", "rev-parse", ref)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get tag hash: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// DiffExitCode runs git diff --exit-code for the given files; non-zero exit returns error.
func (g *Client) DiffExitCode(dir string, files ...string) error {
	args := append([]string{"diff", "--exit-code"}, files...)
	//nolint:gosec // args are internal (e.g. go.sum from CI); no user-controlled paths
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git diff failed")
	}
	return nil
}

// GetLatestCommitMessage returns the subject of the latest commit in dir.
func (g *Client) GetLatestCommitMessage(dir string) (string, error) {
	cmd := exec.Command("git", "log", "-1", "--format=%s")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get latest commit message: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
