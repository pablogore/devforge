package steps

import (
	"fmt"

	"github.com/pablogore/devforge/internal/application"
	"github.com/pablogore/devforge/internal/domain"
)

// PreconditionsStep validates branch, history, and working tree before release.
type PreconditionsStep struct{}

// Name returns the step name.
func (PreconditionsStep) Name() string { return "preconditions" }

// Run executes the preconditions check.
func (PreconditionsStep) Run(ctx *application.Context) error {
	ctx.Log.Info("Validating preconditions", "step", "preconditions")

	hasFullHistory, err := ctx.Git.HasFullHistory(ctx.Workdir)
	if err != nil {
		return fmt.Errorf("failed to check git history: %w", err)
	}
	if !hasFullHistory {
		return domain.ErrShallowCloneDetected
	}

	branch, err := ctx.Git.GetCurrentBranch(ctx.Workdir)
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}
	if branch != "main" {
		return fmt.Errorf("%w: current=%s", domain.ErrNotOnMainBranch, branch)
	}

	isClean, err := ctx.Git.IsWorkingTreeClean(ctx.Workdir)
	if err != nil {
		return fmt.Errorf("failed to check working tree: %w", err)
	}
	if !isClean {
		return domain.ErrWorkingTreeDirty
	}

	return nil
}

// VersionDerivationStep derives the next semantic version from commits since the last tag.
type VersionDerivationStep struct{}

// Name returns the step name.
func (VersionDerivationStep) Name() string { return "version-derivation" }

// Run executes version derivation and stores the result in ctx.Version.
func (VersionDerivationStep) Run(ctx *application.Context) error {
	ctx.Log.Info("Deriving version", "step", "version-derivation")

	lastTag, err := ctx.Git.GetLatestTag(ctx.Workdir)
	if err != nil {
		return fmt.Errorf("failed to get latest tag: %w", err)
	}

	commits, err := ctx.Git.GetCommitsSince(ctx.Workdir, lastTag)
	if err != nil {
		return fmt.Errorf("failed to get commits: %w", err)
	}

	version, err := domain.DeriveNextVersion(commits, lastTag)
	if err != nil {
		if err == domain.ErrNoReleaseableChanges {
			return err
		}
		return fmt.Errorf("version derivation failed: %w", err)
	}

	versionStr := version.String()

	tagHash, err := ctx.Git.GetTagHash(ctx.Workdir, versionStr)
	if err == nil && tagHash != "" {
		return fmt.Errorf("%w: %s", domain.ErrTagAlreadyExists, versionStr)
	}

	version2, err := domain.DeriveNextVersion(commits, lastTag)
	if err != nil {
		return fmt.Errorf("%w: second derivation failed", domain.ErrIdempotencyCheckFailed)
	}
	if version2.String() != versionStr {
		return fmt.Errorf("%w: version1=%s version2=%s", domain.ErrIdempotencyCheckFailed, versionStr, version2.String())
	}

	ctx.Version = versionStr
	return nil
}

// TagCreationStep creates the release git tag.
type TagCreationStep struct{}

// Name returns the step name.
func (TagCreationStep) Name() string { return "create-tag" }

// Run creates the tag at HEAD using ctx.Version.
func (TagCreationStep) Run(ctx *application.Context) error {
	ctx.Log.Info("Creating tag", "version", ctx.Version)
	if err := ctx.Git.CreateTag(ctx.Workdir, ctx.Version); err != nil {
		return err
	}
	return nil
}

// TagVerificationStep verifies the git tag points to HEAD before running goreleaser.
type TagVerificationStep struct{}

// Name returns the step name.
func (TagVerificationStep) Name() string { return "verify-tag" }

// Run verifies the tag commit matches HEAD.
func (TagVerificationStep) Run(ctx *application.Context) error {
	ctx.Log.Info("Verifying tag points to HEAD", "step", "verify-tag", "version", ctx.Version)

	headHash, err := ctx.Git.GetHeadHash(ctx.Workdir)
	if err != nil {
		return fmt.Errorf("failed to get HEAD hash: %w", err)
	}

	tagHash, err := ctx.Git.GetTagHash(ctx.Workdir, ctx.Version)
	if err != nil {
		return fmt.Errorf("failed to get tag hash: %w", err)
	}

	if headHash != tagHash {
		return fmt.Errorf("%w: head=%s tag=%s", domain.ErrTagDoesNotPointToHead, headHash, tagHash)
	}

	return nil
}

// GoreleaserStep runs goreleaser release to build and publish artifacts.
type GoreleaserStep struct{}

// Name returns the step name.
func (GoreleaserStep) Name() string { return "goreleaser" }

// Run executes goreleaser release --clean.
func (GoreleaserStep) Run(ctx *application.Context) error {
	ctx.Log.Info("Running goreleaser", "version", ctx.Version)
	ctx.Log.Info("Executing goreleaser", "step", "goreleaser")
	output, err := ctx.Cmd.Run(ctx.StdCtx, ctx.Workdir, "goreleaser", "release", "--clean")
	if err != nil {
		return fmt.Errorf("%w: %s", domain.ErrReleaseFailed, output)
	}
	return nil
}

func init() {
	application.RegisterStep("preconditions", func() application.Step { return PreconditionsStep{} })
	application.RegisterStep("version-derivation", func() application.Step { return VersionDerivationStep{} })
	application.RegisterStep("create-tag", func() application.Step { return TagCreationStep{} })
	application.RegisterStep("verify-tag", func() application.Step { return TagVerificationStep{} })
	application.RegisterStep("goreleaser", func() application.Step { return GoreleaserStep{} })
}

// ReleaseSteps returns the default release steps in deterministic order.
func ReleaseSteps() []application.Step {
	return []application.Step{
		PreconditionsStep{},
		VersionDerivationStep{},
		TagCreationStep{},
		TagVerificationStep{},
		NewCheckGoreleaserVersionStep(),
		GoreleaserStep{},
	}
}
