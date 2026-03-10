package application

import (
	"github.com/pablogore/devforge/internal/ports"
)

// StepRunner runs a step and logs duration (observability only; does not change step behavior).
// Logs use [devforge] prefix and structured events: step_start, step_success, test_failure, policy_violation, tool_failure, tool_crash, step_failure.
type StepRunner struct {
	log ports.Logger
	clk ports.Clock
}

// NewStepRunner returns a runner that executes steps and logs duration_ms.
func NewStepRunner(log ports.Logger, clk ports.Clock) *StepRunner {
	return &StepRunner{log: log, clk: clk}
}

// Run executes step.Run(ctx) and logs [devforge] STEP START, then on success [devforge] STEP SUCCESS
// or on failure [devforge] TEST FAILURE / POLICY VIOLATION / TOOL FAILURE / TOOL CRASH / STEP FAILURE with event and kind for machine parsing.
func (r *StepRunner) Run(ctx *Context, step Step) error {
	r.log.Info(PipelineLogPrefix+"STEP START", "event", "step_start", "step", step.Name())
	start := r.clk.Now()
	err := step.Run(ctx)
	dur := r.clk.Since(start).Milliseconds()
	if err != nil {
		kind := ClassifyFailure(err)
		switch kind {
		case FailureKindTest:
			r.log.Error(PipelineLogPrefix+"TEST FAILURE", "event", "step_failure", "kind", kind, "step", step.Name(), "duration_ms", dur, "error", err)
		case FailureKindPolicyViolation:
			r.log.Error(PipelineLogPrefix+"POLICY VIOLATION", "event", "step_failure", "kind", kind, "step", step.Name(), "duration_ms", dur, "error", err)
		case FailureKindToolError:
			r.log.Error(PipelineLogPrefix+"TOOL FAILURE", "event", "step_failure", "kind", kind, "step", step.Name(), "duration_ms", dur, "error", err)
		case FailureKindToolCrash:
			r.log.Error(PipelineLogPrefix+"TOOL CRASH", "event", "step_failure", "kind", kind, "step", step.Name(), "duration_ms", dur, "error", err)
		default:
			r.log.Error(PipelineLogPrefix+"STEP FAILURE", "event", "step_failure", "kind", FailureKindUnknown, "step", step.Name(), "duration_ms", dur, "error", err)
		}
		return err
	}
	r.log.Info(PipelineLogPrefix+"STEP SUCCESS", "event", "step_success", "step", step.Name(), "duration_ms", dur)
	return nil
}
