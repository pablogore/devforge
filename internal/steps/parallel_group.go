package steps

import (
	"github.com/pablogore/devforge/internal/application"
	"golang.org/x/sync/errgroup"
)

// ParallelGroupStep runs a fixed slice of steps concurrently; fails on first error and cancels remaining. Deterministic composition.
type ParallelGroupStep struct {
	name  string
	steps []application.Step
}

// NewParallelGroupStep returns a step that runs all steps in parallel (fail-fast). steps is compile-time composed; no dynamic discovery.
func NewParallelGroupStep(name string, steps []application.Step) *ParallelGroupStep {
	return &ParallelGroupStep{name: name, steps: steps}
}

// Name returns the group name for logging.
func (p *ParallelGroupStep) Name() string {
	return p.name
}

// Run executes all sub-steps concurrently via errgroup; returns the first error and cancels remaining. Uses ctx.StdCtx so timeouts cancel running work.
func (p *ParallelGroupStep) Run(ctx *application.Context) error {
	ctx.Log.Info("Running parallel group", "group", p.name)
	g, _ := errgroup.WithContext(ctx.StdCtx)
	for _, step := range p.steps {
		s := step
		g.Go(func() error {
			return s.Run(ctx)
		})
	}
	return g.Wait()
}

// SequentialGroupStep runs a fixed slice of steps one after another; same composition as ParallelGroupStep but deterministic order for testing.
type SequentialGroupStep struct {
	name  string
	steps []application.Step
}

// NewSequentialGroupStep returns a step that runs all sub-steps in slice order (for tests that need deterministic mock order).
func NewSequentialGroupStep(name string, steps []application.Step) *SequentialGroupStep {
	return &SequentialGroupStep{name: name, steps: steps}
}

// Name returns the group name for logging.
func (s *SequentialGroupStep) Name() string {
	return s.name
}

// Run executes each sub-step in order; fails on first error.
func (s *SequentialGroupStep) Run(ctx *application.Context) error {
	ctx.Log.Info("Running parallel group", "group", s.name)
	for _, step := range s.steps {
		if err := step.Run(ctx); err != nil {
			return err
		}
	}
	return nil
}
