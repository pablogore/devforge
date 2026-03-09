package steps

import (
	"context"
	"fmt"

	"github.com/pablogore/devforge/internal/application"
)

// GitInstalledStep checks that git is in PATH.
type GitInstalledStep struct{}

// Name returns the step name.
func (GitInstalledStep) Name() string { return "git-installed" }

// Run executes the check.
func (GitInstalledStep) Run(ctx *application.Context) error {
	_, err := ctx.Cmd.Run(context.Background(), "", "git", "--version")
	if err != nil {
		appendDoctorCheck(ctx, application.CheckResult{Name: "git installed", Passed: false, Message: "git not found in PATH"})
		return fmt.Errorf("git not installed")
	}
	appendDoctorCheck(ctx, application.CheckResult{Name: "git installed", Passed: true, Message: "git found"})
	return nil
}

// GoreleaserInstalledStep checks that goreleaser is in PATH.
type GoreleaserInstalledStep struct{}

// Name returns the step name.
func (GoreleaserInstalledStep) Name() string { return "goreleaser-installed" }

// Run executes the check.
func (GoreleaserInstalledStep) Run(ctx *application.Context) error {
	_, err := ctx.Cmd.Run(context.Background(), "", "goreleaser", "--version")
	if err != nil {
		appendDoctorCheck(ctx, application.CheckResult{Name: "goreleaser installed", Passed: false, Message: "goreleaser not found in PATH"})
		return fmt.Errorf("goreleaser not installed")
	}
	appendDoctorCheck(ctx, application.CheckResult{Name: "goreleaser installed", Passed: true, Message: "goreleaser found"})
	return nil
}

// FullHistoryStep checks that the repo has full git history (no shallow clone).
type FullHistoryStep struct{}

// Name returns the step name.
func (FullHistoryStep) Name() string { return "full-history" }

// Run executes the check.
func (FullHistoryStep) Run(ctx *application.Context) error {
	hasFull, err := ctx.Git.HasFullHistory(ctx.Workdir)
	if err != nil {
		appendDoctorCheck(ctx, application.CheckResult{Name: "full git history", Passed: false, Message: fmt.Sprintf("failed to check: %v", err)})
		return nil
	}
	if !hasFull {
		appendDoctorCheck(ctx, application.CheckResult{Name: "full git history", Passed: false, Message: "shallow clone detected - full history required for release"})
		return nil
	}
	appendDoctorCheck(ctx, application.CheckResult{Name: "full git history", Passed: true, Message: "full history available"})
	return nil
}

// BranchMainStep checks that the current branch is main.
type BranchMainStep struct{}

// Name returns the step name.
func (BranchMainStep) Name() string { return "branch-main" }

// Run executes the check.
func (BranchMainStep) Run(ctx *application.Context) error {
	branch, err := ctx.Git.GetCurrentBranch(ctx.Workdir)
	if err != nil {
		appendDoctorCheck(ctx, application.CheckResult{Name: "on main branch", Passed: false, Message: fmt.Sprintf("failed to get branch: %v", err)})
		return nil
	}
	if branch != "main" {
		appendDoctorCheck(ctx, application.CheckResult{Name: "on main branch", Passed: false, Message: fmt.Sprintf("currently on '%s', expected 'main'", branch)})
		return nil
	}
	appendDoctorCheck(ctx, application.CheckResult{Name: "on main branch", Passed: true, Message: "on main branch"})
	return nil
}

// WorkingTreeCleanStep checks that the working tree has no uncommitted changes.
type WorkingTreeCleanStep struct{}

// Name returns the step name.
func (WorkingTreeCleanStep) Name() string { return "working-tree-clean" }

// Run executes the check.
func (WorkingTreeCleanStep) Run(ctx *application.Context) error {
	isClean, err := ctx.Git.IsWorkingTreeClean(ctx.Workdir)
	if err != nil {
		appendDoctorCheck(ctx, application.CheckResult{Name: "working tree clean", Passed: false, Message: fmt.Sprintf("failed to check: %v", err)})
		return nil
	}
	if !isClean {
		appendDoctorCheck(ctx, application.CheckResult{Name: "working tree clean", Passed: false, Message: "uncommitted changes detected"})
		return nil
	}
	appendDoctorCheck(ctx, application.CheckResult{Name: "working tree clean", Passed: true, Message: "working tree is clean"})
	return nil
}

// TagsAccessibleStep checks that git tags are accessible.
type TagsAccessibleStep struct{}

// Name returns the step name.
func (TagsAccessibleStep) Name() string { return "tags-accessible" }

// Run executes the check.
func (TagsAccessibleStep) Run(ctx *application.Context) error {
	_, err := ctx.Git.GetLatestTag(ctx.Workdir)
	if err != nil {
		appendDoctorCheck(ctx, application.CheckResult{Name: "tags accessible", Passed: false, Message: fmt.Sprintf("failed to access tags: %v", err)})
		return nil
	}
	appendDoctorCheck(ctx, application.CheckResult{Name: "tags accessible", Passed: true, Message: "tags accessible"})
	return nil
}

func appendDoctorCheck(ctx *application.Context, c application.CheckResult) {
	if ctx.DoctorChecks != nil {
		*ctx.DoctorChecks = append(*ctx.DoctorChecks, c)
	}
}

func init() {
	application.RegisterStep("git-installed", func() application.Step { return GitInstalledStep{} })
	application.RegisterStep("goreleaser-installed", func() application.Step { return GoreleaserInstalledStep{} })
	application.RegisterStep("full-history", func() application.Step { return FullHistoryStep{} })
	application.RegisterStep("branch-main", func() application.Step { return BranchMainStep{} })
	application.RegisterStep("working-tree-clean", func() application.Step { return WorkingTreeCleanStep{} })
	application.RegisterStep("tags-accessible", func() application.Step { return TagsAccessibleStep{} })
}

// DoctorSteps returns the default doctor pipeline steps in deterministic order.
func DoctorSteps() []application.Step {
	return []application.Step{
		GitInstalledStep{},
		GoreleaserInstalledStep{},
		FullHistoryStep{},
		BranchMainStep{},
		WorkingTreeCleanStep{},
		TagsAccessibleStep{},
	}
}
