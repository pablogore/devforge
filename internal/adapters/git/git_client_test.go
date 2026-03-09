package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/pablogore/go-specs/specs"
)

// initTempRepo creates a temporary directory with a git repo (one commit, no tag).
// Uses an empty template dir so git init does not create hooks (avoids permission issues in sandboxes).
func initTempRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	emptyTemplate := t.TempDir()
	runGitWithEnv(t, dir, []string{"GIT_TEMPLATE_DIR=" + emptyTemplate}, "init")
	runGit(t, dir, "config", "user.email", "test@test.com")
	runGit(t, dir, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(dir, "f"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "initial")
	return dir
}

// initTempRepoWithTag is like initTempRepo but adds an annotated tag.
func initTempRepoWithTag(t *testing.T, tag string) string {
	t.Helper()
	dir := initTempRepo(t)
	runGit(t, dir, "tag", "-a", tag, "-m", "release "+tag)
	return dir
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	runGitWithEnv(t, dir, nil, args...)
}

func runGitWithEnv(t *testing.T, dir string, env []string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestGitClient(t *testing.T) {
	specs.Describe(t, "GitClient", func(s *specs.Spec) {
		s.It("NewGitClient returns non-nil", func(ctx *specs.Context) {
			g := NewGitClient()
			ctx.Expect(g != nil).To(specs.BeTrue())
		})
		s.It("GetCurrentBranch(non-repo) returns error", func(ctx *specs.Context) {
			g := NewGitClient().(*Client)
			dir := t.TempDir()
			_, err := g.GetCurrentBranch(dir)
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
		s.It("GetLatestTag(non-repo) returns empty or error", func(ctx *specs.Context) {
			g := NewGitClient().(*Client)
			dir := t.TempDir()
			tag, err := g.GetLatestTag(dir)
			if err != nil {
				return
			}
			ctx.Expect(tag).ToEqual("")
		})
		s.It("GetCommitsSince(non-repo) returns error", func(ctx *specs.Context) {
			g := NewGitClient().(*Client)
			dir := t.TempDir()
			_, err := g.GetCommitsSince(dir, "")
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
		s.It("IsWorkingTreeClean(non-repo) returns error", func(ctx *specs.Context) {
			g := NewGitClient().(*Client)
			dir := t.TempDir()
			_, err := g.IsWorkingTreeClean(dir)
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
		s.It("HasFullHistory(non-repo) returns error", func(ctx *specs.Context) {
			g := NewGitClient().(*Client)
			dir := t.TempDir()
			_, err := g.HasFullHistory(dir)
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
		s.It("GetHeadHash(non-repo) returns error", func(ctx *specs.Context) {
			g := NewGitClient().(*Client)
			dir := t.TempDir()
			_, err := g.GetHeadHash(dir)
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
		s.It("GetTagHash(non-repo) returns error", func(ctx *specs.Context) {
			g := NewGitClient().(*Client)
			dir := t.TempDir()
			_, err := g.GetTagHash(dir, "v1.0.0")
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
		s.It("GetLatestCommitMessage(non-repo) returns error", func(ctx *specs.Context) {
			g := NewGitClient().(*Client)
			dir := t.TempDir()
			_, err := g.GetLatestCommitMessage(dir)
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
		s.It("DiffExitCode(non-repo) returns error", func(ctx *specs.Context) {
			g := NewGitClient().(*Client)
			dir := t.TempDir()
			err := g.DiffExitCode(dir, "x")
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
		s.It("CreateTag(non-repo) returns error", func(ctx *specs.Context) {
			g := NewGitClient().(*Client)
			dir := t.TempDir()
			err := g.CreateTag(dir, "v1.0.0")
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
	})
}

func TestClient_GetLatestTag_inRepo(t *testing.T) {
	specs.Describe(t, "GitClient GetLatestTag in repo", func(s *specs.Spec) {
		s.It("returns no error in repo with no tags", func(ctx *specs.Context) {
			dir := initTempRepo(t)
			g := NewGitClient().(*Client)
			tag, err := g.GetLatestTag(dir)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(tag).ToEqual("")
		})
		s.It("returns tag when repo has tag", func(ctx *specs.Context) {
			dir := initTempRepoWithTag(t, "v1.0.0")
			g := NewGitClient().(*Client)
			tag, err := g.GetLatestTag(dir)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(tag).ToEqual("v1.0.0")
		})
	})
}

func TestClient_inRepo(t *testing.T) {
	specs.Describe(t, "GitClient in repo", func(s *specs.Spec) {
		s.It("GetCurrentBranch and other methods succeed", func(ctx *specs.Context) {
			dir := initTempRepoWithTag(t, "v0.1.0")
			g := NewGitClient().(*Client)
			branch, err := g.GetCurrentBranch(dir)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(branch != "").To(specs.BeTrue())
			tag, err := g.GetLatestTag(dir)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(tag).ToEqual("v0.1.0")
			commits, err := g.GetCommitsSince(dir, "")
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(commits) >= 1).To(specs.BeTrue())
		})
		s.It("HasFullHistory in repo returns true", func(ctx *specs.Context) {
			dir := initTempRepo(t)
			g := NewGitClient().(*Client)
			hasFull, err := g.HasFullHistory(dir)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(hasFull).To(specs.BeTrue())
		})
		s.It("GetHeadHash and GetTagHash in repo succeed", func(ctx *specs.Context) {
			dir := initTempRepoWithTag(t, "v0.2.0")
			g := NewGitClient().(*Client)
			headHash, err := g.GetHeadHash(dir)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(headHash) > 0).To(specs.BeTrue())
			tagHash, err := g.GetTagHash(dir, "v0.2.0")
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(tagHash).ToEqual(headHash)
		})
		s.It("DiffExitCode and GetLatestCommitMessage in repo succeed", func(ctx *specs.Context) {
			dir := initTempRepo(t)
			g := NewGitClient().(*Client)
			err := g.DiffExitCode(dir, "f")
			ctx.Expect(err).To(specs.BeNil())
			msg, err := g.GetLatestCommitMessage(dir)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(msg).ToEqual("initial")
		})
	})
}
