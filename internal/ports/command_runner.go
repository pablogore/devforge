package ports

import "context"

// CommandRunner runs shell commands. Context is used for cancellation (e.g. timeouts).
type CommandRunner interface {
	Run(ctx context.Context, dir string, name string, args ...string) (string, error)
	RunCombinedOutput(ctx context.Context, dir string, name string, args ...string) (string, error)
	// RunCombinedOutputWithEnv runs a command with the given environment (e.g. to inject DEVFORGE_PLUGIN_EXECUTION).
	// If env is nil, the process environment is used.
	RunCombinedOutputWithEnv(ctx context.Context, dir string, env []string, name string, args ...string) (string, error)
}
