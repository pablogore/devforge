package profiles

import (
	"testing"

	"github.com/pablogore/go-specs/specs"
)

func TestProfilesRegistry(t *testing.T) {
	specs.Describe(t, "profiles registry", func(s *specs.Spec) {
		s.It("List returns non-empty sorted names including go-lib and go-service", func(ctx *specs.Context) {
			names := List()
			ctx.Expect(len(names) > 0).To(specs.BeTrue())
			ctx.Expect(contains(names, "go-lib")).To(specs.BeTrue())
			ctx.Expect(contains(names, "go-service")).To(specs.BeTrue())
			for i := 1; i < len(names); i++ {
				ctx.Expect(names[i] >= names[i-1]).To(specs.BeTrue())
			}
		})
		s.It("Get(go-lib) returns profile with RunPRWithMode RunRelease RunDoctor", func(ctx *specs.Context) {
			p, ok := Get("go-lib")
			ctx.Expect(ok).To(specs.BeTrue())
			ctx.Expect(p.Name).ToEqual("go-lib")
			ctx.Expect(p.RunPRWithMode != nil).To(specs.BeTrue())
			ctx.Expect(p.RunRelease != nil).To(specs.BeTrue())
			ctx.Expect(p.RunDoctor != nil).To(specs.BeTrue())
		})
		s.It("Get(go-service) returns profile", func(ctx *specs.Context) {
			p, ok := Get("go-service")
			ctx.Expect(ok).To(specs.BeTrue())
			ctx.Expect(p.Name).ToEqual("go-service")
		})
		s.It("Get(nonexistent) returns false", func(ctx *specs.Context) {
			_, ok := Get("nonexistent-profile")
			ctx.Expect(ok).To(specs.BeFalse())
		})
	})
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
