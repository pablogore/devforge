package tools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Exists reports whether a binary with the given name exists in PATH.
func Exists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// installGolangCILintRunner runs the real golangci-lint install; tests can replace it to avoid network.
var installGolangCILintRunner = func(versionDir string) error {
	script := "curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b " + versionDir + " " + GolangCILintVersion
	//nolint:gosec // G204: script is built from internal versionDir and GolangCILintVersion only
	cmd := exec.Command("sh", "-c", script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// installGovulncheckRunner runs the real govulncheck install; tests can replace it to avoid network.
var installGovulncheckRunner = func(versionDir string) error {
	//nolint:gosec // G204: GovulncheckVersion is internal constant; no user input
	cmd := exec.Command("go", "install", "golang.org/x/vuln/cmd/govulncheck@"+GovulncheckVersion)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "GOBIN="+versionDir)
	return cmd.Run()
}

// EnsureTools ensures the bin directory exists, PATH includes it, and both
// golangci-lint and govulncheck are installed into the cache and symlinked into bin.
// Tools are installed once and reused across executions.
func EnsureTools() error {
	if err := ensureBinDir(); err != nil {
		return err
	}
	ensurePath()

	if err := ensureGolangCILint(); err != nil {
		return fmt.Errorf("installing golangci-lint: %w", err)
	}
	if err := ensureGovulncheck(); err != nil {
		return fmt.Errorf("installing govulncheck: %w", err)
	}
	return nil
}

func ensureGolangCILint() error {
	cached := cachedBinaryPath("golangci-lint", GolangCILintVersion, "golangci-lint")
	if pathExists(cached) {
		return ensureSymlink(cached, binSymlinkPath("golangci-lint"))
	}

	_, _ = os.Stderr.Write([]byte("[devforge] Installing golangci-lint " + GolangCILintVersion + "..." + "\n"))
	versionDir := toolVersionDir("golangci-lint", GolangCILintVersion)
	if err := os.MkdirAll(versionDir, 0750); err != nil {
		return err
	}
	if err := installGolangCILintRunner(versionDir); err != nil {
		return err
	}
	return ensureSymlink(cached, binSymlinkPath("golangci-lint"))
}

func ensureGovulncheck() error {
	cached := cachedBinaryPath("govulncheck", GovulncheckVersion, "govulncheck")
	if pathExists(cached) {
		return ensureSymlink(cached, binSymlinkPath("govulncheck"))
	}

	_, _ = os.Stderr.Write([]byte("[devforge] Installing govulncheck " + GovulncheckVersion + "..." + "\n"))
	versionDir := toolVersionDir("govulncheck", GovulncheckVersion)
	if err := os.MkdirAll(versionDir, 0750); err != nil {
		return err
	}
	if err := installGovulncheckRunner(versionDir); err != nil {
		return err
	}
	return ensureSymlink(cached, binSymlinkPath("govulncheck"))
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ensureSymlink creates link pointing to target if link is missing or wrong. Target must exist.
func ensureSymlink(target, link string) error {
	info, err := os.Lstat(link)
	if err != nil {
		if os.IsNotExist(err) {
			return os.Symlink(target, link)
		}
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		dest, err := os.Readlink(link)
		if err == nil && (filepath.Clean(dest) == filepath.Clean(target) || dest == target) {
			return nil
		}
		_ = os.Remove(link)
	}
	return os.Symlink(target, link)
}
