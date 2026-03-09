package application

import (
	"testing"

	"github.com/pablogore/go-specs/specs"
)

func TestStepRegistry(t *testing.T) {
	specs.Describe(t, "step registry", func(s *specs.Spec) {
		s.It("ListSteps returns sorted names and includes registered step", func(ctx *specs.Context) {
			RegisterStep("test-list-steps-step", func() Step { return stubStep{name: "test-list-steps-step"} })
			t.Cleanup(func() { delete(stepRegistry, "test-list-steps-step") })

			names := ListSteps()
			ctx.Expect(len(names) > 0).To(specs.BeTrue())
			ctx.Expect(contains(names, "test-list-steps-step")).To(specs.BeTrue())
			for i := 1; i < len(names); i++ {
				ctx.Expect(names[i] >= names[i-1]).To(specs.BeTrue())
			}
		})
		s.It("GetStep known step returns step", func(ctx *specs.Context) {
			RegisterStep("test-get-step-known", func() Step { return stubStep{name: "test-get-step-known"} })
			t.Cleanup(func() { delete(stepRegistry, "test-get-step-known") })

			step, ok := GetStep("test-get-step-known")
			ctx.Expect(ok).To(specs.BeTrue())
			ctx.Expect(step != nil).To(specs.BeTrue())
			ctx.Expect(step.Name()).ToEqual("test-get-step-known")
		})
		s.It("GetStep unknown step returns false", func(ctx *specs.Context) {
			step, ok := GetStep("nonexistent-step-name-xyz")
			ctx.Expect(ok).To(specs.BeFalse())
			ctx.Expect(step == nil).To(specs.BeTrue())
		})
	})
}

func contains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}
