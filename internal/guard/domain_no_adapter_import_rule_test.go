package guard

import (
	"context"
	"errors"
	"testing"

	"github.com/pablogore/devforge/internal/testkit"
	"github.com/pablogore/go-specs/specs"
)

func TestDomainMustNotImportAdaptersRule(t *testing.T) {
	validJSON := `{"ImportPath":"github.com/foo/internal/domain","Imports":["fmt","strings"]}`
	invalidJSON := `{"ImportPath":"github.com/foo/internal/domain","Imports":["fmt","github.com/foo/internal/adapters/exec"]}`

	specs.Describe(t, "DomainMustNotImportAdaptersRule", func(s *specs.Spec) {
		s.It("Name returns DomainMustNotImportAdapters", func(ctx *specs.Context) {
			r := NewDomainMustNotImportAdaptersRule()
			ctx.Expect(r.Name()).ToEqual("DomainMustNotImportAdapters")
		})
		s.It("Validate covers validation paths", func(ctx *specs.Context) {
			cases := []struct {
				name   string
				output string
				err    error
				want   error
			}{
				{"no output and error (path missing)", "", errors.New("exit 1"), nil},
				{"valid imports", validJSON, nil, nil},
				{"imports adapters", invalidJSON, nil, errDomainImportsAdapters},
			}
			for _, tc := range cases {
				runner := testkit.NewFakeCommandRunner()
				runner.Default = &testkit.CommandResult{Stdout: tc.output, Err: tc.err}
				gCtx := &Context{
					StdCtx:        context.Background(),
					Workdir:       "/wd",
					CommandRunner: runner,
				}
				r := NewDomainMustNotImportAdaptersRule()
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
