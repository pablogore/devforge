package exec

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/pablogore/go-specs/specs"
)

func TestCommandRunner(t *testing.T) {
	specs.Describe(t, "CommandRunner", func(s *specs.Spec) {
		s.It("NewCommandRunner returns non-nil", func(ctx *specs.Context) {
			r := NewCommandRunner()
			ctx.Expect(r != nil).To(specs.BeTrue())
		})
		s.It("Run(true) succeeds with empty output", func(ctx *specs.Context) {
			r := NewCommandRunner().(*CommandRunner)
			dir := t.TempDir()
			out, err := r.Run(context.Background(), dir, "true")
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(out).ToEqual("")
		})
		s.It("Run(echo hello) returns hello", func(ctx *specs.Context) {
			r := NewCommandRunner().(*CommandRunner)
			dir := t.TempDir()
			out, err := r.Run(context.Background(), dir, "echo", "hello")
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(out == "hello\n" || out == "hello\r\n").To(specs.BeTrue())
		})
		s.It("RunCombinedOutput returns combined stdout", func(ctx *specs.Context) {
			r := NewCommandRunner().(*CommandRunner)
			dir := t.TempDir()
			out, err := r.RunCombinedOutput(context.Background(), dir, "echo", "x")
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(out == "x\n" || out == "x\r\n").To(specs.BeTrue())
		})
		s.It("RunCombinedOutputWithEnv passes env", func(ctx *specs.Context) {
			r := NewCommandRunner().(*CommandRunner)
			dir := t.TempDir()
			env := []string{"PATH=" + os.Getenv("PATH"), "TEST_VAR=ok"}
			out, err := r.RunCombinedOutputWithEnv(context.Background(), dir, env, "sh", "-c", "echo $TEST_VAR")
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(out == "ok\n" || out == "ok\r\n").To(specs.BeTrue())
		})
		s.It("RunCombinedOutputWithEnv with nil env uses process env", func(ctx *specs.Context) {
			r := NewCommandRunner().(*CommandRunner)
			dir := t.TempDir()
			out, err := r.RunCombinedOutputWithEnv(context.Background(), dir, nil, "echo", "nilenv")
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(out == "nilenv\n" || out == "nilenv\r\n").To(specs.BeTrue())
		})
		s.It("RunCombinedOutputWithEnv command failure returns output and error", func(ctx *specs.Context) {
			r := NewCommandRunner().(*CommandRunner)
			dir := t.TempDir()
			out, err := r.RunCombinedOutputWithEnv(context.Background(), dir, nil, "false")
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(out).ToEqual("")
		})
		s.It("Run(false) returns error", func(ctx *specs.Context) {
			r := NewCommandRunner().(*CommandRunner)
			dir := t.TempDir()
			_, err := r.Run(context.Background(), dir, "false")
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
		s.It("Run(nonexistent) returns error", func(ctx *specs.Context) {
			r := NewCommandRunner().(*CommandRunner)
			dir := t.TempDir()
			_, err := r.Run(context.Background(), dir, "nonexistent-binary-devforge-xyz")
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
		s.It("Run with dir uses workdir", func(ctx *specs.Context) {
			r := NewCommandRunner().(*CommandRunner)
			dir := t.TempDir()
			_ = os.WriteFile(filepath.Join(dir, "f"), []byte("x"), 0o600)
			out, err := r.Run(context.Background(), dir, "cat", "f")
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(out).ToEqual("x")
		})
	})
}
