package steps

import (
	"context"
	"errors"
	"testing"

	"github.com/pablogore/devforge/internal/application"
	"github.com/pablogore/devforge/internal/guard"
	"github.com/pablogore/devforge/internal/testkit"
	"github.com/pablogore/go-specs/specs"
)

type mockRule struct {
	name string
	err  error
}

func (m *mockRule) Name() string                    { return m.name }
func (m *mockRule) Validate(_ *guard.Context) error { return m.err }

func TestArchitecturalGuardStep(t *testing.T) {
	specs.Describe(t, "ArchitecturalGuardStep", func(s *specs.Spec) {
		s.It("Name returns architectural-guard", func(ctx *specs.Context) {
			s := NewArchitecturalGuardStep(guard.DefaultRules())
			ctx.Expect(s.Name()).ToEqual("architectural-guard")
		})
		s.It("Run empty rules succeeds", func(ctx *specs.Context) {
			s := NewArchitecturalGuardStep([]guard.ArchitecturalRule{})
			appCtx := &application.Context{
				StdCtx:      context.Background(),
				Cmd:         testkit.NewFakeCommandRunner(),
				Git:         &testkit.FakeGitClient{},
				Log:         &testkit.FakeLogger{},
				Workdir:     "/wd",
				ProfileName: "go-lib",
			}
			err := s.Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("Run rule fails returns error", func(ctx *specs.Context) {
			failRule := &mockRule{name: "fail-rule", err: errors.New("validation failed")}
			s := NewArchitecturalGuardStep([]guard.ArchitecturalRule{failRule})
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{
				StdCtx:      context.Background(),
				Cmd:         testkit.NewFakeCommandRunner(),
				Git:         &testkit.FakeGitClient{},
				Log:         log,
				Workdir:     "/wd",
				ProfileName: "go-lib",
			}
			err := s.Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(err.Error()).ToEqual("validation failed")
		})
	})
}
