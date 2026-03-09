package application

import (
	"github.com/pablogore/devforge/internal/config"
)

// Pipeline is a named sequence of steps executed in order.
type Pipeline struct {
	// Name identifies the pipeline (e.g. "pr", "release", "doctor").
	Name string
	// Steps are run in order; pipeline config (enable/disable) may filter them at runtime.
	Steps []Step
}

// filterSteps applies pipeline.enable / pipeline.disable from cfg. If cfg is nil or both are empty, returns steps unchanged.
// Enable (whitelist) takes precedence over Disable when both are set. Order is preserved.
func filterSteps(steps []Step, cfg *config.Config) []Step {
	if cfg == nil {
		return steps
	}
	enable := cfg.Pipeline.Enable
	disable := cfg.Pipeline.Disable
	if len(enable) > 0 {
		allow := make(map[string]bool)
		for _, n := range enable {
			allow[n] = true
		}
		out := make([]Step, 0, len(steps))
		for _, s := range steps {
			if allow[s.Name()] {
				out = append(out, s)
			}
		}
		return out
	}
	if len(disable) > 0 {
		skip := make(map[string]bool)
		for _, n := range disable {
			skip[n] = true
		}
		out := make([]Step, 0, len(steps))
		for _, s := range steps {
			if !skip[s.Name()] {
				out = append(out, s)
			}
		}
		return out
	}
	return steps
}

// Run executes all steps in order using the given runner. Returns on first error.
func (p Pipeline) Run(ctx *Context, runner *StepRunner) error {
	for _, step := range p.Steps {
		if err := runner.Run(ctx, step); err != nil {
			return err
		}
	}
	return nil
}
