package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pablogore/go-specs/specs"
)

func setEnv(key, value string) func() {
	old, had := os.LookupEnv(key)
	_ = os.Setenv(key, value)
	return func() {
		if had {
			_ = os.Setenv(key, old)
		} else {
			_ = os.Unsetenv(key)
		}
	}
}

func TestExists(t *testing.T) {
	specs.Describe(t, "Exists", func(s *specs.Spec) {
		s.It("returns true for binary in PATH", func(ctx *specs.Context) {
			dir := t.TempDir()
			name := "devforge-test-exists-xyz"
			path := filepath.Join(dir, name)
			err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o600)
			ctx.Expect(err).To(specs.BeNil())
			err = os.Chmod(path, 0o755)
			ctx.Expect(err).To(specs.BeNil())
			restore := setEnv("PATH", dir+string(filepath.ListSeparator)+os.Getenv("PATH"))
			defer restore()
			ctx.Expect(Exists(name)).To(specs.BeTrue())
		})
		s.It("returns false for nonexistent binary", func(ctx *specs.Context) {
			ctx.Expect(Exists("devforge-nonexistent-binary-xyz-12345")).To(specs.BeFalse())
		})
	})
}

func Test_pathExists(t *testing.T) {
	specs.Describe(t, "pathExists", func(s *specs.Spec) {
		s.It("returns true for existing dir and file", func(ctx *specs.Context) {
			dir := t.TempDir()
			ctx.Expect(pathExists(dir)).To(specs.BeTrue())
			f := filepath.Join(dir, "f")
			err := os.WriteFile(f, []byte("x"), 0o600)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(pathExists(f)).To(specs.BeTrue())
		})
		s.It("returns false for nonexistent path", func(ctx *specs.Context) {
			dir := t.TempDir()
			ctx.Expect(pathExists(filepath.Join(dir, "nonexistent"))).To(specs.BeFalse())
		})
	})
}

func Test_ensureSymlink(t *testing.T) {
	specs.Describe(t, "ensureSymlink", func(s *specs.Spec) {
		s.It("creates symlink and is idempotent", func(ctx *specs.Context) {
			dir := t.TempDir()
			target := filepath.Join(dir, "target")
			link := filepath.Join(dir, "link")
			err := os.WriteFile(target, []byte("x"), 0o600)
			ctx.Expect(err).To(specs.BeNil())

			err = ensureSymlink(target, link)
			ctx.Expect(err).To(specs.BeNil())
			dest, err := os.Readlink(link)
			ctx.Expect(err).To(specs.BeNil())
			ctx.Expect(filepath.Clean(dest)).ToEqual(filepath.Clean(target))

			err = ensureSymlink(target, link)
			ctx.Expect(err).To(specs.BeNil())
		})
		s.It("replaces symlink when target differs", func(ctx *specs.Context) {
			dir := t.TempDir()
			target := filepath.Join(dir, "target")
			link := filepath.Join(dir, "link")
			other := filepath.Join(dir, "other")
			err := os.WriteFile(target, []byte("x"), 0o600)
			ctx.Expect(err).To(specs.BeNil())
			err = os.WriteFile(other, []byte("y"), 0o600)
			ctx.Expect(err).To(specs.BeNil())
			err = os.Symlink(other, link)
			ctx.Expect(err).To(specs.BeNil())
			err = ensureSymlink(target, link)
			ctx.Expect(err).To(specs.BeNil())
			dest, _ := os.Readlink(link)
			ctx.Expect(filepath.Clean(dest)).ToEqual(filepath.Clean(target))
		})
		s.It("fails when link exists as regular file", func(ctx *specs.Context) {
			dir := t.TempDir()
			target := filepath.Join(dir, "target")
			link := filepath.Join(dir, "link")
			err := os.WriteFile(target, []byte("x"), 0o600)
			ctx.Expect(err).To(specs.BeNil())
			err = os.WriteFile(link, []byte("notalink"), 0o600)
			ctx.Expect(err).To(specs.BeNil())
			err = ensureSymlink(target, link)
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
		s.It("returns error when Lstat fails with non-IsNotExist", func(ctx *specs.Context) {
			dir := t.TempDir()
			restricted := filepath.Join(dir, "restricted")
			err := os.Mkdir(restricted, 0o700)
			ctx.Expect(err).To(specs.BeNil())
			target := filepath.Join(dir, "target")
			link := filepath.Join(restricted, "link")
			err = os.WriteFile(target, []byte("x"), 0o600)
			ctx.Expect(err).To(specs.BeNil())
			err = os.Chmod(restricted, 0o000)
			ctx.Expect(err).To(specs.BeNil())
			defer func() { _ = os.Chmod(restricted, 0o700) }()
			err = ensureSymlink(target, link)
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
		s.It("replaces symlink when Readlink returns different target", func(ctx *specs.Context) {
			dir := t.TempDir()
			target := filepath.Join(dir, "target")
			link := filepath.Join(dir, "link")
			err := os.WriteFile(target, []byte("x"), 0o600)
			ctx.Expect(err).To(specs.BeNil())
			err = os.Symlink(target, link)
			ctx.Expect(err).To(specs.BeNil())
			err = ensureSymlink(target, link)
			ctx.Expect(err).To(specs.BeNil())
			dest, _ := os.Readlink(link)
			ctx.Expect(filepath.Clean(dest)).ToEqual(filepath.Clean(target))
		})
	})
}

func Test_ensureBinDir_and_ensurePath(t *testing.T) {
	specs.Describe(t, "ensureBinDir and ensurePath", func(s *specs.Spec) {
		s.It("ensureBinDir creates bin and ensurePath prepends to PATH", func(ctx *specs.Context) {
			home := t.TempDir()
			restore := setEnv("HOME", home)
			defer restore()

			err := ensureBinDir()
			ctx.Expect(err).To(specs.BeNil())
			bin := toolsBinDir()
			ctx.Expect(pathExists(bin)).To(specs.BeTrue())
			ensurePath()
			path := os.Getenv("PATH")
			ctx.Expect(path != "" && containsPath(path, bin)).To(specs.BeTrue())
		})
	})
}

func Test_ensurePath_whenPATHEmpty(t *testing.T) {
	specs.Describe(t, "ensurePath when PATH empty", func(s *specs.Spec) {
		s.It("sets PATH to bin", func(ctx *specs.Context) {
			home := t.TempDir()
			restoreHome := setEnv("HOME", home)
			defer restoreHome()
			restorePath := setEnv("PATH", "")
			defer restorePath()

			err := ensureBinDir()
			ctx.Expect(err).To(specs.BeNil())
			bin := toolsBinDir()
			ensurePath()
			path := os.Getenv("PATH")
			ctx.Expect(path).ToEqual(bin)
		})
	})
}

func Test_containsPath(t *testing.T) {
	specs.Describe(t, "containsPath", func(s *specs.Spec) {
		s.It("same path returns true", func(ctx *specs.Context) {
			dir := filepath.Clean("/some/bin")
			ctx.Expect(containsPath("/some/bin", dir)).To(specs.BeTrue())
		})
		s.It("path in list returns true", func(ctx *specs.Context) {
			dir := filepath.Clean("/some/bin")
			ctx.Expect(containsPath("/other"+string(filepath.ListSeparator)+"/some/bin", dir)).To(specs.BeTrue())
		})
		s.It("empty returns false", func(ctx *specs.Context) {
			dir := filepath.Clean("/some/bin")
			ctx.Expect(containsPath("", dir)).To(specs.BeFalse())
		})
		s.It("different path returns false", func(ctx *specs.Context) {
			dir := filepath.Clean("/some/bin")
			ctx.Expect(containsPath("/other", dir)).To(specs.BeFalse())
		})
	})
}

func Test_toolVersionDir(t *testing.T) {
	specs.Describe(t, "toolVersionDir", func(s *specs.Spec) {
		s.It("returns path under tools root", func(ctx *specs.Context) {
			home := t.TempDir()
			restore := setEnv("HOME", home)
			defer restore()

			got := toolVersionDir("golangci-lint", "v1.0")
			want := filepath.Join(toolsRoot(), "golangci-lint", "v1.0")
			ctx.Expect(got).ToEqual(want)
		})
	})
}

func Test_cachedBinaryPath(t *testing.T) {
	specs.Describe(t, "cachedBinaryPath", func(s *specs.Spec) {
		s.It("returns path under tools root", func(ctx *specs.Context) {
			home := t.TempDir()
			restore := setEnv("HOME", home)
			defer restore()

			got := cachedBinaryPath("tool", "v1", "binary")
			want := filepath.Join(toolsRoot(), "tool", "v1", "binary")
			ctx.Expect(got).ToEqual(want)
		})
	})
}

func Test_binSymlinkPath(t *testing.T) {
	specs.Describe(t, "binSymlinkPath", func(s *specs.Spec) {
		s.It("returns path under tools bin", func(ctx *specs.Context) {
			home := t.TempDir()
			restore := setEnv("HOME", home)
			defer restore()

			got := binSymlinkPath("golangci-lint")
			want := filepath.Join(toolsRoot(), "bin", "golangci-lint")
			ctx.Expect(got).ToEqual(want)
		})
	})
}

func TestEnsureTools(t *testing.T) {
	specs.Describe(t, "EnsureTools", func(s *specs.Spec) {
		s.It("when cache exists only creates symlinks", func(ctx *specs.Context) {
			home := t.TempDir()
			restore := setEnv("HOME", home)
			defer restore()

			err := ensureBinDir()
			ctx.Expect(err).To(specs.BeNil())
			gocached := cachedBinaryPath("golangci-lint", GolangCILintVersion, "golangci-lint")
			err = os.MkdirAll(filepath.Dir(gocached), 0o750)
			ctx.Expect(err).To(specs.BeNil())
			err = os.WriteFile(gocached, []byte("fake"), 0o600)
			ctx.Expect(err).To(specs.BeNil())
			vulncached := cachedBinaryPath("govulncheck", GovulncheckVersion, "govulncheck")
			err = os.MkdirAll(filepath.Dir(vulncached), 0o750)
			ctx.Expect(err).To(specs.BeNil())
			err = os.WriteFile(vulncached, []byte("fake"), 0o600)
			ctx.Expect(err).To(specs.BeNil())

			err = EnsureTools()
			ctx.Expect(err).To(specs.BeNil())

			linkGo := binSymlinkPath("golangci-lint")
			linkVuln := binSymlinkPath("govulncheck")
			ctx.Expect(pathExists(linkGo)).To(specs.BeTrue())
			ctx.Expect(pathExists(linkVuln)).To(specs.BeTrue())
		})
		s.It("when cache missing runs installers", func(ctx *specs.Context) {
			home := t.TempDir()
			restore := setEnv("HOME", home)
			defer restore()

			err := ensureBinDir()
			ctx.Expect(err).To(specs.BeNil())

			origGo := installGolangCILintRunner
			origVuln := installGovulncheckRunner
			defer func() {
				installGolangCILintRunner = origGo
				installGovulncheckRunner = origVuln
			}()

			installGolangCILintRunner = func(versionDir string) error {
				return os.WriteFile(filepath.Join(versionDir, "golangci-lint"), []byte("fake"), 0o755)
			}
			installGovulncheckRunner = func(versionDir string) error {
				return os.WriteFile(filepath.Join(versionDir, "govulncheck"), []byte("fake"), 0o755)
			}

			err = EnsureTools()
			ctx.Expect(err).To(specs.BeNil())

			linkGo := binSymlinkPath("golangci-lint")
			linkVuln := binSymlinkPath("govulncheck")
			ctx.Expect(pathExists(linkGo)).To(specs.BeTrue())
			ctx.Expect(pathExists(linkVuln)).To(specs.BeTrue())
		})
		s.It("when golangci-lint install fails returns error", func(ctx *specs.Context) {
			home := t.TempDir()
			restore := setEnv("HOME", home)
			defer restore()

			err := ensureBinDir()
			ctx.Expect(err).To(specs.BeNil())

			origGo := installGolangCILintRunner
			defer func() { installGolangCILintRunner = origGo }()

			installGolangCILintRunner = func(_ string) error {
				return os.ErrPermission
			}

			err = EnsureTools()
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(err.Error() != "").To(specs.BeTrue())
		})
		s.It("when govulncheck install fails returns error", func(ctx *specs.Context) {
			home := t.TempDir()
			restore := setEnv("HOME", home)
			defer restore()

			err := ensureBinDir()
			ctx.Expect(err).To(specs.BeNil())
			gocached := cachedBinaryPath("golangci-lint", GolangCILintVersion, "golangci-lint")
			err = os.MkdirAll(filepath.Dir(gocached), 0o750)
			ctx.Expect(err).To(specs.BeNil())
			err = os.WriteFile(gocached, []byte("fake"), 0o600)
			ctx.Expect(err).To(specs.BeNil())

			origVuln := installGovulncheckRunner
			defer func() { installGovulncheckRunner = origVuln }()

			installGovulncheckRunner = func(_ string) error {
				return os.ErrPermission
			}

			err = EnsureTools()
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
		s.It("when ensureBinDir fails returns error", func(ctx *specs.Context) {
			home := t.TempDir()
			devforgePath := filepath.Join(home, ".devforge")
			err := os.WriteFile(devforgePath, []byte("x"), 0o600)
			ctx.Expect(err).To(specs.BeNil())
			restore := setEnv("HOME", home)
			defer restore()

			err = EnsureTools()
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
		s.It("when golangci-lint version dir cannot be created returns error", func(ctx *specs.Context) {
			home := t.TempDir()
			restore := setEnv("HOME", home)
			defer restore()
			toolsDir := filepath.Join(home, ".devforge", "tools")
			ctx.Expect(os.MkdirAll(filepath.Join(toolsDir, "bin"), 0o750)).To(specs.BeNil())
			golangciFile := filepath.Join(toolsDir, "golangci-lint")
			ctx.Expect(os.WriteFile(golangciFile, []byte("x"), 0o600)).To(specs.BeNil())

			err := EnsureTools()
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
		s.It("when govulncheck version dir cannot be created returns error", func(ctx *specs.Context) {
			home := t.TempDir()
			restore := setEnv("HOME", home)
			defer restore()
			ctx.Expect(ensureBinDir()).To(specs.BeNil())
			gocached := cachedBinaryPath("golangci-lint", GolangCILintVersion, "golangci-lint")
			ctx.Expect(os.MkdirAll(filepath.Dir(gocached), 0o750)).To(specs.BeNil())
			ctx.Expect(os.WriteFile(gocached, []byte("x"), 0o600)).To(specs.BeNil())
			govulnFile := filepath.Join(toolsRoot(), "govulncheck")
			ctx.Expect(os.WriteFile(govulnFile, []byte("x"), 0o600)).To(specs.BeNil())

			err := EnsureTools()
			ctx.Expect(err != nil).To(specs.BeTrue())
		})
	})
}

