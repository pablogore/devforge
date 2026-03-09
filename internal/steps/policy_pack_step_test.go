package steps

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/pablogore/devforge/internal/application"
	"github.com/pablogore/devforge/internal/testkit"
	"github.com/pablogore/go-specs/specs"
)

func TestPolicyPackStep(t *testing.T) {
	specs.Describe(t, "PolicyPackStep", func(s *specs.Spec) {
		s.It("Run no policies dir succeeds", func(ctx *specs.Context) {
			dir := t.TempDir()
			appCtx := &application.Context{
				StdCtx:      context.Background(),
				Workdir:     dir,
				ProfileName: "go-lib",
				Cmd:         testkit.NewFakeCommandRunner(),
				Git:         &testkit.FakeGitClient{},
				Log:         &testkit.FakeLogger{},
			}
			runErr := (PolicyPackStep{}).Run(appCtx)
			ctx.Expect(runErr).To(specs.BeNil())
		})
		s.It("Run empty policies dir succeeds", func(ctx *specs.Context) {
			dir := t.TempDir()
			err := os.MkdirAll(filepath.Join(dir, ".syntegrity", "policies"), 0o750)
			ctx.Expect(err).To(specs.BeNil())
			appCtx := &application.Context{
				StdCtx:      context.Background(),
				Workdir:     dir,
				ProfileName: "go-lib",
				Cmd:         testkit.NewFakeCommandRunner(),
				Git:         &testkit.FakeGitClient{},
				Log:         &testkit.FakeLogger{},
			}
			runErr := (PolicyPackStep{}).Run(appCtx)
			ctx.Expect(runErr).To(specs.BeNil())
		})
		s.It("Run with passing policy succeeds", func(ctx *specs.Context) {
			dir := t.TempDir()
			policiesDir := filepath.Join(dir, ".syntegrity", "policies")
			err := os.MkdirAll(policiesDir, 0o750)
			ctx.Expect(err).To(specs.BeNil())
			policyYAML := []byte(`name: test
type: architecture
rules:
  forbid_import: internal/nonexistent/pkg
`)
			err = os.WriteFile(filepath.Join(policiesDir, "arch.yaml"), policyYAML, 0o600)
			ctx.Expect(err).To(specs.BeNil())
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("go", []string{"list", "-json", "./..."}, "", nil)
			appCtx := &application.Context{
				StdCtx:      context.Background(),
				Workdir:     dir,
				ProfileName: "go-lib",
				Cmd:         cmd,
				Git:         &testkit.FakeGitClient{},
				Log:         &testkit.FakeLogger{},
			}
			runErr := (PolicyPackStep{}).Run(appCtx)
			ctx.Expect(runErr).To(specs.BeNil())
		})
	})
}
