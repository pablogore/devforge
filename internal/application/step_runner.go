package application

import (
	"github.com/pablogore/devforge/internal/ports"
)

// StepRunner runs a step and logs duration (observability only; does not change step behavior).
type StepRunner struct {
	log ports.Logger
	clk ports.Clock
}

// NewStepRunner returns a runner that executes steps and logs duration_ms.
func NewStepRunner(log ports.Logger, clk ports.Clock) *StepRunner {
	return &StepRunner{log: log, clk: clk}
}

// Run executes step.Run(ctx) and logs "Step completed" with duration_ms on success, "Step failed" with duration_ms on error.
func (r *StepRunner) Run(ctx *Context, step Step) error {
	start := r.clk.Now()
	err := step.Run(ctx)
	dur := r.clk.Since(start).Milliseconds()
	if err != nil {
		r.log.Error("Step failed", "step", step.Name(), "duration_ms", dur, "error", err)
		return err
	}
	r.log.Info("Step completed", "step", step.Name(), "duration_ms", dur)
	return nil
}
