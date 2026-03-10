package application

import (
	"errors"
	"testing"

	"github.com/pablogore/devforge/internal/testkit"
	"github.com/pablogore/go-specs/specs"
)

func TestRunSteps(t *testing.T) {
	specs.Describe(t, "RunSteps", func(s *specs.Spec) {
		s.It("unknown step returns error", func(ctx *specs.Context) {
			runner := NewStepRunner(nil, testkit.NewFakeClock().Clock())
			err := RunSteps(&Context{}, runner, []string{"nonexistent-step-name"})
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(err.Error() != "").To(specs.BeTrue())
		})
		s.It("known step runs and logs STEP START then STEP SUCCESS", func(ctx *specs.Context) {
			RegisterStep("test-run-steps-step", func() Step { return stubStep{name: "test-run-steps-step"} })
			t.Cleanup(func() { delete(stepRegistry, "test-run-steps-step") })
			log := &testkit.FakeLogger{}
			clk := testkit.NewFakeClock().Clock()
			gCtx := &Context{Log: log, Clock: clk}
			runner := NewStepRunner(log, clk)
			err := RunSteps(gCtx, runner, []string{"test-run-steps-step"})
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(log.LastInfoMsg).ToEqual("[devforge] STEP SUCCESS")
			ctx.Expect(log.InfoCalls >= 2).To(specs.BeTrue())
		})
		s.It("first step succeeds second step fails returns error", func(ctx *specs.Context) {
			RegisterStep("run-steps-ok", func() Step { return stubStep{name: "run-steps-ok"} })
			RegisterStep("run-steps-fail", func() Step { return stubStepFailing{name: "run-steps-fail"} })
			t.Cleanup(func() {
				delete(stepRegistry, "run-steps-ok")
				delete(stepRegistry, "run-steps-fail")
			})
			log := &testkit.FakeLogger{}
			clk := testkit.NewFakeClock().Clock()
			gCtx := &Context{Log: log, Clock: clk}
			runner := NewStepRunner(log, clk)
			err := RunSteps(gCtx, runner, []string{"run-steps-ok", "run-steps-fail"})
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(err.Error()).ToEqual("step failed")
			ctx.Expect(log.ErrorCalls >= 1).To(specs.BeTrue())
		})
	})
}

type stubStepFailing struct{ name string }

func (s stubStepFailing) Name() string { return s.name }

func (s stubStepFailing) Run(_ *Context) error { return errors.New("step failed") }
