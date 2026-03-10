package application

import (
	"context"
	"errors"
	"testing"

	"github.com/pablogore/devforge/internal/domain"
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
		s.It("logs STEP START and STEP SUCCESS on success", func(ctx *specs.Context) {
			log := &testkit.FakeLogger{RecordInfoHistory: true}
			clk := testkit.NewFakeClock().Clock()
			runner := NewStepRunner(log, clk)
			gCtx := &Context{StdCtx: context.Background(), Log: log, Clock: clk}
			err := runner.Run(gCtx, stubStep{name: "stub", err: nil})
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(log.InfoCalls >= 2).To(specs.BeTrue())
			ctx.Expect(log.InfoHistory[0].Msg).ToEqual("[devforge] STEP START")
			ctx.Expect(log.LastInfoMsg).ToEqual("[devforge] STEP SUCCESS")
			ctx.Expect(len(log.LastInfoArgs) >= 4).To(specs.BeTrue())
			ctx.Expect(log.LastInfoArgs[0]).ToEqual("event")
			ctx.Expect(log.LastInfoArgs[2]).ToEqual("step")
			ctx.Expect(log.LastInfoArgs[3]).ToEqual("stub")
		})
		s.It("logs STEP START and STEP FAILURE on generic failure", func(ctx *specs.Context) {
			log := &testkit.FakeLogger{}
			clk := testkit.NewFakeClock().Clock()
			runner := NewStepRunner(log, clk)
			wantErr := errors.New("stub failed")
			gCtx := &Context{StdCtx: context.Background(), Log: log, Clock: clk}
			err := runner.Run(gCtx, stubStep{name: "stub", err: wantErr})
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, wantErr) || err == wantErr).To(specs.BeTrue())
			ctx.Expect(log.LastErrorMsg).ToEqual("[devforge] STEP FAILURE")
			ctx.Expect(len(log.LastErrorArgs) >= 2).To(specs.BeTrue())
		})
		s.It("logs TEST FAILURE when step returns ErrTestFailed", func(ctx *specs.Context) {
			log := &testkit.FakeLogger{}
			clk := testkit.NewFakeClock().Clock()
			runner := NewStepRunner(log, clk)
			gCtx := &Context{StdCtx: context.Background(), Log: log, Clock: clk}
			err := runner.Run(gCtx, stubStep{name: "test", err: domain.ErrTestFailed})
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, domain.ErrTestFailed)).To(specs.BeTrue())
			ctx.Expect(log.LastErrorMsg).ToEqual("[devforge] TEST FAILURE")
		})
		s.It("logs POLICY VIOLATION when step returns ErrFormatting", func(ctx *specs.Context) {
			log := &testkit.FakeLogger{}
			clk := testkit.NewFakeClock().Clock()
			runner := NewStepRunner(log, clk)
			gCtx := &Context{StdCtx: context.Background(), Log: log, Clock: clk}
			err := runner.Run(gCtx, stubStep{name: "gofmt", err: domain.ErrFormatting})
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, domain.ErrFormatting)).To(specs.BeTrue())
			ctx.Expect(log.LastErrorMsg).ToEqual("[devforge] POLICY VIOLATION")
		})
		s.It("logs TOOL FAILURE when step returns ErrToolFailure", func(ctx *specs.Context) {
			log := &testkit.FakeLogger{}
			clk := testkit.NewFakeClock().Clock()
			runner := NewStepRunner(log, clk)
			gCtx := &Context{StdCtx: context.Background(), Log: log, Clock: clk}
			err := runner.Run(gCtx, stubStep{name: "lint", err: domain.ErrToolFailure})
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, domain.ErrToolFailure)).To(specs.BeTrue())
			ctx.Expect(log.LastErrorMsg).ToEqual("[devforge] TOOL FAILURE")
		})
		s.It("logs TOOL CRASH when step returns ErrToolCrash", func(ctx *specs.Context) {
			log := &testkit.FakeLogger{}
			clk := testkit.NewFakeClock().Clock()
			runner := NewStepRunner(log, clk)
			gCtx := &Context{StdCtx: context.Background(), Log: log, Clock: clk}
			err := runner.Run(gCtx, stubStep{name: "lint", err: domain.ErrToolCrash})
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, domain.ErrToolCrash)).To(specs.BeTrue())
			ctx.Expect(log.LastErrorMsg).ToEqual("[devforge] TOOL CRASH")
		})
	})
}
