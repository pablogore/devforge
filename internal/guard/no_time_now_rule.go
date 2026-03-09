package guard

import (
	"errors"
	"strings"
)

var errTimeNowInDomain = errors.New("time.Now() is forbidden in internal/domain")

const noTimeNowRuleName = "NoTimeNowInDomain"

// NoTimeNowInDomainRule forbids time.Now() in internal/domain (deterministic release).
type NoTimeNowInDomainRule struct{}

// NewNoTimeNowInDomainRule returns a rule that fails if time.Now() appears in internal/domain.
func NewNoTimeNowInDomainRule() *NoTimeNowInDomainRule {
	return &NoTimeNowInDomainRule{}
}

// Name returns the rule name.
func (NoTimeNowInDomainRule) Name() string {
	return noTimeNowRuleName
}

// Validate runs git grep for time.Now() under internal/domain; fails if any match.
// Uses only CommandRunner (port). If path does not exist or grep reports fatal, passes.
func (NoTimeNowInDomainRule) Validate(ctx *Context) error {
	out, err := ctx.CommandRunner.RunCombinedOutput(ctx.StdCtx, ctx.Workdir, "git", "grep", "-n", "time.Now()", "--", "internal/domain")
	trimmed := strings.TrimSpace(out)
	if trimmed != "" {
		if strings.Contains(trimmed, "fatal") {
			return nil
		}
		return errTimeNowInDomain
	}
	_ = err
	return nil
}
