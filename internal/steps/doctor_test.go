package steps

import (
	"context"
	"errors"
	"testing"

	"github.com/pablogore/devforge/internal/application"
	"github.com/pablogore/devforge/internal/testkit"
	"github.com/pablogore/go-specs/specs"
)

func TestDoctorSteps(t *testing.T) {
	specs.Describe(t, "DoctorSteps", func(s *specs.Spec) {
		s.It("returns non-empty list in deterministic order", func(ctx *specs.Context) {
			steps := DoctorSteps()
			ctx.Expect(len(steps) > 0).To(specs.BeTrue())
			names := make([]string, len(steps))
			for i, st := range steps {
				names[i] = st.Name()
			}
			ctx.Expect(names[0]).ToEqual("git-installed")
			ctx.Expect(names[1]).ToEqual("goreleaser-installed")
		})
		s.It("returns six steps", func(ctx *specs.Context) {
			steps := DoctorSteps()
			ctx.Expect(len(steps)).ToEqual(6)
		})
	})
}

func TestGitInstalledStep(t *testing.T) {
	specs.Describe(t, "GitInstalledStep", func(s *specs.Spec) {
		s.It("Name returns git-installed", func(ctx *specs.Context) {
			ctx.Expect(GitInstalledStep{}.Name()).ToEqual("git-installed")
		})
		s.It("Run success appends check", func(ctx *specs.Context) {
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("git", []string{"--version"}, "git version 2.x", nil)
			var checks []application.CheckResult
			appCtx := &application.Context{
				StdCtx:       context.Background(),
				Cmd:          cmd,
				DoctorChecks: &checks,
			}
			err := (GitInstalledStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(checks) == 1 && checks[0].Name == "git installed" && checks[0].Passed).To(specs.BeTrue())
		})
		s.It("Run git not found returns error", func(ctx *specs.Context) {
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("git", []string{"--version"}, "", errors.New("exec: not found"))
			var checks []application.CheckResult
			appCtx := &application.Context{StdCtx: context.Background(), Cmd: cmd, DoctorChecks: &checks}
			err := (GitInstalledStep{}).Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(len(checks) == 1 && !checks[0].Passed).To(specs.BeTrue())
		})
	})
}

func TestGoreleaserInstalledStep(t *testing.T) {
	specs.Describe(t, "GoreleaserInstalledStep", func(s *specs.Spec) {
		s.It("Name returns goreleaser-installed", func(ctx *specs.Context) {
			ctx.Expect(GoreleaserInstalledStep{}.Name()).ToEqual("goreleaser-installed")
		})
		s.It("Run success appends check", func(ctx *specs.Context) {
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("goreleaser", []string{"--version"}, "goreleaser version 2.x", nil)
			var checks []application.CheckResult
			appCtx := &application.Context{StdCtx: context.Background(), Cmd: cmd, DoctorChecks: &checks}
			err := (GoreleaserInstalledStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(checks) == 1 && checks[0].Passed).To(specs.BeTrue())
		})
		s.It("Run not found returns error", func(ctx *specs.Context) {
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("goreleaser", []string{"--version"}, "", errors.New("not found"))
			var checks []application.CheckResult
			appCtx := &application.Context{StdCtx: context.Background(), Cmd: cmd, DoctorChecks: &checks}
			err := (GoreleaserInstalledStep{}).Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
	})
}

func TestFullHistoryStep(t *testing.T) {
	specs.Describe(t, "FullHistoryStep", func(s *specs.Spec) {
		s.It("Name returns full-history", func(ctx *specs.Context) {
			ctx.Expect(FullHistoryStep{}.Name()).ToEqual("full-history")
		})
		s.It("Run has full history appends passed", func(ctx *specs.Context) {
			git := &testkit.FakeGitClient{HasFullHistoryOut: true}
			var checks []application.CheckResult
			appCtx := &application.Context{StdCtx: context.Background(), Workdir: "/wd", Git: git, DoctorChecks: &checks}
			err := (FullHistoryStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(checks) == 1 && checks[0].Passed).To(specs.BeTrue())
		})
		s.It("Run shallow appends failed", func(ctx *specs.Context) {
			git := &testkit.FakeGitClient{HasFullHistoryOut: false}
			var checks []application.CheckResult
			appCtx := &application.Context{StdCtx: context.Background(), Workdir: "/wd", Git: git, DoctorChecks: &checks}
			err := (FullHistoryStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(checks) == 1 && !checks[0].Passed).To(specs.BeTrue())
		})
		s.It("Run HasFullHistory error appends failed", func(ctx *specs.Context) {
			git := &testkit.FakeGitClient{HasFullHistoryErr: errors.New("git failed")}
			var checks []application.CheckResult
			appCtx := &application.Context{StdCtx: context.Background(), Workdir: "/wd", Git: git, DoctorChecks: &checks}
			err := (FullHistoryStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(checks) == 1 && !checks[0].Passed).To(specs.BeTrue())
		})
	})
}

func TestBranchMainStep(t *testing.T) {
	specs.Describe(t, "BranchMainStep", func(s *specs.Spec) {
		s.It("Name returns branch-main", func(ctx *specs.Context) {
			ctx.Expect(BranchMainStep{}.Name()).ToEqual("branch-main")
		})
		s.It("Run on main appends passed", func(ctx *specs.Context) {
			git := &testkit.FakeGitClient{GetCurrentBranchOut: "main"}
			var checks []application.CheckResult
			appCtx := &application.Context{StdCtx: context.Background(), Workdir: "/wd", Git: git, DoctorChecks: &checks}
			err := (BranchMainStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(checks) == 1 && checks[0].Passed).To(specs.BeTrue())
		})
		s.It("Run not on main appends failed", func(ctx *specs.Context) {
			git := &testkit.FakeGitClient{GetCurrentBranchOut: "feature"}
			var checks []application.CheckResult
			appCtx := &application.Context{StdCtx: context.Background(), Workdir: "/wd", Git: git, DoctorChecks: &checks}
			err := (BranchMainStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(checks) == 1 && !checks[0].Passed).To(specs.BeTrue())
		})
		s.It("Run GetCurrentBranch error appends failed", func(ctx *specs.Context) {
			git := &testkit.FakeGitClient{GetCurrentBranchErr: errors.New("git failed")}
			var checks []application.CheckResult
			appCtx := &application.Context{StdCtx: context.Background(), Workdir: "/wd", Git: git, DoctorChecks: &checks}
			err := (BranchMainStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(checks) == 1 && !checks[0].Passed).To(specs.BeTrue())
		})
	})
}

func TestWorkingTreeCleanStep(t *testing.T) {
	specs.Describe(t, "WorkingTreeCleanStep", func(s *specs.Spec) {
		s.It("Name returns working-tree-clean", func(ctx *specs.Context) {
			ctx.Expect(WorkingTreeCleanStep{}.Name()).ToEqual("working-tree-clean")
		})
		s.It("Run clean appends passed", func(ctx *specs.Context) {
			git := &testkit.FakeGitClient{IsWorkingTreeCleanOut: true}
			var checks []application.CheckResult
			appCtx := &application.Context{StdCtx: context.Background(), Workdir: "/wd", Git: git, DoctorChecks: &checks}
			err := (WorkingTreeCleanStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(checks) == 1 && checks[0].Passed).To(specs.BeTrue())
		})
		s.It("Run dirty appends failed", func(ctx *specs.Context) {
			git := &testkit.FakeGitClient{IsWorkingTreeCleanOut: false}
			var checks []application.CheckResult
			appCtx := &application.Context{StdCtx: context.Background(), Workdir: "/wd", Git: git, DoctorChecks: &checks}
			err := (WorkingTreeCleanStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(checks) == 1 && !checks[0].Passed).To(specs.BeTrue())
		})
		s.It("Run IsWorkingTreeClean error appends failed", func(ctx *specs.Context) {
			git := &testkit.FakeGitClient{IsWorkingTreeCleanErr: errors.New("git failed")}
			var checks []application.CheckResult
			appCtx := &application.Context{StdCtx: context.Background(), Workdir: "/wd", Git: git, DoctorChecks: &checks}
			err := (WorkingTreeCleanStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(checks) == 1 && !checks[0].Passed).To(specs.BeTrue())
		})
	})
}

func TestTagsAccessibleStep(t *testing.T) {
	specs.Describe(t, "TagsAccessibleStep", func(s *specs.Spec) {
		s.It("Name returns tags-accessible", func(ctx *specs.Context) {
			ctx.Expect(TagsAccessibleStep{}.Name()).ToEqual("tags-accessible")
		})
		s.It("Run tags ok appends passed", func(ctx *specs.Context) {
			git := &testkit.FakeGitClient{GetLatestTagErr: nil}
			var checks []application.CheckResult
			appCtx := &application.Context{StdCtx: context.Background(), Workdir: "/wd", Git: git, DoctorChecks: &checks}
			err := (TagsAccessibleStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(checks) == 1 && checks[0].Passed).To(specs.BeTrue())
		})
		s.It("Run tags error appends failed", func(ctx *specs.Context) {
			git := &testkit.FakeGitClient{GetLatestTagErr: errors.New("no tags")}
			var checks []application.CheckResult
			appCtx := &application.Context{StdCtx: context.Background(), Workdir: "/wd", Git: git, DoctorChecks: &checks}
			err := (TagsAccessibleStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(checks) == 1 && !checks[0].Passed).To(specs.BeTrue())
		})
	})
}
