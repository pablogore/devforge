package guard

import (
	"errors"
	"fmt"
	"strings"
)

// ForbidImport runs go list -json for the given glob and fails if any package
// imports a path containing forbiddenSegment. Used by policy pack evaluation.
func ForbidImport(ctx *Context, listGlob, forbiddenSegment string) error {
	if forbiddenSegment == "" {
		return nil
	}
	out, err := ctx.CommandRunner.RunCombinedOutput(ctx.StdCtx, ctx.Workdir, "go", "list", "-json", listGlob)
	if err != nil || strings.TrimSpace(out) == "" {
		return nil
	}
	errViolation := errors.New("policy: import of " + forbiddenSegment + " forbidden")
	return checkImportsContain(strings.TrimSpace(out), forbiddenSegment, errViolation)
}

// ForbidTimeNow runs git grep for time.Now() under the given path and fails
// if any match is found. Path is e.g. "internal/domain" or "domain" (interpreted as internal/domain).
// Used by policy pack evaluation.
func ForbidTimeNow(ctx *Context, path string) error {
	if path == "" {
		return nil
	}
	if path == "domain" || !strings.Contains(path, "/") {
		path = "internal/" + path
	}
	out, err := ctx.CommandRunner.RunCombinedOutput(ctx.StdCtx, ctx.Workdir, "git", "grep", "-n", "time.Now()", "--", path)
	trimmed := strings.TrimSpace(out)
	if trimmed != "" && !strings.Contains(trimmed, "fatal") {
		return fmt.Errorf("policy: time.Now() is forbidden in %s", path)
	}
	_ = err
	return nil
}
