package steps

import (
	"testing"
	"time"

	"github.com/pablogore/devforge/internal/guard"
	"github.com/pablogore/go-specs/specs"
)

func TestStepNames(t *testing.T) {
	specs.Describe(t, "step names", func(s *specs.Spec) {
		s.It("step Name returns expected", func(ctx *specs.Context) {
			ctx.Expect(GoModTidyStep{}.Name()).ToEqual("go-mod-tidy")
			ctx.Expect(GoFmtStep{}.Name()).ToEqual("gofmt")
			ctx.Expect(GovulnCheckStep{}.Name()).ToEqual("govulncheck")
			ctx.Expect(GoTestStep{}.Name()).ToEqual("test")
			ctx.Expect(GoTestRaceStep{}.Name()).ToEqual("test-race")
			ctx.Expect(CoverageStep{}.Name()).ToEqual("test")
			ctx.Expect(ConventionalCommitStep{}.Name()).ToEqual("conventional-commit")
			ctx.Expect(NewArchitecturalGuardStep(guard.DefaultRules()).Name()).ToEqual("architectural-guard")
			ctx.Expect(PolicyPackStep{}.Name()).ToEqual("policy-pack")
		})
	})
}

func TestGoLibPRSteps(t *testing.T) {
	specs.Describe(t, "GoLibPRSteps", func(s *specs.Spec) {
		s.It("Quick Full Deep return non-empty", func(ctx *specs.Context) {
			timeout := 2 * time.Minute
			quick := GoLibPRStepsQuick(10, timeout)
			full := GoLibPRStepsFull(10, timeout)
			deep := GoLibPRStepsDeep(10, timeout)
			ctx.Expect(len(quick) > 0).To(specs.BeTrue())
			ctx.Expect(len(full) > 0).To(specs.BeTrue())
			ctx.Expect(len(deep) > 0).To(specs.BeTrue())
			ctx.Expect(len(deep) >= len(full)).To(specs.BeTrue())
			ctx.Expect(len(full) >= len(quick)).To(specs.BeTrue())
		})
		s.It("GoLibPRSteps equals GoLibPRStepsFull", func(ctx *specs.Context) {
			timeout := 2 * time.Minute
			ctx.Expect(GoLibPRStepsFull(10, timeout)).ToEqual(GoLibPRSteps(10, timeout))
		})
		s.It("with custom rules returns non-empty", func(ctx *specs.Context) {
			timeout := 2 * time.Minute
			rules := guard.DefaultRules()
			steps := GoLibPRStepsFull(10, timeout, rules...)
			ctx.Expect(len(steps) > 0).To(specs.BeTrue())
		})
	})
}

func TestGoServicePRSteps(t *testing.T) {
	specs.Describe(t, "GoServicePRSteps", func(s *specs.Spec) {
		s.It("Quick Full Deep return non-empty", func(ctx *specs.Context) {
			timeout := 3 * time.Minute
			quick := GoServicePRStepsQuick(10, timeout)
			full := GoServicePRStepsFull(10, timeout)
			deep := GoServicePRStepsDeep(10, timeout)
			ctx.Expect(len(quick) > 0).To(specs.BeTrue())
			ctx.Expect(len(full) > 0).To(specs.BeTrue())
			ctx.Expect(len(deep) > 0).To(specs.BeTrue())
			ctx.Expect(len(deep) >= len(full)).To(specs.BeTrue())
			ctx.Expect(len(full) >= len(quick)).To(specs.BeTrue())
		})
		s.It("GoServicePRSteps equals GoServicePRStepsFull", func(ctx *specs.Context) {
			timeout := 3 * time.Minute
			ctx.Expect(GoServicePRStepsFull(10, timeout)).ToEqual(GoServicePRSteps(10, timeout))
		})
	})
}
