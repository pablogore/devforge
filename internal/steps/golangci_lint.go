package steps

import (
	"fmt"

	"github.com/pablogore/devforge/internal/application"
)

// GolangCILintStep runs golangci-lint run --timeout=5m. Replaces standalone govet, staticcheck, gocyclo, gosec.
type GolangCILintStep struct{}

// Name returns the step name for logging.
func (GolangCILintStep) Name() string {
	return "golangci-lint"
}

// Run executes golangci-lint run --timeout=5m via ctx.Cmd; logs tool name, duration, and result.
// On failure, the returned error includes the linter output so CI logs show the actual lint errors.
// Uses repo config (.golangci.yml etc.) if present; otherwise golangci-lint defaults.
func (GolangCILintStep) Run(ctx *application.Context) error {
	out, err := runTool(ctx, "golangci-lint", "run", "--timeout=5m")
	if err != nil {
		return fmt.Errorf("golangci-lint: %w\n%s", err, out)
	}
	return nil
}

func init() {
	application.RegisterStep("golangci-lint", func() application.Step { return GolangCILintStep{} })
}
