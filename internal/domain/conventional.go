package domain

import (
	"fmt"
	"regexp"
	"strings"
)

// ConventionalCommitPattern is the regex for valid conventional commit subject lines.
const ConventionalCommitPattern = `^(feat|fix|refactor|perf|docs|test|chore|build|ci|revert)(\(.+\))?!?: .+`

// ConventionalCommit holds a parsed commit title.
type ConventionalCommit struct {
	// Title is the commit subject line (e.g. "feat: add X").
	Title string
}

// ValidateConventionalCommit returns an error if title is empty or does not match ConventionalCommitPattern (merge commits are allowed).
func ValidateConventionalCommit(title string) error {
	if title == "" {
		return ErrPRTitleRequired
	}
	if strings.HasPrefix(title, "Merge ") {
		return nil
	}
	matched, err := regexp.MatchString(ConventionalCommitPattern, title)
	if err != nil {
		return fmt.Errorf("regex error: %s", err)
	}
	if !matched {
		return fmt.Errorf("%w: %s", ErrInvalidConventionalCommit, title)
	}
	return nil
}
