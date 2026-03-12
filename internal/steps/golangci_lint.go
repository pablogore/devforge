package steps

import (
	"fmt"

	"github.com/pablogore/devforge/internal/application"
)

// GolangciLintModuleVersion is the pinned golangci-lint v2 cmd package for go run (used by step and tests).
const GolangciLintModuleVersion = "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.0"

// GolangciLintRunArgs are the arguments passed to the linter (run --timeout=5m).
var GolangciLintRunArgs = []string{"run", "--timeout=5m"}

// GolangCILintStep runs golangci-lint v2 run --timeout=5m via "go run <module>@v2.1.0". Replaces standalone govet, staticcheck, gocyclo, gosec.
type GolangCILintStep struct{}

// Name returns the step name for logging.
func (GolangCILintStep) Name() string {
	return "golangci-lint"
}

// Run executes golangci-lint run --timeout=5m via go run with pinned v2.x.x, with retry on crash.
// If golangci-lint crashes (non-zero exit, panic, or fatal error in output), it retries once then fails the pipeline.
// Uses repo config (.golangci.yml etc.) if present; otherwise golangci-lint defaults.
func (GolangCILintStep) Run(ctx *application.Context) error {
	const maxAttempts = 2 // initial run + one retry
	out, err := runGoToolWithRetry(ctx, "golangci-lint", maxAttempts, GolangciLintModuleVersion, GolangciLintRunArgs...)
	if err != nil {
		return fmt.Errorf("golangci-lint: %w\n%s", err, out)
	}
	return nil
}

func init() {
	application.RegisterStep("golangci-lint", func() application.Step { return GolangCILintStep{} })
}
