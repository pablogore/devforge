package steps

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pablogore/devforge/internal/application"
	"github.com/pablogore/devforge/internal/testkit"
	"github.com/pablogore/go-specs/specs"
)

type blockingStep struct {
	name string
}

func (s blockingStep) Name() string { return s.name }

func (s blockingStep) Run(ctx *application.Context) error {
	<-ctx.StdCtx.Done()
	return ctx.StdCtx.Err()
}

type stubStep struct {
	name string
	err  error
}

func (s stubStep) Name() string                     { return s.name }
func (s stubStep) Run(_ *application.Context) error { return s.err }

type commandStep struct{}

func (commandStep) Name() string { return "command" }

func (commandStep) Run(ctx *application.Context) error {
	_, err := ctx.Cmd.RunCombinedOutput(ctx.StdCtx, ctx.Workdir, "echo", "ok")
	return err
}

func TestTimeoutGroupStep(t *testing.T) {
	specs.Describe(t, "TimeoutGroupStep", func(s *specs.Spec) {
		s.It("returns timeout error when step exceeds timeout", func(ctx *specs.Context) {
			log := &testkit.FakeLogger{}
			blocking := &blockingStep{name: "blocking"}
			wrapped := NewTimeoutGroupStep("group", 1*time.Millisecond, blocking)
			appCtx := &application.Context{StdCtx: context.Background(), Log: log}
			err := wrapped.Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, ErrParallelGroupTimeout)).To(specs.BeTrue())
		})
		s.It("succeeds when step completes before timeout", func(ctx *specs.Context) {
			log := &testkit.FakeLogger{}
			stub := stubStep{name: "stub", err: nil}
			wrapped := NewTimeoutGroupStep("group", 10*time.Second, stub)
			appCtx := &application.Context{StdCtx: context.Background(), Log: log}
			err := wrapped.Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("propagates step error", func(ctx *specs.Context) {
			log := &testkit.FakeLogger{}
			wantErr := errors.New("step failed")
			stub := stubStep{name: "stub", err: wantErr}
			wrapped := NewTimeoutGroupStep("group", 10*time.Second, stub)
			appCtx := &application.Context{StdCtx: context.Background(), Log: log}
			err := wrapped.Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(err == wantErr || errors.Is(err, wantErr)).To(specs.BeTrue())
		})
		s.It("command step completes successfully", func(ctx *specs.Context) {
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("echo", []string{"ok"}, "ok", nil)
			log := &testkit.FakeLogger{}
			runStep := commandStep{}
			wrapped := NewTimeoutGroupStep("group", 5*time.Second, runStep)
			appCtx := &application.Context{StdCtx: context.Background(), Cmd: cmd, Log: log, Workdir: "/wd"}
			err := wrapped.Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(cmd.WasCalled("echo", "ok")).To(specs.BeTrue())
		})
	})
}
