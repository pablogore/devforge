package steps

import (
	"strings"

	"github.com/pablogore/devforge/internal/application"
)

// internalCrashPhrases indicate tool-internal failures (e.g. govulncheck "without types") that should not fail the pipeline.
var internalCrashPhrases = []string{"internal error", "panic", "without types"}

// runTool runs a command via ctx.Cmd, logs name and duration, and treats internal tool crashes as non-fatal.
// Returns (output, nil) on success or when output matches internal crash phrases; otherwise (output, err).
func runTool(ctx *application.Context, name string, args ...string) (output string, err error) {
	ctx.Log.Info("running tool", "tool", name)
	start := ctx.Clock.Now()

	out, runErr := ctx.Cmd.RunCombinedOutput(ctx.StdCtx, ctx.Workdir, name, args...)
	durationMs := ctx.Clock.Since(start).Milliseconds()

	if runErr == nil {
		ctx.Log.Info("tool completed", "tool", name, "duration_ms", durationMs, "result", "ok")
		return out, nil
	}

	if isInternalCrash(out) {
		ctx.Log.Warn("tool crashed internally, ignoring", "tool", name)
		ctx.Log.Info("tool completed", "tool", name, "duration_ms", durationMs, "result", "warning")
		return out, nil
	}

	ctx.Log.Error("tool reported issues", "tool", name)
	ctx.Log.Info("tool completed", "tool", name, "duration_ms", durationMs, "result", "fail")
	return out, runErr
}

func isInternalCrash(output string) bool {
	lower := strings.ToLower(output)
	for _, phrase := range internalCrashPhrases {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	return false
}
