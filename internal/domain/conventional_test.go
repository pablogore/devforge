package domain

import (
	"errors"
	"testing"

	"github.com/pablogore/go-specs/specs"
)

func TestValidateConventionalCommit(t *testing.T) {
	specs.Describe(t, "ValidateConventionalCommit", func(s *specs.Spec) {
		s.It("covers validation paths", func(ctx *specs.Context) {
			cases := []struct {
				name    string
				title   string
				wantErr error
			}{
				{"empty", "", ErrPRTitleRequired},
				{"valid feat", "feat: add foo", nil},
				{"valid feat with scope", "feat(api): add endpoint", nil},
				{"valid fix", "fix: correct bug", nil},
				{"valid breaking", "feat!: breaking change", nil},
				{"valid breaking in body", "feat(api): change\n\nBREAKING CHANGE: api", nil},
				{"invalid no type", "add foo", ErrInvalidConventionalCommit},
				{"invalid no colon", "feat add foo", ErrInvalidConventionalCommit},
				{"invalid type", "foo: bar", ErrInvalidConventionalCommit},
				{"refactor", "refactor: cleanup", nil},
				{"perf", "perf: faster", nil},
				{"docs", "docs: readme", nil},
				{"test", "test: unit", nil},
				{"chore", "chore: deps", nil},
				{"build", "build: script", nil},
				{"ci", "ci: workflow", nil},
				{"revert", "revert: feat(api): add", nil},
				{"merge commit", "Merge 38e2f934d890cc9d89843aa4d22e4bb06b779f17 into 0d63a501af1e8df1f50c313f0222bf0120746b9c", nil},
				{"merge with space only", "Merge ", nil},
				{"whitespace only no match", "   ", ErrInvalidConventionalCommit},
			}
			for _, tc := range cases {
				err := ValidateConventionalCommit(tc.title)
				if tc.wantErr != nil {
					ctx.Expect(err != nil).To(specs.BeTrue())
					ctx.Expect(errors.Is(err, tc.wantErr)).To(specs.BeTrue())
				} else {
					ctx.Expect(err).To(specs.BeNil())
				}
			}
		})
	})
}
