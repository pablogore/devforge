package policy

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

const policiesDir = ".syntegrity/policies"

// LoadPolicies reads all .yaml/.yml files from root/.syntegrity/policies and returns
// the parsed policies in deterministic order (by filename). Skips directories and
// non-YAML files. Unreadable files are ignored with a warning. If the directory
// does not exist, returns an empty slice and nil error.
func LoadPolicies(root string) ([]Policy, error) {
	dir := filepath.Join(root, policiesDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading %s: %w", dir, err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		lower := strings.ToLower(name)
		if !strings.HasSuffix(lower, ".yaml") && !strings.HasSuffix(lower, ".yml") {
			continue
		}
		files = append(files, name)
	}
	sort.Strings(files)

	var policies []Policy
	for _, name := range files {
		path := filepath.Join(dir, name)
		//nolint:gosec // G304: path is dir (root+policiesDir) + name from ReadDir; scoped to repo policies
		data, err := os.ReadFile(path)
		if err != nil {
			_, _ = os.Stderr.Write([]byte("warning: skipping policy file " + name + ": " + err.Error() + "\n"))
			continue
		}
		var p Policy
		if err := yaml.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("invalid policy file %s: %w", name, err)
		}
		p.File = name
		policies = append(policies, p)
	}
	return policies, nil
}
