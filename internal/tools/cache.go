package tools

import (
	"os"
	"path/filepath"
)

func toolsRoot() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".devforge", "tools")
}

func toolsBinDir() string {
	return filepath.Join(toolsRoot(), "bin")
}

func ensureBinDir() error {
	return os.MkdirAll(toolsBinDir(), 0750)
}

func ensurePath() {
	bin := toolsBinDir()
	path := os.Getenv("PATH")
	if !containsPath(path, bin) {
		sep := string(filepath.ListSeparator)
		if path == "" {
			_ = os.Setenv("PATH", bin)
		} else {
			_ = os.Setenv("PATH", bin+sep+path)
		}
	}
}

func containsPath(path, dir string) bool {
	dirClean := filepath.Clean(dir)
	for _, entry := range filepath.SplitList(path) {
		if filepath.Clean(entry) == dirClean {
			return true
		}
	}
	return false
}

// toolVersionDir returns the directory for a tool version, e.g. ~/.devforge/tools/golangci-lint/v1.64.8
func toolVersionDir(toolName, version string) string {
	return filepath.Join(toolsRoot(), toolName, version)
}

// cachedBinaryPath returns the path to the binary inside the versioned cache directory.
func cachedBinaryPath(toolName, version, binaryName string) string {
	return filepath.Join(toolVersionDir(toolName, version), binaryName)
}

// binSymlinkPath returns the path of the symlink in the bin directory.
func binSymlinkPath(binaryName string) string {
	return filepath.Join(toolsBinDir(), binaryName)
}
