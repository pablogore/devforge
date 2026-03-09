package steps

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/pablogore/devforge/internal/application"
	"github.com/pablogore/devforge/internal/testkit"
	"github.com/pablogore/go-specs/specs"
)

func TestCheckGoreleaserVersionStep(t *testing.T) {
	specs.Describe(t, "CheckGoreleaserVersionStep", func(s *specs.Spec) {
		s.It("Name returns check-goreleaser-version", func(ctx *specs.Context) {
			st := NewCheckGoreleaserVersionStep()
			ctx.Expect(st.Name()).ToEqual("check-goreleaser-version")
		})
		s.It("Run returns error when goreleaser command fails", func(ctx *specs.Context) {
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("goreleaser", []string{"--version"}, "", errors.New("exec: not found"))
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Cmd: cmd, Log: log, Workdir: "/wd"}
			err := NewCheckGoreleaserVersionStep().Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "goreleaser not available")).To(specs.BeTrue())
		})
		s.It("Run returns error when goreleaser output is empty", func(ctx *specs.Context) {
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("goreleaser", []string{"--version"}, "", nil)
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Cmd: cmd, Log: log, Workdir: "/wd"}
			err := NewCheckGoreleaserVersionStep().Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "goreleaser version output empty")).To(specs.BeTrue())
		})
		s.It("Run succeeds when goreleaser returns valid version", func(ctx *specs.Context) {
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("goreleaser", []string{"--version"}, "goreleaser version 2.3.0\n", nil)
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Cmd: cmd, Log: log, Workdir: "/wd"}
			err := NewCheckGoreleaserVersionStep().Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(log.LastInfoMsg).ToEqual("Goreleaser detected")
		})
		s.It("Run logs version when goreleaser returns version string", func(ctx *specs.Context) {
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("goreleaser", []string{"--version"}, "v2.3.0", nil)
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Cmd: cmd, Log: log, Workdir: "/wd"}
			err := NewCheckGoreleaserVersionStep().Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(log.LastInfoMsg).ToEqual("Goreleaser detected")
		})
	})
}
