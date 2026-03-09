// Package ports defines interfaces for environment, git, logging, and command execution.
package ports

// EnvProvider provides access to environment variables.
type EnvProvider interface {
	Get(key string) string
}
