package guard

import (
	"errors"
	"strings"
)

var errDomainImportsAdapters = errors.New("domain must not import internal/adapters")

const domainNoAdapterRuleName = "DomainMustNotImportAdapters"

// DomainMustNotImportAdaptersRule fails if any package under internal/domain imports internal/adapters.
type DomainMustNotImportAdaptersRule struct{}

// NewDomainMustNotImportAdaptersRule returns a rule that checks domain imports via go list -json.
func NewDomainMustNotImportAdaptersRule() *DomainMustNotImportAdaptersRule {
	return &DomainMustNotImportAdaptersRule{}
}

// Name returns the rule name.
func (DomainMustNotImportAdaptersRule) Name() string {
	return domainNoAdapterRuleName
}

// Validate runs go list -json ./internal/domain/... and fails if any Imports contains /internal/adapters/.
// If domain does not exist (command fails), passes.
func (DomainMustNotImportAdaptersRule) Validate(ctx *Context) error {
	out, err := ctx.CommandRunner.RunCombinedOutput(ctx.StdCtx, ctx.Workdir, "go", "list", "-json", "./internal/domain/...")
	if err != nil && strings.TrimSpace(out) == "" {
		return nil
	}
	trimmed := strings.TrimSpace(out)
	if trimmed == "" {
		return nil
	}
	return checkImportsContain(trimmed, "internal/adapters", errDomainImportsAdapters)
}
