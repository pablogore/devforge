package guard

import (
	"context"
	"errors"
	"testing"

	"github.com/pablogore/devforge/internal/testkit"
	"github.com/pablogore/go-specs/specs"
)

func TestNoCircularImportsRule(t *testing.T) {
	specs.Describe(t, "NoCircularImportsRule", func(s *specs.Spec) {
		s.It("Name returns NoCircularImports", func(ctx *specs.Context) {
			r := NewNoCircularImportsRule()
			ctx.Expect(r.Name()).ToEqual("NoCircularImports")
		})
		s.It("Validate covers validation paths", func(ctx *specs.Context) {
			cases := []struct {
				name   string
				output string
				err    error
				want   error
			}{
				{"success", "pkg list", nil, nil},
				{"cycle in error", "", errors.New("import cycle not allowed"), errCircularImport},
				{"cycle in output", "import cycle not allowed in foo", nil, errCircularImport},
			}
			for _, tc := range cases {
				runner := testkit.NewFakeCommandRunner()
				runner.Default = &testkit.CommandResult{Stdout: tc.output, Err: tc.err}
				gCtx := &Context{
					StdCtx:        context.Background(),
					Workdir:       "/wd",
					CommandRunner: runner,
				}
				r := NewNoCircularImportsRule()
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
