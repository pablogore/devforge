package guard

import (
	"errors"
	"strings"
)

var errCircularImport = errors.New("circular import detected")

const noCircularImportsRuleName = "NoCircularImports"

// NoCircularImportsRule fails if the Go build reports an import cycle.
type NoCircularImportsRule struct{}

// NewNoCircularImportsRule returns a rule that runs go list -deps ./... and fails on import cycle.
func NewNoCircularImportsRule() *NoCircularImportsRule {
	return &NoCircularImportsRule{}
}

// Name returns the rule name.
func (NoCircularImportsRule) Name() string {
	return noCircularImportsRuleName
}

// Validate runs go list -deps ./...; fails if output or error contains "import cycle not allowed".
func (NoCircularImportsRule) Validate(ctx *Context) error {
	out, err := ctx.CommandRunner.RunCombinedOutput(ctx.StdCtx, ctx.Workdir, "go", "list", "-deps", "./...")
	combined := out
	if err != nil {
		combined = out + err.Error()
	}
	if strings.Contains(combined, "import cycle") {
		return errCircularImport
	}
	return nil
}
