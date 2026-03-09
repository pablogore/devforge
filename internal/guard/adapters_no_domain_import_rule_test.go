package guard

import (
	"context"
	"errors"
	"testing"

	"github.com/pablogore/devforge/internal/testkit"
	"github.com/pablogore/go-specs/specs"
)

func TestAdaptersMustNotImportDomainRule(t *testing.T) {
	validJSON := `{"ImportPath":"github.com/foo/internal/adapters/exec","Imports":["os/exec","github.com/foo/internal/ports"]}`
	invalidJSON := `{"ImportPath":"github.com/foo/internal/adapters/exec","Imports":["github.com/foo/internal/domain"]}`

	specs.Describe(t, "AdaptersMustNotImportDomainRule", func(s *specs.Spec) {
		s.It("Name returns AdaptersMustNotImportDomain", func(ctx *specs.Context) {
			r := NewAdaptersMustNotImportDomainRule()
			ctx.Expect(r.Name()).ToEqual("AdaptersMustNotImportDomain")
		})
		s.It("Validate covers validation paths", func(ctx *specs.Context) {
			cases := []struct {
				name   string
				output string
				err    error
				want   error
			}{
				{"no output and error", "", errors.New("exit 1"), nil},
				{"valid imports", validJSON, nil, nil},
				{"imports domain", invalidJSON, nil, errAdaptersImportDomain},
			}
			for _, tc := range cases {
				runner := testkit.NewFakeCommandRunner()
				runner.Default = &testkit.CommandResult{Stdout: tc.output, Err: tc.err}
				gCtx := &Context{
					StdCtx:        context.Background(),
					Workdir:       "/wd",
					CommandRunner: runner,
				}
				r := NewAdaptersMustNotImportDomainRule()
				got := r.Validate(gCtx)
				if tc.want != nil {
					ctx.Expect(got != nil).To(specs.BeTrue())
					ctx.Expect(got == tc.want).To(specs.BeTrue())
				} else {
					ctx.Expect(got).To(specs.BeNil())
				}
			}
		})
	})
}
