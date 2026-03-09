package detection

import (
	"os"
	"path/filepath"
)

// RepoType identifies the kind of repository for profile selection.
type RepoType string

// RepoType constants for profile detection.
const (
	RepoGoLib     RepoType = "go-lib"
	RepoGoService RepoType = "go-service"
)

// exists returns true if path exists (file or directory).
func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DetectProfile returns the profile name for the repository at workdir.
// If workdir has go.mod and a cmd/ directory, returns "go-service"; otherwise "go-lib".
// If there is no go.mod, returns "go-lib" as the default.
func DetectProfile(workdir string) string {
	if exists(filepath.Join(workdir, "go.mod")) {
		if exists(filepath.Join(workdir, "cmd")) {
			return string(RepoGoService)
		}
		return string(RepoGoLib)
	}
	return string(RepoGoLib)
}
