// Package mocks provides testify mocks for ports (CommandRunner, EnvProvider, GitClient, Logger).
package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockCommandRunner is a testify mock for ports.CommandRunner.
type MockCommandRunner struct {
	mock.Mock
}

// Run mocks CommandRunner.Run.
func (m *MockCommandRunner) Run(ctx context.Context, dir string, name string, args ...string) (string, error) {
	argsCopy := make([]interface{}, 0, len(args)+3)
	argsCopy = append(argsCopy, ctx, dir, name)
	for _, a := range args {
		argsCopy = append(argsCopy, a)
	}
	got := m.Called(argsCopy...)
	return got.String(0), got.Error(1)
}

// RunCombinedOutput mocks CommandRunner.RunCombinedOutput.
func (m *MockCommandRunner) RunCombinedOutput(ctx context.Context, dir string, name string, args ...string) (string, error) {
	argsCopy := make([]interface{}, 0, len(args)+3)
	argsCopy = append(argsCopy, ctx, dir, name)
	for _, a := range args {
		argsCopy = append(argsCopy, a)
	}
	got := m.Called(argsCopy...)
	return got.String(0), got.Error(1)
}

// RunCombinedOutputWithEnv mocks CommandRunner.RunCombinedOutputWithEnv.
func (m *MockCommandRunner) RunCombinedOutputWithEnv(ctx context.Context, dir string, env []string, name string, args ...string) (string, error) {
	argsCopy := make([]interface{}, 0, len(args)+4)
	argsCopy = append(argsCopy, ctx, dir, env, name)
	for _, a := range args {
		argsCopy = append(argsCopy, a)
	}
	got := m.Called(argsCopy...)
	return got.String(0), got.Error(1)
}
