package mocks

import (
	"github.com/stretchr/testify/mock"
)

// MockGitClient is a testify mock for ports.GitClient.
type MockGitClient struct {
	mock.Mock
}

// GetCurrentBranch mocks GitClient.GetCurrentBranch.
func (m *MockGitClient) GetCurrentBranch(dir string) (string, error) {
	got := m.Called(dir)
	return got.String(0), got.Error(1)
}

// GetLatestTag mocks GitClient.GetLatestTag.
func (m *MockGitClient) GetLatestTag(dir string) (string, error) {
	got := m.Called(dir)
	return got.String(0), got.Error(1)
}

// GetCommitsSince mocks GitClient.GetCommitsSince.
func (m *MockGitClient) GetCommitsSince(dir, sinceTag string) ([]string, error) {
	got := m.Called(dir, sinceTag)
	if got.Get(0) == nil {
		return nil, got.Error(1)
	}
	return got.Get(0).([]string), got.Error(1)
}

// CreateTag mocks GitClient.CreateTag.
func (m *MockGitClient) CreateTag(dir, version string) error {
	return m.Called(dir, version).Error(0)
}

// IsWorkingTreeClean mocks GitClient.IsWorkingTreeClean.
func (m *MockGitClient) IsWorkingTreeClean(dir string) (bool, error) {
	got := m.Called(dir)
	return got.Bool(0), got.Error(1)
}

// HasFullHistory mocks GitClient.HasFullHistory.
func (m *MockGitClient) HasFullHistory(dir string) (bool, error) {
	got := m.Called(dir)
	return got.Bool(0), got.Error(1)
}

// GetHeadHash mocks GitClient.GetHeadHash.
func (m *MockGitClient) GetHeadHash(dir string) (string, error) {
	got := m.Called(dir)
	return got.String(0), got.Error(1)
}

// GetTagHash mocks GitClient.GetTagHash.
func (m *MockGitClient) GetTagHash(dir, tag string) (string, error) {
	got := m.Called(dir, tag)
	return got.String(0), got.Error(1)
}

// DiffExitCode mocks GitClient.DiffExitCode.
func (m *MockGitClient) DiffExitCode(dir string, files ...string) error {
	args := make([]interface{}, 0, len(files)+1)
	args = append(args, dir)
	for _, f := range files {
		args = append(args, f)
	}
	return m.Called(args...).Error(0)
}

// GetLatestCommitMessage mocks GitClient.GetLatestCommitMessage.
func (m *MockGitClient) GetLatestCommitMessage(dir string) (string, error) {
	got := m.Called(dir)
	return got.String(0), got.Error(1)
}
