package steps

import (
	"fmt"
	"strings"

	"github.com/pablogore/devforge/internal/application"
)

// CheckGoreleaserVersionStep verifies that goreleaser is available before release steps.
type CheckGoreleaserVersionStep struct{}

// NewCheckGoreleaserVersionStep returns a new CheckGoreleaserVersionStep.
func NewCheckGoreleaserVersionStep() *CheckGoreleaserVersionStep {
	return &CheckGoreleaserVersionStep{}
}

// Name returns the step name for logging and registry.
func (CheckGoreleaserVersionStep) Name() string { return "check-goreleaser-version" }

func init() {
	application.RegisterStep("check-goreleaser-version", func() application.Step { return NewCheckGoreleaserVersionStep() })
	application.RegisterStep("goreleaser-check", func() application.Step { return NewCheckGoreleaserVersionStep() })
}

// Run executes goreleaser --version and fails if the binary is missing or returns an error.
func (CheckGoreleaserVersionStep) Run(ctx *application.Context) error {
	ctx.Log.Info("Checking goreleaser availability")

	out, err := ctx.Cmd.RunCombinedOutput(ctx.StdCtx, ctx.Workdir, "goreleaser", "--version")
	if err != nil {
		return fmt.Errorf("goreleaser not available: %w", err)
	}

	version := strings.TrimSpace(out)
	if version == "" {
		return fmt.Errorf("goreleaser version output empty")
	}

	ctx.Log.Info("Goreleaser detected", "version", version)
	return nil
}
