package steps

import (
	"github.com/pablogore/devforge/internal/application"
	"github.com/pablogore/devforge/internal/guard"
)

// ArchitecturalGuardStep runs a deterministic list of architectural rules.
// Rules are executed in slice order; execution fails on first error.
type ArchitecturalGuardStep struct {
	rules []guard.ArchitecturalRule
}

// NewArchitecturalGuardStep returns a step that runs the given rules in order.
func NewArchitecturalGuardStep(rules []guard.ArchitecturalRule) *ArchitecturalGuardStep {
	return &ArchitecturalGuardStep{rules: rules}
}

func init() {
	application.RegisterStep("architectural-guard", func() application.Step { return NewArchitecturalGuardStep(guard.DefaultRules()) })
}

// Name returns the step name for logging.
func (s *ArchitecturalGuardStep) Name() string {
	return "architectural-guard"
}

// Run executes each rule in deterministic order. It logs the rule name before
// execution and fails immediately on first validation error.
func (s *ArchitecturalGuardStep) Run(ctx *application.Context) error {
	gCtx := &guard.Context{
		StdCtx:        ctx.StdCtx,
		Workdir:       ctx.Workdir,
		Profile:       ctx.ProfileName,
		GitClient:     ctx.Git,
		CommandRunner: ctx.Cmd,
		Logger:        ctx.Log,
	}
	for _, rule := range s.rules {
		ctx.Log.Info("Running architectural rule", "rule", rule.Name())
		if err := rule.Validate(gCtx); err != nil {
			return err
		}
	}
	return nil
}
