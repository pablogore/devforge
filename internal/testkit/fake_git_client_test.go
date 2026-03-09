package testkit

import (
	"errors"
	"testing"

	"github.com/pablogore/go-specs/specs"
)

func TestFakeGitClient(t *testing.T) {
	specs.Describe(t, "FakeGitClient", func(s *specs.Spec) {
		s.It("GetCurrentBranch returns Out and Err", func(ctx *specs.Context) {
			f := &FakeGitClient{GetCurrentBranchOut: "main", GetCurrentBranchErr: nil}
			out, err := f.GetCurrentBranch("/wd")
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(out).ToEqual("main")
			f.GetCurrentBranchErr = errors.New("not a repo")
			_, err = f.GetCurrentBranch("/wd")
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
		s.It("GetLatestTag returns Out and Err", func(ctx *specs.Context) {
			f := &FakeGitClient{GetLatestTagOut: "v1.0.0"}
			out, err := f.GetLatestTag("/wd")
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(out).ToEqual("v1.0.0")
		})
		s.It("GetCommitsSince returns slice and Err", func(ctx *specs.Context) {
			f := &FakeGitClient{GetCommitsSinceOut: []string{"feat: x"}}
			out, err := f.GetCommitsSince("/wd", "")
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(out)).ToEqual(1)
			ctx.Expect(out[0]).ToEqual("feat: x")
		})
		s.It("CreateTag returns Err", func(ctx *specs.Context) {
			f := &FakeGitClient{}
			ctx.Expect(f.CreateTag("/wd", "v1.0.0")).To(specs.BeNil())
			f.CreateTagErr = errors.New("tag exists")
			ctx.Expect(f.CreateTag("/wd", "v1.0.0") != nil).To(specs.BeTrue())
		})
		s.It("IsWorkingTreeClean returns Out and Err", func(ctx *specs.Context) {
			f := &FakeGitClient{IsWorkingTreeCleanOut: true}
			ok, err := f.IsWorkingTreeClean("/wd")
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(ok).To(specs.BeTrue())
		})
		s.It("HasFullHistory returns Out and Err", func(ctx *specs.Context) {
			f := &FakeGitClient{HasFullHistoryOut: false}
			ok, err := f.HasFullHistory("/wd")
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(ok).To(specs.BeFalse())
		})
		s.It("GetHeadHash and GetTagHash return Out and Err", func(ctx *specs.Context) {
			f := &FakeGitClient{GetHeadHashOut: "abc123", GetTagHashOut: "abc123"}
			out, err := f.GetHeadHash("/wd")
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(out).ToEqual("abc123")
			out, err = f.GetTagHash("/wd", "v1.0.0")
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(out).ToEqual("abc123")
		})
		s.It("GetTagHashResponses consumed in order", func(ctx *specs.Context) {
			f := &FakeGitClient{
				GetTagHashResponses: []GitResponse{{Out: "a", Err: nil}, {Out: "b", Err: nil}},
			}
			out1, _ := f.GetTagHash("/wd", "v1")
			out2, _ := f.GetTagHash("/wd", "v1")
			ctx.Expect(out1).ToEqual("a")
			ctx.Expect(out2).ToEqual("b")
		})
		s.It("DiffExitCode returns Err", func(ctx *specs.Context) {
			f := &FakeGitClient{}
			ctx.Expect(f.DiffExitCode("/wd", "f")).To(specs.BeNil())
			f.DiffExitCodeErr = errors.New("diff")
			ctx.Expect(f.DiffExitCode("/wd", "f") != nil).To(specs.BeTrue())
		})
		s.It("GetLatestCommitMessage returns Out and Err", func(ctx *specs.Context) {
			f := &FakeGitClient{GetLatestCommitMessageOut: "feat: add"}
			out, err := f.GetLatestCommitMessage("/wd")
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(out).ToEqual("feat: add")
		})
	})
}
