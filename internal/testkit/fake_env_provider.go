package testkit

import "github.com/pablogore/devforge/internal/ports"

// FakeEnvProvider implements ports.EnvProvider using a map. Get returns env[key] or "".
type FakeEnvProvider struct {
	Env map[string]string
}

// NewFakeEnvProvider returns an EnvProvider with the given map (nil = empty).
func NewFakeEnvProvider(env map[string]string) *FakeEnvProvider {
	if env == nil {
		env = make(map[string]string)
	}
	return &FakeEnvProvider{Env: env}
}

// Get returns the value for key or "".
func (f *FakeEnvProvider) Get(key string) string {
	return f.Env[key]
}

var _ ports.EnvProvider = (*FakeEnvProvider)(nil)
