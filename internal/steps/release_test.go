package steps

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/pablogore/devforge/internal/application"
	"github.com/pablogore/devforge/internal/domain"
	"github.com/pablogore/devforge/internal/testkit"
	"github.com/pablogore/go-specs/specs"
)

func TestVersionDerivationStep(t *testing.T) {
	specs.Describe(t, "VersionDerivationStep", func(s *specs.Spec) {
		s.It("Run success sets version", func(ctx *specs.Context) {
			dir := t.TempDir()
			git := &testkit.FakeGitClient{}
			git.GetLatestTagOut = "v1.0.0"
			git.GetLatestTagErr = nil
			git.GetCommitsSinceOut = []string{"feat: add feature"}
			git.GetCommitsSinceErr = nil
			git.GetTagHashErr = errors.New("tag not found")
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{
				StdCtx:  context.Background(),
				Workdir: dir,
				Git:     git,
				Log:     log,
			}
			err := (VersionDerivationStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(appCtx.Version).ToEqual("v1.1.0")
		})
		s.It("Run no releaseable changes returns ErrNoReleaseableChanges", func(ctx *specs.Context) {
			dir := t.TempDir()
			git := &testkit.FakeGitClient{}
			git.GetLatestTagOut = "v1.0.0"
			git.GetCommitsSinceOut = []string{"chore: deps"}
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Workdir: dir, Git: git, Log: log}
			err := (VersionDerivationStep{}).Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, domain.ErrNoReleaseableChanges)).To(specs.BeTrue())
		})
		s.It("Run tag already exists returns ErrTagAlreadyExists", func(ctx *specs.Context) {
			dir := t.TempDir()
			git := &testkit.FakeGitClient{}
			git.GetLatestTagOut = "v1.0.0"
			git.GetCommitsSinceOut = []string{"feat: x"}
			git.GetTagHashOut = "abc123"
			git.GetTagHashErr = nil
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Workdir: dir, Git: git, Log: log}
			err := (VersionDerivationStep{}).Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, domain.ErrTagAlreadyExists)).To(specs.BeTrue())
		})
		s.It("Run GetLatestTag error returns error", func(ctx *specs.Context) {
			dir := t.TempDir()
			git := &testkit.FakeGitClient{}
			git.GetLatestTagErr = errors.New("no tags")
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Workdir: dir, Git: git, Log: log}
			err := (VersionDerivationStep{}).Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "latest tag")).To(specs.BeTrue())
		})
		s.It("Run GetCommitsSince error returns error", func(ctx *specs.Context) {
			dir := t.TempDir()
			git := &testkit.FakeGitClient{}
			git.GetLatestTagOut = "v1.0.0"
			git.GetCommitsSinceErr = errors.New("log failed")
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Workdir: dir, Git: git, Log: log}
			err := (VersionDerivationStep{}).Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
	})
}

func TestPreconditionsStep(t *testing.T) {
	specs.Describe(t, "PreconditionsStep", func(s *specs.Spec) {
		s.It("Run success when on main and clean", func(ctx *specs.Context) {
			dir := t.TempDir()
			git := &testkit.FakeGitClient{}
			git.HasFullHistoryOut = true
			git.GetCurrentBranchOut = "main"
			git.IsWorkingTreeCleanOut = true
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Workdir: dir, Git: git, Log: log}
			err := (PreconditionsStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("Run shallow clone returns ErrShallowCloneDetected", func(ctx *specs.Context) {
			dir := t.TempDir()
			git := &testkit.FakeGitClient{}
			git.HasFullHistoryOut = false
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Workdir: dir, Git: git, Log: log}
			err := (PreconditionsStep{}).Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, domain.ErrShallowCloneDetected)).To(specs.BeTrue())
		})
		s.It("Run not on main returns ErrNotOnMainBranch", func(ctx *specs.Context) {
			dir := t.TempDir()
			git := &testkit.FakeGitClient{}
			git.HasFullHistoryOut = true
			git.GetCurrentBranchOut = "feature"
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Workdir: dir, Git: git, Log: log}
			err := (PreconditionsStep{}).Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, domain.ErrNotOnMainBranch)).To(specs.BeTrue())
		})
		s.It("Run working tree dirty returns ErrWorkingTreeDirty", func(ctx *specs.Context) {
			dir := t.TempDir()
			git := &testkit.FakeGitClient{}
			git.HasFullHistoryOut = true
			git.GetCurrentBranchOut = "main"
			git.IsWorkingTreeCleanOut = false
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Workdir: dir, Git: git, Log: log}
			err := (PreconditionsStep{}).Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, domain.ErrWorkingTreeDirty)).To(specs.BeTrue())
		})
		s.It("Run HasFullHistory error returns error", func(ctx *specs.Context) {
			dir := t.TempDir()
			git := &testkit.FakeGitClient{}
			git.HasFullHistoryErr = errors.New("git failed")
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Workdir: dir, Git: git, Log: log}
			err := (PreconditionsStep{}).Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "history")).To(specs.BeTrue())
		})
		s.It("Run GetCurrentBranch error returns error", func(ctx *specs.Context) {
			dir := t.TempDir()
			git := &testkit.FakeGitClient{}
			git.HasFullHistoryOut = true
			git.GetCurrentBranchErr = errors.New("branch failed")
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Workdir: dir, Git: git, Log: log}
			err := (PreconditionsStep{}).Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
		s.It("Run IsWorkingTreeClean error returns error", func(ctx *specs.Context) {
			dir := t.TempDir()
			git := &testkit.FakeGitClient{}
			git.HasFullHistoryOut = true
			git.GetCurrentBranchOut = "main"
			git.IsWorkingTreeCleanErr = errors.New("status failed")
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Workdir: dir, Git: git, Log: log}
			err := (PreconditionsStep{}).Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
	})
}

func TestTagCreationStep(t *testing.T) {
	specs.Describe(t, "TagCreationStep", func(s *specs.Spec) {
		s.It("Name returns create-tag", func(ctx *specs.Context) {
			ctx.Expect((TagCreationStep{}).Name()).ToEqual("create-tag")
		})
		s.It("Run success", func(ctx *specs.Context) {
			dir := t.TempDir()
			git := &testkit.FakeGitClient{}
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Workdir: dir, Git: git, Log: log, Version: "v1.1.0"}
			err := (TagCreationStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("Run create tag fails returns error", func(ctx *specs.Context) {
			dir := t.TempDir()
			git := &testkit.FakeGitClient{}
			git.CreateTagErr = fmt.Errorf("tag failed")
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Workdir: dir, Git: git, Log: log, Version: "v1.1.0"}
			err := (TagCreationStep{}).Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "tag failed")).To(specs.BeTrue())
		})
	})
}

func TestReleaseSteps(t *testing.T) {
	specs.Describe(t, "ReleaseSteps", func(s *specs.Spec) {
		s.It("returns non-empty list in order", func(ctx *specs.Context) {
			steps := ReleaseSteps()
			ctx.Expect(len(steps) > 0).To(specs.BeTrue())
			ctx.Expect(steps[0].Name()).ToEqual("preconditions")
			ctx.Expect(steps[1].Name()).ToEqual("version-derivation")
		})
		s.It("includes goreleaser and verify-tag", func(ctx *specs.Context) {
			steps := ReleaseSteps()
			names := make([]string, len(steps))
			for i, st := range steps {
				names[i] = st.Name()
			}
			ctx.Expect(len(names) >= 6).To(specs.BeTrue())
		})
	})
}

func TestTagVerificationStep(t *testing.T) {
	specs.Describe(t, "TagVerificationStep", func(s *specs.Spec) {
		s.It("Name returns verify-tag", func(ctx *specs.Context) {
			ctx.Expect((TagVerificationStep{}).Name()).ToEqual("verify-tag")
		})
		s.It("Run success when head and tag hash match", func(ctx *specs.Context) {
			dir := t.TempDir()
			git := &testkit.FakeGitClient{}
			git.GetHeadHashOut = "abc123"
			git.GetTagHashOut = "abc123"
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Workdir: dir, Git: git, Log: log, Version: "v1.0.0"}
			err := (TagVerificationStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("Run head hash error returns error", func(ctx *specs.Context) {
			dir := t.TempDir()
			git := &testkit.FakeGitClient{}
			git.GetHeadHashErr = errors.New("no head")
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Workdir: dir, Git: git, Log: log, Version: "v1.0.0"}
			err := (TagVerificationStep{}).Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
		s.It("Run tag hash error returns error", func(ctx *specs.Context) {
			dir := t.TempDir()
			git := &testkit.FakeGitClient{}
			git.GetHeadHashOut = "abc"
			git.GetTagHashErr = errors.New("no tag")
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Workdir: dir, Git: git, Log: log, Version: "v1.0.0"}
			err := (TagVerificationStep{}).Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
		s.It("Run mismatch returns ErrTagDoesNotPointToHead", func(ctx *specs.Context) {
			dir := t.TempDir()
			git := &testkit.FakeGitClient{}
			git.GetHeadHashOut = "abc"
			git.GetTagHashOut = "def"
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Workdir: dir, Git: git, Log: log, Version: "v1.0.0"}
			err := (TagVerificationStep{}).Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, domain.ErrTagDoesNotPointToHead)).To(specs.BeTrue())
		})
	})
}

func TestGoreleaserStep(t *testing.T) {
	specs.Describe(t, "GoreleaserStep", func(s *specs.Spec) {
		s.It("Run success", func(ctx *specs.Context) {
			dir := t.TempDir()
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("goreleaser", []string{"release", "--clean"}, "", nil)
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Cmd: cmd, Workdir: dir, Log: log, Version: "v1.0.0"}
			err := (GoreleaserStep{}).Run(appCtx)
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("Run fails returns error", func(ctx *specs.Context) {
			dir := t.TempDir()
			cmd := testkit.NewFakeCommandRunner()
			cmd.Stub("goreleaser", []string{"release", "--clean"}, "error", fmt.Errorf("goreleaser failed"))
			log := &testkit.FakeLogger{}
			appCtx := &application.Context{StdCtx: context.Background(), Cmd: cmd, Workdir: dir, Log: log, Version: "v1.0.0"}
			err := (GoreleaserStep{}).Run(appCtx)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "release failed")).To(specs.BeTrue())
		})
	})
}
