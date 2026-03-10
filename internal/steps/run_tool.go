package steps

import (
	"fmt"
	"strings"

	"github.com/pablogore/devforge/internal/application"
	"github.com/pablogore/devforge/internal/domain"
)

const devforgePrefix = "[devforge] "

// crashPhrases indicate tool output that signals an internal crash (panic or fatal error).
// Used to detect golangci-lint crashes so the pipeline can retry and then fail instead of ignoring.
var crashPhrases = []string{"panic:", "fatal error"}

// isToolCrash reports whether the tool output indicates an internal crash (panic/fatal).
// Only output is used so that lint failures (non-zero exit, no panic) are not retried.
// Crash triggers retry then pipeline failure; tool failure (e.g. lint issues) fails immediately.
func isToolCrash(output string, _ error) bool {
	lower := strings.ToLower(output)
	for _, phrase := range crashPhrases {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	return false
}

// crashReason returns a short reason for logging (non-zero exit vs panic/fatal in output).
func crashReason(output string, runErr error) string {
	if runErr != nil {
		return "non-zero exit"
	}
	lower := strings.ToLower(output)
	for _, phrase := range crashPhrases {
		if strings.Contains(lower, phrase) {
			return "panic/fatal error in output"
		}
	}
	return "unknown"
}

// crashReasonShort returns a human-readable suffix for "[devforge] <tool> crashed (suffix)".
func crashReasonShort(output string, runErr error) string {
	if runErr != nil {
		return "non-zero exit"
	}
	lower := strings.ToLower(output)
	for _, phrase := range crashPhrases {
		if strings.Contains(lower, phrase) {
			return "panic"
		}
	}
	return "unknown"
}

// runTool runs a command once via ctx.Cmd, logs [devforge] TOOL START/SUCCESS/FAILURE/CRASH, and returns output and error.
// It does not retry or ignore crashes; use runToolWithRetry for golangci-lint.
func runTool(ctx *application.Context, name string, args ...string) (output string, err error) {
	ctx.Log.Info(devforgePrefix+"TOOL START", "event", "tool_start", "tool", name)
	ctx.Log.Info(devforgePrefix+"running "+name, "event", "tool_start", "tool", name)
	start := ctx.Clock.Now()

	out, runErr := ctx.Cmd.RunCombinedOutput(ctx.StdCtx, ctx.Workdir, name, args...)
	durationMs := ctx.Clock.Since(start).Milliseconds()

	if runErr == nil && !isToolCrash(out, runErr) {
		ctx.Log.Info(devforgePrefix+"TOOL SUCCESS", "event", "tool_success", "tool", name, "duration_ms", durationMs)
		return out, nil
	}
	if runErr != nil && !isToolCrash(out, runErr) {
		ctx.Log.Error(devforgePrefix+"TOOL FAILURE", "event", "tool_failure", "tool", name, "duration_ms", durationMs, "error", runErr)
		return out, fmt.Errorf("%w: %v", domain.ErrToolFailure, runErr)
	}

	// Crash without exec error (e.g. panic recovered but exit 0)
	reason := crashReason(out, runErr)
	short := crashReasonShort(out, runErr)
	ctx.Log.Error(devforgePrefix+"TOOL CRASH", "event", "tool_crash", "tool", name, "reason", short, "duration_ms", durationMs)
	ctx.Log.Error(devforgePrefix+name+" crashed ("+short+")", "event", "tool_crash", "tool", name, "reason", reason, "duration_ms", durationMs)
	return out, fmt.Errorf("%w: %s: %s", domain.ErrToolCrash, name, reason)
}

// runToolWithRetry runs the tool up to maxAttempts times. If a crash is detected (non-zero exit,
// panic, or fatal error in output), it logs [devforge] TOOL CRASH, retrying (n/m), then fails the pipeline.
func runToolWithRetry(ctx *application.Context, name string, maxAttempts int, args ...string) (output string, err error) {
	var lastOut string
	var lastErr error
	var lastReason string

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		ctx.Log.Info(devforgePrefix+"TOOL START", "event", "tool_start", "tool", name)
		ctx.Log.Info(devforgePrefix+"running "+name, "event", "tool_start", "tool", name)
		start := ctx.Clock.Now()

		out, runErr := ctx.Cmd.RunCombinedOutput(ctx.StdCtx, ctx.Workdir, name, args...)
		durationMs := ctx.Clock.Since(start).Milliseconds()
		lastOut = out
		lastErr = runErr

		if runErr == nil && !isToolCrash(out, runErr) {
			ctx.Log.Info(devforgePrefix+"TOOL SUCCESS", "event", "tool_success", "tool", name, "duration_ms", durationMs)
			return out, nil
		}
		if runErr != nil && !isToolCrash(out, runErr) {
			ctx.Log.Error(devforgePrefix+"TOOL FAILURE", "event", "tool_failure", "tool", name, "duration_ms", durationMs, "error", runErr)
			return out, fmt.Errorf("%w: %v", domain.ErrToolFailure, runErr)
		}

		reason := crashReason(out, runErr)
		short := crashReasonShort(out, runErr)
		lastReason = reason
		ctx.Log.Error(devforgePrefix+"TOOL CRASH", "event", "tool_crash", "tool", name, "reason", short, "duration_ms", durationMs)
		ctx.Log.Error(devforgePrefix+name+" crashed ("+short+")", "event", "tool_crash", "tool", name, "reason", reason, "duration_ms", durationMs)

		if attempt < maxAttempts {
			retryNum := attempt
			retryTotal := maxAttempts - 1
			ctx.Log.Info(devforgePrefix+"retrying ("+fmt.Sprint(retryNum)+"/"+fmt.Sprint(retryTotal)+")", "event", "retry", "tool", name, "attempt", attempt, "max_retries", retryTotal)
		} else {
			ctx.Log.Error(devforgePrefix+name+" failed after retry — failing pipeline", "event", "tool_failure", "tool", name)
			if lastErr != nil {
				return lastOut, fmt.Errorf("%w: %v", domain.ErrToolCrash, lastErr)
			}
			return lastOut, fmt.Errorf("%w: %s: %s", domain.ErrToolCrash, name, lastReason)
		}
	}

	if lastErr != nil {
		return lastOut, fmt.Errorf("%w: %v", domain.ErrToolCrash, lastErr)
	}
	return lastOut, fmt.Errorf("%w: %s: %s", domain.ErrToolCrash, name, crashReason(lastOut, lastErr))
}

// runGoToolWithRetry runs a tool via "go run <moduleVersion> <toolArgs...>" with retry on crash.
// displayName is used for all [devforge] log lines. Pins the tool to moduleVersion (e.g. github.com/foo/v2@v2.1.0).
// Same retry and crash semantics as runToolWithRetry; crashes fail the pipeline after one retry.
func runGoToolWithRetry(ctx *application.Context, displayName string, maxAttempts int, moduleVersion string, toolArgs ...string) (output string, err error) {
	goArgs := make([]string, 0, 2+len(toolArgs))
	goArgs = append(goArgs, "run", moduleVersion)
	goArgs = append(goArgs, toolArgs...)
	var lastOut string
	var lastErr error
	var lastReason string

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		ctx.Log.Info(devforgePrefix+"TOOL START", "event", "tool_start", "tool", displayName)
		ctx.Log.Info(devforgePrefix+"running "+displayName, "event", "tool_start", "tool", displayName)
		start := ctx.Clock.Now()

		out, runErr := ctx.Cmd.RunCombinedOutput(ctx.StdCtx, ctx.Workdir, "go", goArgs...)
		durationMs := ctx.Clock.Since(start).Milliseconds()
		lastOut = out
		lastErr = runErr

		if runErr == nil && !isToolCrash(out, runErr) {
			ctx.Log.Info(devforgePrefix+"TOOL SUCCESS", "event", "tool_success", "tool", displayName, "duration_ms", durationMs)
			return out, nil
		}
		if runErr != nil && !isToolCrash(out, runErr) {
			ctx.Log.Error(devforgePrefix+"TOOL FAILURE", "event", "tool_failure", "tool", displayName, "duration_ms", durationMs, "error", runErr)
			return out, fmt.Errorf("%w: %v", domain.ErrToolFailure, runErr)
		}

		reason := crashReason(out, runErr)
		short := crashReasonShort(out, runErr)
		lastReason = reason
		ctx.Log.Error(devforgePrefix+"TOOL CRASH", "event", "tool_crash", "tool", displayName, "reason", short, "duration_ms", durationMs)
		ctx.Log.Error(devforgePrefix+displayName+" crashed ("+short+")", "event", "tool_crash", "tool", displayName, "reason", reason, "duration_ms", durationMs)

		if attempt < maxAttempts {
			retryNum := attempt
			retryTotal := maxAttempts - 1
			ctx.Log.Info(devforgePrefix+"retrying ("+fmt.Sprint(retryNum)+"/"+fmt.Sprint(retryTotal)+")", "event", "retry", "tool", displayName, "attempt", attempt, "max_retries", retryTotal)
		} else {
			ctx.Log.Error(devforgePrefix+displayName+" failed after retry — failing pipeline", "event", "tool_failure", "tool", displayName)
			if lastErr != nil {
				return lastOut, fmt.Errorf("%w: %v", domain.ErrToolCrash, lastErr)
			}
			return lastOut, fmt.Errorf("%w: %s: %s", domain.ErrToolCrash, displayName, lastReason)
		}
	}

	if lastErr != nil {
		return lastOut, fmt.Errorf("%w: %v", domain.ErrToolCrash, lastErr)
	}
	return lastOut, fmt.Errorf("%w: %s: %s", domain.ErrToolCrash, displayName, crashReason(lastOut, lastErr))
}
