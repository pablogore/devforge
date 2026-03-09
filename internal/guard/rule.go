package guard

import (
	"context"

	"github.com/pablogore/devforge/internal/ports"
)

// ArchitecturalRule is a single compile-time architectural constraint.
// Implementations validate against Context and return an error on violation.
type ArchitecturalRule interface {
	Name() string
	Validate(ctx *Context) error
}

// Context holds the dependencies and inputs required for rule validation.
// Only ports are used; no adapters.
type Context struct {
	// StdCtx is propagated for cancellation (e.g. timeouts).
	StdCtx context.Context
	// Workdir is the repository root for git and command execution.
	Workdir string
	// Profile is the CI profile name (e.g. "go-lib", "go-service").
	Profile string
	// GitClient runs git commands in Workdir.
	GitClient ports.GitClient
	// CommandRunner runs shell commands for rule checks.
	CommandRunner ports.CommandRunner
	// Logger is used for rule output.
	Logger ports.Logger
}
