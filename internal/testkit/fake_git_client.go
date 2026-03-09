package testkit

import "github.com/pablogore/devforge/internal/ports"

// GitResponse holds (Out, Err) for string-returning git methods.
type GitResponse struct {
	Out string
	Err error
}

// FakeGitClient implements ports.GitClient with configurable return values per method.
// Set the Out/Err fields before the test; each method returns the corresponding value.
// For methods called multiple times in one run (e.g. GetTagHash), set the Responses slice
// and the fake returns Responses[i] on the i-th call (then reuses last or zero).
type FakeGitClient struct {
	GetCurrentBranchOut    string
	GetCurrentBranchErr    error
	GetLatestTagOut        string
	GetLatestTagErr        error
	GetCommitsSinceOut     []string
	GetCommitsSinceErr     error
	CreateTagErr           error
	IsWorkingTreeCleanOut  bool
	IsWorkingTreeCleanErr  error
	HasFullHistoryOut      bool
	HasFullHistoryErr      error
	GetHeadHashOut         string
	GetHeadHashErr         error
	GetTagHashOut          string
	GetTagHashErr          error
	GetTagHashResponses    []GitResponse // if set, consumed in order per GetTagHash call
	getTagHashIdx          int
	DiffExitCodeErr        error
	GetLatestCommitMessageOut string
	GetLatestCommitMessageErr error
}

func (f *FakeGitClient) GetCurrentBranch(string) (string, error) {
	return f.GetCurrentBranchOut, f.GetCurrentBranchErr
}

func (f *FakeGitClient) GetLatestTag(string) (string, error) {
	return f.GetLatestTagOut, f.GetLatestTagErr
}

func (f *FakeGitClient) GetCommitsSince(string, string) ([]string, error) {
	return f.GetCommitsSinceOut, f.GetCommitsSinceErr
}

func (f *FakeGitClient) CreateTag(string, string) error {
	return f.CreateTagErr
}

func (f *FakeGitClient) IsWorkingTreeClean(string) (bool, error) {
	return f.IsWorkingTreeCleanOut, f.IsWorkingTreeCleanErr
}

func (f *FakeGitClient) HasFullHistory(string) (bool, error) {
	return f.HasFullHistoryOut, f.HasFullHistoryErr
}

func (f *FakeGitClient) GetHeadHash(string) (string, error) {
	return f.GetHeadHashOut, f.GetHeadHashErr
}

func (f *FakeGitClient) GetTagHash(string, string) (string, error) {
	if len(f.GetTagHashResponses) > 0 {
		if f.getTagHashIdx < len(f.GetTagHashResponses) {
			r := f.GetTagHashResponses[f.getTagHashIdx]
			f.getTagHashIdx++
			return r.Out, r.Err
		}
		return "", nil
	}
	return f.GetTagHashOut, f.GetTagHashErr
}

func (f *FakeGitClient) DiffExitCode(dir string, files ...string) error {
	return f.DiffExitCodeErr
}

func (f *FakeGitClient) GetLatestCommitMessage(string) (string, error) {
	return f.GetLatestCommitMessageOut, f.GetLatestCommitMessageErr
}

var _ ports.GitClient = (*FakeGitClient)(nil)
