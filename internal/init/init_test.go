package init

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pablogore/go-specs/specs"
)

func TestInitRepository(t *testing.T) {
	specs.Describe(t, "InitRepository", func(s *specs.Spec) {
		s.It("empty repo creates .syntegrity, .syntegrity.yml, .golangci.yml", func(ctx *specs.Context) {
			root := t.TempDir()
			result, err := InitRepository(root, false)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(result != nil).To(specs.BeTrue())
			ctx.Expect(dirExists(root, ".syntegrity")).To(specs.BeTrue())
			ctx.Expect(dirExists(root, ".syntegrity/policies")).To(specs.BeTrue())
			ctx.Expect(fileExists(root, ".syntegrity.yml")).To(specs.BeTrue())
			ctx.Expect(fileExists(root, ".golangci.yml")).To(specs.BeTrue())
			ctx.Expect(sliceContains(result.Created, ".syntegrity.yml")).To(specs.BeTrue())
			ctx.Expect(sliceContains(result.Created, ".golangci.yml")).To(specs.BeTrue())
		})
		s.It("existing .syntegrity policies are preserved", func(ctx *specs.Context) {
			root := t.TempDir()
			policiesDirPath := filepath.Join(root, ".syntegrity", "policies")
			ctx.Expect(os.MkdirAll(policiesDirPath, 0o750)).To(specs.BeNil())
			existingPolicy := filepath.Join(policiesDirPath, "custom.yaml")
			ctx.Expect(os.WriteFile(existingPolicy, []byte("name: custom\ntype: custom\n"), 0o600)).To(specs.BeNil())

			result, err := InitRepository(root, false)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(result != nil).To(specs.BeTrue())

			data, err := os.ReadFile(existingPolicy)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(string(data)).ToEqual("name: custom\ntype: custom\n")
			ctx.Expect(sliceContains(result.Created, ".syntegrity.yml")).To(specs.BeTrue())
		})
		s.It("does not overwrite existing config", func(ctx *specs.Context) {
			root := t.TempDir()
			configPath := filepath.Join(root, ".syntegrity.yml")
			ctx.Expect(os.WriteFile(configPath, []byte("mode: quick\nprofile: go-lib\n"), 0o600)).To(specs.BeNil())
			golangciPath := filepath.Join(root, ".golangci.yml")
			ctx.Expect(os.WriteFile(golangciPath, []byte("linters:\n  enable: [errcheck]\n"), 0o600)).To(specs.BeNil())

			result, err := InitRepository(root, false)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(result != nil).To(specs.BeTrue())

			configData, err := os.ReadFile(configPath)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(string(configData)).ToEqual("mode: quick\nprofile: go-lib\n")
			golangciData, err := os.ReadFile(golangciPath)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(string(golangciData)).ToEqual("linters:\n  enable: [errcheck]\n")
			for _, p := range result.Created {
				ctx.Expect(p != ".syntegrity.yml").To(specs.BeTrue())
				ctx.Expect(p != ".golangci.yml").To(specs.BeTrue())
			}
		})
		s.It("force overwrites existing config", func(ctx *specs.Context) {
			root := t.TempDir()
			configPath := filepath.Join(root, ".syntegrity.yml")
			ctx.Expect(os.WriteFile(configPath, []byte("mode: quick\n"), 0o600)).To(specs.BeNil())
			golangciPath := filepath.Join(root, ".golangci.yml")
			ctx.Expect(os.WriteFile(golangciPath, []byte("linters:\n  enable: [errcheck]\n"), 0o600)).To(specs.BeNil())

			result, err := InitRepository(root, true)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(result != nil).To(specs.BeTrue())
			ctx.Expect(sliceContains(result.Created, ".syntegrity.yml")).To(specs.BeTrue())
			ctx.Expect(sliceContains(result.Created, ".golangci.yml")).To(specs.BeTrue())

			configData, err := os.ReadFile(configPath)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(string(configData)).ToEqual("mode: full\n")
			golangciData, err := os.ReadFile(golangciPath)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(strings.Contains(string(golangciData), "version: \"2\"")).To(specs.BeTrue())
		})
		s.It("suggests workflow when .github/workflows exists but no devforge", func(ctx *specs.Context) {
			root := t.TempDir()
			workflowsDir := filepath.Join(root, ".github", "workflows")
			err := os.MkdirAll(workflowsDir, 0o750)
			ctx.Expect(err).To(specs.BeNil())
			err = os.WriteFile(filepath.Join(workflowsDir, "ci.yml"), []byte("name: other\non: push\n"), 0o600)
			ctx.Expect(err).To(specs.BeNil())
			result, err := InitRepository(root, false)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(result != nil).To(specs.BeTrue())
			ctx.Expect(len(result.Suggestion) > 0).To(specs.BeTrue())
			ctx.Expect(strings.Contains(result.Suggestion, "devforge")).To(specs.BeTrue())
		})
		s.It("MkdirAll failure returns error", func(ctx *specs.Context) {
			root := t.TempDir()
			syntegrityPath := filepath.Join(root, ".syntegrity")
			ctx.Expect(os.WriteFile(syntegrityPath, []byte("x"), 0o600)).To(specs.BeNil())
			result, err := InitRepository(root, false)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(result == nil).To(specs.BeTrue())
		})
		s.It("WriteFile config failure returns error", func(ctx *specs.Context) {
			root := t.TempDir()
			policiesDir := filepath.Join(root, ".syntegrity", "policies")
			ctx.Expect(os.MkdirAll(policiesDir, 0o750)).To(specs.BeNil())
			configPath := filepath.Join(root, ".syntegrity.yml")
			ctx.Expect(os.Mkdir(configPath, 0o750)).To(specs.BeNil())
			result, err := InitRepository(root, true)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(result == nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), ".syntegrity.yml")).To(specs.BeTrue())
		})
		s.It("WriteFile golangci failure returns error", func(ctx *specs.Context) {
			root := t.TempDir()
			policiesDir := filepath.Join(root, ".syntegrity", "policies")
			ctx.Expect(os.MkdirAll(policiesDir, 0o750)).To(specs.BeNil())
			configPath := filepath.Join(root, ".syntegrity.yml")
			ctx.Expect(os.WriteFile(configPath, []byte("profile: go-lib\n"), 0o600)).To(specs.BeNil())
			golangciPath := filepath.Join(root, ".golangci.yml")
			ctx.Expect(os.Mkdir(golangciPath, 0o750)).To(specs.BeNil())
			result, err := InitRepository(root, true)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(result == nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(err.Error(), ".golangci.yml")).To(specs.BeTrue())
		})
		s.It("generates policies when issues exist", func(ctx *specs.Context) {
			root := t.TempDir()
			domainDir := filepath.Join(root, "internal", "domain")
			ctx.Expect(os.MkdirAll(domainDir, 0o750)).To(specs.BeNil())
			ctx.Expect(os.WriteFile(filepath.Join(domainDir, "pkg.go"), []byte("package domain\nimport _ \"internal/adapters\"\n"), 0o600)).To(specs.BeNil())

			result, err := InitRepository(root, false)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(result != nil).To(specs.BeTrue())
			ctx.Expect(sliceContains(result.Created, ".syntegrity/policies/architecture.yaml")).To(specs.BeTrue())

			archPath := filepath.Join(root, ".syntegrity", "policies", "architecture.yaml")
			info, err := os.Stat(archPath)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(info.IsDir() == false).To(specs.BeTrue())
			data, err := os.ReadFile(archPath)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(strings.Contains(string(data), "forbid_import")).To(specs.BeTrue())
			ctx.Expect(strings.Contains(string(data), "internal/adapters")).To(specs.BeTrue())
		})
		s.It("suggestion contains devforge when no workflow", func(ctx *specs.Context) {
			root := t.TempDir()
			workflowsDir := filepath.Join(root, ".github", "workflows")
			ctx.Expect(os.MkdirAll(workflowsDir, 0o750)).To(specs.BeNil())
			ctx.Expect(os.WriteFile(filepath.Join(workflowsDir, "other.yml"), []byte("name: Other\n"), 0o600)).To(specs.BeNil())

			result, err := InitRepository(root, false)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(result != nil).To(specs.BeTrue())
			ctx.Expect(strings.Contains(result.Suggestion, "devforge")).To(specs.BeTrue())
			ctx.Expect(strings.Contains(result.Suggestion, "devforge/devforge")).To(specs.BeTrue())
		})
	})
}

func TestExistingConfigFiles(t *testing.T) {
	specs.Describe(t, "ExistingConfigFiles", func(s *specs.Spec) {
		s.It("returns empty for empty dir", func(ctx *specs.Context) {
			root := t.TempDir()
			ctx.Expect(len(ExistingConfigFiles(root))).ToEqual(0)
		})
		s.It("returns .syntegrity.yml when present", func(ctx *specs.Context) {
			root := t.TempDir()
			ctx.Expect(os.WriteFile(filepath.Join(root, ".syntegrity.yml"), nil, 0o600)).To(specs.BeNil())
			ctx.Expect(ExistingConfigFiles(root)).ToEqual([]string{".syntegrity.yml"})
		})
		s.It("returns both when both present", func(ctx *specs.Context) {
			root := t.TempDir()
			ctx.Expect(os.WriteFile(filepath.Join(root, ".syntegrity.yml"), nil, 0o600)).To(specs.BeNil())
			ctx.Expect(os.WriteFile(filepath.Join(root, ".golangci.yml"), nil, 0o600)).To(specs.BeNil())
			got := ExistingConfigFiles(root)
			ctx.Expect(len(got)).ToEqual(2)
			ctx.Expect(sliceContains(got, ".syntegrity.yml")).To(specs.BeTrue())
			ctx.Expect(sliceContains(got, ".golangci.yml")).To(specs.BeTrue())
		})
	})
}

func TestHasSyntegrityWorkflow(t *testing.T) {
	specs.Describe(t, "hasSyntegrityWorkflow", func(s *specs.Spec) {
		s.It("returns true when workflow uses devforge", func(ctx *specs.Context) {
			root := t.TempDir()
			workflowsDir := filepath.Join(root, ".github", "workflows")
			ctx.Expect(os.MkdirAll(workflowsDir, 0o750)).To(specs.BeNil())
			ctx.Expect(os.WriteFile(filepath.Join(workflowsDir, "ci.yml"), []byte("name: CI\njobs:\n  syntegrity:\n    steps:\n      - uses: devforge/devforge@v1\n"), 0o600)).To(specs.BeNil())
			ctx.Expect(hasSyntegrityWorkflow(workflowsDir)).To(specs.BeTrue())
		})
		s.It("returns true when content contains devforge", func(ctx *specs.Context) {
			root := t.TempDir()
			workflowsDir := filepath.Join(root, ".github", "workflows")
			ctx.Expect(os.MkdirAll(workflowsDir, 0o750)).To(specs.BeNil())
			ctx.Expect(os.WriteFile(filepath.Join(workflowsDir, "ci.yml"), []byte("devforge\n"), 0o600)).To(specs.BeNil())
			ctx.Expect(hasSyntegrityWorkflow(workflowsDir)).To(specs.BeTrue())
		})
		s.It("returns false when no syntegrity workflow", func(ctx *specs.Context) {
			root := t.TempDir()
			workflowsDir := filepath.Join(root, ".github", "workflows")
			ctx.Expect(os.MkdirAll(workflowsDir, 0o750)).To(specs.BeNil())
			ctx.Expect(os.WriteFile(filepath.Join(workflowsDir, "other.yml"), []byte("name: Other\n"), 0o600)).To(specs.BeNil())
			ctx.Expect(hasSyntegrityWorkflow(workflowsDir)).To(specs.BeFalse())
		})
		s.It("returns false when read dir fails", func(ctx *specs.Context) {
			ctx.Expect(hasSyntegrityWorkflow(filepath.Join(t.TempDir(), "nonexistent"))).To(specs.BeFalse())
		})
		s.It("skips dirs and non-yaml and finds devforge", func(ctx *specs.Context) {
			root := t.TempDir()
			workflowsDir := filepath.Join(root, ".github", "workflows")
			ctx.Expect(os.MkdirAll(workflowsDir, 0o750)).To(specs.BeNil())
			ctx.Expect(os.MkdirAll(filepath.Join(workflowsDir, "subdir"), 0o750)).To(specs.BeNil())
			ctx.Expect(os.WriteFile(filepath.Join(workflowsDir, "readme.txt"), []byte("not yaml"), 0o600)).To(specs.BeNil())
			ctx.Expect(os.WriteFile(filepath.Join(workflowsDir, "ci.yml"), []byte("uses: devforge/devforge@v1\n"), 0o600)).To(specs.BeNil())
			ctx.Expect(hasSyntegrityWorkflow(workflowsDir)).To(specs.BeTrue())
		})
		s.It("skips entry when ReadFile fails", func(ctx *specs.Context) {
			root := t.TempDir()
			workflowsDir := filepath.Join(root, ".github", "workflows")
			ctx.Expect(os.MkdirAll(workflowsDir, 0o750)).To(specs.BeNil())
			ctx.Expect(os.Mkdir(filepath.Join(workflowsDir, "ci.yml"), 0o750)).To(specs.BeNil())
			ctx.Expect(os.WriteFile(filepath.Join(workflowsDir, "real.yml"), []byte("devforge\n"), 0o600)).To(specs.BeNil())
			ctx.Expect(hasSyntegrityWorkflow(workflowsDir)).To(specs.BeTrue())
		})
	})
}

func dirExists(root, rel string) bool {
	info, err := os.Stat(filepath.Join(root, rel))
	return err == nil && info.IsDir()
}

func fileExists(root, rel string) bool {
	info, err := os.Stat(filepath.Join(root, rel))
	return err == nil && !info.IsDir()
}

func sliceContains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
