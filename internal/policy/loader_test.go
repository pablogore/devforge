package policy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pablogore/go-specs/specs"
)

func TestLoadPolicies(t *testing.T) {
	specs.Describe(t, "LoadPolicies", func(s *specs.Spec) {
		s.It("returns empty when dir does not exist", func(ctx *specs.Context) {
			dir := t.TempDir()
			root := filepath.Join(dir, "nonexistent")
			policies, err := LoadPolicies(root)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(policies) == 0).To(specs.BeTrue())
		})
		s.It("returns empty when policies dir is empty", func(ctx *specs.Context) {
			root := t.TempDir()
			_ = os.MkdirAll(filepath.Join(root, ".devforge", "policies"), 0o750)
			policies, err := LoadPolicies(root)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(policies) == 0).To(specs.BeTrue())
		})
		s.It("loads valid YAML policy", func(ctx *specs.Context) {
			root := t.TempDir()
			policyDir := filepath.Join(root, ".devforge", "policies")
			_ = os.MkdirAll(policyDir, 0o750)
			_ = os.WriteFile(filepath.Join(policyDir, "a.yaml"), []byte(`
name: test-policy
type: architectural
rules:
  forbid_import: "forbidden/pkg"
`), 0o600)
			policies, err := LoadPolicies(root)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(policies)).ToEqual(1)
			ctx.Expect(policies[0].File).ToEqual("a.yaml")
			ctx.Expect(policies[0].Name).ToEqual("test-policy")
			ctx.Expect(policies[0].Type).ToEqual("architectural")
			ctx.Expect(policies[0].Rules["forbid_import"]).ToEqual("forbidden/pkg")
		})
		s.It("skips non-YAML files", func(ctx *specs.Context) {
			root := t.TempDir()
			policyDir := filepath.Join(root, ".devforge", "policies")
			_ = os.MkdirAll(policyDir, 0o750)
			_ = os.WriteFile(filepath.Join(policyDir, "readme.txt"), []byte("text"), 0o600)
			policies, err := LoadPolicies(root)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(policies) == 0).To(specs.BeTrue())
		})
		s.It("returns policies in deterministic order", func(ctx *specs.Context) {
			root := t.TempDir()
			policyDir := filepath.Join(root, ".devforge", "policies")
			_ = os.MkdirAll(policyDir, 0o750)
			_ = os.WriteFile(filepath.Join(policyDir, "z.yaml"), []byte("name: z\n"), 0o600)
			_ = os.WriteFile(filepath.Join(policyDir, "a.yaml"), []byte("name: a\n"), 0o600)
			policies, err := LoadPolicies(root)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(policies)).ToEqual(2)
			ctx.Expect(policies[0].File).ToEqual("a.yaml")
			ctx.Expect(policies[1].File).ToEqual("z.yaml")
		})
		s.It("returns error for invalid YAML", func(ctx *specs.Context) {
			root := t.TempDir()
			policyDir := filepath.Join(root, ".devforge", "policies")
			_ = os.MkdirAll(policyDir, 0o750)
			_ = os.WriteFile(filepath.Join(policyDir, "bad.yaml"), []byte("invalid: [[["), 0o600)
			_, err := LoadPolicies(root)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "bad.yaml")).To(specs.BeTrue())
		})
		s.It("skips unreadable file and returns readable ones", func(ctx *specs.Context) {
			root := t.TempDir()
			policyDir := filepath.Join(root, ".devforge", "policies")
			_ = os.MkdirAll(policyDir, 0o750)
			_ = os.WriteFile(filepath.Join(policyDir, "good.yaml"), []byte("name: good\ntype: test\n"), 0o600)
			unreadable := filepath.Join(policyDir, "unreadable.yaml")
			_ = os.WriteFile(unreadable, []byte("name: x\n"), 0o600)
			_ = os.Chmod(unreadable, 0o000)
			defer func() { _ = os.Chmod(unreadable, 0o600) }()
			policies, err := LoadPolicies(root)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(policies)).ToEqual(1)
			ctx.Expect(policies[0].Name).ToEqual("good")
		})
	})
}
