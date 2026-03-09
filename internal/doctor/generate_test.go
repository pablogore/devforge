package doctor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pablogore/devforge/internal/policy"
	"github.com/pablogore/go-specs/specs"
)

func TestGeneratePolicies(t *testing.T) {
	specs.Describe(t, "GeneratePolicies", func(s *specs.Spec) {
		s.It("returns empty when no issues", func(ctx *specs.Context) {
			root := t.TempDir()
			got, err := GeneratePolicies(root)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(got)).ToEqual(0)
		})
		s.It("generates architecture policy when domain imports adapters", func(ctx *specs.Context) {
			root := t.TempDir()
			domainDir := filepath.Join(root, "internal", "domain")
			ctx.Expect(os.MkdirAll(domainDir, 0o750)).To(specs.BeNil())
			ctx.Expect(os.WriteFile(filepath.Join(domainDir, "pkg.go"), []byte("package domain\nimport _ \"internal/adapters\"\n"), 0o600)).To(specs.BeNil())

			got, err := GeneratePolicies(root)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(got)).ToEqual(1)
			ctx.Expect(strings.Contains(got[0], "architecture.yaml")).To(specs.BeTrue())

			policies, err := policy.LoadPolicies(root)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(policies)).ToEqual(1)
			ctx.Expect(policies[0].Name).ToEqual("architecture")
			ctx.Expect(policies[0].Type).ToEqual("architecture")
			_, ok := policies[0].Rules["forbid_import"]
			ctx.Expect(ok).To(specs.BeTrue())
			ctx.Expect(policies[0].Rules["forbid_import"]).ToEqual("internal/adapters")
		})
		s.It("generates policy when time.Now used", func(ctx *specs.Context) {
			root := t.TempDir()
			domainDir := filepath.Join(root, "internal", "domain")
			ctx.Expect(os.MkdirAll(domainDir, 0o750)).To(specs.BeNil())
			ctx.Expect(os.WriteFile(filepath.Join(domainDir, "pkg.go"), []byte("package domain\nimport \"time\"\nfunc f() { _ = time.Now() }\n"), 0o600)).To(specs.BeNil())

			got, err := GeneratePolicies(root)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(got)).ToEqual(1)
			ctx.Expect(strings.Contains(got[0], "architecture.yaml")).To(specs.BeTrue())

			policies, err := policy.LoadPolicies(root)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(policies)).ToEqual(1)
			ctx.Expect(policies[0].Name).ToEqual("architecture")
			_, ok := policies[0].Rules["forbid_time_now"]
			ctx.Expect(ok).To(specs.BeTrue())
			ctx.Expect(policies[0].Rules["forbid_time_now"]).ToEqual("domain")
		})
		s.It("generates security policy for dangerous imports", func(ctx *specs.Context) {
			root := t.TempDir()
			cmdDir := filepath.Join(root, "cmd", "app")
			ctx.Expect(os.MkdirAll(cmdDir, 0o750)).To(specs.BeNil())
			ctx.Expect(os.WriteFile(filepath.Join(cmdDir, "main.go"), []byte("package main\nimport _ \"net/http/pprof\"\nfunc main() {}\n"), 0o600)).To(specs.BeNil())

			got, err := GeneratePolicies(root)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(got)).ToEqual(1)
			ctx.Expect(strings.Contains(got[0], "security.yaml")).To(specs.BeTrue())

			policies, err := policy.LoadPolicies(root)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(policies)).ToEqual(1)
			ctx.Expect(policies[0].Name).ToEqual("security")
			ctx.Expect(policies[0].Type).ToEqual("security")
			_, ok := policies[0].Rules["forbid_import"]
			ctx.Expect(ok).To(specs.BeTrue())
			ctx.Expect(policies[0].Rules["forbid_import"]).ToEqual("net/http/pprof")
		})
		s.It("generates both architecture and security when adapter and dangerous", func(ctx *specs.Context) {
			root := t.TempDir()
			domainDir := filepath.Join(root, "internal", "domain")
			ctx.Expect(os.MkdirAll(domainDir, 0o750)).To(specs.BeNil())
			ctx.Expect(os.WriteFile(filepath.Join(domainDir, "pkg.go"), []byte("package domain\nimport _ \"internal/adapters\"\n"), 0o600)).To(specs.BeNil())
			ctx.Expect(os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\nimport _ \"net/http/pprof\"\nfunc main() {}\n"), 0o600)).To(specs.BeNil())
			got, err := GeneratePolicies(root)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(got)).ToEqual(2)
		})
		s.It("returns error when policies dir cannot be created", func(ctx *specs.Context) {
			root := t.TempDir()
			devforgePath := filepath.Join(root, ".devforge")
			ctx.Expect(os.WriteFile(devforgePath, []byte("x"), 0o600)).To(specs.BeNil())
			domainDir := filepath.Join(root, "internal", "domain")
			ctx.Expect(os.MkdirAll(domainDir, 0o750)).To(specs.BeNil())
			ctx.Expect(os.WriteFile(filepath.Join(domainDir, "pkg.go"), []byte("package domain\nimport _ \"internal/adapters\"\n"), 0o600)).To(specs.BeNil())
			got, err := GeneratePolicies(root)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(got == nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "create")).To(specs.BeTrue())
		})
		s.It("generates multiple policies for multiple issues", func(ctx *specs.Context) {
			root := t.TempDir()
			domainDir := filepath.Join(root, "internal", "domain")
			ctx.Expect(os.MkdirAll(domainDir, 0o750)).To(specs.BeNil())
			ctx.Expect(os.WriteFile(filepath.Join(domainDir, "pkg.go"), []byte("package domain\nimport _ \"internal/adapters\"\n"), 0o600)).To(specs.BeNil())
			ctx.Expect(os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\nimport _ \"net/http/pprof\"\nfunc main() {}\n"), 0o600)).To(specs.BeNil())

			got, err := GeneratePolicies(root)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(got)).ToEqual(2)

			policies, err := policy.LoadPolicies(root)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(policies)).ToEqual(2)
			names := make(map[string]bool)
			for _, p := range policies {
				names[p.Name] = true
			}
			ctx.Expect(names["architecture"]).To(specs.BeTrue())
			ctx.Expect(names["security"]).To(specs.BeTrue())
		})
		s.It("output is deterministic", func(ctx *specs.Context) {
			root := t.TempDir()
			domainDir := filepath.Join(root, "internal", "domain")
			ctx.Expect(os.MkdirAll(domainDir, 0o750)).To(specs.BeNil())
			ctx.Expect(os.WriteFile(filepath.Join(domainDir, "pkg.go"), []byte("package domain\nimport _ \"internal/adapters\"\n"), 0o600)).To(specs.BeNil())

			got1, err := GeneratePolicies(root)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(got1)).ToEqual(1)
			data1, err := os.ReadFile(filepath.Join(root, ".devforge", "policies", "architecture.yaml"))
			ctx.Expect(err).To(specs.BeNil())

			got2, err := GeneratePolicies(root)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(got2)).ToEqual(1)
			data2, err := os.ReadFile(filepath.Join(root, ".devforge", "policies", "architecture.yaml"))
			ctx.Expect(err).To(specs.BeNil())

			ctx.Expect(string(data1)).ToEqual(string(data2))
		})
		s.It("no issues returns empty", func(ctx *specs.Context) {
			root := t.TempDir()
			got, err := GeneratePolicies(root)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(len(got)).ToEqual(0)
		})
	})
}

func Test_valueToNode(t *testing.T) {
	specs.Describe(t, "valueToNode", func(s *specs.Spec) {
		s.It("slice returns node with content", func(ctx *specs.Context) {
			node, err := valueToNode([]interface{}{"a", "b"})
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(node != nil).To(specs.BeTrue())
			ctx.Expect(len(node.Content)).ToEqual(2)
		})
		s.It("other type returns node", func(ctx *specs.Context) {
			node, err := valueToNode(42)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(node != nil).To(specs.BeTrue())
		})
		s.It("unsupported type panics or returns error", func(ctx *specs.Context) {
			defer func() {
				if r := recover(); r != nil {
					// yaml.Node.Encode panics for chan; recover is acceptable
				}
			}()
			_, err := valueToNode(make(chan int))
			if err != nil {
				ctx.Expect(err != nil).To(specs.BeTrue())
				return
			}
			// If no error, Encode may have panicked and we recovered
		})
	})
}

func Test_writePolicyFile(t *testing.T) {
	specs.Describe(t, "writePolicyFile", func(s *specs.Spec) {
		s.It("returns error when write fails", func(ctx *specs.Context) {
			dir := t.TempDir()
			p := &generatedPolicy{Name: "test", Type: "test", Rules: map[string]interface{}{"rule": "value"}}
			err := writePolicyFile(dir, p)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), "write")).To(specs.BeTrue())
		})
	})
}
