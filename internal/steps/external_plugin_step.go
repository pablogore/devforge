package steps

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pablogore/devforge/internal/application"
)

// ExternalPluginStep runs a standalone binary named forge-plugin-<name> discovered via PATH.
// It implements the Step interface for external plugins; the binary is executed with workdir set.
type ExternalPluginStep struct {
	// name is the plugin name (suffix after "forge-plugin-"), e.g. "security".
	name string
}

// Name returns the step name for logging and registry, e.g. "plugin-security".
func (s *ExternalPluginStep) Name() string {
	return "plugin-" + s.name
}

// Run executes the plugin binary (forge-plugin-<name>) in ctx.Workdir.
// If ctx.ExternalPluginConfig has an entry for this plugin with Enabled false, the step is skipped.
// Otherwise DEVFORGE_PLUGIN_EXECUTION and optionally DEVFORGE_PLUGIN_CONFIG (JSON params) are set in the environment.
func (s *ExternalPluginStep) Run(ctx *application.Context) error {
	if ctx.ExternalPluginConfig != nil {
		if cfg, ok := ctx.ExternalPluginConfig[s.name]; ok && !cfg.Enabled {
			ctx.Log.Info("plugin skipped (disabled in config)", "plugin", s.name)
			return nil
		}
	}

	binary := "forge-plugin-" + s.name
	ctx.Log.Info("running plugin", "plugin", s.name)
	env := append(os.Environ(), "DEVFORGE_PLUGIN_EXECUTION=1")

	if ctx.ExternalPluginConfig != nil {
		if cfg, ok := ctx.ExternalPluginConfig[s.name]; ok && len(cfg.Params) > 0 {
			cfgJSON, err := json.Marshal(cfg.Params)
			if err == nil {
				env = append(env, "DEVFORGE_PLUGIN_CONFIG="+string(cfgJSON))
			}
		}
	}

	out, err := ctx.Cmd.RunCombinedOutputWithEnv(ctx.StdCtx, ctx.Workdir, env, binary)
	if err != nil {
		return fmt.Errorf("plugin %s failed: %s", s.name, string(out))
	}
	return nil
}
