package profiles

import (
	"context"

	"github.com/pablogore/devforge/internal/adapters/clock"
	"github.com/pablogore/devforge/internal/adapters/env"
	"github.com/pablogore/devforge/internal/adapters/exec"
	"github.com/pablogore/devforge/internal/adapters/git"
	"github.com/pablogore/devforge/internal/adapters/logger"
	"github.com/pablogore/devforge/internal/application"
)

// RunSteps builds a PR-like execution context (adapters, workdir) and runs the named steps in order.
// Used by the "run" CLI command to execute individual steps locally.
func RunSteps(workdir string, stepNames []string) error {
	cmd := exec.NewCommandRunner()
	gitClient := git.NewGitClient()
	envProvider := env.NewEnvProvider()
	log := logger.New("info", "text")
	clk := clock.NewRealClock()

	ctx := &application.Context{
		StdCtx:  context.Background(),
		Cmd:     cmd,
		Git:     gitClient,
		Env:     envProvider,
		Log:     log,
		Clock:   clk,
		Workdir: workdir,
	}
	runner := application.NewStepRunner(log, clk)
	return application.RunSteps(ctx, runner, stepNames)
}
