//nolint:revive // var-naming: package name describes coverage policy resolution; stdlib conflict accepted
package coverage

import (
	"context"
	"errors"
	"testing"

	"github.com/pablogore/devforge/internal/testkit"
	"github.com/pablogore/go-specs/specs"
)

func TestValidateCoveragePatterns(t *testing.T) {
	specs.Describe(t, "ValidateCoveragePatterns", func(s *specs.Spec) {
		s.It("accepts nil or single pattern", func(ctx *specs.Context) {
			ctx.Expect(ValidateCoveragePatterns(nil)).To(specs.BeNil())
			ctx.Expect(ValidateCoveragePatterns([]string{"*"})).To(specs.BeNil())
			ctx.Expect(ValidateCoveragePatterns([]string{"internal/domain"})).To(specs.BeNil())
			ctx.Expect(ValidateCoveragePatterns([]string{"internal/*", "cmd/*"})).To(specs.BeNil())
		})
		s.It("returns ErrWildcardWithOthers when * is mixed with others", func(ctx *specs.Context) {
			err := ValidateCoveragePatterns([]string{"*", "internal/domain"})
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, ErrWildcardWithOthers)).To(specs.BeTrue())
			err = ValidateCoveragePatterns([]string{"internal/domain", "*"})
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, ErrWildcardWithOthers)).To(specs.BeTrue())
		})
	})
}

func TestBuildCoverPkgFlag(t *testing.T) {
	specs.Describe(t, "BuildCoverPkgFlag", func(s *specs.Spec) {
		s.It("covers flag paths", func(ctx *specs.Context) {
			ctx.Expect(BuildCoverPkgFlag(nil)).ToEqual("")
			ctx.Expect(BuildCoverPkgFlag([]string{})).ToEqual("")
			ctx.Expect(BuildCoverPkgFlag([]string{"internal/domain"})).ToEqual("internal/domain")
			ctx.Expect(BuildCoverPkgFlag([]string{"internal/domain", "internal/application"})).ToEqual("internal/domain,internal/application")
		})
	})
}

func TestResolveCoveragePackages(t *testing.T) {
	stdCtx := context.Background()
	specs.Describe(t, "ResolveCoveragePackages", func(s *specs.Spec) {
		s.It("wildcard returns all excluding vendor testdata examples generated mocks", func(ctx *specs.Context) {
			out := "github.com/foo/internal/domain\ngithub.com/foo/internal/vendor/x\ngithub.com/foo/internal/application\n" +
				"github.com/foo/internal/testdata/y\ngithub.com/foo/examples/demo\ngithub.com/foo/internal/generated/z\n" +
				"github.com/foo/internal/mocks\n"
			runner := testkit.NewFakeCommandRunner()
			runner.Stub("go", []string{"list", "./..."}, out, nil)
			got, err := ResolveCoveragePackages(stdCtx, "/wd", []string{"*"}, runner)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(got) == 2).To(specs.BeTrue())
			ctx.Expect(got[0] == "github.com/foo/internal/domain" || got[1] == "github.com/foo/internal/domain").To(specs.BeTrue())
			ctx.Expect(got[0] == "github.com/foo/internal/application" || got[1] == "github.com/foo/internal/application").To(specs.BeTrue())
			ctx.Expect(runner.WasCalled("go", "list", "./...")).To(specs.BeTrue())
		})
		s.It("glob pattern matches", func(ctx *specs.Context) {
			out := "github.com/foo/internal/domain\ngithub.com/foo/internal/application\ngithub.com/foo/cmd/cli\n"
			runner := testkit.NewFakeCommandRunner()
			runner.Stub("go", []string{"list", "./..."}, out, nil)
			got, err := ResolveCoveragePackages(stdCtx, "/wd", []string{"internal/*"}, runner)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(got)).ToEqual(2)
		})
		s.It("explicit pattern suffix match", func(ctx *specs.Context) {
			out := "github.com/foo/internal/domain\ngithub.com/foo/internal/application\n"
			runner := testkit.NewFakeCommandRunner()
			runner.Stub("go", []string{"list", "./..."}, out, nil)
			got, err := ResolveCoveragePackages(stdCtx, "/wd", []string{"internal/domain"}, runner)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(got).ToEqual([]string{"github.com/foo/internal/domain"})
		})
		s.It("list fails returns error", func(ctx *specs.Context) {
			runner := testkit.NewFakeCommandRunner()
			runner.Stub("go", []string{"list", "./..."}, "", errors.New("go list failed"))
			got, err := ResolveCoveragePackages(stdCtx, "/wd", []string{"*"}, runner)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(got == nil).To(specs.BeTrue())
		})
		s.It("empty lines ignored", func(ctx *specs.Context) {
			out := "pkg/a\n\npkg/b\n  \n"
			runner := testkit.NewFakeCommandRunner()
			runner.Stub("go", []string{"list", "./..."}, out, nil)
			got, err := ResolveCoveragePackages(stdCtx, "/wd", []string{"*"}, runner)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(got)).ToEqual(2)
			ctx.Expect(got[0] == "pkg/a" || got[1] == "pkg/a").To(specs.BeTrue())
			ctx.Expect(got[0] == "pkg/b" || got[1] == "pkg/b").To(specs.BeTrue())
		})
		s.It("exact full path match", func(ctx *specs.Context) {
			out := "github.com/foo/internal/domain\ngithub.com/foo/internal/application\n"
			runner := testkit.NewFakeCommandRunner()
			runner.Stub("go", []string{"list", "./..."}, out, nil)
			got, err := ResolveCoveragePackages(stdCtx, "/wd", []string{"github.com/foo/internal/domain"}, runner)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(got).ToEqual([]string{"github.com/foo/internal/domain"})
		})
		s.It("validatePatterns fails without calling runner", func(ctx *specs.Context) {
			runner := testkit.NewFakeCommandRunner()
			runner.Default = &testkit.CommandResult{Err: errors.New("unexpected")}
			got, err := ResolveCoveragePackages(stdCtx, "/wd", []string{"*", "internal/domain"}, runner)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, ErrWildcardWithOthers)).To(specs.BeTrue())
			ctx.Expect(got == nil).To(specs.BeTrue())
		})
		s.It("dot-slash patterns from .devforge.yml match go list import paths", func(ctx *specs.Context) {
			out := "github.com/getsyntegrity/kit-core/domain\ngithub.com/getsyntegrity/kit-core/errorschain\ngithub.com/getsyntegrity/kit-core/fflags\n"
			runner := testkit.NewFakeCommandRunner()
			runner.Stub("go", []string{"list", "./..."}, out, nil)
			got, err := ResolveCoveragePackages(stdCtx, "/wd", []string{"./domain", "./errorschain", "./fflags"}, runner)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(got)).ToEqual(3)
			ctx.Expect(contains(got, "github.com/getsyntegrity/kit-core/domain")).To(specs.BeTrue())
			ctx.Expect(contains(got, "github.com/getsyntegrity/kit-core/errorschain")).To(specs.BeTrue())
			ctx.Expect(contains(got, "github.com/getsyntegrity/kit-core/fflags")).To(specs.BeTrue())
		})
	})
}

func contains(s []string, x string) bool {
	for _, v := range s {
		if v == x {
			return true
		}
	}
	return false
}

func TestMatchPackage_CoverageGaps(t *testing.T) {
	specs.Describe(t, "matchPackage", func(s *specs.Spec) {
		s.It("exact match returns true", func(ctx *specs.Context) {
			ctx.Expect(matchPackage("internal/domain", "internal/domain")).To(specs.BeTrue())
		})
		s.It("suffix match returns true", func(ctx *specs.Context) {
			ctx.Expect(matchPackage("internal/domain", "github.com/foo/repo/internal/domain")).To(specs.BeTrue())
		})
		s.It("no match returns false", func(ctx *specs.Context) {
			ctx.Expect(matchPackage("internal/domain", "internal/application")).To(specs.BeFalse())
		})
		s.It("glob pattern matches full path", func(ctx *specs.Context) {
			ctx.Expect(matchPackage("internal/*", "internal/domain")).To(specs.BeTrue())
		})
		s.It("glob pattern matches suffix", func(ctx *specs.Context) {
			ctx.Expect(matchPackage("internal/*", "github.com/foo/repo/internal/domain")).To(specs.BeTrue())
		})
		s.It("dot-slash pattern matches go list import path (e.g. .devforge.yml packages)", func(ctx *specs.Context) {
			ctx.Expect(matchPackage("./domain", "github.com/getsyntegrity/kit-core/domain")).To(specs.BeTrue())
			ctx.Expect(matchPackage("./errorschain", "github.com/getsyntegrity/kit-core/errorschain")).To(specs.BeTrue())
			ctx.Expect(matchPackage("./internal/domain", "github.com/foo/repo/internal/domain")).To(specs.BeTrue())
		})
	})
}
