package mocks

import (
	"github.com/pablogore/devforge/internal/ports"
	"github.com/stretchr/testify/mock"
)

// MockLogger is a testify mock for ports.Logger.
type MockLogger struct {
	mock.Mock
}

// Debug mocks Logger.Debug.
func (m *MockLogger) Debug(msg string, args ...any) {
	m.Called(append([]any{msg}, args...)...)
}

// Info mocks Logger.Info.
func (m *MockLogger) Info(msg string, args ...any) {
	m.Called(append([]any{msg}, args...)...)
}

// Warn mocks Logger.Warn.
func (m *MockLogger) Warn(msg string, args ...any) {
	m.Called(append([]any{msg}, args...)...)
}

// Error mocks Logger.Error.
func (m *MockLogger) Error(msg string, args ...any) {
	m.Called(append([]any{msg}, args...)...)
}

// With mocks Logger.With.
func (m *MockLogger) With(args ...any) ports.Logger {
	got := m.Called(args...)
	if got.Get(0) == nil {
		return nil
	}
	return got.Get(0).(ports.Logger)
}

// Sync mocks Logger.Sync.
func (m *MockLogger) Sync() error {
	return m.Called().Error(0)
}
