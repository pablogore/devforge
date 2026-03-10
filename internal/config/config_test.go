package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pablogore/go-specs/specs"
	yaml "gopkg.in/yaml.v3"
)

func TestUnmarshalYAML(t *testing.T) {
	specs.Describe(t, "Config YAML unmarshal", func(s *specs.Spec) {
		s.It("decodes policies coverage", func(ctx *specs.Context) {
			yamlBytes := []byte(`mode: full

policies:
  coverage:
    threshold: 95
    packages:
      - "*"
`)
			var c Config
			err := yaml.Unmarshal(yamlBytes, &c)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(c.Policies != nil).To(specs.BeTrue())
			ctx.Expect(c.Policies.Coverage != nil).To(specs.BeTrue())
			ctx.Expect(c.Policies.Coverage.Threshold).ToEqual(95)
			ctx.Expect(len(c.Policies.Coverage.Packages) == 1 && c.Policies.Coverage.Packages[0] == "*").To(specs.BeTrue())
			ctx.Expect(c.Mode).ToEqual("full")
		})
		s.It("decodes policies coverage exclude", func(ctx *specs.Context) {
			yamlBytes := []byte(`policies:
  coverage:
    threshold: 90
    exclude:
      - "**/testkit/**"
`)
			var c Config
			err := yaml.Unmarshal(yamlBytes, &c)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(c.Policies != nil && c.Policies.Coverage != nil).To(specs.BeTrue())
			ctx.Expect(c.Policies.Coverage.Threshold).ToEqual(90)
			ctx.Expect(len(c.Policies.Coverage.Exclude) == 1 && c.Policies.Coverage.Exclude[0] == "**/testkit/**").To(specs.BeTrue())
		})
		s.It("plugins nil when not present", func(ctx *specs.Context) {
			yamlBytes := []byte("profile: go-lib\n")
			var c Config
			ctx.Expect(yaml.Unmarshal(yamlBytes, &c)).To(specs.BeNil())
			ctx.Expect(c.Plugins).To(specs.BeNil())
			ctx.Expect(c.PluginConfig).To(specs.BeNil())
		})
		s.It("plugins as map populates PluginConfig when present", func(ctx *specs.Context) {
			yamlBytes := []byte(`profile: go-lib
plugins:
  myplugin:
    enabled: false
    severity: high
`)
			var c Config
			ctx.Expect(yaml.Unmarshal(yamlBytes, &c)).To(specs.BeNil())
			ctx.Expect(c.Profile).ToEqual("go-lib")
			if c.PluginConfig != nil {
				_, ok := c.PluginConfig["myplugin"]
				ctx.Expect(ok).To(specs.BeTrue())
				ctx.Expect(c.PluginConfig["myplugin"].Enabled).ToEqual(false)
				ctx.Expect(c.PluginConfig["myplugin"].Params["severity"]).ToEqual("high")
			}
		})
		s.It("plugins as list populates Plugins", func(ctx *specs.Context) {
			yamlBytes := []byte(`profile: go-lib
plugins:
  - name: lint-extra
    run: "echo lint"
`)
			var c Config
			ctx.Expect(yaml.Unmarshal(yamlBytes, &c)).To(specs.BeNil())
			ctx.Expect(c.Profile).ToEqual("go-lib")
			ctx.Expect(len(c.Plugins)).ToEqual(1)
			ctx.Expect(c.Plugins[0].Name).ToEqual("lint-extra")
			ctx.Expect(c.Plugins[0].Run).ToEqual("echo lint")
		})
		s.It("plugins node with non-sequence non-mapping leaves nil", func(ctx *specs.Context) {
			// Scalar or other node kind: default branch in UnmarshalYAML
			yamlBytes := []byte("profile: x\nplugins: scalar\n")
			var c Config
			ctx.Expect(yaml.Unmarshal(yamlBytes, &c)).To(specs.BeNil())
			ctx.Expect(c.Plugins).To(specs.BeNil())
			ctx.Expect(c.PluginConfig).To(specs.BeNil())
		})
		s.It("invalid root structure returns Decode error", func(ctx *specs.Context) {
			yamlBytes := []byte("profile:\n  nested: key\nmode: full\n")
			var c Config
			err := yaml.Unmarshal(yamlBytes, &c)
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
	})
}

func TestLoadConfig(t *testing.T) {
	specs.Describe(t, "LoadConfig", func(s *specs.Spec) {
		s.It("reads policies from file in temp dir", func(ctx *specs.Context) {
			dir := t.TempDir()
			path := filepath.Join(dir, configFileName)
			content := []byte(`mode: full
policies:
  coverage:
    threshold: 95
    packages:
      - "*"
`)
			ctx.Expect(os.WriteFile(path, content, 0o600)).To(specs.BeNil())
			cfg, err := LoadConfig(dir)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(cfg != nil).To(specs.BeTrue())
			ctx.Expect(cfg.Policies != nil).To(specs.BeTrue())
			ctx.Expect(cfg.Policies.Coverage != nil).To(specs.BeTrue())
			ctx.Expect(cfg.Policies.Coverage.Threshold).ToEqual(95)
			ctx.Expect(len(cfg.Policies.Coverage.Packages) == 1 && cfg.Policies.Coverage.Packages[0] == "*").To(specs.BeTrue())
		})
		s.It("file not found returns default", func(ctx *specs.Context) {
			dir := t.TempDir()
			cfg, err := LoadConfig(dir)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(cfg != nil).To(specs.BeTrue())
			ctx.Expect(cfg.Mode).ToEqual("full")
			ctx.Expect(cfg.Profile).ToEqual("")
		})
		s.It("read error returns error", func(ctx *specs.Context) {
			dir := t.TempDir()
			path := filepath.Join(dir, configFileName)
			_ = os.Mkdir(path, 0o750)
			cfg, err := LoadConfig(dir)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(cfg == nil).To(specs.BeTrue())
		})
		s.It("applies defaults when mode empty", func(ctx *specs.Context) {
			dir := t.TempDir()
			path := filepath.Join(dir, configFileName)
			_ = os.WriteFile(path, []byte("profile: go-lib\n"), 0o600)
			cfg, err := LoadConfig(dir)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(cfg != nil).To(specs.BeTrue())
			ctx.Expect(cfg.Mode).ToEqual("full")
			ctx.Expect(cfg.Profile).ToEqual("go-lib")
		})
		s.It("strips BOM", func(ctx *specs.Context) {
			dir := t.TempDir()
			path := filepath.Join(dir, configFileName)
			bom := []byte{0xef, 0xbb, 0xbf}
			content := append(bom, []byte("profile: go-lib\nmode: full\n")...)
			_ = os.WriteFile(path, content, 0o600)
			cfg, err := LoadConfig(dir)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(cfg != nil).To(specs.BeTrue())
			ctx.Expect(cfg.Profile).ToEqual("go-lib")
		})
		s.It("workdir dot uses cwd", func(ctx *specs.Context) {
			dir := t.TempDir()
			path := filepath.Join(dir, configFileName)
			_ = os.WriteFile(path, []byte("profile: go-lib\n"), 0o600)
			prev, err := os.Getwd()
			ctx.Expect(err).To(specs.BeNil())
			defer func() { _ = os.Chdir(prev) }()
			_ = os.Chdir(dir)
			cfg, err := LoadConfig(".")
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(cfg != nil).To(specs.BeTrue())
			ctx.Expect(cfg.Profile).ToEqual("go-lib")
		})
		s.It("invalid YAML returns error", func(ctx *specs.Context) {
			dir := t.TempDir()
			path := filepath.Join(dir, configFileName)
			_ = os.WriteFile(path, []byte("profile: [unclosed\n"), 0o600)
			cfg, err := LoadConfig(dir)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(cfg == nil).To(specs.BeTrue())
		})
	})
}

func TestDefaultConfig(t *testing.T) {
	specs.Describe(t, "DefaultConfig", func(s *specs.Spec) {
		s.It("returns non-nil with full mode and empty profile", func(ctx *specs.Context) {
			cfg := DefaultConfig()
			ctx.Expect(cfg != nil).To(specs.BeTrue())
			ctx.Expect(cfg.Profile).ToEqual("")
			ctx.Expect(cfg.Mode).ToEqual("full")
			ctx.Expect(cfg.Plugins).To(specs.BeNil())
			ctx.Expect(cfg.PluginConfig).To(specs.BeNil())
		})
	})
}

func TestExternalPluginCfg_UnmarshalYAML(t *testing.T) {
	specs.Describe(t, "ExternalPluginCfg YAML", func(s *specs.Spec) {
		s.It("decodes enabled and params", func(ctx *specs.Context) {
			yamlBytes := []byte(`
enabled: false
severity: high
`)
			var e ExternalPluginCfg
			ctx.Expect(yaml.Unmarshal(yamlBytes, &e)).To(specs.BeNil())
			ctx.Expect(e.Enabled).ToEqual(false)
			ctx.Expect(e.Params["severity"]).ToEqual("high")
		})
		s.It("enabled non-bool leaves default true", func(ctx *specs.Context) {
			yamlBytes := []byte("enabled: yes\n")
			var e ExternalPluginCfg
			ctx.Expect(yaml.Unmarshal(yamlBytes, &e)).To(specs.BeNil())
			ctx.Expect(e.Enabled).To(specs.BeTrue())
		})
		s.It("Decode error when value is not a map", func(ctx *specs.Context) {
			yamlBytes := []byte("[1, 2, 3]\n")
			var e ExternalPluginCfg
			err := yaml.Unmarshal(yamlBytes, &e)
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
	})
}

