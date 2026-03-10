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
	"github.com/pablogore/devforge/internal/config"
	"github.com/pablogore/devforge/internal/domain"
	"github.com/pablogore/devforge/internal/testkit"
	"github.com/pablogore/go-specs/specs"
)

func TestGoTestRaceStep_Run(t *testing.T) {
	specs.Describe(t, "GoTestRaceStep.Run", func(s *specs.Spec) {
		s.It("succeeds when go test -race succeeds", func(ctx *specs.Context) {
			dir := t.TempDir()
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("go", []string{"test", "-race", "./..."}, "", nil)
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Cmd: cmd, Workdir: dir, Log: log}
			err := (GoTestRaceStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("returns error when test fails", func(ctx *specs.Context) {
			dir := t.TempDir()
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("go", []string{"test", "-race", "./..."}, "fail", domain.ErrTestFailed)
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Cmd: cmd, Workdir: dir, Log: log}
			err := (GoTestRaceStep{}).Run(appCtx)
			ctx.Expect(errors.Is(err, domain.ErrTestFailed)).To(specs.BeTrue())
		})
	})
}

func TestGoTestStep_Run(t *testing.T) {
	specs.Describe(t, "GoTestStep.Run", func(s *specs.Spec) {
		s.It("with coverage packages resolved logs policy", func(ctx *specs.Context) {
			dir := t.TempDir()
			cmd := testkit.NewFakeCommandRunner()
			coverPkg := "pkg1,pkg2"
			cmd.Stub("go", []string{"test", "-race", "-coverprofile=coverage.out", "-coverpkg=" + coverPkg, "./..."}, "", nil)
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{
				StdCtx:                   context.Background(),
				Cmd:                       cmd,
				Workdir:                   dir,
				Log:                       log,
				CoverPkg:                  coverPkg,
				CoveragePackagesResolved:  []string{"pkg1", "pkg2"},
				CoverageThreshold:         95,
				Config:                    &config.Config{Policies: &config.Policies{Coverage: &config.CoveragePolicy{Packages: []string{"internal/*"}}}},
			}
			err := (GoTestStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("empty CoverPkg uses default", func(ctx *specs.Context) {
			dir := t.TempDir()
			defaultPkg := "./internal/domain,./internal/application"
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("go", []string{"test", "-race", "-coverprofile=coverage.out", "-coverpkg=" + defaultPkg, "./..."}, "", nil)
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Cmd: cmd, Workdir: dir, Log: log, CoverPkg: ""}
			err := (GoTestStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("uses specs runner when available and logs skip message", func(ctx *specs.Context) {
			dir := t.TempDir()
			orig := specsRunnerAvailableFunc
			defer func() { specsRunnerAvailableFunc = orig }()
			specsRunnerAvailableFunc = func(*application.Context) bool { return true }
			cmd := testkit.NewFakeCommandRunner()
			defaultPkg := "./internal/domain,./internal/application"
			cmd.Stub("specs", []string{"run", "--", "-race", "-coverprofile=coverage.out", "-coverpkg=" + defaultPkg, "./..."}, "", nil)
			log := &testkit.FakeLogger{RecordInfoHistory: true}
			appCtx := &application.Context{StdCtx: context.Background(), Cmd: cmd, Workdir: dir, Log: log, CoverPkg: ""}
			err := (GoTestStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(cmd.WasCalled("specs", "run", "--", "-race", "-coverprofile=coverage.out", "-coverpkg="+defaultPkg, "./...")).To(specs.BeTrue())
			var foundRunner, foundSkip bool
			for _, c := range log.InfoHistory {
				if strings.Contains(c.Msg, "tests executed via specs runner") {
					foundRunner = true
				}
				if strings.Contains(c.Msg, "skipping go test duplicate run") {
					foundSkip = true
				}
			}
			ctx.Expect(foundRunner).To(specs.BeTrue())
			ctx.Expect(foundSkip).To(specs.BeTrue())
		})
	})
}

func TestIntegrationTestsStep_Run(t *testing.T) {
	specs.Describe(t, "IntegrationTestsStep.Run", func(s *specs.Spec) {
		s.It("skips when no build tag", func(ctx *specs.Context) {
			dir := t.TempDir()
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Workdir: dir, Log: log}
			err := (IntegrationTestsStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(log.LastInfoMsg).ToEqual("Skipping integration tests (no //go:build integration found)")
		})
		s.It("runs when build tag present", func(ctx *specs.Context) {
			dir := t.TempDir()
			f := filepath.Join(dir, "integration_test.go")
			err := os.WriteFile(f, []byte("//go:build integration\npackage p\n"), 0o600)
			ctx.Expect(err).To(specs.BeNil())
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("go", []string{"test", "-tags=integration", "-count=1", "./..."}, "", nil)
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Cmd: cmd, Workdir: dir, Log: log}
			err = (IntegrationTestsStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("with integrationtest dir runs that pkg", func(ctx *specs.Context) {
			dir := t.TempDir()
			err := os.MkdirAll(filepath.Join(dir, "integrationtest"), 0o750)
			ctx.Expect(err).To(specs.BeNil())
			err = os.WriteFile(filepath.Join(dir, "integrationtest", "test.go"), []byte("//go:build integration\npackage integrationtest\n"), 0o600)
			ctx.Expect(err).To(specs.BeNil())
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("go", []string{"test", "-tags=integration", "-count=1", "./integrationtest/..."}, "", nil)
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Cmd: cmd, Workdir: dir, Log: log}
			err = (IntegrationTestsStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
		})
	})
}

func TestIntegrationTestsStep_hasIntegrationBuildTag(t *testing.T) {
	specs.Describe(t, "hasIntegrationBuildTag", func(s *specs.Spec) {
		s.It("empty dir returns false", func(ctx *specs.Context) {
			dir := t.TempDir()
			ctx.Expect(hasIntegrationBuildTag(dir)).To(specs.BeFalse())
		})
		s.It("dir with //go:build integration returns true", func(ctx *specs.Context) {
			dir := t.TempDir()
			err := os.WriteFile(filepath.Join(dir, "x.go"), []byte("//go:build integration\npackage p\n"), 0o600)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(hasIntegrationBuildTag(dir)).To(specs.BeTrue())
		})
		s.It("skips vendor and dot dirs", func(ctx *specs.Context) {
			dir := t.TempDir()
			err := os.MkdirAll(filepath.Join(dir, "vendor"), 0o750)
			ctx.Expect(err).To(specs.BeNil())
			err = os.MkdirAll(filepath.Join(dir, ".git"), 0o750)
			ctx.Expect(err).To(specs.BeNil())
			err = os.WriteFile(filepath.Join(dir, "vendor", "x.go"), []byte("//go:build integration\npackage p\n"), 0o600)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(hasIntegrationBuildTag(dir)).To(specs.BeFalse())
			err = os.WriteFile(filepath.Join(dir, "p.go"), []byte("//go:build integration\npackage p\n"), 0o600)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(hasIntegrationBuildTag(dir)).To(specs.BeTrue())
		})
		s.It("detects legacy // +build integration", func(ctx *specs.Context) {
			dir := t.TempDir()
			err := os.WriteFile(filepath.Join(dir, "int_test.go"), []byte("// +build integration\n\npackage p\n"), 0o600)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(hasIntegrationBuildTag(dir)).To(specs.BeTrue())
		})
	})
}

func TestGoFmtStep_Run(t *testing.T) {
	specs.Describe(t, "GoFmtStep.Run", func(s *specs.Spec) {
		s.It("success when gofmt passes", func(ctx *specs.Context) {
			dir := t.TempDir()
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("gofmt", []string{"-s", "-l", "."}, "", nil)
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Cmd: cmd, Workdir: dir, Log: log}
			err := (GoFmtStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("error when unformatted files", func(ctx *specs.Context) {
			dir := t.TempDir()
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("gofmt", []string{"-s", "-l", "."}, "foo.go\n", nil)
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Cmd: cmd, Workdir: dir, Log: log}
			err := (GoFmtStep{}).Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, domain.ErrFormatting)).To(specs.BeTrue())
		})
	})
}

func TestGovulnCheckStep_Run(t *testing.T) {
	specs.Describe(t, "GovulnCheckStep.Run", func(s *specs.Spec) {
		s.It("success", func(ctx *specs.Context) {
			dir := t.TempDir()
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("govulncheck", []string{"-json", "./..."}, `{"Vulns":null}`, nil)
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Cmd: cmd, Workdir: dir, Log: log}
			err := (GovulnCheckStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("with output error returns error containing output", func(ctx *specs.Context) {
			dir := t.TempDir()
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("govulncheck", []string{"-json", "./..."}, "some output", errors.New("exit 1"))
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Cmd: cmd, Workdir: dir, Log: log}
			err := (GovulnCheckStep{}).Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "some output")).To(specs.BeTrue())
		})
		s.It("no output error returns not in PATH message", func(ctx *specs.Context) {
			dir := t.TempDir()
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("govulncheck", []string{"-json", "./..."}, "", errors.New("not in PATH"))
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Cmd: cmd, Workdir: dir, Log: log}
			err := (GovulnCheckStep{}).Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "not installed or not in PATH")).To(specs.BeTrue())
		})
	})
}

// golangciLintGoArgs returns the "go" command args used by GolangCILintStep (for test stubbing).
func golangciLintGoArgs() []string {
	a := make([]string, 0, 2+len(GolangciLintRunArgs))
	a = append(a, "run", GolangciLintModuleVersion)
	a = append(a, GolangciLintRunArgs...)
	return a
}

func TestGolangCILintStep_Run(t *testing.T) {
	specs.Describe(t, "GolangCILintStep.Run", func(s *specs.Spec) {
		s.It("success when tool exits ok", func(ctx *specs.Context) {
			dir := t.TempDir()
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("go", golangciLintGoArgs(), "", nil)
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{
				StdCtx:  context.Background(),
				Cmd:     cmd,
				Workdir: dir,
				Log:     log,
				Clock:   clock.NewRealClock(),
			}
			err := (GolangCILintStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("crash on first run then success on retry passes", func(ctx *specs.Context) {
			dir := t.TempDir()
			cmd := testkit.NewFakeCommandRunner()
			cmd.Enqueue("go", golangciLintGoArgs(), "panic: something", errors.New("exit 2"))
			cmd.Enqueue("go", golangciLintGoArgs(), "", nil)
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{
				StdCtx:  context.Background(),
				Cmd:     cmd,
				Workdir: dir,
				Log:     log,
				Clock:   clock.NewRealClock(),
			}
			err := (GolangCILintStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("crash on both runs fails pipeline after retry", func(ctx *specs.Context) {
			dir := t.TempDir()
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("go", golangciLintGoArgs(), "internal error: panic: runtime error", errors.New("exit 1"))
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{
				StdCtx:  context.Background(),
				Cmd:     cmd,
				Workdir: dir,
				Log:     log,
				Clock:   clock.NewRealClock(),
			}
			err := (GolangCILintStep{}).Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "golangci-lint")).To(specs.BeTrue())
		})
		s.It("returns error when tool fails with lint issues (no retry)", func(ctx *specs.Context) {
			dir := t.TempDir()
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("go", golangciLintGoArgs(), "file.go:10: some lint error", errors.New("exit 1"))
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{
				StdCtx:  context.Background(),
				Cmd:     cmd,
				Workdir: dir,
				Log:     log,
				Clock:   clock.NewRealClock(),
			}
			err := (GolangCILintStep{}).Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "golangci-lint")).To(specs.BeTrue())
		})
	})
}

func TestCoverageStep_Run(t *testing.T) {
	specs.Describe(t, "CoverageStep.Run", func(s *specs.Spec) {
		s.It("passes when above threshold", func(ctx *specs.Context) {
			dir := t.TempDir()
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("go", []string{"tool", "cover", "-func=coverage.out"}, "total:\t(statements)\t95.0%\n", nil)
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{
				StdCtx:            context.Background(),
				Cmd:               cmd,
				Workdir:           dir,
				Log:               log,
				CoverageThreshold: 95,
			}
			err := (CoverageStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("fails when below threshold", func(ctx *specs.Context) {
			dir := t.TempDir()
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("go", []string{"tool", "cover", "-func=coverage.out"}, "total:\t(statements)\t80.0%\n", nil)
			appCtx := &application.Context{
				StdCtx:            context.Background(),
				Cmd:               cmd,
				Workdir:           dir,
				Log:               &testkit.FakeLogger{},
				CoverageThreshold: 95,
			}
			err := (CoverageStep{}).Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "80%")).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "95%")).To(specs.BeTrue())
		})
		s.It("parse fails returns ErrCoverageParse", func(ctx *specs.Context) {
			dir := t.TempDir()
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("go", []string{"tool", "cover", "-func=coverage.out"}, "", errors.New("no coverage.out"))
			appCtx := &application.Context{StdCtx: context.Background(), Cmd: cmd, Workdir: dir, Log: &testkit.FakeLogger{}}
			err := (CoverageStep{}).Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, domain.ErrCoverageParse)).To(specs.BeTrue())
		})
	})
}

func TestConventionalCommitStep_Run(t *testing.T) {
	specs.Describe(t, "ConventionalCommitStep.Run", func(s *specs.Spec) {
		s.It("title override succeeds", func(ctx *specs.Context) {
			appCtx := &application.Context{
				StdCtx:        context.Background(),
				Workdir:       t.TempDir(),
				Log:           &testkit.FakeLogger{},
				TitleOverride: "feat: add feature",
			}
			err := (ConventionalCommitStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("PR title from env succeeds", func(ctx *specs.Context) {
			env := testkit.NewFakeEnvProvider(map[string]string{"PR_TITLE": "fix: bug fix"})
			appCtx := &application.Context{
				StdCtx:  context.Background(),
				Workdir: t.TempDir(),
				Log:     &testkit.FakeLogger{},
				Env:     env,
			}
			err := (ConventionalCommitStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("from git commit succeeds", func(ctx *specs.Context) {
			dir := t.TempDir()
			env := testkit.NewFakeEnvProvider(map[string]string{"PR_TITLE": ""})
			git := &testkit.FakeGitClient{GetLatestCommitMessageOut: "docs: update readme", GetLatestCommitMessageErr: nil}
			appCtx := &application.Context{
				StdCtx:  context.Background(),
				Workdir: dir,
				Log:     &testkit.FakeLogger{},
				Env:     env,
				Git:     git,
			}
			err := (ConventionalCommitStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("empty title fails with ErrPRTitleRequired", func(ctx *specs.Context) {
			env := testkit.NewFakeEnvProvider(map[string]string{"PR_TITLE": ""})
			git := &testkit.FakeGitClient{GetLatestCommitMessageOut: "", GetLatestCommitMessageErr: errors.New("no commit")}
			appCtx := &application.Context{
				StdCtx:  context.Background(),
				Workdir: t.TempDir(),
				Log:     &testkit.FakeLogger{},
				Env:     env,
				Git:     git,
			}
			err := (ConventionalCommitStep{}).Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, domain.ErrPRTitleRequired)).To(specs.BeTrue())
		})
	})
}

func TestGoModTidyStep_Run(t *testing.T) {
	specs.Describe(t, "GoModTidyStep.Run", func(s *specs.Spec) {
		s.It("success when no go.sum", func(ctx *specs.Context) {
			dir := t.TempDir()
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("go", []string{"mod", "tidy"}, "", nil)
			cmd.Stub("test", []string{"-e", "go.sum"}, "", errors.New("no go.sum"))
			git := &testkit.FakeGitClient{DiffExitCodeErr: nil}
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{
				StdCtx:  context.Background(),
				Cmd:     cmd,
				Git:     git,
				Workdir: dir,
				Log:     log,
			}
			err := (GoModTidyStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("success when go.sum exists and clean", func(ctx *specs.Context) {
			dir := t.TempDir()
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("go", []string{"mod", "tidy"}, "", nil)
			cmd.Stub("test", []string{"-e", "go.sum"}, "", nil)
			git := &testkit.FakeGitClient{DiffExitCodeErr: nil}
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Cmd: cmd, Git: git, Workdir: dir, Log: log}
			err := (GoModTidyStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("returns error when go mod tidy fails", func(ctx *specs.Context) {
			dir := t.TempDir()
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("go", []string{"mod", "tidy"}, "", errors.New("tidy failed"))
			git := &testkit.FakeGitClient{}
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Cmd: cmd, Git: git, Workdir: dir, Log: log}
			err := (GoModTidyStep{}).Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, domain.ErrModNotTidy)).To(specs.BeFalse())
		})
		s.It("returns ErrModNotTidy when go.mod has diff", func(ctx *specs.Context) {
			dir := t.TempDir()
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("go", []string{"mod", "tidy"}, "", nil)
			cmd.Stub("test", []string{"-e", "go.sum"}, "", errors.New("no go.sum"))
			git := &testkit.FakeGitClient{DiffExitCodeErr: errors.New("dirty go.mod")}
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Cmd: cmd, Git: git, Workdir: dir, Log: log}
			err := (GoModTidyStep{}).Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, domain.ErrModNotTidy)).To(specs.BeTrue())
		})
	})
}
