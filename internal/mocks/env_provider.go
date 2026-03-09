package mocks

import (
	"github.com/stretchr/testify/mock"
)

// MockEnvProvider is a testify mock for ports.EnvProvider.
type MockEnvProvider struct {
	mock.Mock
}

// Get mocks EnvProvider.Get.
func (m *MockEnvProvider) Get(key string) string {
	return m.Called(key).String(0)
}
