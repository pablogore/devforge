package config

import yaml "gopkg.in/yaml.v3"

// PipelineConfig lets repositories enable or disable pipeline steps. If both Enable and Disable are set, Enable takes precedence (whitelist).
type PipelineConfig struct {
	// Disable lists step names to remove from the pipeline.
	Disable []string `yaml:"disable"`
	// Enable, if non-empty, is a whitelist: only these step names run.
	Enable []string `yaml:"enable"`
}

// CoveragePolicy configures coverage validation (optional; when absent, profile defaults apply).
type CoveragePolicy struct {
	// Threshold is the minimum coverage percentage (0–100).
	Threshold int `yaml:"threshold"`
	// Packages are package patterns: "*" (all), "internal/*", or explicit paths like "internal/domain".
	Packages []string `yaml:"packages"`
}

// Policies holds optional policy overrides from .syntegrity.yml.
type Policies struct {
	// Coverage configures coverage threshold and package patterns; nil means use profile defaults.
	Coverage *CoveragePolicy `yaml:"coverage"`
}

// Config holds repository-level DevForge configuration from .syntegrity.yml.
type Config struct {
	// Profile is the CI profile name (e.g. "go-lib", "go-service"); empty means default.
	Profile string `yaml:"profile"`
	// Mode is the PR pipeline mode: "quick", "full", or "deep"; default when empty is "full".
	Mode string `yaml:"mode"`
	// Pipeline enables or disables specific steps.
	Pipeline PipelineConfig `yaml:"pipeline"`
	// Plugins are run-command plugin entries (set when "plugins" is a list in YAML).
	Plugins []PluginConfig `yaml:"-"`
	// PluginConfig is per-plugin config for external binaries (set when "plugins" is a map in YAML).
	PluginConfig map[string]ExternalPluginCfg `yaml:"-"`
	// Policies holds optional coverage (and future) policy overrides.
	Policies *Policies `yaml:"policies"`
}

// pluginsWire holds the raw YAML node for "plugins" so it can be decoded as either a sequence or a mapping.
type pluginsWire struct {
	N *yaml.Node
}

// UnmarshalYAML stores the value node so the caller can decode it as list or map.
func (p *pluginsWire) UnmarshalYAML(value *yaml.Node) error {
	p.N = value
	return nil
}

// configWire is used to unmarshal .syntegrity.yml so "plugins" can be either a list or a map.
type configWire struct {
	Profile  string           `yaml:"profile"`
	Mode     string           `yaml:"mode"`
	Pipeline PipelineConfig   `yaml:"pipeline"`
	Plugins  *pluginsWire     `yaml:"plugins"`
	Policies *Policies        `yaml:"policies"`
}

// UnmarshalYAML lets "plugins" be either a list (run-command plugins) or a map (external plugin config).
func (c *Config) UnmarshalYAML(value *yaml.Node) error {
	var w configWire
	if err := value.Decode(&w); err != nil {
		return err
	}
	c.Profile = w.Profile
	c.Mode = w.Mode
	c.Pipeline = w.Pipeline
	c.Policies = w.Policies
	c.Plugins = nil
	c.PluginConfig = nil
	if w.Plugins == nil || w.Plugins.N == nil {
		return nil
	}
	switch w.Plugins.N.Kind {
	case yaml.SequenceNode:
		return w.Plugins.N.Decode(&c.Plugins)
	case yaml.MappingNode:
		return w.Plugins.N.Decode(&c.PluginConfig)
	default:
		return nil
	}
}

// PluginConfig describes a single run-command plugin entry (bash -c).
type PluginConfig struct {
	// Name is the plugin identifier (e.g. "lint-extra").
	Name string `yaml:"name"`
	// Run is the shell command (executed via bash -c).
	Run string `yaml:"run"`
}

// ExternalPluginCfg is the config for one external plugin (forge-plugin-<name>).
// Enabled defaults to true when not set. Other keys (e.g. severity) are in Params and passed as DEVFORGE_PLUGIN_CONFIG JSON.
type ExternalPluginCfg struct {
	// Enabled turns the plugin on or off; default true when unset.
	Enabled bool `yaml:"enabled"`
	// Params holds plugin-specific keys (e.g. severity); passed as JSON to the plugin.
	Params map[string]interface{} `yaml:"-"`
}

// UnmarshalYAML implements custom unmarshaling so we capture "enabled" and put all other keys into Params.
func (e *ExternalPluginCfg) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	e.Enabled = true
	e.Params = make(map[string]interface{})
	for k, v := range raw {
		if k == "enabled" {
			if b, ok := v.(bool); ok {
				e.Enabled = b
			}
			continue
		}
		e.Params[k] = v
	}
	return nil
}
