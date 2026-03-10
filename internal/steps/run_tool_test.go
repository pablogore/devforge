package steps

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/pablogore/devforge/internal/application"
	"github.com/pablogore/devforge/internal/domain"
	"github.com/pablogore/devforge/internal/testkit"
	"github.com/pablogore/go-specs/specs"
)

func TestIsToolCrash(t *testing.T) {
	specs.Describe(t, "isToolCrash", func(s *specs.Spec) {
		s.It("returns false on non-zero exit with no panic in output (lint failure)", func(ctx *specs.Context) {
			ctx.Expect(isToolCrash("", errors.New("exit 1"))).To(specs.BeFalse())
			ctx.Expect(isToolCrash("file.go:10: lint error", errors.New("exit 1"))).To(specs.BeFalse())
		})
		s.It("returns true when output contains panic:", func(ctx *specs.Context) {
			ctx.Expect(isToolCrash("panic: runtime error", nil)).To(specs.BeTrue())
			ctx.Expect(isToolCrash("PANIC: something", nil)).To(specs.BeTrue())
		})
		s.It("returns true when output contains fatal error", func(ctx *specs.Context) {
			ctx.Expect(isToolCrash("fatal error: something", nil)).To(specs.BeTrue())
			ctx.Expect(isToolCrash("Fatal error", nil)).To(specs.BeTrue())
		})
		s.It("returns false on success with clean output", func(ctx *specs.Context) {
			ctx.Expect(isToolCrash("", nil)).To(specs.BeFalse())
			ctx.Expect(isToolCrash("file.go:10: some lint", nil)).To(specs.BeFalse())
		})
	})
}

func TestRunToolWithRetry(t *testing.T) {
	specs.Describe(t, "runToolWithRetry", func(s *specs.Spec) {
		s.It("succeeds on first run", func(ctx *specs.Context) {
			dir := t.TempDir()
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("tool", []string{"run"}, "", nil)
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{
				StdCtx:  context.Background(),
				Cmd:     cmd,
				Workdir: dir,
				Log:     log,
				Clock:   testkit.NewFakeClock().Clock(),
			}
			out, err := runToolWithRetry(appCtx, "tool", 2, "run")
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(out).ToEqual("")
		})
		s.It("fails after retry when both runs crash", func(ctx *specs.Context) {
			dir := t.TempDir()
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("tool", []string{"run"}, "panic: crash", errors.New("exit 1"))
			log := &testkit.FakeLogger{RecordInfoHistory: true}
			appCtx := &application.Context{
				StdCtx:  context.Background(),
				Cmd:     cmd,
				Workdir: dir,
				Log:     log,
				Clock:   testkit.NewFakeClock().Clock(),
			}
			out, err := runToolWithRetry(appCtx, "tool", 2, "run")
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(out).ToEqual("panic: crash")
			ctx.Expect(errors.Is(err, domain.ErrToolCrash)).To(specs.BeTrue())
			// Machine-readable TOOL START and retry line present
			var hasToolStart, hasRetry bool
			for _, c := range log.InfoHistory {
				if c.Msg == "[devforge] TOOL START" {
					hasToolStart = true
				}
				if strings.HasPrefix(c.Msg, "[devforge] retrying (") {
					hasRetry = true
				}
			}
			ctx.Expect(hasToolStart).To(specs.BeTrue())
			ctx.Expect(hasRetry).To(specs.BeTrue())
		})
	})
}
