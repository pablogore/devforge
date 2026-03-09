package testkit

import (
	"context"
	"errors"
	"testing"

	"github.com/pablogore/go-specs/specs"
)

func TestFakeCommandRunner_Stub(t *testing.T) {
	specs.Describe(t, "FakeCommandRunner.Stub", func(s *specs.Spec) {
		s.It("returns stubbed result for exact name and args", func(ctx *specs.Context) {
			runner := NewFakeCommandRunner()
			runner.Stub("git", []string{"status"}, "clean", nil)

			out, err := runner.RunCombinedOutput(context.Background(), "/wd", "git", "status")

			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(out).ToEqual("clean")
			ctx.Expect(runner.CallCount()).ToEqual(1)
			ctx.Expect(runner.WasCalled("git", "status")).To(specs.BeTrue())
		})
		s.It("returns stubbed error when stub has error", func(ctx *specs.Context) {
			runner := NewFakeCommandRunner()
			wantErr := errors.New("exit 1")
			runner.Stub("go", []string{"list", "./..."}, "", wantErr)

			out, err := runner.RunCombinedOutput(context.Background(), "/repo", "go", "list", "./...")

			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(err == wantErr).To(specs.BeTrue())
			ctx.Expect(out).ToEqual("")
		})
		s.It("records dir in Calls", func(ctx *specs.Context) {
			runner := NewFakeCommandRunner()
			runner.Stub("git", []string{"status"}, "ok", nil)
			_, _ = runner.Run(context.Background(), "/my/dir", "git", "status")

			call, ok := runner.LastCall()
			ctx.Expect(ok).To(specs.BeTrue())
			ctx.Expect(call.Dir).ToEqual("/my/dir")
			ctx.Expect(call.Name).ToEqual("git")
			ctx.Expect(len(call.Args)).ToEqual(1)
			ctx.Expect(call.Args[0]).ToEqual("status")
		})
	})
}

func TestFakeCommandRunner_Enqueue(t *testing.T) {
	specs.Describe(t, "FakeCommandRunner.Enqueue", func(s *specs.Spec) {
		s.It("returns queued results in order for repeated calls", func(ctx *specs.Context) {
			runner := NewFakeCommandRunner()
			runner.Enqueue("git", []string{"status"}, "dirty", nil)
			runner.Enqueue("git", []string{"status"}, "clean", nil)

			out1, err1 := runner.RunCombinedOutput(context.Background(), "/wd", "git", "status")
			out2, err2 := runner.RunCombinedOutput(context.Background(), "/wd", "git", "status")

			ctx.Expect(err1).To(specs.BeNil())
			ctx.Expect(out1).ToEqual("dirty")
			ctx.Expect(err2).To(specs.BeNil())
			ctx.Expect(out2).ToEqual("clean")
			ctx.Expect(runner.CallCount()).ToEqual(2)
		})
		s.It("returns queued error on second call", func(ctx *specs.Context) {
			runner := NewFakeCommandRunner()
			runner.Enqueue("go", []string{"build"}, "", nil)
			runner.Enqueue("go", []string{"build"}, "", errors.New("build failed"))

			_, err1 := runner.RunCombinedOutput(context.Background(), "", "go", "build")
			_, err2 := runner.RunCombinedOutput(context.Background(), "", "go", "build")

			ctx.Expect(err1).To(specs.BeNil())
			ctx.Expect(err2 != nil).To(specs.BeTrue())
		})
	})
}

func TestFakeCommandRunner_Recording(t *testing.T) {
	specs.Describe(t, "FakeCommandRunner recording", func(s *specs.Spec) {
		s.It("CallCount returns number of calls", func(ctx *specs.Context) {
			runner := NewFakeCommandRunner()
			runner.Stub("a", nil, "", nil)
			runner.Stub("b", nil, "", nil)
			_, _ = runner.Run(context.Background(), "", "a")
			_, _ = runner.Run(context.Background(), "", "b")
			ctx.Expect(runner.CallCount()).ToEqual(2)
		})
		s.It("LastCall returns zero value when no calls", func(ctx *specs.Context) {
			runner := NewFakeCommandRunner()
			_, ok := runner.LastCall()
			ctx.Expect(ok).To(specs.BeFalse())
		})
		s.It("WasCalled returns false for different args", func(ctx *specs.Context) {
			runner := NewFakeCommandRunner()
			runner.Stub("go", []string{"list"}, "", nil)
			_, _ = runner.RunCombinedOutput(context.Background(), "", "go", "list")
			ctx.Expect(runner.WasCalled("go", "list")).To(specs.BeTrue())
			ctx.Expect(runner.WasCalled("go", "build")).To(specs.BeFalse())
		})
	})
}

func TestFakeCommandRunner_Default(t *testing.T) {
	specs.Describe(t, "FakeCommandRunner.Default", func(s *specs.Spec) {
		s.It("returns default for unstubbed command", func(ctx *specs.Context) {
			runner := NewFakeCommandRunner()
			runner.Default = &CommandResult{Stdout: "", Err: errors.New("unexpected command")}

			_, err := runner.RunCombinedOutput(context.Background(), "/wd", "unknown", "cmd")

			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(err.Error()).ToEqual("unexpected command")
		})
		s.It("returns zero value when no stub and Default is nil", func(ctx *specs.Context) {
			runner := NewFakeCommandRunner()
			out, err := runner.RunCombinedOutput(context.Background(), "", "anything")
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(out).ToEqual("")
		})
		s.It("RequireNoUnexpectedCalls returns error when Default was used", func(ctx *specs.Context) {
			runner := NewFakeCommandRunner()
			runner.Default = &CommandResult{Err: errors.New("unexpected")}
			_, _ = runner.Run(context.Background(), "", "x")
			err := runner.RequireNoUnexpectedCalls()
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(err.Error()).ToEqual("fake_command_runner: one or more calls had no stub or queued result (used Default)")
		})
		s.It("RequireNoUnexpectedCalls returns nil when all calls stubbed", func(ctx *specs.Context) {
			runner := NewFakeCommandRunner()
			runner.Stub("git", []string{"status"}, "ok", nil)
			_, _ = runner.Run(context.Background(), "", "git", "status")
			ctx.Expect(runner.RequireNoUnexpectedCalls()).To(specs.BeNil())
		})
	})
}

func TestFakeCommandRunner_LegacyResponses(t *testing.T) {
	specs.Describe(t, "FakeCommandRunner legacy Responses", func(s *specs.Spec) {
		s.It("consumes Responses in order when no stub or queue", func(ctx *specs.Context) {
			runner := &FakeCommandRunner{
				exactStub: make(map[string]CommandResult),
				Responses: []CmdResponse{
					{Out: "first", Err: nil},
					{Out: "second", Err: nil},
				},
			}
			out1, _ := runner.RunCombinedOutput(context.Background(), "", "a")
			out2, _ := runner.RunCombinedOutput(context.Background(), "", "b")
			ctx.Expect(out1).ToEqual("first")
			ctx.Expect(out2).ToEqual("second")
		})
	})
}

func TestFakeCommandRunner_RunCombinedOutputWithEnv(t *testing.T) {
	specs.Describe(t, "FakeCommandRunner.RunCombinedOutputWithEnv", func(s *specs.Spec) {
		s.It("delegates to run with dir and name/args, env ignored for lookup", func(ctx *specs.Context) {
			runner := NewFakeCommandRunner()
			runner.Stub("go", []string{"build"}, "built", nil)
			out, err := runner.RunCombinedOutputWithEnv(context.Background(), "/wd", []string{"GOOS=linux"}, "go", "build")
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(out).ToEqual("built")
			call, ok := runner.LastCall()
			ctx.Expect(ok).To(specs.BeTrue())
			ctx.Expect(call.Dir).ToEqual("/wd")
			ctx.Expect(call.Name).ToEqual("go")
		})
	})
}

func TestFakeCommandRunner_Reset(t *testing.T) {
	specs.Describe(t, "FakeCommandRunner.Reset", func(s *specs.Spec) {
		s.It("clears Calls and defaultUsed, keeps stubs", func(ctx *specs.Context) {
			runner := NewFakeCommandRunner()
			runner.Stub("git", []string{"status"}, "ok", nil)
			_, _ = runner.Run(context.Background(), "/d", "git", "status")
			ctx.Expect(runner.CallCount()).ToEqual(1)
			runner.Reset()
			ctx.Expect(runner.CallCount()).ToEqual(0)
			_, ok := runner.LastCall()
			ctx.Expect(ok).To(specs.BeFalse())
			out, err := runner.Run(context.Background(), "/d", "git", "status")
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(out).ToEqual("ok")
		})
	})
}
