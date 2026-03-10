package application

import (
	"context"

	"github.com/pablogore/devforge/internal/config"
	"github.com/pablogore/devforge/internal/coverage"
	"github.com/pablogore/devforge/internal/ports"
)

// PRUsecase runs the PR validation pipeline.
type PRUsecase struct {
	commandRunner     ports.CommandRunner
	gitClient         ports.GitClient
	envProvider       ports.EnvProvider
	coverageThreshold float64
	logger            ports.Logger
	clock             ports.Clock
	profileName       string
	runner            *StepRunner
	pipeline          Pipeline
}

// NewPRUsecase returns a new PRUsecase.
func NewPRUsecase(cmd ports.CommandRunner, git ports.GitClient, env ports.EnvProvider, coverageThreshold float64, logger ports.Logger, clock ports.Clock, profileName string, pipeline Pipeline) *PRUsecase {
	return &PRUsecase{
		commandRunner:     cmd,
		gitClient:         git,
		envProvider:       env,
		coverageThreshold: coverageThreshold,
		logger:            logger,
		clock:             clock,
		profileName:       profileName,
		runner:            NewStepRunner(logger, clock),
		pipeline:          pipeline,
	}
}

// Run runs the PR pipeline using the latest commit message for conventional-commit validation.
func (u *PRUsecase) Run(workdir, baseRef string) error {
	return u.RunWithTitle(workdir, baseRef, "")
}

// RunWithTitle runs the PR pipeline with an optional title override for conventional-commit.
func (u *PRUsecase) RunWithTitle(workdir, baseRef, titleOverride string) error {
	return u.RunWithTitleAndPluginConfig(workdir, baseRef, titleOverride, nil, nil)
}

// RunWithTitleAndPluginConfig runs PR validation with optional per-plugin config and pipeline step filtering from .devforge.yml.
// pluginConfig is keyed by plugin name; nil means no config (all discovered plugins run normally). cfg is the loaded config; nil means default pipeline (no filtering).
func (u *PRUsecase) RunWithTitleAndPluginConfig(workdir, _ string, titleOverride string, pluginConfig map[string]ExternalPluginConfig, cfg *config.Config) error {
	start := u.clock.Now()
	u.logger.Info("Starting PR validation", "workdir", workdir, "profile", u.profileName)

	stdCtx := context.Background()
	ctx := &Context{
		StdCtx:               stdCtx,
		Cmd:                  u.commandRunner,
		Git:                  u.gitClient,
		Env:                  u.envProvider,
		Log:                  u.logger,
		Clock:                u.clock,
		Workdir:              workdir,
		CoverageThreshold:    u.coverageThreshold,
		TitleOverride:        titleOverride,
		ProfileName:          u.profileName,
		ExternalPluginConfig: pluginConfig,
		Config:               cfg,
	}
	if cfg != nil && cfg.Policies != nil && cfg.Policies.Coverage != nil {
		pol := cfg.Policies.Coverage
		if pol.Threshold > 0 {
			ctx.CoverageThreshold = float64(pol.Threshold)
		}
		if len(pol.Packages) > 0 {
			if err := coverage.ValidateCoveragePatterns(pol.Packages); err != nil {
				return err
			}
			resolved, err := coverage.ResolveCoveragePackages(stdCtx, workdir, pol.Packages, u.commandRunner)
			if err != nil {
				return err
			}
			ctx.CoverPkg = coverage.BuildCoverPkgFlag(resolved)
			ctx.CoveragePackagesResolved = resolved
		} else if pol.Threshold > 0 {
			// Threshold set but no packages: use all module packages so coverage is measured repo-wide (e.g. kit-core with no internal/).
			resolved, err := coverage.ResolveCoveragePackages(stdCtx, workdir, []string{"*"}, u.commandRunner)
			if err != nil {
				return err
			}
			ctx.CoverPkg = coverage.BuildCoverPkgFlag(resolved)
			ctx.CoveragePackagesResolved = resolved
		}
		if pol.Threshold > 0 || len(pol.Packages) > 0 {
			u.logger.Info("Coverage policy applied from .devforge.yml", "threshold", ctx.CoverageThreshold, "packages", pol.Packages)
		}
	} else if cfg != nil {
		u.logger.Info("Using profile default coverage threshold", "threshold", ctx.CoverageThreshold, "reason", "no policies.coverage in .devforge.yml or file not found")
	}
	steps := filterSteps(u.pipeline.Steps, cfg)
	for _, step := range steps {
		if err := u.runner.Run(ctx, step); err != nil {
			totalMs := u.clock.Since(start).Milliseconds()
			u.logger.Error("PR validation failed", "step", step.Name(), "error", err.Error(), "total_duration_ms", totalMs)
			return err
		}
	}

	totalMs := u.clock.Since(start).Milliseconds()
	u.logger.Info("PR validation completed", "profile", u.profileName, "total_duration_ms", totalMs)
	return nil
}
