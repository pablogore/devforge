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

type goLibDeps struct {
	prUsecase      *application.PRUsecase
	releaseUsecase *application.ReleaseUsecase
	doctorUsecase  *application.DoctorUsecase
}

func newGoLibDeps() *goLibDeps {
	cmd := exec.NewCommandRunner()
	gitClient := git.NewGitClient()
	envProvider := env.NewEnvProvider()
	log := logger.New("info", "text")
	clk := clock.NewRealClock()
	prPipeline, _ := application.GetPipeline("pr")
	releasePipeline, _ := application.GetPipeline("release")
	doctorPipeline, _ := application.GetPipeline("doctor")
	prUsecase := application.NewPRUsecase(cmd, gitClient, envProvider, 94.0, log, clk, "go-lib", prPipeline)
	releaseUsecase := application.NewReleaseUsecase(cmd, gitClient, envProvider, log, clk, releasePipeline)
	doctorUsecase := application.NewDoctorUsecase(gitClient, cmd, log, clk, doctorPipeline)
	return &goLibDeps{
		prUsecase:      prUsecase,
		releaseUsecase: releaseUsecase,
		doctorUsecase:  doctorUsecase,
	}
}

func goLibPRStepsForMode(mode application.RunMode) []application.Step {
	th := GoLibComplexityThreshold()
	timeout := GoLibStaticAnalysisTimeout()
	rules := GoLibCustomRules()
	switch mode {
	case application.ModeQuick:
		return steps.GoLibPRStepsQuick(th, timeout, rules...)
	case application.ModeDeep:
		return steps.GoLibPRStepsDeep(th, timeout, rules...)
	default:
		return steps.GoLibPRStepsFull(th, timeout, rules...)
	}
}

// GoLibComplexityThreshold returns the cyclomatic complexity threshold for go-lib (compile-time).
func GoLibComplexityThreshold() int {
	return 15
}

// GoLibStaticAnalysisTimeout returns the timeout for the static-analysis parallel group (compile-time default 2m).
func GoLibStaticAnalysisTimeout() time.Duration {
	return 2 * time.Minute
}

// GoLibCustomRules returns profile-specific architectural rules (compile-time extension). Default is none.
func GoLibCustomRules() []guard.ArchitecturalRule {
	return nil
}

// RunGoLibPR runs PR validation with default (full) mode.
func RunGoLibPR(workdir, baseRef string) error {
	return RunGoLibPRWithMode(workdir, baseRef, "", application.ModeFull, nil)
}

// RunGoLibPRWithTitle runs PR validation with full mode and the given title for conventional-commit validation.
func RunGoLibPRWithTitle(workdir, baseRef, title string) error {
	return RunGoLibPRWithMode(workdir, baseRef, title, application.ModeFull, nil)
}

// RunGoLibPRWithMode runs PR validation with the given mode (quick, full, deep). Steps are selected at compile time per mode.
func RunGoLibPRWithMode(workdir, baseRef, title string, mode application.RunMode, cfg *config.Config) error {
	cmd := exec.NewCommandRunner()
	gitClient := git.NewGitClient()
	envProvider := env.NewEnvProvider()
	log := logger.New("info", "text")
	clk := clock.NewRealClock()
	stepList := goLibPRStepsForMode(mode)
	prUsecase := application.NewPRUsecase(cmd, gitClient, envProvider, 94.0, log, clk, "go-lib", application.Pipeline{Name: "pr", Steps: stepList})
	pluginConfig := externalPluginConfigFrom(cfg)
	if title != "" {
		return prUsecase.RunWithTitleAndPluginConfig(workdir, baseRef, title, pluginConfig, cfg)
	}
	return prUsecase.RunWithTitleAndPluginConfig(workdir, baseRef, "", pluginConfig, cfg)
}

func externalPluginConfigFrom(cfg *config.Config) map[string]application.ExternalPluginConfig {
	if cfg == nil || len(cfg.PluginConfig) == 0 {
		return nil
	}
	out := make(map[string]application.ExternalPluginConfig, len(cfg.PluginConfig))
	for name, c := range cfg.PluginConfig {
		params := c.Params
		if params == nil {
			params = make(map[string]interface{})
		}
		out[name] = application.ExternalPluginConfig{Enabled: c.Enabled, Params: params}
	}
	return out
}

// RunGoLibRelease runs the full release pipeline for the go-lib profile and returns the new version.
func RunGoLibRelease(workdir string) (string, error) {
	deps := newGoLibDeps()
	result, err := deps.releaseUsecase.Run(workdir)
	if err != nil {
		return "", fmt.Errorf("go-lib release failed: %w", err)
	}
	return result.Version, nil
}

// ValidateGoLibVersion derives the next semantic version for the go-lib profile without creating a tag (dry-run).
func ValidateGoLibVersion(workdir string) (string, error) {
	deps := newGoLibDeps()
	return deps.releaseUsecase.ValidateVersionDerivation(workdir)
}

// RunGoLibDoctor runs the doctor pipeline and returns aggregated check results for the go-lib profile.
func RunGoLibDoctor(workdir string) (*application.DoctorResult, error) {
	deps := newGoLibDeps()
	return deps.doctorUsecase.Run(workdir)
}

func init() {
	application.RegisterPipeline(application.Pipeline{
		Name:  "pr",
		Steps: append(steps.GoLibPRStepsFull(GoLibComplexityThreshold(), GoLibStaticAnalysisTimeout(), GoLibCustomRules()...), steps.DiscoveredPluginSteps()...),
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
		Name:          "go-lib",
		RunPRWithMode: RunGoLibPRWithMode,
		RunRelease:    RunGoLibRelease,
		RunDoctor:     RunGoLibDoctor,
	})
}
