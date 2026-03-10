//nolint:revive // var-naming: package name describes coverage policy resolution; stdlib conflict accepted
package coverage

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	"github.com/pablogore/devforge/internal/ports"
)

// ErrWildcardWithOthers is returned when packages contains "*" and other entries.
var ErrWildcardWithOthers = errors.New("coverage packages: \"*\" must be the only entry when used; cannot mix with other patterns")

// Excluded path segments: packages under these are filtered out of go list ./...
// mocks: test-only helpers; not required for coverage threshold.
var excludedDirNames = map[string]bool{
	"vendor": true, "testdata": true, "examples": true, "generated": true, "mocks": true,
}

// ValidateCoveragePatterns returns an error if patterns contain "*" together with other entries.
func ValidateCoveragePatterns(patterns []string) error {
	if len(patterns) <= 1 {
		return nil
	}
	for _, p := range patterns {
		if p == "*" {
			return ErrWildcardWithOthers
		}
	}
	return nil
}

// ResolveCoveragePackages resolves patterns to a list of package import paths.
// patterns: ["*"] → all packages from go list ./... (excluding vendor, testdata, examples, generated).
// patterns containing "*" → glob match against that list.
// Otherwise → treat each as an explicit package path (included if present in module).
// Call ValidateCoveragePatterns before calling this.
func ResolveCoveragePackages(ctx context.Context, workdir string, patterns []string, cmd ports.CommandRunner) ([]string, error) {
	if err := ValidateCoveragePatterns(patterns); err != nil {
		return nil, err
	}
	all, err := listModulePackages(ctx, workdir, cmd)
	if err != nil {
		return nil, err
	}
	if len(patterns) == 1 && patterns[0] == "*" {
		return all, nil
	}
	var out []string
	for _, pkg := range all {
		for _, pat := range patterns {
			if matchPackage(pat, pkg) {
				out = append(out, pkg)
				break
			}
		}
	}
	return out, nil
}

// listModulePackages runs go list ./... and returns import paths, excluding vendor/testdata/examples/generated.
func listModulePackages(ctx context.Context, workdir string, cmd ports.CommandRunner) ([]string, error) {
	out, err := cmd.RunCombinedOutput(ctx, workdir, "go", "list", "./...")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	var pkgs []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if excludedPackage(line) {
			continue
		}
		pkgs = append(pkgs, line)
	}
	return pkgs, nil
}

func excludedPackage(importPath string) bool {
	parts := strings.Split(importPath, "/")
	for _, p := range parts {
		if excludedDirNames[p] {
			return true
		}
	}
	return false
}

// matchPackage returns true if pattern matches the package import path.
// pattern may be exact ("internal/domain", "./domain", or full import path) or glob ("internal/*", "./internal/*").
func matchPackage(pattern, pkg string) bool {
	// Normalize "./foo" to "foo" so patterns from .devforge.yml match go list import paths.
	if strings.HasPrefix(pattern, "./") {
		pattern = pattern[2:]
	}
	if pattern == "" {
		return false
	}
	if !strings.Contains(pattern, "*") {
		if pattern == pkg {
			return true
		}
		// Match suffix so "internal/domain" or "domain" matches "github.com/foo/repo/internal/domain" or ".../domain".
		return strings.HasSuffix(pkg, "/"+pattern) || strings.HasSuffix(pkg, pattern)
	}
	// Glob: try full path first, then suffix (e.g. "internal/*" vs ".../internal/domain").
	pat := filepath.ToSlash(pattern)
	if matched, _ := filepath.Match(pat, pkg); matched {
		return true
	}
	prefix := pattern[:strings.Index(pattern, "*")]
	if prefix != "" && strings.Contains(pkg, prefix) {
		if i := strings.Index(pkg, prefix); i >= 0 {
			suffix := pkg[i:]
			matched, _ := filepath.Match(pat, suffix)
			return matched
		}
	}
	return false
}

// BuildCoverPkgFlag returns the -coverpkg flag value (comma-separated package list).
// Uses package paths as returned by go list (import paths); go test accepts them for -coverpkg.
func BuildCoverPkgFlag(packages []string) string {
	return strings.Join(packages, ",")
}
