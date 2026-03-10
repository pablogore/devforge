package application

import (
	"errors"

	"github.com/pablogore/devforge/internal/domain"
)

// PipelineLogPrefix is the standard prefix for DevForge pipeline log lines (human- and machine-readable).
const PipelineLogPrefix = "[devforge] "

// FailureKind classifies a step failure for structured logging.
type FailureKind string

const (
	// FailureKindTest indicates go test or specs run failed.
	FailureKindTest FailureKind = "test_failure"
	// FailureKindPolicyViolation indicates governance/validation failed (conventional commit, coverage, guard, etc.).
	FailureKindPolicyViolation FailureKind = "policy_violation"
	// FailureKindToolError indicates a tool exited with non-zero (e.g. lint failures).
	FailureKindToolError FailureKind = "tool_error"
	// FailureKindToolCrash indicates a tool crashed (panic/fatal in output or unexpected failure).
	FailureKindToolCrash FailureKind = "tool_crash"
	// FailureKindUnknown is used when the error cannot be classified.
	FailureKindUnknown FailureKind = "unknown"
)

// ClassifyFailure returns the failure kind for the given step error for structured logging.
// Used to emit [devforge] TEST FAILURE, POLICY VIOLATION, TOOL FAILURE, TOOL CRASH, or STEP FAILURE with kind=.
func ClassifyFailure(err error) FailureKind {
	if err == nil {
		return FailureKindUnknown
	}
	if errors.Is(err, domain.ErrTestFailed) {
		return FailureKindTest
	}
	if errors.Is(err, domain.ErrModNotTidy) ||
		errors.Is(err, domain.ErrFormatting) ||
		errors.Is(err, domain.ErrPRTitleRequired) ||
		errors.Is(err, domain.ErrInvalidConventionalCommit) ||
		errors.Is(err, domain.ErrCoverageParse) ||
		errors.Is(err, domain.ErrCoverageTooLow) ||
		errors.Is(err, domain.ErrGovulncheckHighOrCritical) {
		return FailureKindPolicyViolation
	}
	if errors.Is(err, domain.ErrToolCrash) {
		return FailureKindToolCrash
	}
	if errors.Is(err, domain.ErrToolFailure) {
		return FailureKindToolError
	}
	return FailureKindUnknown
}
