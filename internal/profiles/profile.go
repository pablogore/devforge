package profiles

import (
	"github.com/pablogore/devforge/internal/application"
	"github.com/pablogore/devforge/internal/config"
)

// Profile defines the entry points for a CI/CD profile (pr, release, doctor).
type Profile struct {
	// Name is the profile identifier (e.g. "go-lib", "go-service").
	Name string

	// RunPRWithMode runs PR validation with the given mode and optional config.
	RunPRWithMode func(workdir, baseRef, title string, mode application.RunMode, cfg *config.Config) error
	// RunRelease runs the release pipeline and returns the new version.
	RunRelease func(workdir string) (string, error)
	// RunDoctor runs the doctor pipeline and returns check results.
	RunDoctor func(workdir string) (*application.DoctorResult, error)
}
