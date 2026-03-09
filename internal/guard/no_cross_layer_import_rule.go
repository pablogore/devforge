package guard

import (
	"errors"
	"strings"
)

var (
	errCrossLayerDomain = errors.New("domain must not import application, adapters, or profiles")
	errCrossLayerApp    = errors.New("application must not import adapters")
	errCrossLayerPorts  = errors.New("ports must not import domain, application, or adapters")
)

const noCrossLayerRuleName = "NoCrossLayerImports"

// NoCrossLayerImportRule enforces allowed dependency direction; fails on first violation.
type NoCrossLayerImportRule struct{}

// NewNoCrossLayerImportRule returns a rule that checks domain, application, and ports imports.
func NewNoCrossLayerImportRule() *NoCrossLayerImportRule {
	return &NoCrossLayerImportRule{}
}

// Name returns the rule name.
func (NoCrossLayerImportRule) Name() string {
	return noCrossLayerRuleName
}

// Validate runs go list -json for domain, application, ports; fails if any forbidden import is found.
func (NoCrossLayerImportRule) Validate(ctx *Context) error {
	if e := validateDomainImports(ctx); e != nil {
		return e
	}
	if e := validateApplicationImports(ctx); e != nil {
		return e
	}
	if e := validatePortsImports(ctx); e != nil {
		return e
	}
	return nil
}

func validateDomainImports(ctx *Context) error {
	out, err := ctx.CommandRunner.RunCombinedOutput(ctx.StdCtx, ctx.Workdir, "go", "list", "-json", "./internal/domain/...")
	if err != nil || strings.TrimSpace(out) == "" {
		return nil
	}
	trimmed := strings.TrimSpace(out)
	if e := checkImportsContain(trimmed, "internal/application", errCrossLayerDomain); e != nil {
		return e
	}
	if e := checkImportsContain(trimmed, "internal/adapters", errCrossLayerDomain); e != nil {
		return e
	}
	return checkImportsContain(trimmed, "internal/profiles", errCrossLayerDomain)
}

func validateApplicationImports(ctx *Context) error {
	out, err := ctx.CommandRunner.RunCombinedOutput(ctx.StdCtx, ctx.Workdir, "go", "list", "-json", "./internal/application/...")
	if err != nil || strings.TrimSpace(out) == "" {
		return nil
	}
	return checkImportsContain(strings.TrimSpace(out), "internal/adapters", errCrossLayerApp)
}

func validatePortsImports(ctx *Context) error {
	out, err := ctx.CommandRunner.RunCombinedOutput(ctx.StdCtx, ctx.Workdir, "go", "list", "-json", "./internal/ports/...")
	if err != nil || strings.TrimSpace(out) == "" {
		return nil
	}
	trimmed := strings.TrimSpace(out)
	if e := checkImportsContain(trimmed, "internal/domain", errCrossLayerPorts); e != nil {
		return e
	}
	if e := checkImportsContain(trimmed, "internal/application", errCrossLayerPorts); e != nil {
		return e
	}
	return checkImportsContain(trimmed, "internal/adapters", errCrossLayerPorts)
}
