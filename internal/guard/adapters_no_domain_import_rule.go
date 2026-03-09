package guard

import (
	"errors"
	"strings"
)

var errAdaptersImportDomain = errors.New("adapters must not import internal/domain")

const adaptersNoDomainRuleName = "AdaptersMustNotImportDomain"

// AdaptersMustNotImportDomainRule fails if any package under internal/adapters imports internal/domain.
type AdaptersMustNotImportDomainRule struct{}

// NewAdaptersMustNotImportDomainRule returns a rule that checks adapter imports via go list -json.
func NewAdaptersMustNotImportDomainRule() *AdaptersMustNotImportDomainRule {
	return &AdaptersMustNotImportDomainRule{}
}

// Name returns the rule name.
func (AdaptersMustNotImportDomainRule) Name() string {
	return adaptersNoDomainRuleName
}

// Validate runs go list -json ./internal/adapters/... and fails if any Imports contains /internal/domain.
func (AdaptersMustNotImportDomainRule) Validate(ctx *Context) error {
	out, err := ctx.CommandRunner.RunCombinedOutput(ctx.StdCtx, ctx.Workdir, "go", "list", "-json", "./internal/adapters/...")
	if err != nil && strings.TrimSpace(out) == "" {
		return nil
	}
	trimmed := strings.TrimSpace(out)
	if trimmed == "" {
		return nil
	}
	return checkImportsContain(trimmed, "internal/domain", errAdaptersImportDomain)
}
