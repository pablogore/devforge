package steps

import (
	"testing"
	"time"

	"github.com/pablogore/devforge/internal/application"
	"github.com/pablogore/devforge/internal/guard"
	"github.com/pablogore/go-specs/specs"
)

func stepNames(s []application.Step) []string {
	names := make([]string, 0, len(s))
	for _, st := range s {
		names = append(names, st.Name())
	}
	return names
}

func sliceContains(names []string, x string) bool {
	for _, n := range names {
		if n == x {
			return true
		}
	}
	return false
}

func TestGoLibPRStepsQuick(t *testing.T) {
	specs.Describe(t, "GoLibPRStepsQuick", func(s *specs.Spec) {
		s.It("excludes heavy steps", func(ctx *specs.Context) {
			names := stepNames(GoLibPRStepsQuick(15, 2*time.Minute))
			ctx.Expect(sliceContains(names, "go-mod-tidy")).To(specs.BeTrue())
			ctx.Expect(sliceContains(names, "conventional-commit")).To(specs.BeTrue())
			ctx.Expect(sliceContains(names, "architectural-guard")).To(specs.BeTrue())
			ctx.Expect(sliceContains(names, "static-analysis")).To(specs.BeTrue())
			ctx.Expect(sliceContains(names, "test")).To(specs.BeFalse())
		})
	})
}

func TestGoLibPRStepsFull(t *testing.T) {
	specs.Describe(t, "GoLibPRStepsFull", func(s *specs.Spec) {
		s.It("includes security and test", func(ctx *specs.Context) {
			names := stepNames(GoLibPRStepsFull(15, 2*time.Minute))
			ctx.Expect(sliceContains(names, "go-mod-tidy")).To(specs.BeTrue())
			ctx.Expect(sliceContains(names, "static-analysis")).To(specs.BeTrue())
			ctx.Expect(sliceContains(names, "test")).To(specs.BeTrue())
			ctx.Expect(sliceContains(names, "govulncheck")).To(specs.BeTrue())
		})
	})
}

func TestGoLibPRStepsDeep(t *testing.T) {
	specs.Describe(t, "GoLibPRStepsDeep", func(s *specs.Spec) {
		s.It("includes race step and static-analysis", func(ctx *specs.Context) {
			names := stepNames(GoLibPRStepsDeep(15, 2*time.Minute))
			ctx.Expect(sliceContains(names, "test-race")).To(specs.BeTrue())
			ctx.Expect(sliceContains(names, "static-analysis")).To(specs.BeTrue())
		})
	})
}

func TestGoServicePRStepsQuickDeep(t *testing.T) {
	specs.Describe(t, "GoServicePRSteps Quick and Deep", func(s *specs.Spec) {
		s.It("Quick includes static-analysis and excludes test", func(ctx *specs.Context) {
			names := stepNames(GoServicePRStepsQuick(20, 3*time.Minute))
			ctx.Expect(sliceContains(names, "static-analysis")).To(specs.BeTrue())
			ctx.Expect(sliceContains(names, "test")).To(specs.BeFalse())
		})
		s.It("Deep includes test-race", func(ctx *specs.Context) {
			names := stepNames(GoServicePRStepsDeep(20, 3*time.Minute))
			ctx.Expect(sliceContains(names, "test-race")).To(specs.BeTrue())
		})
	})
}

func TestGoLibPRSteps_EqualsFull(t *testing.T) {
	specs.Describe(t, "GoLibPRSteps equals Full", func(s *specs.Spec) {
		s.It("same length and step names", func(ctx *specs.Context) {
			timeout := 2 * time.Minute
			full := GoLibPRStepsFull(15, timeout, guard.DefaultRules()...)
			legacy := GoLibPRSteps(15, timeout)
			ctx.Expect(len(legacy)).ToEqual(len(full))
			for i := range full {
				ctx.Expect(legacy[i].Name()).ToEqual(full[i].Name())
			}
		})
	})
}
