package application

import (
	"testing"

	"github.com/pablogore/devforge/internal/testkit"
	"github.com/pablogore/go-specs/specs"
)

func TestPipelineRegistry(t *testing.T) {
	specs.Describe(t, "pipeline registry", func(s *specs.Spec) {
		s.It("RegisterPipeline GetPipeline ListPipelines roundtrip", func(ctx *specs.Context) {
			p := Pipeline{Name: "test-registry-pipeline", Steps: []Step{stubStep{name: "x"}}}
			RegisterPipeline(p)
			t.Cleanup(func() { delete(pipelineRegistry, "test-registry-pipeline") })

			got, ok := GetPipeline("test-registry-pipeline")
			ctx.Expect(ok).To(specs.BeTrue())
			ctx.Expect(got.Name).ToEqual("test-registry-pipeline")
			names := ListPipelines()
			ctx.Expect(contains(names, "test-registry-pipeline")).To(specs.BeTrue())
		})
		s.It("GetPipeline unknown returns false", func(ctx *specs.Context) {
			_, ok := GetPipeline("nonexistent-pipeline-name")
			ctx.Expect(ok).To(specs.BeFalse())
		})
		s.It("RunPipeline unknown returns error", func(ctx *specs.Context) {
			runner := NewStepRunner(nil, testkit.NewFakeClock().Clock())
			err := RunPipeline("unknown-pipeline", &Context{}, runner)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(err.Error() != "").To(specs.BeTrue())
		})
		s.It("RunPipeline success runs steps and logs", func(ctx *specs.Context) {
			log := &testkit.FakeLogger{}
			clk := testkit.NewFakeClock().Clock()
			RegisterPipeline(Pipeline{Name: "run-test-pipeline", Steps: []Step{stubStep{name: "s1"}}})
			t.Cleanup(func() { delete(pipelineRegistry, "run-test-pipeline") })
			gCtx := &Context{StdCtx: nil, Log: log, Clock: clk}
			runner := NewStepRunner(log, clk)
			err := RunPipeline("run-test-pipeline", gCtx, runner)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(log.LastInfoMsg).ToEqual("Step completed")
			ctx.Expect(log.InfoCalls >= 1).To(specs.BeTrue())
		})
	})
}
