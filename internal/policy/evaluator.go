package policy

import (
	"fmt"

	"github.com/pablogore/devforge/internal/guard"
)

// ruleValues returns one or more string values from a rule value (string or []interface{}).
// Preserves backward compatibility: single string → []string{s}; slice → elements as strings.
func ruleValues(v interface{}) []string {
	if v == nil {
		return nil
	}
	if s, ok := v.(string); ok {
		return []string{s}
	}
	if slice, ok := v.([]interface{}); ok {
		var out []string
		for _, e := range slice {
			if s, ok := e.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

// severityIsWarning returns true if the policy severity is "warning"; otherwise treat as error.
func severityIsWarning(p Policy) bool {
	return p.Severity == "warning"
}

// Evaluate runs all rules in the given policies using the guard context.
// Returns on first violation with severity=error (default). severity=warning logs but does not fail.
// Known rules: forbid_import, forbid_time_now.
// Rule values may be a single string (backward compatible) or a list of strings.
func Evaluate(ctx *guard.Context, policies []Policy) error {
	for _, p := range policies {
		for ruleName, ruleValue := range p.Rules {
			values := ruleValues(ruleValue)
			if len(values) == 0 {
				continue
			}
			switch ruleName {
			case "forbid_import":
				for _, value := range values {
					if err := guard.ForbidImport(ctx, "./...", value); err != nil {
						msg := fmt.Sprintf("policy violation %s (%s): rule=%s value=%s: %v", p.Name, p.File, ruleName, value, err)
						if severityIsWarning(p) {
							ctx.Logger.Info("policy warning (non-fatal)", "message", msg)
							continue
						}
						return fmt.Errorf("%s: %w", msg, err)
					}
				}
			case "forbid_time_now":
				for _, value := range values {
					if err := guard.ForbidTimeNow(ctx, value); err != nil {
						msg := fmt.Sprintf("policy violation %s (%s): rule=%s value=%s: %v", p.Name, p.File, ruleName, value, err)
						if severityIsWarning(p) {
							ctx.Logger.Info("policy warning (non-fatal)", "message", msg)
							continue
						}
						return fmt.Errorf("%s: %w", msg, err)
					}
				}
			default:
				// unknown rule: skip (extensible later)
			}
		}
	}
	return nil
}
