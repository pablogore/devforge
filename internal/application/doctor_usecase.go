package application

import (
	"context"

	"github.com/pablogore/devforge/internal/ports"
)

// DoctorResult holds the result of running doctor checks.
type DoctorResult struct {
	// Checks are the individual check outcomes (e.g. git-installed, branch-main).
	Checks []CheckResult
}

// CheckResult holds a single doctor check outcome.
type CheckResult struct {
	// Name is the step or check name.
	Name string
	// Passed is true if the check succeeded.
	Passed bool
	// Message is a short description or error text.
	Message string
}

// DoctorUsecase runs the doctor pipeline and collects check results.
type DoctorUsecase struct {
	gitClient     ports.GitClient
	commandRunner ports.CommandRunner
	logger        ports.Logger
	clock         ports.Clock
	runner        *StepRunner
	pipeline      Pipeline
}

// NewDoctorUsecase returns a new DoctorUsecase.
func NewDoctorUsecase(git ports.GitClient, cmd ports.CommandRunner, logger ports.Logger, clock ports.Clock, pipeline Pipeline) *DoctorUsecase {
	return &DoctorUsecase{
		gitClient:     git,
		commandRunner: cmd,
		logger:        logger,
		clock:         clock,
		runner:        NewStepRunner(logger, clock),
		pipeline:      pipeline,
	}
}

// Run executes the doctor pipeline and returns aggregated check results.
func (u *DoctorUsecase) Run(workdir string) (*DoctorResult, error) {
	u.logger.Info("Running doctor checks", "workdir", workdir)

	checks := []CheckResult{}
	ctx := &Context{
		StdCtx:       context.Background(),
		Cmd:          u.commandRunner,
		Git:          u.gitClient,
		Log:          u.logger,
		Clock:        u.clock,
		Workdir:      workdir,
		DoctorChecks: &checks,
	}

	for _, step := range u.pipeline.Steps {
		if err := u.runner.Run(ctx, step); err != nil {
			u.logger.Warn("Doctor check failed", "check", step.Name())
			return &DoctorResult{Checks: checks}, err
		}
	}

	result := &DoctorResult{Checks: checks}
	allPassed := true
	for _, check := range result.Checks {
		if !check.Passed {
			allPassed = false
			u.logger.Warn("Doctor check failed", "check", check.Name)
		}
	}

	if allPassed {
		u.logger.Info("All doctor checks passed")
	} else {
		u.logger.Warn("Some doctor checks failed")
	}

	return result, nil
}
