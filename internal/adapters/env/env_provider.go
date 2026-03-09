package env

import (
	"os"

	"github.com/pablogore/devforge/internal/ports"
)

// Provider is the default adapter implementing ports.EnvProvider using os.Getenv.
type Provider struct{}

// NewEnvProvider returns a new Provider (implements ports.EnvProvider).
func NewEnvProvider() ports.EnvProvider {
	return &Provider{}
}

// Get returns the value of the environment variable named by key.
func (e *Provider) Get(key string) string {
	return os.Getenv(key)
}
