package application

import (
	"context"
	"errors"
	"testing"

	"github.com/pablogore/devforge/internal/config"
	"github.com/pablogore/devforge/internal/testkit"
	"github.com/pablogore/go-specs/specs"
)

func TestFilterSteps(t *testing.T) {
	specs.Describe(t, "filterSteps", func(s *specs.Spec) {
		s.It("nil config returns unchanged", func(ctx *specs.Context) {
			steps := []Step{stubStep{name: "a"}, stubStep{name: "b"}}
			got := filterSteps(steps, nil)
			ctx.Expect(len(got)).ToEqual(2)
			ctx.Expect(got[0].Name()).ToEqual("a")
			ctx.Expect(got[1].Name()).ToEqual("b")
		})
		s.It("empty enable/disable returns unchanged", func(ctx *specs.Context) {
			steps := []Step{stubStep{name: "a"}}
			got := filterSteps(steps, &config.Config{})
			ctx.Expect(len(got)).ToEqual(1)
			ctx.Expect(got[0].Name()).ToEqual("a")
		})
		s.It("enable whitelist filters to allowed names", func(ctx *specs.Context) {
			steps := []Step{stubStep{name: "a"}, stubStep{name: "b"}, stubStep{name: "c"}}
			cfg := &config.Config{Pipeline: config.PipelineConfig{Enable: []string{"b", "a"}}}
			got := filterSteps(steps, cfg)
			ctx.Expect(len(got)).ToEqual(2)
			ctx.Expect(got[0].Name()).ToEqual("a")
			ctx.Expect(got[1].Name()).ToEqual("b")
		})
		s.It("disable filters out names", func(ctx *specs.Context) {
			steps := []Step{stubStep{name: "a"}, stubStep{name: "b"}, stubStep{name: "c"}}
			cfg := &config.Config{Pipeline: config.PipelineConfig{Disable: []string{"b"}}}
			got := filterSteps(steps, cfg)
			ctx.Expect(len(got)).ToEqual(2)
			ctx.Expect(got[0].Name()).ToEqual("a")
			ctx.Expect(got[1].Name()).ToEqual("c")
		})
	})
}

func TestPipeline_Run(t *testing.T) {
	const durationMs = int64(50)
	specs.Describe(t, "Pipeline.Run", func(s *specs.Spec) {
		s.It("success runs all steps and logs", func(ctx *specs.Context) {
			log := &testkit.FakeLogger{RecordInfoHistory: true}
			clk := testkit.NewFakeClock().Clock()
			gCtx := &Context{StdCtx: context.Background(), Log: log, Clock: clk}
			runner := NewStepRunner(log, clk)
			p := Pipeline{Name: "test", Steps: []Step{stubStep{name: "s1", err: nil}, stubStep{name: "s2", err: nil}}}
			err := p.Run(gCtx, runner)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(log.InfoCalls >= 2).To(specs.BeTrue())
			var stepCompleted int
			for _, c := range log.InfoHistory {
				if c.Msg == "Step completed" {
					stepCompleted++
				}
			}
			ctx.Expect(stepCompleted).ToEqual(2)
		})
		s.It("first step fails returns error and logs", func(ctx *specs.Context) {
			wantErr := errors.New("stub error")
			log := &testkit.FakeLogger{}
			clk := testkit.NewFakeClock().Clock()
			gCtx := &Context{StdCtx: context.Background(), Log: log, Clock: clk}
			runner := NewStepRunner(log, clk)
			p := Pipeline{Name: "test", Steps: []Step{stubStep{name: "s1", err: wantErr}, stubStep{name: "s2"}}}
			err := p.Run(gCtx, runner)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(err == wantErr || errors.Is(err, wantErr)).To(specs.BeTrue())
			ctx.Expect(log.LastErrorMsg).ToEqual("Step failed")
			ctx.Expect(log.ErrorCalls >= 1).To(specs.BeTrue())
		})
	})
}
