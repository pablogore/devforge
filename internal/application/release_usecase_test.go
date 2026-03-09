package application_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/pablogore/devforge/internal/application"
	"github.com/pablogore/devforge/internal/domain"
	"github.com/pablogore/devforge/internal/steps"
	"github.com/pablogore/devforge/internal/testkit"
	"github.com/pablogore/go-specs/specs"
)

func TestReleaseUsecase(t *testing.T) {
	workdir := "/wd"
	pipeline := application.Pipeline{Name: "release", Steps: steps.ReleaseSteps()}

	specs.Describe(t, "ReleaseUsecase", func(s *specs.Spec) {
		s.It("NewReleaseUsecase returns non-nil", func(ctx *specs.Context) {
			cmd := &testkit.FakeCommandRunner{}
			git := &testkit.FakeGitClient{}
			env := testkit.NewFakeEnvProvider(nil)
			log := &testkit.FakeLogger{}
			u := application.NewReleaseUsecase(cmd, git, env, log, testkit.NewFakeClock().Clock(), pipeline)
			ctx.Expect(u != nil).To(specs.BeTrue())
		})

		s.It("Run success returns version and commit msg", func(ctx *specs.Context) {
			cmd := &testkit.FakeCommandRunner{
				Responses: []testkit.CmdResponse{
					{Out: "goreleaser version 2.3.0", Err: nil},
					{Out: "", Err: nil},
				},
			}
			git := &testkit.FakeGitClient{
				HasFullHistoryOut:      true,
				GetCurrentBranchOut:   "main",
				IsWorkingTreeCleanOut:  true,
				GetLatestTagOut:        "v1.0.0",
				GetCommitsSinceOut:     []string{"feat: add"},
				GetTagHashResponses:    []testkit.GitResponse{{Out: "", Err: errors.New("tag not found")}, {Out: "abc123", Err: nil}},
				CreateTagErr:           nil,
				GetHeadHashOut:         "abc123",
			}
			log := &testkit.FakeLogger{}
			u := application.NewReleaseUsecase(cmd, git, testkit.NewFakeEnvProvider(nil), log, testkit.NewFakeClock().Clock(), pipeline)
			result, err := u.Run(workdir)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(result != nil).To(specs.BeTrue())
			ctx.Expect(result.Version).ToEqual("v1.1.0")
			ctx.Expect(strings.Contains(result.CommitMsg, "v1.1.0")).To(specs.BeTrue())
		})

		s.It("Run shallow clone returns ErrShallowCloneDetected", func(ctx *specs.Context) {
			git := &testkit.FakeGitClient{HasFullHistoryOut: false}
			log := &testkit.FakeLogger{}
			u := application.NewReleaseUsecase(nil, git, nil, log, testkit.NewFakeClock().Clock(), pipeline)
			result, err := u.Run(workdir)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(result == nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, domain.ErrShallowCloneDetected)).To(specs.BeTrue())
		})

		s.It("Run not on main returns ErrNotOnMainBranch", func(ctx *specs.Context) {
			git := &testkit.FakeGitClient{
				HasFullHistoryOut:    true,
				GetCurrentBranchOut: "develop",
			}
			log := &testkit.FakeLogger{}
			u := application.NewReleaseUsecase(nil, git, nil, log, testkit.NewFakeClock().Clock(), pipeline)
			_, err := u.Run(workdir)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, domain.ErrNotOnMainBranch)).To(specs.BeTrue())
		})

		s.It("Run working tree dirty returns ErrWorkingTreeDirty", func(ctx *specs.Context) {
			git := &testkit.FakeGitClient{
				HasFullHistoryOut:     true,
				GetCurrentBranchOut:   "main",
				IsWorkingTreeCleanOut:  false,
			}
			log := &testkit.FakeLogger{}
			u := application.NewReleaseUsecase(nil, git, nil, log, testkit.NewFakeClock().Clock(), pipeline)
			_, err := u.Run(workdir)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, domain.ErrWorkingTreeDirty)).To(specs.BeTrue())
		})

		s.It("Run no releaseable commits returns ErrNoReleaseableChanges", func(ctx *specs.Context) {
			git := &testkit.FakeGitClient{
				HasFullHistoryOut:    true,
				GetCurrentBranchOut: "main",
				IsWorkingTreeCleanOut: true,
				GetLatestTagOut:      "v1.0.0",
				GetCommitsSinceOut:   []string{"chore: deps"},
			}
			log := &testkit.FakeLogger{}
			u := application.NewReleaseUsecase(nil, git, nil, log, testkit.NewFakeClock().Clock(), pipeline)
			_, err := u.Run(workdir)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, domain.ErrNoReleaseableChanges)).To(specs.BeTrue())
		})

		s.It("Run tag already exists returns ErrTagAlreadyExists", func(ctx *specs.Context) {
			git := &testkit.FakeGitClient{
				HasFullHistoryOut:    true,
				GetCurrentBranchOut: "main",
				IsWorkingTreeCleanOut: true,
				GetLatestTagOut:      "v1.0.0",
				GetCommitsSinceOut:   []string{"feat: add"},
				GetTagHashOut:        "abc123",
				GetTagHashErr:        nil,
			}
			log := &testkit.FakeLogger{}
			u := application.NewReleaseUsecase(nil, git, nil, log, testkit.NewFakeClock().Clock(), pipeline)
			_, err := u.Run(workdir)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, domain.ErrTagAlreadyExists)).To(specs.BeTrue())
		})

		s.It("Run create tag fails returns error containing tag creation", func(ctx *specs.Context) {
			git := &testkit.FakeGitClient{
				HasFullHistoryOut:   true,
				GetCurrentBranchOut: "main",
				IsWorkingTreeCleanOut: true,
				GetLatestTagOut:     "v1.0.0",
				GetCommitsSinceOut:  []string{"feat: add"},
				GetTagHashResponses: []testkit.GitResponse{{Out: "", Err: errors.New("no tag")}},
				CreateTagErr:        errors.New("create failed"),
			}
			log := &testkit.FakeLogger{}
			u := application.NewReleaseUsecase(nil, git, nil, log, testkit.NewFakeClock().Clock(), pipeline)
			_, err := u.Run(workdir)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "tag")).To(specs.BeTrue())
		})

		s.It("Run tag does not point to HEAD returns ErrTagDoesNotPointToHead", func(ctx *specs.Context) {
			git := &testkit.FakeGitClient{
				HasFullHistoryOut:    true,
				GetCurrentBranchOut: "main",
				IsWorkingTreeCleanOut: true,
				GetLatestTagOut:      "v1.0.0",
				GetCommitsSinceOut:   []string{"feat: add"},
				GetTagHashResponses:  []testkit.GitResponse{{Out: "", Err: errors.New("no tag")}, {Out: "def456", Err: nil}},
				CreateTagErr:         nil,
				GetHeadHashOut:       "abc123",
			}
			log := &testkit.FakeLogger{}
			u := application.NewReleaseUsecase(nil, git, nil, log, testkit.NewFakeClock().Clock(), pipeline)
			_, err := u.Run(workdir)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, domain.ErrTagDoesNotPointToHead)).To(specs.BeTrue())
		})

		s.It("Run version derivation fails returns error", func(ctx *specs.Context) {
			git := &testkit.FakeGitClient{
				HasFullHistoryOut:    true,
				GetCurrentBranchOut:  "main",
				IsWorkingTreeCleanOut: true,
				GetLatestTagOut:      "not-a-version",
				GetCommitsSinceOut:   []string{"feat: x"},
			}
			log := &testkit.FakeLogger{}
			u := application.NewReleaseUsecase(nil, git, nil, log, testkit.NewFakeClock().Clock(), pipeline)
			_, err := u.Run(workdir)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "version derivation")).To(specs.BeTrue())
		})
		s.It("Run goreleaser fails returns ErrReleaseFailed", func(ctx *specs.Context) {
			cmd := &testkit.FakeCommandRunner{
				Responses: []testkit.CmdResponse{
					{Out: "goreleaser version 2.3.0", Err: nil},
					{Out: "goreleaser error", Err: errors.New("failed")},
				},
			}
			git := &testkit.FakeGitClient{
				HasFullHistoryOut:   true,
				GetCurrentBranchOut: "main",
				IsWorkingTreeCleanOut: true,
				GetLatestTagOut:     "v1.0.0",
				GetCommitsSinceOut:  []string{"feat: add"},
				GetTagHashResponses: []testkit.GitResponse{{Out: "", Err: errors.New("no tag")}, {Out: "abc123", Err: nil}},
				CreateTagErr:        nil,
				GetHeadHashOut:      "abc123",
			}
			log := &testkit.FakeLogger{}
			u := application.NewReleaseUsecase(cmd, git, testkit.NewFakeEnvProvider(nil), log, testkit.NewFakeClock().Clock(), pipeline)
			_, err := u.Run(workdir)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, domain.ErrReleaseFailed)).To(specs.BeTrue())
		})
		s.It("Run check-goreleaser-version fails returns error", func(ctx *specs.Context) {
			cmd := &testkit.FakeCommandRunner{}
			cmd.Stub("goreleaser", []string{"--version"}, "", errors.New("not found"))
			git := &testkit.FakeGitClient{
				HasFullHistoryOut:    true,
				GetCurrentBranchOut:  "main",
				IsWorkingTreeCleanOut: true,
				GetLatestTagOut:      "v1.0.0",
				GetCommitsSinceOut:   []string{"feat: add"},
				GetTagHashResponses:   []testkit.GitResponse{{Out: "", Err: errors.New("no tag")}, {Out: "abc123", Err: nil}},
				CreateTagErr:         nil,
				GetHeadHashOut:       "abc123",
			}
			log := &testkit.FakeLogger{}
			u := application.NewReleaseUsecase(cmd, git, testkit.NewFakeEnvProvider(nil), log, testkit.NewFakeClock().Clock(), pipeline)
			_, err := u.Run(workdir)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "goreleaser check")).To(specs.BeTrue())
		})

		s.It("Run has full history fails returns error", func(ctx *specs.Context) {
			git := &testkit.FakeGitClient{
				HasFullHistoryOut: false,
				HasFullHistoryErr: errors.New("check failed"),
			}
			log := &testkit.FakeLogger{}
			u := application.NewReleaseUsecase(nil, git, nil, log, testkit.NewFakeClock().Clock(), pipeline)
			_, err := u.Run(workdir)
			ctx.Expect(err != nil).To(specs.BeTrue())
		})

		s.It("ValidateVersionDerivation returns derived version", func(ctx *specs.Context) {
			git := &testkit.FakeGitClient{
				GetLatestTagOut:   "v1.0.0",
				GetCommitsSinceOut: []string{"feat: add"},
			}
			log := &testkit.FakeLogger{}
			u := application.NewReleaseUsecase(nil, git, nil, log, testkit.NewFakeClock().Clock(), pipeline)
			version, err := u.ValidateVersionDerivation(workdir)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(version).ToEqual("v1.1.0")
		})

		s.It("ValidateVersionDerivation no releaseable changes returns ErrNoReleaseableChanges", func(ctx *specs.Context) {
			git := &testkit.FakeGitClient{
				GetLatestTagOut:   "v1.0.0",
				GetCommitsSinceOut: []string{"chore: deps"},
			}
			log := &testkit.FakeLogger{}
			u := application.NewReleaseUsecase(nil, git, nil, log, testkit.NewFakeClock().Clock(), pipeline)
			_, err := u.ValidateVersionDerivation(workdir)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, domain.ErrNoReleaseableChanges)).To(specs.BeTrue())
		})

		s.It("ValidateVersionDerivation invalid tag format returns ErrInvalidLastTag", func(ctx *specs.Context) {
			git := &testkit.FakeGitClient{
				GetLatestTagOut:   "not-a-version",
				GetCommitsSinceOut: []string{"feat: x"},
			}
			log := &testkit.FakeLogger{}
			u := application.NewReleaseUsecase(nil, git, nil, log, testkit.NewFakeClock().Clock(), pipeline)
			_, err := u.ValidateVersionDerivation(workdir)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, domain.ErrInvalidLastTag)).To(specs.BeTrue())
		})
		s.It("ValidateVersionDerivation GetLatestTag error returns error", func(ctx *specs.Context) {
			git := &testkit.FakeGitClient{GetLatestTagErr: errors.New("git failed")}
			log := &testkit.FakeLogger{}
			u := application.NewReleaseUsecase(nil, git, nil, log, testkit.NewFakeClock().Clock(), pipeline)
			_, err := u.ValidateVersionDerivation(workdir)
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
		s.It("ValidateVersionDerivation GetCommitsSince error returns error", func(ctx *specs.Context) {
			git := &testkit.FakeGitClient{
				GetLatestTagOut:   "v1.0.0",
				GetCommitsSinceErr: errors.New("git failed"),
			}
			log := &testkit.FakeLogger{}
			u := application.NewReleaseUsecase(nil, git, nil, log, testkit.NewFakeClock().Clock(), pipeline)
			_, err := u.ValidateVersionDerivation(workdir)
			ctx.Expect(err != nil).To(specs.BeTrue())
		})

		s.It("Run unknown step fails returns step error", func(ctx *specs.Context) {
			log := &testkit.FakeLogger{}
			pipeline := application.Pipeline{Name: "release", Steps: []application.Step{releaseFailingStep{name: "unknown-step"}}}
			u := application.NewReleaseUsecase(nil, nil, nil, log, testkit.NewFakeClock().Clock(), pipeline)
			_, err := u.Run(workdir)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "step failed")).To(specs.BeTrue())
		})
	})
}

type releaseFailingStep struct{ name string }

func (s releaseFailingStep) Name() string                      { return s.name }
func (s releaseFailingStep) Run(_ *application.Context) error { return errors.New("step failed") }
