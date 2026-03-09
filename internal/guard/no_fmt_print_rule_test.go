package guard

import (
	"context"
	"errors"
	"testing"

	"github.com/pablogore/devforge/internal/testkit"
	"github.com/pablogore/go-specs/specs"
)

func TestNoFmtPrintOutsideCmdRule(t *testing.T) {
	specs.Describe(t, "NoFmtPrintOutsideCmdRule", func(s *specs.Spec) {
		s.It("Name returns NoFmtPrintOutsideCmd", func(ctx *specs.Context) {
			r := NewNoFmtPrintOutsideCmdRule()
			ctx.Expect(r.Name()).ToEqual("NoFmtPrintOutsideCmd")
		})
		s.It("Validate covers validation paths", func(ctx *specs.Context) {
			cases := []struct {
				output string
				err    error
				want   error
			}{
				{"", errors.New("exit 1"), nil},
				{"", nil, nil},
				{"cmd/main.go:5: fmt.Println(\"x\")", nil, nil},
				{"internal/profiles/foo.go:1: fmt.Printf(\"x\")", nil, errFmtPrintOutsideCmd},
				{"cmd/a.go:1: x\ninternal/profiles/b.go:2: fmt.Println()", nil, errFmtPrintOutsideCmd},
				{"fatal: not a repo", nil, nil},
				{"internal/foo/bar_test.go:10: fmt.Println()", nil, nil},
				{"internal/guard/other.go:1: fmt.Println()", nil, nil},
				{"internal/foo.go:1: fmt.Fprintf(os.Stderr, \"x\")", nil, nil},
				{"internal\\profiles\\p.go:1: fmt.Printf(\"x\")", nil, errFmtPrintOutsideCmd},
			}
			for _, c := range cases {
				runner := testkit.NewFakeCommandRunner()
				runner.Default = &testkit.CommandResult{Stdout: c.output, Err: c.err}
				gCtx := &Context{
					StdCtx:        context.Background(),
					Workdir:       "/wd",
					CommandRunner: runner,
				}
				r := NewNoFmtPrintOutsideCmdRule()
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

func Test_pathFromGrepLine(t *testing.T) {
	specs.Describe(t, "pathFromGrepLine", func(s *specs.Spec) {
		s.It("covers path extraction paths", func(ctx *specs.Context) {
			cases := []struct {
				line string
				want string
			}{
				{"cmd/main.go:10: content", "cmd/main.go"},
				{"internal/foo.go:1: x", "internal/foo.go"},
				{"no-colon", ""},
				{"", ""},
			}
			for _, c := range cases {
				got := pathFromGrepLine(c.line)
				ctx.Expect(got).ToEqual(c.want)
			}
		})
	})
}

func Test_underCmd(t *testing.T) {
	specs.Describe(t, "underCmd", func(s *specs.Spec) {
		s.It("covers path under cmd/ paths", func(ctx *specs.Context) {
			cases := []struct {
				path string
				want bool
			}{
				{"cmd/main.go", true},
				{"cmd/foo/bar.go", true},
				{"internal/cmd/bar.go", true},
				{"internal/profiles/foo.go", false},
				{"domain/foo.go", false},
			}
			for _, c := range cases {
				got := underCmd(c.path)
				ctx.Expect(got).ToEqual(c.want)
			}
		})
	})
}
