package steps

import (
	"context"
	"errors"
	"testing"

	"github.com/pablogore/devforge/internal/application"
	"github.com/pablogore/devforge/internal/testkit"
	"github.com/pablogore/go-specs/specs"
)

type mockStep struct {
	name string
	run  func(*application.Context) error
}

func (m mockStep) Name() string { return m.name }
func (m mockStep) Run(ctx *application.Context) error {
	if m.run != nil {
		return m.run(ctx)
	}
	return nil
}

func TestParallelGroupStep(t *testing.T) {
	specs.Describe(t, "ParallelGroupStep", func(s *specs.Spec) {
		s.It("Name returns group name", func(ctx *specs.Context) {
			st := NewParallelGroupStep("my-group", nil)
			ctx.Expect(st.Name()).ToEqual("my-group")
		})
		s.It("Run all succeed group succeeds", func(ctx *specs.Context) {
			log := &testkit.FakeLogger{}
			steps := []application.Step{
				mockStep{name: "a", run: func(*application.Context) error { return nil }},
				mockStep{name: "b", run: func(*application.Context) error { return nil }},
			}
			appCtx := &application.Context{StdCtx: context.Background(), Log: log, Workdir: "/wd"}
			err := NewParallelGroupStep("test", steps).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("Run one fails group fails", func(ctx *specs.Context) {
			wantErr := errors.New("step failed")
			log := &testkit.FakeLogger{}
			steps := []application.Step{
				mockStep{name: "a", run: func(*application.Context) error { return nil }},
				mockStep{name: "b", run: func(*application.Context) error { return wantErr }},
				mockStep{name: "c", run: func(*application.Context) error { return nil }},
			}
			appCtx := &application.Context{StdCtx: context.Background(), Log: log, Workdir: "/wd"}
			err := NewParallelGroupStep("test", steps).Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(err == wantErr || errors.Is(err, wantErr)).To(specs.BeTrue())
		})
		s.It("Run multiple fail returns one error", func(ctx *specs.Context) {
			errA := errors.New("err A")
			errB := errors.New("err B")
			log := &testkit.FakeLogger{}
			steps := []application.Step{
				mockStep{name: "a", run: func(*application.Context) error { return errA }},
				mockStep{name: "b", run: func(*application.Context) error { return errB }},
			}
			appCtx := &application.Context{StdCtx: context.Background(), Log: log, Workdir: "/wd"}
			err := NewParallelGroupStep("test", steps).Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(err == errA || err == errB).To(specs.BeTrue())
		})
		s.It("Run empty slice succeeds", func(ctx *specs.Context) {
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Log: log, Workdir: "/wd"}
			err := NewParallelGroupStep("empty", nil).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("Run concurrent steps does not panic", func(ctx *specs.Context) {
			log := &testkit.FakeLogger{}
			n := 10
			steps := make([]application.Step, n)
			for i := 0; i < n; i++ {
				steps[i] = mockStep{name: "s", run: func(*application.Context) error { return nil }}
			}
			appCtx := &application.Context{StdCtx: context.Background(), Log: log, Workdir: "/wd"}
			err := NewParallelGroupStep("concurrent", steps).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
		})
	})
}
