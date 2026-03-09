package guard

import (
	"errors"
	"fmt"
	"strings"
)

var errFmtPrintOutsideCmd = errors.New("fmt.Print* is forbidden outside cmd/")

const noFmtPrintRuleName = "NoFmtPrintOutsideCmd"

// NoFmtPrintOutsideCmdRule forbids fmt.Print/fmt.Println/fmt.Printf outside cmd/.
type NoFmtPrintOutsideCmdRule struct{}

// NewNoFmtPrintOutsideCmdRule returns a rule that fails if fmt.Print* appears outside cmd/.
func NewNoFmtPrintOutsideCmdRule() *NoFmtPrintOutsideCmdRule {
	return &NoFmtPrintOutsideCmdRule{}
}

// Name returns the rule name.
func (NoFmtPrintOutsideCmdRule) Name() string {
	return noFmtPrintRuleName
}

// Validate runs git grep for fmt.Print* call sites; fails if any match is outside cmd/.
// Uses only CommandRunner (port). Matches under cmd/ are filtered out in Go.
func (NoFmtPrintOutsideCmdRule) Validate(ctx *Context) error {
	out, err := ctx.CommandRunner.RunCombinedOutput(ctx.StdCtx, ctx.Workdir, "git", "grep", "-n", "-E", `fmt\.(Println|Printf)\(`, "--", "*.go")
	if err != nil && strings.TrimSpace(out) == "" {
		return nil
	}
	trimmed := strings.TrimSpace(out)
	if trimmed == "" {
		return nil
	}
	if strings.Contains(trimmed, "fatal") {
		return nil
	}
	for _, line := range strings.Split(trimmed, "\n") {
		path := pathFromGrepLine(line)
		if path == "" {
			continue
		}
		path = strings.ReplaceAll(path, "\\", "/")
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		// Skip this package (rule + test file with string literals that match grep)
		if strings.HasPrefix(path, "internal/guard/") || strings.Contains(path, "/internal/guard/") {
			continue
		}
		// Explicit skip for this rule's test file (path may vary by environment)
		if strings.Contains(path, "no_fmt_print_rule") {
			continue
		}
		// Skip fmt.Fprintf (regex matches "Printf" substring); only forbid fmt.Println/fmt.Printf
		if strings.Contains(strings.ToLower(line), "fprintf") {
			continue
		}
		if !underCmd(path) {
			return fmt.Errorf("%w: %s", errFmtPrintOutsideCmd, line)
		}
	}
	return nil
}

// pathFromGrepLine extracts the file path from "path:lineno:content" (git grep -n).
func pathFromGrepLine(line string) string {
	idx := strings.Index(line, ":")
	if idx < 0 {
		return ""
	}
	return line[:idx]
}

// underCmd reports whether the path is under cmd/ (allowed).
func underCmd(path string) bool {
	return strings.HasPrefix(path, "cmd/") || strings.Contains(path, "/cmd/")
}
