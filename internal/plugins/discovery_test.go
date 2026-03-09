package plugins

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pablogore/go-specs/specs"
)

func sliceContains(names []string, x string) bool {
	for _, n := range names {
		if n == x {
			return true
		}
	}
	return false
}

func setEnv(key, value string) func() {
	old, had := os.LookupEnv(key)
	if value == "" {
		_ = os.Unsetenv(key)
	} else {
		_ = os.Setenv(key, value)
	}
	return func() {
		if had {
			_ = os.Setenv(key, old)
		} else {
			_ = os.Unsetenv(key)
		}
	}
}

func TestDiscover(t *testing.T) {
	specs.Describe(t, "Discover", func(s *specs.Spec) {
		s.It("returns nil when DEVFORGE_PLUGIN_EXECUTION is set", func(ctx *specs.Context) {
			restore := setEnv("DEVFORGE_PLUGIN_EXECUTION", "1")
			defer restore()
			got := Discover()
			ctx.Expect(got == nil).To(specs.BeTrue())
		})
		s.It("returns nil when PATH is empty", func(ctx *specs.Context) {
			restore := setEnv("PATH", "")
			defer restore()
			restore2 := setEnv("DEVFORGE_PLUGIN_EXECUTION", "")
			defer restore2()
			got := Discover()
			ctx.Expect(got == nil).To(specs.BeTrue())
		})
		s.It("finds executable in PATH", func(ctx *specs.Context) {
			dir := t.TempDir()
			pluginPath := filepath.Join(dir, "forge-plugin-test-discovery")
			err := os.WriteFile(pluginPath, []byte("#!/bin/sh\nexit 0"), 0o600)
			ctx.Expect(err).To(specs.BeNil())
			//nolint:gosec // G302: executable bit required so Discover() finds the plugin in PATH
			err = os.Chmod(pluginPath, 0o755)
			ctx.Expect(err).To(specs.BeNil())
			oldPath := os.Getenv("PATH")
			_ = os.Setenv("PATH", dir)
			defer func() { _ = os.Setenv("PATH", oldPath) }()
			_ = os.Unsetenv("DEVFORGE_PLUGIN_EXECUTION")

			got := Discover()
			ctx.Expect(sliceContains(got, "test-discovery")).To(specs.BeTrue())
		})
		s.It("skips empty path entry and non-executable", func(ctx *specs.Context) {
			dir := t.TempDir()
			pluginPath := filepath.Join(dir, "forge-plugin-ok")
			err := os.WriteFile(pluginPath, []byte("x"), 0o600)
			ctx.Expect(err).To(specs.BeNil())
			//nolint:gosec // G302: executable bit required so Discover() finds the plugin in PATH
			err = os.Chmod(pluginPath, 0o755)
			ctx.Expect(err).To(specs.BeNil())
			noExec := filepath.Join(dir, "forge-plugin-skip")
			err = os.WriteFile(noExec, []byte("x"), 0o600)
			ctx.Expect(err).To(specs.BeNil())
			err = os.MkdirAll(filepath.Join(dir, "forge-plugin-dir"), 0o750)
			ctx.Expect(err).To(specs.BeNil())

			restore := setEnv("PATH", string(filepath.ListSeparator)+dir)
			defer restore()
			restore2 := setEnv("DEVFORGE_PLUGIN_EXECUTION", "")
			defer restore2()

			got := Discover()
			ctx.Expect(sliceContains(got, "ok")).To(specs.BeTrue())
			ctx.Expect(sliceContains(got, "skip")).To(specs.BeFalse())
			ctx.Expect(sliceContains(got, "dir")).To(specs.BeFalse())
		})
		s.It("skips symlink and empty plugin name and deduplicates", func(ctx *specs.Context) {
			dir := t.TempDir()
			pluginPath := filepath.Join(dir, "forge-plugin-visible")
			err := os.WriteFile(pluginPath, []byte("x"), 0o600)
			ctx.Expect(err).To(specs.BeNil())
			err = os.Chmod(pluginPath, 0o755)
			ctx.Expect(err).To(specs.BeNil())
			emptyName := filepath.Join(dir, "forge-plugin-")
			err = os.WriteFile(emptyName, []byte("x"), 0o600)
			ctx.Expect(err).To(specs.BeNil())
			err = os.Chmod(emptyName, 0o755)
			ctx.Expect(err).To(specs.BeNil())
			symlinkPath := filepath.Join(dir, "forge-plugin-symlink")
			err = os.Symlink(pluginPath, symlinkPath)
			ctx.Expect(err).To(specs.BeNil())

			restore := setEnv("PATH", dir)
			defer restore()
			restore2 := setEnv("DEVFORGE_PLUGIN_EXECUTION", "")
			defer restore2()

			got := Discover()
			ctx.Expect(sliceContains(got, "visible")).To(specs.BeTrue())
			ctx.Expect(sliceContains(got, "symlink")).To(specs.BeFalse())
		})
		s.It("skips path that is not a directory and non-prefix files", func(ctx *specs.Context) {
			dir := t.TempDir()
			pluginPath := filepath.Join(dir, "forge-plugin-found")
			err := os.WriteFile(pluginPath, []byte("x"), 0o600)
			ctx.Expect(err).To(specs.BeNil())
			err = os.Chmod(pluginPath, 0o755)
			ctx.Expect(err).To(specs.BeNil())
			fileAsPath := filepath.Join(dir, "not-a-dir")
			err = os.WriteFile(fileAsPath, []byte("x"), 0o600)
			ctx.Expect(err).To(specs.BeNil())
			otherBinary := filepath.Join(dir, "other-binary")
			err = os.WriteFile(otherBinary, []byte("x"), 0o600)
			ctx.Expect(err).To(specs.BeNil())
			err = os.Chmod(otherBinary, 0o755)
			ctx.Expect(err).To(specs.BeNil())

			pathWithEmpty := dir + string(filepath.ListSeparator) + "" + string(filepath.ListSeparator) + fileAsPath
			restore := setEnv("PATH", pathWithEmpty)
			defer restore()
			restore2 := setEnv("DEVFORGE_PLUGIN_EXECUTION", "")
			defer restore2()

			got := Discover()
			ctx.Expect(sliceContains(got, "found")).To(specs.BeTrue())
			ctx.Expect(sliceContains(got, "other")).To(specs.BeFalse())
		})
	})
}
