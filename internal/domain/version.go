package domain

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Version holds a semantic version (major, minor, patch).
type Version struct {
	// Major is the major version number.
	Major int
	// Minor is the minor version number.
	Minor int
	// Patch is the patch version number.
	Patch int
}

// String returns the version as vMajor.Minor.Patch.
func (v Version) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// ParseVersion parses a tag string (e.g. v1.0.0) into Version. Returns ErrInvalidVersionFormat on invalid input.
func ParseVersion(tag string) (Version, error) {
	tag = strings.TrimPrefix(tag, "v")
	parts := strings.Split(tag, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("%w: %s", ErrInvalidVersionFormat, tag)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return Version{}, fmt.Errorf("%w: %s", ErrInvalidVersionFormat, tag)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return Version{}, fmt.Errorf("%w: %s", ErrInvalidVersionFormat, tag)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return Version{}, fmt.Errorf("%w: %s", ErrInvalidVersionFormat, tag)
	}

	return Version{Major: major, Minor: minor, Patch: patch}, nil
}

// CommitType is the conventional commit type (feat, fix, etc.).
type CommitType string

// CommitType constants for conventional commits.
const (
	CommitTypeFeature  CommitType = "feat"
	CommitTypeFix      CommitType = "fix"
	CommitTypePerf     CommitType = "perf"
	CommitTypeRefactor CommitType = "refactor"
	CommitTypeDocs     CommitType = "docs"
	CommitTypeTest     CommitType = "test"
	CommitTypeChore    CommitType = "chore"
	CommitTypeBuild    CommitType = "build"
	CommitTypeCI       CommitType = "ci"
	CommitTypeRevert   CommitType = "revert"
)

// BumpType is the version bump kind (none, patch, minor, major).
type BumpType int

// BumpType constants for version increments.
const (
	BumpTypeNone  BumpType = 0
	BumpTypePatch BumpType = 1
	BumpTypeMinor BumpType = 2
	BumpTypeMajor BumpType = 3
)

// ParseCommitType extracts the conventional commit type from a message subject.
func ParseCommitType(msg string) (CommitType, bool) {
	pattern := `^(feat|fix|refactor|perf|docs|test|chore|build|ci|revert)(\(.+\))?!?:`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(msg)
	if len(matches) < 2 {
		return "", false
	}
	return CommitType(matches[1]), true
}

// HasBreakingChange reports whether the message indicates a breaking change.
func (ct CommitType) HasBreakingChange(msg string) bool {
	return strings.Contains(msg, "!") || strings.Contains(msg, "BREAKING CHANGE:")
}

// ToBumpType returns the version bump for this commit type and message.
func (ct CommitType) ToBumpType(msg string) BumpType {
	if ct.HasBreakingChange(msg) {
		return BumpTypeMajor
	}
	switch ct {
	case CommitTypeFeature:
		return BumpTypeMinor
	case CommitTypeFix, CommitTypePerf, CommitTypeRefactor:
		return BumpTypePatch
	default:
		return BumpTypeNone
	}
}

// DeriveNextVersion computes the next semantic version from commits since lastTag.
func DeriveNextVersion(commits []string, lastTag string) (Version, error) {
	var current Version
	var err error

	if lastTag == "" {
		current = Version{Major: 0, Minor: 1, Patch: 0}
	} else {
		current, err = ParseVersion(lastTag)
		if err != nil {
			return Version{}, fmt.Errorf("%w: %s", ErrInvalidLastTag, lastTag)
		}
	}

	maxBump := BumpTypeNone

	for _, commit := range commits {
		commitType, ok := ParseCommitType(commit)
		if !ok {
			continue
		}
		bump := commitType.ToBumpType(commit)
		if bump > maxBump {
			maxBump = bump
		}
	}

	if maxBump == BumpTypeNone {
		return Version{}, ErrNoReleaseableChanges
	}

	switch maxBump {
	case BumpTypeMajor:
		current.Major++
		current.Minor = 0
		current.Patch = 0
	case BumpTypeMinor:
		current.Minor++
		current.Patch = 0
	case BumpTypePatch:
		current.Patch++
	}

	return current, nil
}

// ValidateVersionFormat checks that version is a valid vMajor.Minor.Patch string.
func ValidateVersionFormat(version string) error {
	pattern := `^v\d+\.\d+\.\d+$`
	if !regexp.MustCompile(pattern).MatchString(version) {
		return fmt.Errorf("%w: %s", ErrInvalidVersionFormat, version)
	}
	return nil
}
