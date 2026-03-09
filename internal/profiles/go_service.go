package profiles

import (
	"fmt"
	"time"

	"github.com/pablogore/devforge/internal/adapters/clock"
	"github.com/pablogore/devforge/internal/adapters/env"
	"github.com/pablogore/devforge/internal/adapters/exec"
	"github.com/pablogore/devforge/internal/adapters/git"
	"github.com/pablogore/devforge/internal/adapters/logger"
	"github.com/pablogore/devforge/internal/application"
	"github.com/pablogore/devforge/internal/config"
	"github.com/pablogore/devforge/internal/guard"
	"github.com/pablogore/devforge/internal/steps"
)

type goServiceDeps struct {
	prUsecase      *application.PRUsecase
	releaseUsecase *application.ReleaseUsecase
	doctorUsecase  *application.DoctorUsecase
}

func newGoServiceDeps() *goServiceDeps {
	cmd := exec.NewCommandRunner()
	gitClient := git.NewGitClient()
	envProvider := env.NewEnvProvider()
	log := logger.New("info", "text")
	clk := clock.NewRealClock()
	prPipeline, _ := application.GetPipeline("go-service-pr")
	releasePipeline, _ := application.GetPipeline("release")
	doctorPipeline, _ := application.GetPipeline("doctor")
	prUsecase := application.NewPRUsecase(cmd, gitClient, envProvider, 80.0, log, clk, "go-service", prPipeline)
	releaseUsecase := application.NewReleaseUsecase(cmd, gitClient, envProvider, log, clk, releasePipeline)
	doctorUsecase := application.NewDoctorUsecase(gitClient, cmd, log, clk, doctorPipeline)
	return &goServiceDeps{
		prUsecase:      prUsecase,
		releaseUsecase: releaseUsecase,
		doctorUsecase:  doctorUsecase,
	}
}

func goServicePRStepsForMode(mode application.RunMode) []application.Step {
	th := GoServiceComplexityThreshold()
	timeout := GoServiceStaticAnalysisTimeout()
	rules := GoServiceCustomRules()
	switch mode {
	case application.ModeQuick:
		return steps.GoServicePRStepsQuick(th, timeout, rules...)
	case application.ModeDeep:
		return steps.GoServicePRStepsDeep(th, timeout, rules...)
	default:
		return steps.GoServicePRStepsFull(th, timeout, rules...)
	}
}

// GoServiceComplexityThreshold returns the cyclomatic complexity threshold for go-service (compile-time).
func GoServiceComplexityThreshold() int {
	return 20
}

// GoServiceStaticAnalysisTimeout returns the timeout for the static-analysis parallel group (compile-time default 3m).
func GoServiceStaticAnalysisTimeout() time.Duration {
	return 3 * time.Minute
}

// GoServiceCustomRules returns profile-specific architectural rules (compile-time extension). Default is none.
func GoServiceCustomRules() []guard.ArchitecturalRule {
	return nil
}

// RunGoServicePR runs PR validation with default (full) mode.
func RunGoServicePR(workdir, baseRef string) error {
	return RunGoServicePRWithMode(workdir, baseRef, "", application.ModeFull, nil)
}

// RunGoServicePRWithTitle runs PR validation with full mode and the given title for conventional-commit validation.
func RunGoServicePRWithTitle(workdir, baseRef, title string) error {
	return RunGoServicePRWithMode(workdir, baseRef, title, application.ModeFull, nil)
}

// RunGoServicePRWithMode runs PR validation with the given mode (quick, full, deep).
func RunGoServicePRWithMode(workdir, baseRef, title string, mode application.RunMode, cfg *config.Config) error {
	cmd := exec.NewCommandRunner()
	gitClient := git.NewGitClient()
	envProvider := env.NewEnvProvider()
	log := logger.New("info", "text")
	clk := clock.NewRealClock()
	stepList := goServicePRStepsForMode(mode)
	prUsecase := application.NewPRUsecase(cmd, gitClient, envProvider, 80.0, log, clk, "go-service", application.Pipeline{Name: "go-service-pr", Steps: stepList})
	pluginConfig := externalPluginConfigFrom(cfg)
	if title != "" {
		return prUsecase.RunWithTitleAndPluginConfig(workdir, baseRef, title, pluginConfig, cfg)
	}
	return prUsecase.RunWithTitleAndPluginConfig(workdir, baseRef, "", pluginConfig, cfg)
}

// RunGoServiceRelease runs the full release pipeline (preconditions, version derivation, tag, goreleaser) and returns the new version.
func RunGoServiceRelease(workdir string) (string, error) {
	deps := newGoServiceDeps()
	result, err := deps.releaseUsecase.Run(workdir)
	if err != nil {
		return "", fmt.Errorf("go-service release failed: %w", err)
	}
	return result.Version, nil
}

// RunGoServiceDoctor runs the doctor pipeline and returns aggregated check results for the go-service profile.
func RunGoServiceDoctor(workdir string) (*application.DoctorResult, error) {
	deps := newGoServiceDeps()
	return deps.doctorUsecase.Run(workdir)
}

func init() {
	application.RegisterPipeline(application.Pipeline{
		Name:  "go-service-pr",
		Steps: append(steps.GoServicePRStepsFull(GoServiceComplexityThreshold(), GoServiceStaticAnalysisTimeout(), GoServiceCustomRules()...), steps.DiscoveredPluginSteps()...),
	})
	application.RegisterPipeline(application.Pipeline{
		Name:  "release",
		Steps: steps.ReleaseSteps(),
	})
	application.RegisterPipeline(application.Pipeline{
		Name:  "doctor",
		Steps: steps.DoctorSteps(),
	})
	Register(Profile{
		Name:          "go-service",
		RunPRWithMode: RunGoServicePRWithMode,
		RunRelease:    RunGoServiceRelease,
		RunDoctor:     RunGoServiceDoctor,
	})
}
