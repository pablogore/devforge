package application

import (
	"context"

	"github.com/pablogore/devforge/internal/config"
	"github.com/pablogore/devforge/internal/ports"
)

// Step is a single validation or execution step in a PR or release flow.
type Step interface {
	Name() string
	Run(ctx *Context) error
}

// ExternalPluginConfig is the runtime config for one external plugin (forge-plugin-<name>).
// Set by the use case from .syntegrity.yml plugin_config; used by ExternalPluginStep to skip disabled plugins and pass params as DEVFORGE_PLUGIN_CONFIG.
type ExternalPluginConfig struct {
	// Enabled controls whether the plugin is run; false skips it.
	Enabled bool
	// Params are passed to the plugin as DEVFORGE_PLUGIN_CONFIG JSON.
	Params map[string]interface{}
}

// Context holds dependencies and configuration for step execution.
// Steps receive a pointer so they can set derived values (e.g. Version for release).
type Context struct {
	// StdCtx is propagated for cancellation (e.g. timeouts); use cases default it to context.Background().
	StdCtx context.Context
	// Cmd runs shell commands (e.g. go test, golangci-lint).
	Cmd ports.CommandRunner
	// Git runs git operations in Workdir.
	Git ports.GitClient
	// Env reads environment variables (explicit CI inputs only).
	Env ports.EnvProvider
	// Log is used for step observability.
	Log ports.Logger
	// Clock is used for step duration and timing.
	Clock ports.Clock
	// Workdir is the repository root for commands and git.
	Workdir string
	// CoverageThreshold is the minimum required coverage (e.g. 94.0).
	CoverageThreshold float64
	// CoverPkg is the -coverpkg flag value (comma-separated packages). Empty means use profile default.
	CoverPkg string
	// CoveragePackagesResolved is the list of resolved packages (for logging when policy is applied).
	CoveragePackagesResolved []string
	// TitleOverride is the PR title used for conventional-commit validation when non-empty.
	TitleOverride string
	// Version is set by release steps (e.g. version-derivation) for downstream steps.
	Version string
	// ProfileName is the CI profile (e.g. "go-lib", "go-service").
	ProfileName string
	// DoctorChecks is used by doctor steps to append CheckResult; set by DoctorUsecase before running.
	DoctorChecks *[]CheckResult
	// ExternalPluginConfig maps plugin name (e.g. "security") to config; nil means no config (run all discovered plugins normally).
	ExternalPluginConfig map[string]ExternalPluginConfig
	// Config is the loaded .syntegrity.yml; nil means use default pipeline (no step filtering).
	Config *config.Config
}
