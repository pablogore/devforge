package steps

import (
	"context"
	"errors"
	"time"

	"github.com/pablogore/devforge/internal/application"
)

// ErrParallelGroupTimeout is returned when a timed parallel group exceeds its deadline.
var ErrParallelGroupTimeout = errors.New("parallel group timeout exceeded")

// TimeoutGroupStep wraps a step (typically a ParallelGroupStep) with a timeout. When the timeout is exceeded,
// the derived context is cancelled so running work (e.g. sub-steps) is cancelled; no goroutines keep running.
type TimeoutGroupStep struct {
	name    string
	timeout time.Duration
	step    application.Step
}

// NewTimeoutGroupStep returns a step that runs the given step with a deadline. Cancels sub-steps on timeout.
func NewTimeoutGroupStep(name string, timeout time.Duration, step application.Step) *TimeoutGroupStep {
	return &TimeoutGroupStep{name: name, timeout: timeout, step: step}
}

// Name returns the group name for logging.
func (t *TimeoutGroupStep) Name() string {
	return t.name
}

// Run creates a derived context with the timeout, clones Context with StdCtx set to that context,
// runs the wrapped step, and returns errParallelGroupTimeout if the deadline is exceeded.
func (t *TimeoutGroupStep) Run(ctx *application.Context) error {
	ctx.Log.Info("Starting timed parallel group", "group", t.name, "timeout_ms", t.timeout.Milliseconds())
	derived, cancel := context.WithTimeout(ctx.StdCtx, t.timeout)
	defer cancel()

	clone := *ctx
	clone.StdCtx = derived

	err := t.step.Run(&clone)
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrParallelGroupTimeout
	}
	if derived.Err() == context.DeadlineExceeded {
		return ErrParallelGroupTimeout
	}
	return err
}
