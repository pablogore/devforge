package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/pablogore/devforge/internal/domain"
	"github.com/pablogore/devforge/internal/ports"
)

// ReleaseUsecase runs the release pipeline (preconditions, version derivation, tag, goreleaser).
type ReleaseUsecase struct {
	commandRunner ports.CommandRunner
	gitClient     ports.GitClient
	envProvider   ports.EnvProvider
	logger        ports.Logger
	clock         ports.Clock
	runner        *StepRunner
	pipeline      Pipeline
}

// NewReleaseUsecase returns a new ReleaseUsecase.
func NewReleaseUsecase(cmd ports.CommandRunner, git ports.GitClient, env ports.EnvProvider, logger ports.Logger, clock ports.Clock, pipeline Pipeline) *ReleaseUsecase {
	return &ReleaseUsecase{
		commandRunner: cmd,
		gitClient:     git,
		envProvider:   env,
		logger:        logger,
		clock:         clock,
		runner:        NewStepRunner(logger, clock),
		pipeline:      pipeline,
	}
}

// ReleaseResult holds the version and commit message produced by a successful release.
type ReleaseResult struct {
	// Version is the new semantic version tag (e.g. v1.2.3).
	Version string
	// CommitMsg is the subject of the release commit (e.g. for changelog).
	CommitMsg string
}

// Run executes the release pipeline and returns the new version on success.
func (u *ReleaseUsecase) Run(workdir string) (*ReleaseResult, error) {
	start := u.clock.Now()
	u.logger.Info("Starting release", "workdir", workdir)

	ctx := &Context{
		StdCtx:  context.Background(),
		Cmd:     u.commandRunner,
		Git:     u.gitClient,
		Env:     u.envProvider,
		Log:     u.logger,
		Clock:   u.clock,
		Workdir: workdir,
	}
	for _, step := range u.pipeline.Steps {
		if err := u.runner.Run(ctx, step); err != nil {
			totalMs := u.clock.Since(start).Milliseconds()
			u.logger.Error("Release failed", "step", step.Name(), "error", err.Error(), "total_duration_ms", totalMs)
			switch step.Name() {
			case "preconditions":
				return nil, fmt.Errorf("precondition failed: %w", err)
			case "version-derivation":
				return nil, fmt.Errorf("version derivation failed: %w", err)
			case "create-tag":
				return nil, fmt.Errorf("tag creation failed: %w", err)
			case "verify-tag":
				return nil, fmt.Errorf("tag verification failed: %w", err)
			case "check-goreleaser-version":
				return nil, fmt.Errorf("goreleaser check failed: %w", err)
			case "goreleaser":
				return nil, fmt.Errorf("goreleaser failed: %w", err)
			default:
				return nil, err
			}
		}
	}

	totalMs := u.clock.Since(start).Milliseconds()
	u.logger.Info("Release completed", "version", ctx.Version, "total_duration_ms", totalMs)

	return &ReleaseResult{
		Version:   ctx.Version,
		CommitMsg: fmt.Sprintf("Release %s", ctx.Version),
	}, nil
}

// ValidateVersionDerivation derives the next version without creating a tag; used for dry-run or validation.
func (u *ReleaseUsecase) ValidateVersionDerivation(workdir string) (string, error) {
	lastTag, err := u.gitClient.GetLatestTag(workdir)
	if err != nil {
		return "", err
	}

	commits, err := u.gitClient.GetCommitsSince(workdir, lastTag)
	if err != nil {
		return "", err
	}

	version, err := domain.DeriveNextVersion(commits, lastTag)
	if err != nil {
		if err == domain.ErrNoReleaseableChanges {
			return "", err
		}
		return "", err
	}

	versionStr := version.String()
	if !strings.HasPrefix(versionStr, "v") {
		return "", domain.ErrInvalidVersionFormat
	}

	if err := domain.ValidateVersionFormat(versionStr); err != nil {
		return "", err
	}

	return versionStr, nil
}
