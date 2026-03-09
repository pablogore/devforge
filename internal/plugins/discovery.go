package plugins

import (
	"os"
	"path/filepath"
	"strings"
)

const binaryPrefix = "forge-plugin-"

// Discover returns plugin names for executables found in PATH with the prefix "forge-plugin-".
// Each name is the suffix after the prefix (e.g. "security" for forge-plugin-security).
// Order is deterministic by walking PATH order and deduplicating by name (first occurrence wins).
func Discover() []string {
	// Prevent plugin recursion.
	// When a plugin executes forge, it sets DEVFORGE_PLUGIN_EXECUTION.
	// If the variable is present, plugin discovery is skipped so plugins are not
	// executed recursively.
	if os.Getenv("DEVFORGE_PLUGIN_EXECUTION") != "" {
		return nil
	}
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return nil
	}
	paths := filepath.SplitList(pathEnv)
	seen := make(map[string]bool)
	var plugins []string
	for _, p := range paths {
		if p == "" {
			continue
		}
		entries, err := os.ReadDir(p)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if !strings.HasPrefix(name, binaryPrefix) {
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			mode := info.Mode()
			if !mode.IsRegular() {
				continue
			}
			if mode&0111 == 0 {
				continue
			}
			pluginName := strings.TrimPrefix(name, binaryPrefix)
			if pluginName != "" && !seen[pluginName] {
				seen[pluginName] = true
				plugins = append(plugins, pluginName)
			}
		}
	}
	return plugins
}
