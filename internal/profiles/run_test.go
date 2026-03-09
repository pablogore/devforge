package profiles

import (
	"strings"
	"testing"
	"time"

	"github.com/pablogore/devforge/internal/application"
	"github.com/pablogore/devforge/internal/config"
	"github.com/pablogore/go-specs/specs"
)

func TestRunSteps(t *testing.T) {
	specs.Describe(t, "RunSteps", func(s *specs.Spec) {
		s.It("returns error for unknown step", func(ctx *specs.Context) {
			err := RunSteps(".", []string{"nonexistent-step-name"})
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "unknown step")).To(specs.BeTrue())
		})
		s.It("succeeds with nil or empty names", func(ctx *specs.Context) {
			ctx.Expect(RunSteps(".", nil)).To(specs.BeNil())
			ctx.Expect(RunSteps(".", []string{})).To(specs.BeNil())
		})
	})
}

func TestProfileEntryPoints(t *testing.T) {
	specs.Describe(t, "profile entry points", func(s *specs.Spec) {
		s.It("RunGoLibPRWithTitle exercises path", func(ctx *specs.Context) {
			dir := t.TempDir()
			err := RunGoLibPRWithTitle(dir, "main", "feat: test")
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
		s.It("RunGoLibDoctor exercises path", func(ctx *specs.Context) {
			dir := t.TempDir()
			_, _ = RunGoLibDoctor(dir)
		})
		s.It("ValidateGoLibVersion exercises path", func(ctx *specs.Context) {
			dir := t.TempDir()
			_, _ = ValidateGoLibVersion(dir)
		})
		s.It("GoLibComplexityThreshold and timeouts are positive", func(ctx *specs.Context) {
			ctx.Expect(GoLibComplexityThreshold() > 0).To(specs.BeTrue())
			ctx.Expect(GoLibStaticAnalysisTimeout() > time.Duration(0)).To(specs.BeTrue())
			_ = GoLibCustomRules()
		})
		s.It("RunGoServicePRWithTitle exercises path", func(ctx *specs.Context) {
			dir := t.TempDir()
			err := RunGoServicePRWithTitle(dir, "main", "feat: test")
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
		s.It("RunGoLibPR exercises path", func(ctx *specs.Context) {
			dir := t.TempDir()
			err := RunGoLibPR(dir, "main")
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
		s.It("RunGoLibRelease exercises path", func(ctx *specs.Context) {
			dir := t.TempDir()
			_, err := RunGoLibRelease(dir)
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
		s.It("RunGoLibPRWithMode with plugin config exercises path", func(ctx *specs.Context) {
			dir := t.TempDir()
			cfg := &config.Config{
				PluginConfig: map[string]config.ExternalPluginCfg{
					"myplugin": {Enabled: true, Params: map[string]interface{}{"k": "v"}},
				},
			}
			err := RunGoLibPRWithMode(dir, "main", "feat: x", application.ModeQuick, cfg)
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
		s.It("RunGoServicePR exercises path", func(ctx *specs.Context) {
			dir := t.TempDir()
			err := RunGoServicePR(dir, "main")
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
		s.It("RunGoServiceRelease exercises path", func(ctx *specs.Context) {
			dir := t.TempDir()
			_, err := RunGoServiceRelease(dir)
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
		s.It("RunGoServiceDoctor exercises path", func(ctx *specs.Context) {
			dir := t.TempDir()
			_, _ = RunGoServiceDoctor(dir)
		})
		s.It("RunGoServicePRWithMode Deep exercises path", func(ctx *specs.Context) {
			dir := t.TempDir()
			err := RunGoServicePRWithMode(dir, "main", "", application.ModeDeep, nil)
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
		s.It("GoServiceComplexityThreshold and timeouts are positive", func(ctx *specs.Context) {
			ctx.Expect(GoServiceComplexityThreshold() > 0).To(specs.BeTrue())
			ctx.Expect(GoServiceStaticAnalysisTimeout() > time.Duration(0)).To(specs.BeTrue())
			_ = GoServiceCustomRules()
		})
	})
}
