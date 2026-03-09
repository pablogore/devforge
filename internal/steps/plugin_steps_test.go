package steps

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pablogore/devforge/internal/adapters/clock"
	"github.com/pablogore/devforge/internal/application"
	"github.com/pablogore/devforge/internal/testkit"
	"github.com/pablogore/go-specs/specs"
)

type pluginTestStub struct {
	name string
	err  error
}

func (s *pluginTestStub) Name() string                     { return s.name }
func (s *pluginTestStub) Run(_ *application.Context) error { return s.err }

func TestPluginStep(t *testing.T) {
	specs.Describe(t, "PluginStep", func(s *specs.Spec) {
		s.It("Name returns plugin name", func(ctx *specs.Context) {
			st := NewPluginStep("my-plugin", "true")
			ctx.Expect(st.Name()).ToEqual("my-plugin")
		})
		s.It("Run executes command and succeeds", func(ctx *specs.Context) {
			dir := t.TempDir()
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("bash", []string{"-c", "true"}, "", nil)
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{
				StdCtx:  context.Background(),
				Cmd:     cmd,
				Workdir: dir,
				Log:     log,
				Clock:   clock.NewRealClock(),
			}
			err := NewPluginStep("my-plugin", "true").Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(log.InfoCalls >= 1).To(specs.BeTrue())
		})
	})
}

func TestExternalPluginStep(t *testing.T) {
	specs.Describe(t, "ExternalPluginStep", func(s *specs.Spec) {
		s.It("Name returns plugin-<name>", func(ctx *specs.Context) {
			st := &ExternalPluginStep{name: "security"}
			ctx.Expect(st.Name()).ToEqual("plugin-security")
		})
		s.It("Run disabled in config skips", func(ctx *specs.Context) {
			st := &ExternalPluginStep{name: "disabled-plugin"}
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{
				StdCtx:  context.Background(),
				Log:     log,
				Workdir: t.TempDir(),
				ExternalPluginConfig: map[string]application.ExternalPluginConfig{
					"disabled-plugin": {Enabled: false},
				},
			}
			err := st.Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(log.LastInfoMsg).ToEqual("plugin skipped (disabled in config)")
		})
		s.It("Run enabled executes plugin", func(ctx *specs.Context) {
			dir := t.TempDir()
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("forge-plugin-myplugin", []string{}, "", nil)
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Cmd: cmd, Log: log, Workdir: dir}
			err := (&ExternalPluginStep{name: "myplugin"}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("Run with params calls plugin", func(ctx *specs.Context) {
			dir := t.TempDir()
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("forge-plugin-myplugin", []string{}, "", nil)
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{
				StdCtx:  context.Background(),
				Cmd:     cmd,
				Log:     log,
				Workdir: dir,
				ExternalPluginConfig: map[string]application.ExternalPluginConfig{
					"myplugin": {Enabled: true, Params: map[string]interface{}{"severity": "high"}},
				},
			}
			err := (&ExternalPluginStep{name: "myplugin"}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(cmd.WasCalled("forge-plugin-myplugin")).To(specs.BeTrue())
		})
		s.It("Run plugin fails returns error", func(ctx *specs.Context) {
			dir := t.TempDir()
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("forge-plugin-myplugin", []string{}, "plugin stderr", errors.New("exit 1"))
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Cmd: cmd, Log: log, Workdir: dir}
			err := (&ExternalPluginStep{name: "myplugin"}).Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "myplugin")).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "plugin stderr")).To(specs.BeTrue())
		})
	})
}

func TestGolangCILintStep_Name(t *testing.T) {
	specs.Describe(t, "GolangCILintStep", func(s *specs.Spec) {
		s.It("Name returns golangci-lint", func(ctx *specs.Context) {
			ctx.Expect(GolangCILintStep{}.Name()).ToEqual("golangci-lint")
		})
	})
}

func TestSequentialGroupStep(t *testing.T) {
	specs.Describe(t, "SequentialGroupStep", func(s *specs.Spec) {
		s.It("Name and Run", func(ctx *specs.Context) {
			stub := &pluginTestStub{name: "b", err: nil}
			seq := NewSequentialGroupStep("seq1", []application.Step{stub})
			ctx.Expect(seq.Name()).ToEqual("seq1")
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Log: log, Workdir: t.TempDir()}
			err := seq.Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("Run second step fails propagates error", func(ctx *specs.Context) {
			errFail := errors.New("second failed")
			stubs := []application.Step{
				&pluginTestStub{name: "a", err: nil},
				&pluginTestStub{name: "b", err: errFail},
			}
			seq := NewSequentialGroupStep("seq1", stubs)
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Log: log, Workdir: t.TempDir()}
			err := seq.Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(err == errFail || errors.Is(err, errFail)).To(specs.BeTrue())
		})
	})
}

func TestParallelGroupStep_Name_and_Run(t *testing.T) {
	specs.Describe(t, "ParallelGroupStep name and run", func(s *specs.Spec) {
		s.It("Name and Run", func(ctx *specs.Context) {
			stub := &pluginTestStub{name: "a", err: nil}
			p := NewParallelGroupStep("group1", []application.Step{stub})
			ctx.Expect(p.Name()).ToEqual("group1")
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Log: log, Workdir: t.TempDir()}
			err := p.Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
		})
	})
}

func TestDiscoveredPluginSteps(t *testing.T) {
	specs.Describe(t, "DiscoveredPluginSteps", func(s *specs.Spec) {
		s.It("returns steps when PATH has plugin", func(ctx *specs.Context) {
			dir := t.TempDir()
			pluginPath := filepath.Join(dir, "forge-plugin-foo")
			ctx.Expect(os.WriteFile(pluginPath, []byte("#!/bin/sh\nexit 0"), 0o755)).To(specs.BeNil())
			origPath := os.Getenv("PATH")
			origPluginExec := os.Getenv("DEVFORGE_PLUGIN_EXECUTION")
			defer func() {
				_ = os.Setenv("PATH", origPath)
				_ = os.Setenv("DEVFORGE_PLUGIN_EXECUTION", origPluginExec)
			}()
			ctx.Expect(os.Setenv("PATH", dir)).To(specs.BeNil())
			ctx.Expect(os.Unsetenv("DEVFORGE_PLUGIN_EXECUTION")).To(specs.BeNil())

			steps := DiscoveredPluginSteps()
			ctx.Expect(len(steps)).ToEqual(1)
			ctx.Expect(steps[0].Name()).ToEqual("plugin-foo")
		})
	})
}
