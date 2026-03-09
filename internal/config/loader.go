package config

import (
	"bytes"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v3"
)

// BOM is the UTF-8 byte order mark; stripping it allows YAML to parse when the file was saved with BOM.
var bom = []byte{0xef, 0xbb, 0xbf}

const configFileName = ".syntegrity.yml"

// DefaultConfig returns a config with default values: Profile "", Mode "full", Plugins nil, PluginConfig nil.
func DefaultConfig() *Config {
	return &Config{
		Profile:      "",
		Mode:         "full",
		Plugins:      nil,
		PluginConfig: nil,
	}
}

// LoadConfig looks for .syntegrity.yml inside workdir. If the file does not exist,
// returns DefaultConfig(). If it exists, parses YAML and returns the config, applying
// defaults for missing fields (Mode "full" if empty, Plugins empty if nil).
// When workdir is ".", it is resolved to the process current working directory so
// the config is always read from the directory from which the CLI was invoked.
func LoadConfig(workdir string) (*Config, error) {
	if workdir == "." {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		workdir = cwd
	}
	path := filepath.Join(workdir, configFileName)
	//nolint:gosec // G304: path is workdir + constant configFileName; workdir is from application/CI, not user input
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}
	data = bytes.TrimPrefix(data, bom)

	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	applyDefaults(&c)
	return &c, nil
}

func applyDefaults(c *Config) {
	if c.Mode == "" {
		c.Mode = "full"
	}
	if c.Plugins == nil {
		c.Plugins = []PluginConfig{}
	}
	if c.PluginConfig == nil {
		c.PluginConfig = map[string]ExternalPluginCfg{}
	}
}
