package guard

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/pablogore/devforge/internal/testkit"
	"github.com/pablogore/go-specs/specs"
)

func TestForbidImport(t *testing.T) {
	specs.Describe(t, "ForbidImport", func(s *specs.Spec) {
		s.It("returns nil when forbiddenSegment is empty", func(ctx *specs.Context) {
			runner := testkit.NewFakeCommandRunner()
			gCtx := &Context{
				StdCtx:        context.Background(),
				Workdir:       "/wd",
				CommandRunner: runner,
			}
			err := ForbidImport(gCtx, "./...", "")
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("returns nil when go list fails or output empty", func(ctx *specs.Context) {
			runner := testkit.NewFakeCommandRunner()
			runner.Default = &testkit.CommandResult{Stdout: "", Err: errors.New("exit 1")}
			gCtx := &Context{StdCtx: context.Background(), Workdir: "/wd", CommandRunner: runner}
			err := ForbidImport(gCtx, "./...", "internal/domain")
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("returns error when imports contain forbidden segment", func(ctx *specs.Context) {
			out := `{"ImportPath":"pkg","Imports":["internal/domain"]}`
			runner := testkit.NewFakeCommandRunner()
			runner.Default = &testkit.CommandResult{Stdout: out, Err: nil}
			gCtx := &Context{StdCtx: context.Background(), Workdir: "/wd", CommandRunner: runner}
			err := ForbidImport(gCtx, "./...", "internal/domain")
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "forbidden")).To(specs.BeTrue())
		})
		s.It("returns nil when imports do not contain segment", func(ctx *specs.Context) {
			out := `{"ImportPath":"pkg","Imports":["os","internal/ports"]}`
			runner := testkit.NewFakeCommandRunner()
			runner.Default = &testkit.CommandResult{Stdout: out, Err: nil}
			gCtx := &Context{StdCtx: context.Background(), Workdir: "/wd", CommandRunner: runner}
			err := ForbidImport(gCtx, "./...", "internal/domain")
			ctx.Expect(err).To(specs.BeNil())
		})
	})
}

func TestForbidTimeNow(t *testing.T) {
	specs.Describe(t, "ForbidTimeNow", func(s *specs.Spec) {
		s.It("returns nil when path is empty", func(ctx *specs.Context) {
			runner := testkit.NewFakeCommandRunner()
			gCtx := &Context{StdCtx: context.Background(), Workdir: "/wd", CommandRunner: runner}
			err := ForbidTimeNow(gCtx, "")
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("normalizes domain to internal/domain", func(ctx *specs.Context) {
			runner := testkit.NewFakeCommandRunner()
			runner.Default = &testkit.CommandResult{Stdout: "", Err: nil}
			gCtx := &Context{StdCtx: context.Background(), Workdir: "/wd", CommandRunner: runner}
			err := ForbidTimeNow(gCtx, "domain")
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("returns nil when output contains fatal", func(ctx *specs.Context) {
			runner := testkit.NewFakeCommandRunner()
			runner.Default = &testkit.CommandResult{Stdout: "fatal: not a git repo", Err: nil}
			gCtx := &Context{StdCtx: context.Background(), Workdir: "/wd", CommandRunner: runner}
			err := ForbidTimeNow(gCtx, "internal/domain")
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("returns error when time.Now() match found", func(ctx *specs.Context) {
			runner := testkit.NewFakeCommandRunner()
			runner.Default = &testkit.CommandResult{Stdout: "version.go:42: time.Now()", Err: nil}
			gCtx := &Context{StdCtx: context.Background(), Workdir: "/wd", CommandRunner: runner}
			err := ForbidTimeNow(gCtx, "internal/domain")
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "time.Now()")).To(specs.BeTrue())
		})
		s.It("returns nil when output empty", func(ctx *specs.Context) {
			runner := testkit.NewFakeCommandRunner()
			runner.Default = &testkit.CommandResult{Stdout: "", Err: nil}
			gCtx := &Context{StdCtx: context.Background(), Workdir: "/wd", CommandRunner: runner}
			err := ForbidTimeNow(gCtx, "internal/domain")
			ctx.Expect(err).To(specs.BeNil())
		})
	})
}
