package application

import (
	"context"
	"errors"
	"testing"

	"github.com/pablogore/devforge/internal/testkit"
	"github.com/pablogore/go-specs/specs"
)

const stepRunnerTestDurationMs = int64(50)

// stubStep is a minimal Step for testing StepRunner (no real behavior).
type stubStep struct {
	name string
	err  error
}

func (s stubStep) Name() string         { return s.name }
func (s stubStep) Run(_ *Context) error { return s.err }

func TestStepRunner(t *testing.T) {
	specs.Describe(t, "StepRunner", func(s *specs.Spec) {
		s.It("logs duration on success", func(ctx *specs.Context) {
			log := &testkit.FakeLogger{}
			clk := testkit.NewFakeClock().Clock()
			runner := NewStepRunner(log, clk)
			gCtx := &Context{StdCtx: context.Background(), Log: log, Clock: clk}
			err := runner.Run(gCtx, stubStep{name: "stub", err: nil})
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(log.LastInfoMsg).ToEqual("Step completed")
			ctx.Expect(len(log.LastInfoArgs) >= 4).To(specs.BeTrue())
			ctx.Expect(log.LastInfoArgs[0]).ToEqual("step")
			ctx.Expect(log.LastInfoArgs[1]).ToEqual("stub")
			ctx.Expect(log.LastInfoArgs[2]).ToEqual("duration_ms")
			ctx.Expect(log.LastInfoArgs[3]).ToEqual(stepRunnerTestDurationMs)
		})
		s.It("logs error and returns step error on failure", func(ctx *specs.Context) {
			log := &testkit.FakeLogger{}
			clk := testkit.NewFakeClock().Clock()
			runner := NewStepRunner(log, clk)
			wantErr := errors.New("stub failed")
			gCtx := &Context{StdCtx: context.Background(), Log: log, Clock: clk}
			err := runner.Run(gCtx, stubStep{name: "stub", err: wantErr})
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, wantErr) || err == wantErr).To(specs.BeTrue())
			ctx.Expect(log.LastErrorMsg).ToEqual("Step failed")
			ctx.Expect(len(log.LastErrorArgs) >= 2).To(specs.BeTrue())
			ctx.Expect(log.LastErrorArgs[0]).ToEqual("step")
			ctx.Expect(log.LastErrorArgs[1]).ToEqual("stub")
		})
	})
}
