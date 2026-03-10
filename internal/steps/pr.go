package steps

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/pablogore/devforge/internal/application"
	"github.com/pablogore/devforge/internal/domain"
	"github.com/pablogore/devforge/internal/guard"
)

const goSpecsModule = "github.com/pablogore/go-specs"

// GoModTidyStep runs go mod tidy and fails if go.mod or go.sum change.
type GoModTidyStep struct{}

// Name returns the step name.
func (GoModTidyStep) Name() string { return "go-mod-tidy" }

// Run executes go mod tidy and checks for diffs.
func (GoModTidyStep) Run(ctx *application.Context) error {
	ctx.Log.Info("Running go mod tidy", "step", "go-mod-tidy")
	_, err := ctx.Cmd.RunCombinedOutput(ctx.StdCtx, ctx.Workdir, "go", "mod", "tidy")
	if err != nil {
		return fmt.Errorf("go mod tidy failed: %w", err)
	}
	if err := ctx.Git.DiffExitCode(ctx.Workdir, "go.mod"); err != nil {
		return domain.ErrModNotTidy
	}
	if _, err := ctx.Cmd.Run(ctx.StdCtx, ctx.Workdir, "test", "-e", "go.sum"); err == nil {
		if err := ctx.Git.DiffExitCode(ctx.Workdir, "go.sum"); err != nil {
			return domain.ErrModNotTidy
		}
	}
	return nil
}

// GoFmtStep checks that code is formatted with gofmt -s -l.
type GoFmtStep struct{}

// Name returns the step name.
func (GoFmtStep) Name() string { return "gofmt" }

// Run runs gofmt and fails if any file would change.
func (GoFmtStep) Run(ctx *application.Context) error {
	ctx.Log.Info("Running gofmt check", "step", "gofmt")
	out, err := ctx.Cmd.RunCombinedOutput(ctx.StdCtx, ctx.Workdir, "gofmt", "-s", "-l", ".")
	if err != nil {
		return domain.ErrFormatting
	}
	if len(strings.TrimSpace(out)) > 0 {
		return domain.ErrFormatting
	}
	return nil
}

// GovulnCheckStep runs govulncheck -json ./... and fails on HIGH/CRITICAL.
type GovulnCheckStep struct{}

// Name returns the step name.
func (GovulnCheckStep) Name() string { return "govulncheck" }

// Run executes govulncheck and validates output.
func (GovulnCheckStep) Run(ctx *application.Context) error {
	ctx.Log.Info("Running govulncheck", "step", "govulncheck")
	out, err := ctx.Cmd.RunCombinedOutput(ctx.StdCtx, ctx.Workdir, "govulncheck", "-json", "./...")
	if err != nil {
		if len(out) > 0 {
			return fmt.Errorf("govulncheck failed: %w\n%s", err, out)
		}
		return fmt.Errorf("govulncheck failed (not installed or not in PATH?): %w", err)
	}
	if err := domain.ValidateGovulncheckOutput(out); err != nil {
		return fmt.Errorf("govulncheck: %w\n%s", err, out)
	}
	ctx.Log.Info("Govulncheck passed", "step", "govulncheck")
	return nil
}

const defaultCoverPkg = "./internal/domain,./internal/application"

// specsRunnerAvailableFunc is used by GoTestStep to decide specs vs go test; tests can override.
var specsRunnerAvailableFunc = func(ctx *application.Context) bool {
	if _, err := exec.LookPath("specs"); err != nil {
		return false
	}
	_, err := ctx.Cmd.RunCombinedOutput(ctx.StdCtx, ctx.Workdir, "go", "list", "-m", goSpecsModule)
	return err == nil
}

// GoTestStep runs tests via specs runner (if available) or go test, with race and coverage. Produces coverage.out to avoid duplicate test runs.
type GoTestStep struct{}

// Name returns the step name.
func (GoTestStep) Name() string { return "test" }

// Run runs tests once: if specs CLI exists and repo has go-specs, runs specs with race and coverage; otherwise runs go test. Output is coverage.out.
func (GoTestStep) Run(ctx *application.Context) error {
	coverPkg := ctx.CoverPkg
	if coverPkg == "" {
		coverPkg = defaultCoverPkg
	}
	if len(ctx.CoveragePackagesResolved) > 0 {
		packagesFromPolicy := "profile default"
		if ctx.Config != nil && ctx.Config.Policies != nil && ctx.Config.Policies.Coverage != nil {
			packagesFromPolicy = strings.Join(ctx.Config.Policies.Coverage.Packages, ", ")
		}
		ctx.Log.Info("Coverage policy applied", "threshold", ctx.CoverageThreshold, "packages", packagesFromPolicy)
		ctx.Log.Info("Expanded packages", "list", strings.Join(ctx.CoveragePackagesResolved, ", "))
	}

	if specsRunnerAvailableFunc(ctx) {
		ctx.Log.Info("[devforge] tests executed via specs runner")
		ctx.Log.Info("[devforge] skipping go test duplicate run")
		// specs run forwards args after -- to go test; -race and -coverprofile ensure race and coverage.out
		args := []string{"run", "--", "-race", "-coverprofile=coverage.out", "-coverpkg=" + coverPkg, "./..."}
		_, err := ctx.Cmd.RunCombinedOutput(ctx.StdCtx, ctx.Workdir, "specs", args...)
		if err != nil {
			return domain.ErrTestFailed
		}
		return nil
	}

	ctx.Log.Info("Running go test", "step", "test", "flags", "race,coverage")
	_, err := ctx.Cmd.RunCombinedOutput(ctx.StdCtx, ctx.Workdir, "go", "test", "-race", "-coverprofile=coverage.out", "-coverpkg="+coverPkg, "./...")
	if err != nil {
		return domain.ErrTestFailed
	}
	return nil
}

// GoTestRaceStep runs go test -race ./... (heavier runtime check for deep mode). Does not produce coverage.
type GoTestRaceStep struct{}

// Name returns the step name.
func (GoTestRaceStep) Name() string { return "test-race" }

// Run executes go test -race ./...
func (GoTestRaceStep) Run(ctx *application.Context) error {
	ctx.Log.Info("Running go test with race", "step", "test-race")
	_, err := ctx.Cmd.RunCombinedOutput(ctx.StdCtx, ctx.Workdir, "go", "test", "-race", "./...")
	if err != nil {
		return domain.ErrTestFailed
	}
	return nil
}

// CoverageStep parses coverage.out and enforces the configured coverage threshold.
type CoverageStep struct{}

// Name returns "test" so failure logging matches original behavior (test+coverage reported as step=test).
func (CoverageStep) Name() string { return "test" }

// Run parses coverage and fails if below threshold.
func (CoverageStep) Run(ctx *application.Context) error {
	coverOut, err := ctx.Cmd.RunCombinedOutput(ctx.StdCtx, ctx.Workdir, "go", "tool", "cover", "-func=coverage.out")
	if err != nil {
		return domain.ErrCoverageParse
	}
	result, err := domain.ParseCoverageFromFunc(coverOut)
	if err != nil {
		return err
	}
	if err := domain.ValidateCoverage(result.Percentage, ctx.CoverageThreshold); err != nil {
		return err
	}
	excluded := ctx.CoverageExcludedCount
	ctx.Log.Info("Coverage check passed",
		"coverage", result.Percentage,
		"threshold", ctx.CoverageThreshold,
		"excluded_packages", excluded)
	ctx.Log.Info("Coverage: "+formatPct(result.Percentage)+" | Threshold: "+formatPct(ctx.CoverageThreshold)+" | Excluded packages: "+formatInt(excluded))
	return nil
}

func formatPct(f float64) string {
	return fmt.Sprintf("%.1f%%", f)
}

func formatInt(n int) string {
	return fmt.Sprintf("%d", n)
}

// ConventionalCommitStep validates PR title or latest commit message against conventional commit format.
type ConventionalCommitStep struct{}

// Name returns the step name.
func (ConventionalCommitStep) Name() string { return "conventional-commit" }

// Run validates the title/commit message.
func (ConventionalCommitStep) Run(ctx *application.Context) error {
	prTitle := ctx.TitleOverride
	if prTitle == "" {
		prTitle = ctx.Env.Get("PR_TITLE")
	}
	if prTitle == "" {
		msg, err := ctx.Git.GetLatestCommitMessage(ctx.Workdir)
		if err == nil && msg != "" {
			prTitle = msg
		}
	}
	if prTitle == "" {
		return domain.ErrPRTitleRequired
	}
	return domain.ValidateConventionalCommit(prTitle)
}

func init() {
	application.RegisterStep("go-mod-tidy", func() application.Step { return GoModTidyStep{} })
	application.RegisterStep("gofmt", func() application.Step { return GoFmtStep{} })
	application.RegisterStep("govulncheck", func() application.Step { return GovulnCheckStep{} })
	application.RegisterStep("test", func() application.Step { return GoTestStep{} })
	application.RegisterStep("test-race", func() application.Step { return GoTestRaceStep{} })
	application.RegisterStep("conventional-commit", func() application.Step { return ConventionalCommitStep{} })
}

// GoLibPRSteps returns the default PR steps for go-lib profile. Same as Full; callers pass staticAnalysisTimeout (e.g. 2m).
func GoLibPRSteps(complexityThreshold int, staticAnalysisTimeout time.Duration, customRules ...guard.ArchitecturalRule) []application.Step {
	return GoLibPRStepsFull(complexityThreshold, staticAnalysisTimeout, customRules...)
}

// GoLibPRStepsQuick returns fast structural checks only; static-analysis group is wrapped with timeout.
func GoLibPRStepsQuick(_ int, staticAnalysisTimeout time.Duration, customRules ...guard.ArchitecturalRule) []application.Step {
	return []application.Step{
		GoModTidyStep{},
		ConventionalCommitStep{},
		NewArchitecturalGuardStep(append(guard.DefaultRules(), customRules...)),
		PolicyPackStep{},
		NewTimeoutGroupStep("static-analysis", staticAnalysisTimeout, GolangCILintStep{}),
	}
}

// GoLibPRStepsFull returns the full PR pipeline: static-analysis (golangci-lint) then security (govulncheck). Deterministic order.
func GoLibPRStepsFull(_ int, staticAnalysisTimeout time.Duration, customRules ...guard.ArchitecturalRule) []application.Step {
	return []application.Step{
		GoModTidyStep{},
		ConventionalCommitStep{},
		NewArchitecturalGuardStep(append(guard.DefaultRules(), customRules...)),
		PolicyPackStep{},
		NewTimeoutGroupStep("static-analysis", staticAnalysisTimeout, GolangCILintStep{}),
		GovulnCheckStep{},
		GoTestStep{},
		CoverageStep{},
	}
}

// GoLibPRStepsDeep returns full pipeline + test-race; static-analysis then security. Deterministic order.
func GoLibPRStepsDeep(_ int, staticAnalysisTimeout time.Duration, customRules ...guard.ArchitecturalRule) []application.Step {
	return []application.Step{
		GoModTidyStep{},
		ConventionalCommitStep{},
		NewArchitecturalGuardStep(append(guard.DefaultRules(), customRules...)),
		PolicyPackStep{},
		NewTimeoutGroupStep("static-analysis", staticAnalysisTimeout, GolangCILintStep{}),
		GovulnCheckStep{},
		GoTestStep{},
		CoverageStep{},
		GoTestRaceStep{},
	}
}

// GoServicePRSteps returns the default PR steps for go-service. Callers pass staticAnalysisTimeout (e.g. 3m).
func GoServicePRSteps(complexityThreshold int, staticAnalysisTimeout time.Duration, customRules ...guard.ArchitecturalRule) []application.Step {
	return GoServicePRStepsFull(complexityThreshold, staticAnalysisTimeout, customRules...)
}

// GoServicePRStepsQuick returns quick steps for go-service; static-analysis wrapped with timeout.
func GoServicePRStepsQuick(_ int, staticAnalysisTimeout time.Duration, customRules ...guard.ArchitecturalRule) []application.Step {
	return []application.Step{
		GoModTidyStep{},
		ConventionalCommitStep{},
		NewArchitecturalGuardStep(append(guard.DefaultRules(), customRules...)),
		PolicyPackStep{},
		NewTimeoutGroupStep("static-analysis", staticAnalysisTimeout, GolangCILintStep{}),
	}
}

// GoServicePRStepsFull returns full PR steps for go-service; static-analysis then security.
func GoServicePRStepsFull(_ int, staticAnalysisTimeout time.Duration, customRules ...guard.ArchitecturalRule) []application.Step {
	return []application.Step{
		GoModTidyStep{},
		ConventionalCommitStep{},
		NewArchitecturalGuardStep(append(guard.DefaultRules(), customRules...)),
		PolicyPackStep{},
		NewTimeoutGroupStep("static-analysis", staticAnalysisTimeout, GolangCILintStep{}),
		GovulnCheckStep{},
		GoTestStep{},
		CoverageStep{},
	}
}

// GoServicePRStepsDeep returns full + test-race + integration-tests for go-service; static-analysis then security.
func GoServicePRStepsDeep(_ int, staticAnalysisTimeout time.Duration, customRules ...guard.ArchitecturalRule) []application.Step {
	return []application.Step{
		GoModTidyStep{},
		ConventionalCommitStep{},
		NewArchitecturalGuardStep(append(guard.DefaultRules(), customRules...)),
		PolicyPackStep{},
		NewTimeoutGroupStep("static-analysis", staticAnalysisTimeout, GolangCILintStep{}),
		GovulnCheckStep{},
		GoTestStep{},
		CoverageStep{},
		GoTestRaceStep{},
		IntegrationTestsStep{},
	}
}
