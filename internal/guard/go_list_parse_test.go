package guard

import (
	"errors"
	"testing"

	"github.com/pablogore/go-specs/specs"
)

func Test_pathContainsSegment(t *testing.T) {
	specs.Describe(t, "pathContainsSegment", func(s *specs.Spec) {
		s.It("covers path and segment combinations", func(ctx *specs.Context) {
			cases := []struct {
				path    string
				segment string
				want    bool
			}{
				{"github.com/foo/internal/domain", "internal/domain", true},
				{"github.com/foo/internal/domain/bar", "internal/domain", true},
				{"internal/domain", "internal/domain", true},
				{"github.com/foo/internal/domainutils", "internal/domain", false},
				{"github.com/foo/internal/domainfoo", "internal/domain", false},
				{"github.com/foo/internal/adapters", "internal/adapters", true},
				{"github.com/foo/internal/adapters/exec", "internal/adapters", true},
				{"github.com/foo/internal/adaptershelper", "internal/adapters", false},
				{"fmt", "internal/domain", false},
				{"", "internal/domain", false},
				{"github.com/foo/bar", "", false},
			}
			for _, tc := range cases {
				got := pathContainsSegment(tc.path, tc.segment)
				ctx.Expect(got).ToEqual(tc.want)
			}
		})
	})
}

func Test_checkImportsContain(t *testing.T) {
	good := `{"ImportPath":"pkg","Imports":["fmt"]}`
	bad := `{"ImportPath":"pkg","Imports":["fmt","foo/internal/adapters"]}`
	customErr := errors.New("custom")

	specs.Describe(t, "checkImportsContain", func(s *specs.Spec) {
		s.It("returns nil when no match", func(ctx *specs.Context) {
			err := checkImportsContain(good, "internal/adapters", customErr)
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("returns custom error when match", func(ctx *specs.Context) {
			err := checkImportsContain(bad, "internal/adapters", customErr)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(err == customErr).To(specs.BeTrue())
		})
		s.It("returns error for multiple packages when one matches", func(ctx *specs.Context) {
			two := good + "\n" + bad
			err := checkImportsContain(two, "internal/adapters", customErr)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(err == customErr).To(specs.BeTrue())
		})
		s.It("does not match domainutils for segment internal/domain", func(ctx *specs.Context) {
			withDomainUtils := `{"ImportPath":"pkg","Imports":["github.com/foo/internal/domainutils"]}`
			err := checkImportsContain(withDomainUtils, "internal/domain", customErr)
			ctx.Expect(err).To(specs.BeNil())
		})
	})
}
