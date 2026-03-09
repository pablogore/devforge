package steps

import (
	"github.com/pablogore/devforge/internal/application"
	"github.com/pablogore/devforge/internal/guard"
	"github.com/pablogore/devforge/internal/policy"
)

func init() {
	application.RegisterStep("policy-pack", func() application.Step { return PolicyPackStep{} })
}

// PolicyPackStep runs policy packs loaded from .devforge/policies/.
// If the directory does not exist or no policies are loaded, the step is a no-op.
type PolicyPackStep struct{}

// Name returns the step name for logging and registry.
func (PolicyPackStep) Name() string {
	return "policy-pack"
}

// Run loads policies from ctx.Workdir/.devforge/policies and evaluates them.
// Returns nil if no policies exist or all pass; returns error on first violation.
func (PolicyPackStep) Run(ctx *application.Context) error {
	policies, err := policy.LoadPolicies(ctx.Workdir)
	if err != nil {
		return err
	}
	if len(policies) == 0 {
		return nil
	}
	gCtx := &guard.Context{
		StdCtx:        ctx.StdCtx,
		Workdir:       ctx.Workdir,
		Profile:       ctx.ProfileName,
		CommandRunner: ctx.Cmd,
		Logger:        ctx.Log,
		GitClient:     ctx.Git,
	}
	return policy.Evaluate(gCtx, policies)
}
