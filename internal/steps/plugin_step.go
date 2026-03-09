package steps

import (
	"fmt"

	"github.com/pablogore/devforge/internal/application"
)

// PluginStep runs a user-defined command via bash -c "<command>". Used for plugin entries from .devforge.yml.
// Build from config with NewPluginStep(config.Name, config.Run).
type PluginStep struct {
	name    string // plugin name for logging (unexported to avoid shadowing Name())
	command string
}

// NewPluginStep returns a step that runs the given command; name is the plugin name for logging.
func NewPluginStep(name, command string) *PluginStep {
	return &PluginStep{name: name, command: command}
}

// Name returns the plugin name for the Step interface.
func (s *PluginStep) Name() string {
	return s.name
}

// Run executes the plugin command via bash -c, logs name and duration, and fails the pipeline on non-zero exit.
func (s *PluginStep) Run(ctx *application.Context) error {
	ctx.Log.Info("running plugin", "plugin", s.name)
	start := ctx.Clock.Now()

	_, err := ctx.Cmd.RunCombinedOutput(ctx.StdCtx, ctx.Workdir, "bash", "-c", s.command)
	durationMs := ctx.Clock.Since(start).Milliseconds()

	ctx.Log.Info("plugin completed", "plugin", s.name, "duration_ms", durationMs)
	if err != nil {
		return fmt.Errorf("plugin %s: %w", s.name, err)
	}
	return nil
}
