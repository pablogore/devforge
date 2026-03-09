package exec

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/pablogore/devforge/internal/ports"
)

// CommandRunner runs shell commands in a given directory via os/exec.
type CommandRunner struct{}

// NewCommandRunner returns a new CommandRunner.
func NewCommandRunner() ports.CommandRunner {
	return &CommandRunner{}
}

// Run runs the named command with args in dir and returns stdout. Stderr is not captured.
func (r *CommandRunner) Run(ctx context.Context, dir string, name string, args ...string) (string, error) {
	//nolint:gosec // G204: name and args are from internal callers (steps, guard); no shell injection
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%s %v failed: %s", name, args, out)
	}
	return string(out), nil
}

// RunCombinedOutput runs the named command with args in dir and returns combined stdout and stderr.
func (r *CommandRunner) RunCombinedOutput(ctx context.Context, dir string, name string, args ...string) (string, error) {
	return r.RunCombinedOutputWithEnv(ctx, dir, nil, name, args...)
}

// RunCombinedOutputWithEnv runs the command with the given env (or process env if nil) and returns combined output.
func (r *CommandRunner) RunCombinedOutputWithEnv(ctx context.Context, dir string, env []string, name string, args ...string) (string, error) {
	//nolint:gosec // G204: name and args are from internal callers (steps, guard, plugins); no shell injection
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	if env != nil {
		cmd.Env = env
	} else {
		cmd.Env = os.Environ()
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("%s %v failed", name, args)
	}
	return string(out), nil
}
