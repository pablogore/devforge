package detection

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pablogore/go-specs/specs"
)

func TestDetectProfile(t *testing.T) {
	specs.Describe(t, "DetectProfile", func(s *specs.Spec) {
		s.It("returns go-lib when go.mod present and no cmd", func(ctx *specs.Context) {
			dir := t.TempDir()
			ctx.Expect(os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module foo\n"), 0o600)).To(specs.BeNil())
			got := DetectProfile(dir)
			ctx.Expect(got).ToEqual(string(RepoGoLib))
		})
		s.It("returns go-service when go.mod and cmd exist", func(ctx *specs.Context) {
			dir := t.TempDir()
			ctx.Expect(os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module foo\n"), 0o600)).To(specs.BeNil())
			ctx.Expect(os.MkdirAll(filepath.Join(dir, "cmd"), 0o750)).To(specs.BeNil())
			got := DetectProfile(dir)
			ctx.Expect(got).ToEqual(string(RepoGoService))
		})
		s.It("returns go-lib when no go.mod", func(ctx *specs.Context) {
			dir := t.TempDir()
			got := DetectProfile(dir)
			ctx.Expect(got).ToEqual(string(RepoGoLib))
		})
		s.It("returns go-lib when cmd exists but no go.mod", func(ctx *specs.Context) {
			dir := t.TempDir()
			ctx.Expect(os.MkdirAll(filepath.Join(dir, "cmd"), 0o750)).To(specs.BeNil())
			got := DetectProfile(dir)
			ctx.Expect(got).ToEqual(string(RepoGoLib))
		})
	})
}
