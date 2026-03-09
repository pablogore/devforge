package doctor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pablogore/go-specs/specs"
)

func TestGenerateSuggestions(t *testing.T) {
	specs.Describe(t, "GenerateSuggestions", func(s *specs.Spec) {
		s.It("returns empty when no issues", func(ctx *specs.Context) {
			root := t.TempDir()
			got := GenerateSuggestions(root)
			ctx.Expect(len(got)).ToEqual(0)
		})
		s.It("returns architecture suggestion when domain imports adapters", func(ctx *specs.Context) {
			root := t.TempDir()
			domainDir := filepath.Join(root, "internal", "domain")
			ctx.Expect(os.MkdirAll(domainDir, 0o750)).To(specs.BeNil())
			ctx.Expect(os.WriteFile(filepath.Join(domainDir, "pkg.go"), []byte("package domain\nimport _ \"internal/adapters\"\n"), 0o600)).To(specs.BeNil())
			got := GenerateSuggestions(root)
			ctx.Expect(len(got)).ToEqual(1)
			ctx.Expect(got[0].File).ToEqual("architecture.yaml")
			ctx.Expect(sliceContains(got[0].Rules, "forbid_import: internal/adapters")).To(specs.BeTrue())
		})
		s.It("returns domain suggestion when time.Now used", func(ctx *specs.Context) {
			root := t.TempDir()
			domainDir := filepath.Join(root, "internal", "domain")
			ctx.Expect(os.MkdirAll(domainDir, 0o750)).To(specs.BeNil())
			ctx.Expect(os.WriteFile(filepath.Join(domainDir, "pkg.go"), []byte("package domain\nimport \"time\"\nfunc f() { _ = time.Now() }\n"), 0o600)).To(specs.BeNil())
			got := GenerateSuggestions(root)
			ctx.Expect(len(got)).ToEqual(1)
			ctx.Expect(got[0].File).ToEqual("domain.yaml")
			ctx.Expect(got[0].Rules).ToEqual([]string{"forbid_time_now: domain"})
		})
		s.It("returns security suggestion for dangerous imports", func(ctx *specs.Context) {
			root := t.TempDir()
			cmdDir := filepath.Join(root, "cmd", "app")
			ctx.Expect(os.MkdirAll(cmdDir, 0o750)).To(specs.BeNil())
			ctx.Expect(os.WriteFile(filepath.Join(cmdDir, "main.go"), []byte("package main\nimport _ \"net/http/pprof\"\nfunc main() {}\n"), 0o600)).To(specs.BeNil())
			got := GenerateSuggestions(root)
			ctx.Expect(len(got)).ToEqual(1)
			ctx.Expect(got[0].File).ToEqual("security.yaml")
			ctx.Expect(len(got[0].Rules) > 0 && strings.Contains(got[0].Rules[0], "forbid_import")).To(specs.BeTrue())
		})
	})
}

func sliceContains(slice []string, sub string) bool {
	for _, s := range slice {
		if strings.Contains(s, sub) || s == sub {
			return true
		}
	}
	return false
}
