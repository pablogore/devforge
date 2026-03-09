package application_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/pablogore/devforge/internal/application"
	"github.com/pablogore/devforge/internal/config"
	"github.com/pablogore/devforge/internal/coverage"
	"github.com/pablogore/devforge/internal/domain"
	"github.com/pablogore/devforge/internal/guard"
	"github.com/pablogore/devforge/internal/steps"
	"github.com/pablogore/devforge/internal/testkit"
	"github.com/pablogore/go-specs/specs"
)

func goLibPRStepsForTest(_ int) []application.Step {
	return []application.Step{
		steps.GoModTidyStep{},
		steps.ConventionalCommitStep{},
		steps.NewArchitecturalGuardStep(guard.DefaultRules()),
		steps.NewTimeoutGroupStep("static-analysis", 2*time.Minute, steps.GolangCILintStep{}),
		steps.GovulnCheckStep{},
		steps.GoTestStep{},
		steps.CoverageStep{},
	}
}

// prSuccessResponses is the sequence of (out, err) for a full PR success (go mod tidy, test -e, greps, go list x5, golangci, govulncheck, go test, go tool cover).
func prSuccessResponses() []testkit.CmdResponse {
	return []testkit.CmdResponse{
		{Out: "", Err: nil},                                           // go mod tidy
		{Out: "", Err: errors.New("no go.sum")},                     // test -e go.sum
		{Out: "", Err: errors.New("exit 1")},                         // git grep time.Now
		{Out: "", Err: errors.New("exit 1")},                         // git grep fmt
		{Out: "", Err: nil},                                           // go list -deps
		{Out: `{"ImportPath":"x","Imports":[]}`, Err: nil},          // go list domain
		{Out: `{"ImportPath":"x","Imports":[]}`, Err: nil},          // go list domain (2nd)
		{Out: `{"ImportPath":"x","Imports":[]}`, Err: nil},          // adapters
		{Out: `{"ImportPath":"x","Imports":[]}`, Err: nil},          // application
		{Out: `{"ImportPath":"x","Imports":[]}`, Err: nil},          // ports
		{Out: "", Err: nil},                                           // golangci-lint
		{Out: "", Err: nil},                                           // govulncheck
		{Out: "", Err: nil},                                           // go test
		{Out: "total: (statements) 96.0%", Err: nil},                 // go tool cover
	}
}

func TestPRUsecase(t *testing.T) {
	workdir := "/wd"
	profile := "go-lib"
	pipeline := application.Pipeline{Name: "pr", Steps: goLibPRStepsForTest(15)}

	specs.Describe(t, "PRUsecase", func(s *specs.Spec) {
		s.It("Run success completes with no error", func(ctx *specs.Context) {
			cmd := &testkit.FakeCommandRunner{Responses: prSuccessResponses()}
			git := &testkit.FakeGitClient{}
			env := testkit.NewFakeEnvProvider(map[string]string{"PR_TITLE": "feat: add feature"})
			log := &testkit.FakeLogger{}
			u := application.NewPRUsecase(cmd, git, env, 95.0, log, testkit.NewFakeClock().Clock(), profile, pipeline)
			err := u.Run(workdir, "main")
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("RunWithTitleAndPluginConfig with cfg and no policies logs profile default", func(ctx *specs.Context) {
			cmd := &testkit.FakeCommandRunner{Responses: prSuccessResponses()}
			git := &testkit.FakeGitClient{}
			env := testkit.NewFakeEnvProvider(nil)
			log := &testkit.FakeLogger{}
			u := application.NewPRUsecase(cmd, git, env, 95.0, log, testkit.NewFakeClock().Clock(), profile, pipeline)
			cfg := &config.Config{}
			err := u.RunWithTitleAndPluginConfig(workdir, "main", "feat: x", nil, cfg)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(strings.Contains(log.LastInfoMsg, "completed") || log.InfoCalls >= 1).To(specs.BeTrue())
		})
		s.It("RunWithTitleAndPluginConfig with cfg.Policies.Coverage applies threshold", func(ctx *specs.Context) {
			cmd := &testkit.FakeCommandRunner{Responses: prSuccessResponses()}
			git := &testkit.FakeGitClient{}
			env := testkit.NewFakeEnvProvider(nil)
			log := &testkit.FakeLogger{}
			u := application.NewPRUsecase(cmd, git, env, 95.0, log, testkit.NewFakeClock().Clock(), profile, pipeline)
			cfg := &config.Config{
				Policies: &config.Policies{
					Coverage: &config.CoveragePolicy{Threshold: 90, Packages: []string{}},
				},
			}
			err := u.RunWithTitleAndPluginConfig(workdir, "main", "feat: x", nil, cfg)
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("RunWithTitleAndPluginConfig with invalid coverage patterns returns error", func(ctx *specs.Context) {
			cmd := &testkit.FakeCommandRunner{Responses: prSuccessResponses()}
			git := &testkit.FakeGitClient{}
			env := testkit.NewFakeEnvProvider(nil)
			log := &testkit.FakeLogger{}
			u := application.NewPRUsecase(cmd, git, env, 95.0, log, testkit.NewFakeClock().Clock(), profile, pipeline)
			cfg := &config.Config{
				Policies: &config.Policies{
					Coverage: &config.CoveragePolicy{Threshold: 90, Packages: []string{"*", "internal/domain"}},
				},
			}
			err := u.RunWithTitleAndPluginConfig(workdir, "main", "feat: x", nil, cfg)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, coverage.ErrWildcardWithOthers)).To(specs.BeTrue())
		})
		s.It("RunWithTitleAndPluginConfig ResolveCoveragePackages error returns error", func(ctx *specs.Context) {
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("go", []string{"list", "./..."}, "", errors.New("go list failed"))
			git := &testkit.FakeGitClient{}
			env := testkit.NewFakeEnvProvider(nil)
			log := &testkit.FakeLogger{}
			u := application.NewPRUsecase(cmd, git, env, 95.0, log, testkit.NewFakeClock().Clock(), profile, pipeline)
			cfg := &config.Config{
				Policies: &config.Policies{
					Coverage: &config.CoveragePolicy{Threshold: 90, Packages: []string{"internal/domain"}},
				},
			}
			err := u.RunWithTitleAndPluginConfig(workdir, "main", "feat: x", nil, cfg)
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
		s.It("RunWithTitleAndPluginConfig with cfg.Pipeline.Enable filters steps", func(ctx *specs.Context) {
			resp := []testkit.CmdResponse{{Out: "", Err: nil}, {Out: "", Err: errors.New("no go.sum")}}
			cmd := &testkit.FakeCommandRunner{Responses: resp}
			git := &testkit.FakeGitClient{}
			env := testkit.NewFakeEnvProvider(nil)
			log := &testkit.FakeLogger{}
			pipelineSingle := application.Pipeline{Name: "pr", Steps: goLibPRStepsForTest(5)}
			u := application.NewPRUsecase(cmd, git, env, 95.0, log, testkit.NewFakeClock().Clock(), profile, pipelineSingle)
			cfg := &config.Config{Pipeline: config.PipelineConfig{Enable: []string{"go-mod-tidy"}}}
			err := u.RunWithTitleAndPluginConfig(workdir, "main", "", nil, cfg)
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("NewPRUsecase returns non-nil", func(ctx *specs.Context) {
			cmd := &testkit.FakeCommandRunner{}
			git := &testkit.FakeGitClient{}
			env := testkit.NewFakeEnvProvider(nil)
			log := &testkit.FakeLogger{}
			u := application.NewPRUsecase(cmd, git, env, 95.0, log, testkit.NewFakeClock().Clock(), profile, pipeline)
			ctx.Expect(u != nil).To(specs.BeTrue())
		})

		s.It("RunWithTitle success completes with no error", func(ctx *specs.Context) {
			cmd := &testkit.FakeCommandRunner{Responses: prSuccessResponses()}
			git := &testkit.FakeGitClient{}
			env := testkit.NewFakeEnvProvider(nil)
			log := &testkit.FakeLogger{}
			u := application.NewPRUsecase(cmd, git, env, 95.0, log, testkit.NewFakeClock().Clock(), profile, pipeline)
			err := u.RunWithTitle(workdir, "main", "feat: add feature")
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(log.LastInfoMsg).ToEqual("PR validation completed")
		})

		s.It("RunWithTitle go mod tidy fails returns error containing tidy", func(ctx *specs.Context) {
			resp := prSuccessResponses()
			resp[0] = testkit.CmdResponse{Out: "error", Err: errors.New("tidy failed")}
			cmd := &testkit.FakeCommandRunner{Responses: resp}
			git := &testkit.FakeGitClient{}
			env := testkit.NewFakeEnvProvider(nil)
			log := &testkit.FakeLogger{}
			u := application.NewPRUsecase(cmd, git, env, 95.0, log, testkit.NewFakeClock().Clock(), profile, pipeline)
			err := u.RunWithTitle(workdir, "main", "")
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "tidy")).To(specs.BeTrue())
		})

		s.It("RunWithTitle mod not tidy go.mod diff returns ErrModNotTidy", func(ctx *specs.Context) {
			cmd := &testkit.FakeCommandRunner{Responses: prSuccessResponses()}
			git := &testkit.FakeGitClient{DiffExitCodeErr: domain.ErrModNotTidy}
			env := testkit.NewFakeEnvProvider(nil)
			log := &testkit.FakeLogger{}
			u := application.NewPRUsecase(cmd, git, env, 95.0, log, testkit.NewFakeClock().Clock(), profile, pipeline)
			err := u.RunWithTitle(workdir, "main", "")
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, domain.ErrModNotTidy)).To(specs.BeTrue())
		})

		s.It("RunWithTitle golangci-lint fails returns error containing golangci-lint", func(ctx *specs.Context) {
			resp := prSuccessResponses()
			resp[10] = testkit.CmdResponse{Out: "vet error", Err: errors.New("lint failed")} // golangci is index 10
			cmd := &testkit.FakeCommandRunner{Responses: resp}
			git := &testkit.FakeGitClient{}
			env := testkit.NewFakeEnvProvider(map[string]string{"PR_TITLE": "feat: x"})
			log := &testkit.FakeLogger{}
			u := application.NewPRUsecase(cmd, git, env, 95.0, log, testkit.NewFakeClock().Clock(), profile, pipeline)
			err := u.RunWithTitle(workdir, "main", "")
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "golangci-lint")).To(specs.BeTrue())
		})

		s.It("RunWithTitle govulncheck fails returns error containing govulncheck", func(ctx *specs.Context) {
			resp := prSuccessResponses()
			resp[11] = testkit.CmdResponse{Out: "Vulnerability found\n", Err: errors.New("exit status 1")}
			cmd := &testkit.FakeCommandRunner{Responses: resp}
			git := &testkit.FakeGitClient{}
			env := testkit.NewFakeEnvProvider(map[string]string{"PR_TITLE": "feat: x"})
			log := &testkit.FakeLogger{}
			u := application.NewPRUsecase(cmd, git, env, 95.0, log, testkit.NewFakeClock().Clock(), profile, pipeline)
			err := u.RunWithTitle(workdir, "main", "")
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "govulncheck")).To(specs.BeTrue())
		})

		s.It("RunWithTitle test fails returns ErrTestFailed", func(ctx *specs.Context) {
			resp := prSuccessResponses()
			resp[12] = testkit.CmdResponse{Out: "", Err: errors.New("test failed")}
			cmd := &testkit.FakeCommandRunner{Responses: resp}
			git := &testkit.FakeGitClient{}
			env := testkit.NewFakeEnvProvider(map[string]string{"PR_TITLE": "feat: x"})
			log := &testkit.FakeLogger{}
			u := application.NewPRUsecase(cmd, git, env, 95.0, log, testkit.NewFakeClock().Clock(), profile, pipeline)
			err := u.RunWithTitle(workdir, "main", "")
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, domain.ErrTestFailed)).To(specs.BeTrue())
		})

		s.It("RunWithTitle coverage below threshold returns error", func(ctx *specs.Context) {
			resp := prSuccessResponses()
			resp[13] = testkit.CmdResponse{Out: "total: (statements) 80.0%", Err: nil} // below 95
			cmd := &testkit.FakeCommandRunner{Responses: resp}
			git := &testkit.FakeGitClient{}
			env := testkit.NewFakeEnvProvider(map[string]string{"PR_TITLE": "feat: x"})
			log := &testkit.FakeLogger{}
			u := application.NewPRUsecase(cmd, git, env, 95.0, log, testkit.NewFakeClock().Clock(), profile, pipeline)
			err := u.RunWithTitle(workdir, "main", "")
			ctx.Expect(err != nil).To(specs.BeTrue())
		})

		s.It("RunWithTitle conventional commit missing in PR title returns error", func(ctx *specs.Context) {
			cmd := &testkit.FakeCommandRunner{Responses: prSuccessResponses()}
			git := &testkit.FakeGitClient{}
			env := testkit.NewFakeEnvProvider(map[string]string{"PR_TITLE": "not conventional"})
			log := &testkit.FakeLogger{}
			u := application.NewPRUsecase(cmd, git, env, 95.0, log, testkit.NewFakeClock().Clock(), profile, pipeline)
			err := u.RunWithTitle(workdir, "main", "")
			ctx.Expect(err != nil).To(specs.BeTrue())
		})

	})
}
