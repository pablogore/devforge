package domain

import "errors"

// Sentinel errors used by domain and application layers.
var (
	ErrModNotTidy                = errors.New("go.mod or go.sum not tidy")
	ErrFormatting                = errors.New("gofmt check failed")
	ErrVetFailed                 = errors.New("go vet failed")
	ErrTestFailed                = errors.New("go test failed")
	ErrCoverageParse             = errors.New("could not parse coverage from output")
	ErrCoverageTooLow            = errors.New("coverage below required threshold")
	ErrPRTitleRequired           = errors.New("PR_TITLE environment variable is required")
	ErrInvalidConventionalCommit = errors.New("invalid commit message format")
	ErrInvalidVersionFormat      = errors.New("invalid version format")
	ErrInvalidLastTag            = errors.New("invalid last tag format")
	ErrNoReleaseableChanges      = errors.New("no releaseable changes found")
	ErrNotOnMainBranch           = errors.New("not on main branch")
	ErrWorkingTreeDirty          = errors.New("working tree is dirty")
	ErrShallowCloneDetected      = errors.New("shallow clone detected - full history required")
	ErrTagAlreadyExists          = errors.New("tag already exists")
	ErrTagDoesNotPointToHead     = errors.New("tag does not point to HEAD")
	ErrVersionNotDerivable       = errors.New("version cannot be derived")
	ErrReleaseFailed             = errors.New("release failed")
	ErrIdempotencyCheckFailed    = errors.New("idempotency check failed - version not deterministic")
	ErrGovulncheckHighOrCritical = errors.New("govulncheck found HIGH or CRITICAL vulnerabilities")
)
