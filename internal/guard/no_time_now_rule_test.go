package guard

import (
	"context"
	"errors"
	"testing"

	"github.com/pablogore/devforge/internal/testkit"
	"github.com/pablogore/go-specs/specs"
)

func TestNoTimeNowInDomainRule(t *testing.T) {
	specs.Describe(t, "NoTimeNowInDomainRule", func(s *specs.Spec) {
		s.It("Name returns NoTimeNowInDomain", func(ctx *specs.Context) {
			r := NewNoTimeNowInDomainRule()
			ctx.Expect(r.Name()).ToEqual("NoTimeNowInDomain")
		})
		s.It("Validate covers validation paths", func(ctx *specs.Context) {
			cases := []struct {
				output string
				err    error
				want   error
			}{
				{"", errors.New("exit 1"), nil},
				{"", nil, nil},
				{"internal/domain/foo.go:10: time.Now()", nil, errTimeNowInDomain},
				{"internal/domain/a.go:1: x\ndef.go:2: time.Now()", nil, errTimeNowInDomain},
				{"fatal: not a git repo", errors.New("fatal"), nil},
				{"fatal: internal/domain not found", nil, nil},
			}
			for _, c := range cases {
				runner := testkit.NewFakeCommandRunner()
				runner.Default = &testkit.CommandResult{Stdout: c.output, Err: c.err}
				gCtx := &Context{
					StdCtx:        context.Background(),
					Workdir:       "/wd",
					CommandRunner: runner,
				}
				r := NewNoTimeNowInDomainRule()
				got := r.Validate(gCtx)
				if c.want != nil {
					ctx.Expect(got != nil).To(specs.BeTrue())
					ctx.Expect(errors.Is(got, c.want)).To(specs.BeTrue())
				} else {
					ctx.Expect(got).To(specs.BeNil())
				}
			}
		})
	})
}
