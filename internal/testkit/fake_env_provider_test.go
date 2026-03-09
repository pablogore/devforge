package testkit

import (
	"testing"

	"github.com/pablogore/go-specs/specs"
)

func TestFakeEnvProvider(t *testing.T) {
	specs.Describe(t, "FakeEnvProvider", func(s *specs.Spec) {
		s.It("NewFakeEnvProvider with nil uses empty map", func(ctx *specs.Context) {
			p := NewFakeEnvProvider(nil)
			ctx.Expect(p != nil).To(specs.BeTrue())
			ctx.Expect(p.Get("X")).ToEqual("")
		})
		s.It("Get returns value for key", func(ctx *specs.Context) {
			p := NewFakeEnvProvider(map[string]string{"HOME": "/home", "PATH": "/bin"})
			ctx.Expect(p.Get("HOME")).ToEqual("/home")
			ctx.Expect(p.Get("PATH")).ToEqual("/bin")
		})
		s.It("Get returns empty string for missing key", func(ctx *specs.Context) {
			p := NewFakeEnvProvider(map[string]string{"A": "1"})
			ctx.Expect(p.Get("B")).ToEqual("")
		})
	})
}
