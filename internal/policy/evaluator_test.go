package policy

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/pablogore/devforge/internal/guard"
	"github.com/pablogore/devforge/internal/ports"
	"github.com/pablogore/devforge/internal/testkit"
	"github.com/pablogore/go-specs/specs"
)

// fakePolicyLogger records Info calls for assertions.
type fakePolicyLogger struct {
	lastInfoMsg string
	lastInfoKv  []any
}

func (f *fakePolicyLogger) Debug(string, ...any) {}
func (f *fakePolicyLogger) Warn(string, ...any) {}
func (f *fakePolicyLogger) Info(msg string, args ...any) {
	f.lastInfoMsg = msg
	f.lastInfoKv = args
}
func (f *fakePolicyLogger) Error(string, ...any) {}
func (f *fakePolicyLogger) With(...any) ports.Logger {
	return f
}
func (f *fakePolicyLogger) Sync() error { return nil }

func TestEvaluate(t *testing.T) {
	stdCtx := context.Background()
	specs.Describe(t, "Evaluate", func(s *specs.Spec) {
		s.It("empty policies returns nil", func(ctx *specs.Context) {
			gCtx := &guard.Context{StdCtx: stdCtx, Logger: &fakePolicyLogger{}}
			ctx.Expect(Evaluate(gCtx, nil)).To(specs.BeNil())
			ctx.Expect(Evaluate(gCtx, []Policy{})).To(specs.BeNil())
		})
		s.It("unknown rule is skipped", func(ctx *specs.Context) {
			gCtx := &guard.Context{StdCtx: stdCtx, Logger: &fakePolicyLogger{}}
			policies := []Policy{{
				File: "test.yaml",
				Name: "p1",
				Rules: map[string]interface{}{
					"unknown_rule": "value",
				},
			}}
			ctx.Expect(Evaluate(gCtx, policies)).To(specs.BeNil())
		})
		s.It("forbid_import empty values skipped", func(ctx *specs.Context) {
			gCtx := &guard.Context{StdCtx: stdCtx, Logger: &fakePolicyLogger{}}
			policies := []Policy{{
				File: "test.yaml",
				Rules: map[string]interface{}{
					"forbid_import": []interface{}{},
				},
			}}
			ctx.Expect(Evaluate(gCtx, policies)).To(specs.BeNil())
		})
		s.It("forbid_import when go list fails returns nil", func(ctx *specs.Context) {
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("go", []string{"list", "-json", "./..."}, "", errors.New("go list failed"))
			gCtx := &guard.Context{StdCtx: stdCtx, Workdir: "/wd", CommandRunner: cmd, Logger: &fakePolicyLogger{}}
			policies := []Policy{{
				File: "p.yaml",
				Name: "p",
				Rules: map[string]interface{}{
					"forbid_import": "forbidden/pkg",
				},
			}}
			ctx.Expect(Evaluate(gCtx, policies)).To(specs.BeNil())
		})
		s.It("forbid_import violation returns error", func(ctx *specs.Context) {
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("go", []string{"list", "-json", "./..."}, `{"ImportPath":"x","Imports":["forbidden/pkg"]}`, nil)
			gCtx := &guard.Context{StdCtx: stdCtx, Workdir: "/wd", CommandRunner: cmd, Logger: &fakePolicyLogger{}}
			policies := []Policy{{
				File: "p.yaml",
				Name: "p",
				Rules: map[string]interface{}{
					"forbid_import": "forbidden/pkg",
				},
			}}
			err := Evaluate(gCtx, policies)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "forbid_import")).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "forbidden/pkg")).To(specs.BeTrue())
		})
		s.It("forbid_import severity warning logs and does not fail", func(ctx *specs.Context) {
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("go", []string{"list", "-json", "./..."}, `{"ImportPath":"x","Imports":["forbidden/pkg"]}`, nil)
			log := &fakePolicyLogger{}
			gCtx := &guard.Context{StdCtx: stdCtx, Workdir: "/wd", CommandRunner: cmd, Logger: log}
			policies := []Policy{{
				File:     "p.yaml",
				Name:     "p",
				Severity: "warning",
				Rules: map[string]interface{}{
					"forbid_import": "forbidden/pkg",
				},
			}}
			ctx.Expect(Evaluate(gCtx, policies)).To(specs.BeNil())
			ctx.Expect(log.lastInfoMsg).ToEqual("policy warning (non-fatal)")
		})
		s.It("forbid_time_now empty path skipped", func(ctx *specs.Context) {
			gCtx := &guard.Context{StdCtx: stdCtx, Logger: &fakePolicyLogger{}}
			policies := []Policy{{
				File: "p.yaml",
				Rules: map[string]interface{}{
					"forbid_time_now": "",
				},
			}}
			ctx.Expect(Evaluate(gCtx, policies)).To(specs.BeNil())
		})
		s.It("forbid_time_now no match returns nil", func(ctx *specs.Context) {
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("git", []string{"grep", "-n", "time.Now()", "--", "internal/domain"}, "", nil)
			gCtx := &guard.Context{StdCtx: stdCtx, Workdir: "/wd", CommandRunner: cmd, Logger: &fakePolicyLogger{}}
			policies := []Policy{{
				File: "p.yaml",
				Rules: map[string]interface{}{
					"forbid_time_now": "domain",
				},
			}}
			ctx.Expect(Evaluate(gCtx, policies)).To(specs.BeNil())
		})
		s.It("forbid_time_now violation returns error", func(ctx *specs.Context) {
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("git", []string{"grep", "-n", "time.Now()", "--", "internal/domain"}, "file.go:10: time.Now()", nil)
			gCtx := &guard.Context{StdCtx: stdCtx, Workdir: "/wd", CommandRunner: cmd, Logger: &fakePolicyLogger{}}
			policies := []Policy{{
				File: "p.yaml",
				Rules: map[string]interface{}{
					"forbid_time_now": "domain",
				},
			}}
			err := Evaluate(gCtx, policies)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "time.Now()")).To(specs.BeTrue())
		})
		s.It("ruleValues single string path", func(ctx *specs.Context) {
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("go", []string{"list", "-json", "./..."}, "", nil)
			gCtx := &guard.Context{StdCtx: stdCtx, Workdir: "/wd", CommandRunner: cmd, Logger: &fakePolicyLogger{}}
			policies := []Policy{{
				File: "p.yaml",
				Rules: map[string]interface{}{
					"forbid_import": "pkg",
				},
			}}
			ctx.Expect(Evaluate(gCtx, policies)).To(specs.BeNil())
		})
		s.It("ruleValues nil skipped", func(ctx *specs.Context) {
			gCtx := &guard.Context{StdCtx: stdCtx, Logger: &fakePolicyLogger{}}
			policies := []Policy{{
				File: "p.yaml",
				Rules: map[string]interface{}{
					"forbid_import": nil,
				},
			}}
			ctx.Expect(Evaluate(gCtx, policies)).To(specs.BeNil())
		})
		s.It("ruleValues slice path", func(ctx *specs.Context) {
			cmd := testkit.NewFakeCommandRunner()
			cmd.Enqueue("go", []string{"list", "-json", "./..."}, "", nil)
			cmd.Enqueue("go", []string{"list", "-json", "./..."}, "", nil)
			gCtx := &guard.Context{StdCtx: stdCtx, Workdir: "/wd", CommandRunner: cmd, Logger: &fakePolicyLogger{}}
			policies := []Policy{{
				File: "p.yaml",
				Rules: map[string]interface{}{
					"forbid_import": []interface{}{"a", "b"},
				},
			}}
			ctx.Expect(Evaluate(gCtx, policies)).To(specs.BeNil())
		})
	})
}
