package env

import (
	"os"
	"testing"

	"github.com/pablogore/go-specs/specs"
)

func TestEnvProvider(t *testing.T) {
	specs.Describe(t, "EnvProvider", func(s *specs.Spec) {
		s.It("NewEnvProvider returns non-nil", func(ctx *specs.Context) {
			p := NewEnvProvider()
			ctx.Expect(p != nil).To(specs.BeTrue())
		})
		s.It("Get returns set value", func(ctx *specs.Context) {
			key := "SYNTEGRITY_TEST_ENV_KEY_12345"
			val := "test-value"
			_ = os.Setenv(key, val)
			defer os.Unsetenv(key)

			p := NewEnvProvider().(*Provider)
			got := p.Get(key)
			ctx.Expect(got).ToEqual(val)
		})
		s.It("Get returns empty for missing key", func(ctx *specs.Context) {
			p := NewEnvProvider().(*Provider)
			got := p.Get("SYNTEGRITY_NONEXISTENT_VAR_98765")
			ctx.Expect(got).ToEqual("")
		})
	})
}
