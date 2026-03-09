package application_test

import (
	"errors"
	"testing"

	"github.com/pablogore/devforge/internal/application"
	"github.com/pablogore/devforge/internal/steps"
	"github.com/pablogore/devforge/internal/testkit"
	"github.com/pablogore/go-specs/specs"
)

func TestDoctorUsecase(t *testing.T) {
	workdir := "/wd"
	pipeline := application.Pipeline{Name: "doctor", Steps: steps.DoctorSteps()}

	specs.Describe(t, "DoctorUsecase", func(s *specs.Spec) {
		s.It("NewDoctorUsecase returns non-nil", func(ctx *specs.Context) {
			git := &testkit.FakeGitClient{}
			cmd := &testkit.FakeCommandRunner{}
			log := &testkit.FakeLogger{}
			u := application.NewDoctorUsecase(git, cmd, log, testkit.NewFakeClock().Clock(), pipeline)
			ctx.Expect(u != nil).To(specs.BeTrue())
		})

		s.It("Run all pass returns result with 6 passed checks", func(ctx *specs.Context) {
			cmd := &testkit.FakeCommandRunner{
				Responses: []testkit.CmdResponse{
					{Out: "git version 2.0", Err: nil},
					{Out: "goreleaser version 1.0", Err: nil},
				},
			}
			git := &testkit.FakeGitClient{
				HasFullHistoryOut:     true,
				GetCurrentBranchOut:  "main",
				IsWorkingTreeCleanOut: true,
				GetLatestTagOut:       "v1.0.0",
			}
			log := &testkit.FakeLogger{}
			u := application.NewDoctorUsecase(git, cmd, log, testkit.NewFakeClock().Clock(), pipeline)
			result, err := u.Run(workdir)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(result != nil).To(specs.BeTrue())
			ctx.Expect(len(result.Checks)).ToEqual(6)
			for _, c := range result.Checks {
				ctx.Expect(c.Passed).To(specs.BeTrue())
			}
			ctx.Expect(log.LastInfoMsg).ToEqual("All doctor checks passed")
		})

		s.It("Run git not installed returns error and one failed check", func(ctx *specs.Context) {
			cmd := &testkit.FakeCommandRunner{
				Responses: []testkit.CmdResponse{{Out: "", Err: errors.New("not found")}},
			}
			git := &testkit.FakeGitClient{}
			log := &testkit.FakeLogger{}
			u := application.NewDoctorUsecase(git, cmd, log, testkit.NewFakeClock().Clock(), pipeline)
			result, err := u.Run(workdir)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(result != nil).To(specs.BeTrue())
			ctx.Expect(len(result.Checks)).ToEqual(1)
			ctx.Expect(result.Checks[0].Passed).To(specs.BeFalse())
			ctx.Expect(result.Checks[0].Name).ToEqual("git installed")
		})

		s.It("Run goreleaser not installed returns error", func(ctx *specs.Context) {
			cmd := &testkit.FakeCommandRunner{
				Responses: []testkit.CmdResponse{
					{Out: "ok", Err: nil},
					{Out: "", Err: errors.New("not found")},
				},
			}
			git := &testkit.FakeGitClient{}
			log := &testkit.FakeLogger{}
			u := application.NewDoctorUsecase(git, cmd, log, testkit.NewFakeClock().Clock(), pipeline)
			result, err := u.Run(workdir)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(len(result.Checks) >= 2).To(specs.BeTrue())
			ctx.Expect(result.Checks[1].Passed).To(specs.BeFalse())
		})

		s.It("Run shallow clone passes but full history check fails", func(ctx *specs.Context) {
			cmd := &testkit.FakeCommandRunner{
				Responses: []testkit.CmdResponse{
					{Out: "ok", Err: nil},
					{Out: "ok", Err: nil},
				},
			}
			git := &testkit.FakeGitClient{
				HasFullHistoryOut:     false,
				GetCurrentBranchOut:  "main",
				IsWorkingTreeCleanOut: true,
				GetLatestTagOut:       "v1.0.0",
			}
			log := &testkit.FakeLogger{}
			u := application.NewDoctorUsecase(git, cmd, log, testkit.NewFakeClock().Clock(), pipeline)
			result, err := u.Run(workdir)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(result != nil).To(specs.BeTrue())
			var c *application.CheckResult
			for i := range result.Checks {
				if result.Checks[i].Name == "full git history" {
					c = &result.Checks[i]
					break
				}
			}
			ctx.Expect(c != nil).To(specs.BeTrue())
			ctx.Expect(c.Passed).To(specs.BeFalse())
		})

		s.It("Run not on main passes but branch check fails", func(ctx *specs.Context) {
			cmd := &testkit.FakeCommandRunner{
				Responses: []testkit.CmdResponse{
					{Out: "ok", Err: nil},
					{Out: "ok", Err: nil},
				},
			}
			git := &testkit.FakeGitClient{
				HasFullHistoryOut:     true,
				GetCurrentBranchOut:  "develop",
				IsWorkingTreeCleanOut: true,
				GetLatestTagOut:       "v1.0.0",
			}
			log := &testkit.FakeLogger{}
			u := application.NewDoctorUsecase(git, cmd, log, testkit.NewFakeClock().Clock(), pipeline)
			result, err := u.Run(workdir)
			ctx.Expect(err).To(specs.BeNil())
			var c *application.CheckResult
			for i := range result.Checks {
				if result.Checks[i].Name == "on main branch" {
					c = &result.Checks[i]
					break
				}
			}
			ctx.Expect(c != nil).To(specs.BeTrue())
			ctx.Expect(c.Passed).To(specs.BeFalse())
		})

		s.It("Run working tree dirty fails working tree check", func(ctx *specs.Context) {
			cmd := &testkit.FakeCommandRunner{
				Responses: []testkit.CmdResponse{
					{Out: "ok", Err: nil},
					{Out: "ok", Err: nil},
				},
			}
			git := &testkit.FakeGitClient{
				HasFullHistoryOut:     true,
				GetCurrentBranchOut:  "main",
				IsWorkingTreeCleanOut: false,
				GetLatestTagOut:       "v1.0.0",
			}
			log := &testkit.FakeLogger{}
			u := application.NewDoctorUsecase(git, cmd, log, testkit.NewFakeClock().Clock(), pipeline)
			result, _ := u.Run(workdir)
			var c *application.CheckResult
			for i := range result.Checks {
				if result.Checks[i].Name == "working tree clean" {
					c = &result.Checks[i]
					break
				}
			}
			ctx.Expect(c != nil).To(specs.BeTrue())
			ctx.Expect(c.Passed).To(specs.BeFalse())
		})

		s.It("Run full history check fails returns failed check", func(ctx *specs.Context) {
			cmd := &testkit.FakeCommandRunner{
				Responses: []testkit.CmdResponse{
					{Out: "ok", Err: nil},
					{Out: "ok", Err: nil},
				},
			}
			git := &testkit.FakeGitClient{
				HasFullHistoryOut:     false,
				HasFullHistoryErr:    errors.New("check err"),
				GetCurrentBranchOut:  "main",
				IsWorkingTreeCleanOut: true,
				GetLatestTagOut:       "v1.0.0",
			}
			log := &testkit.FakeLogger{}
			u := application.NewDoctorUsecase(git, cmd, log, testkit.NewFakeClock().Clock(), pipeline)
			result, _ := u.Run(workdir)
			var c *application.CheckResult
			for i := range result.Checks {
				if result.Checks[i].Name == "full git history" {
					c = &result.Checks[i]
					break
				}
			}
			ctx.Expect(c != nil).To(specs.BeTrue())
			ctx.Expect(c.Passed).To(specs.BeFalse())
			ctx.Expect(len(c.Message) > 0).To(specs.BeTrue())
		})

		s.It("Run get branch fails fails branch check", func(ctx *specs.Context) {
			cmd := &testkit.FakeCommandRunner{
				Responses: []testkit.CmdResponse{
					{Out: "ok", Err: nil},
					{Out: "ok", Err: nil},
				},
			}
			git := &testkit.FakeGitClient{
				HasFullHistoryOut:     true,
				GetCurrentBranchErr:  errors.New("branch err"),
				IsWorkingTreeCleanOut: true,
				GetLatestTagOut:       "v1.0.0",
			}
			log := &testkit.FakeLogger{}
			u := application.NewDoctorUsecase(git, cmd, log, testkit.NewFakeClock().Clock(), pipeline)
			result, _ := u.Run(workdir)
			var c *application.CheckResult
			for i := range result.Checks {
				if result.Checks[i].Name == "on main branch" {
					c = &result.Checks[i]
					break
				}
			}
			ctx.Expect(c != nil).To(specs.BeTrue())
			ctx.Expect(c.Passed).To(specs.BeFalse())
		})

		s.It("Run working tree check fails", func(ctx *specs.Context) {
			cmd := &testkit.FakeCommandRunner{
				Responses: []testkit.CmdResponse{
					{Out: "ok", Err: nil},
					{Out: "ok", Err: nil},
				},
			}
			git := &testkit.FakeGitClient{
				HasFullHistoryOut:      true,
				GetCurrentBranchOut:   "main",
				IsWorkingTreeCleanOut: false,
				IsWorkingTreeCleanErr: errors.New("status err"),
				GetLatestTagOut:        "v1.0.0",
			}
			log := &testkit.FakeLogger{}
			u := application.NewDoctorUsecase(git, cmd, log, testkit.NewFakeClock().Clock(), pipeline)
			result, _ := u.Run(workdir)
			var c *application.CheckResult
			for i := range result.Checks {
				if result.Checks[i].Name == "working tree clean" {
					c = &result.Checks[i]
					break
				}
			}
			ctx.Expect(c != nil).To(specs.BeTrue())
			ctx.Expect(c.Passed).To(specs.BeFalse())
		})

		s.It("Run tags not accessible fails tags check", func(ctx *specs.Context) {
			cmd := &testkit.FakeCommandRunner{
				Responses: []testkit.CmdResponse{
					{Out: "ok", Err: nil},
					{Out: "ok", Err: nil},
				},
			}
			git := &testkit.FakeGitClient{
				HasFullHistoryOut:     true,
				GetCurrentBranchOut:  "main",
				IsWorkingTreeCleanOut: true,
				GetLatestTagErr:      errors.New("no tags"),
			}
			log := &testkit.FakeLogger{}
			u := application.NewDoctorUsecase(git, cmd, log, testkit.NewFakeClock().Clock(), pipeline)
			result, _ := u.Run(workdir)
			var c *application.CheckResult
			for i := range result.Checks {
				if result.Checks[i].Name == "tags accessible" {
					c = &result.Checks[i]
					break
				}
			}
			ctx.Expect(c != nil).To(specs.BeTrue())
			ctx.Expect(c.Passed).To(specs.BeFalse())
		})
	})
}
