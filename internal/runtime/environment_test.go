//nolint:var-naming // package name matches adapter purpose; stdlib conflict accepted (same as environment.go)
package runtime

import (
	"os"
	"testing"

	"github.com/pablogore/go-specs/specs"
)

func TestDetectEnvironment(t *testing.T) {
	save := func(keys ...string) map[string]string {
		m := make(map[string]string)
		for _, k := range keys {
			if v, ok := os.LookupEnv(k); ok {
				m[k] = v
			}
		}
		return m
	}
	restore := func(m map[string]string) {
		for k := range m {
			_ = os.Setenv(k, m[k])
		}
	}
	unsetEnv := func(keys ...string) {
		for _, k := range keys {
			_ = os.Unsetenv(k)
		}
	}

	specs.Describe(t, "DetectEnvironment", func(s *specs.Spec) {
		s.It("IsLocal when no CI vars", func(ctx *specs.Context) {
			unsetEnv("CI", "GITHUB_ACTIONS", "GITLAB_CI")
			defer restore(save("CI", "GITHUB_ACTIONS", "GITLAB_CI"))
			env := DetectEnvironment()
			ctx.Expect(env.IsCI).ToEqual(false)
			ctx.Expect(env.IsGitHub).ToEqual(false)
			ctx.Expect(env.IsGitLab).ToEqual(false)
			ctx.Expect(env.IsLocal).To(specs.BeTrue())
		})
		s.It("IsCI when CI set", func(ctx *specs.Context) {
			unsetEnv("CI", "GITHUB_ACTIONS", "GITLAB_CI")
			defer restore(save("CI", "GITHUB_ACTIONS", "GITLAB_CI"))
			_ = os.Setenv("CI", "true")
			defer os.Unsetenv("CI")
			env := DetectEnvironment()
			ctx.Expect(env.IsCI).To(specs.BeTrue())
			ctx.Expect(env.IsLocal).ToEqual(false)
		})
		s.It("IsGitHub when GITHUB_ACTIONS set", func(ctx *specs.Context) {
			unsetEnv("CI", "GITHUB_ACTIONS", "GITLAB_CI")
			defer restore(save("CI", "GITHUB_ACTIONS", "GITLAB_CI"))
			_ = os.Setenv("GITHUB_ACTIONS", "true")
			defer os.Unsetenv("GITHUB_ACTIONS")
			env := DetectEnvironment()
			ctx.Expect(env.IsGitHub).To(specs.BeTrue())
		})
		s.It("IsGitLab when GITLAB_CI set", func(ctx *specs.Context) {
			unsetEnv("CI", "GITHUB_ACTIONS", "GITLAB_CI")
			defer restore(save("CI", "GITHUB_ACTIONS", "GITLAB_CI"))
			_ = os.Setenv("GITLAB_CI", "true")
			defer os.Unsetenv("GITLAB_CI")
			env := DetectEnvironment()
			ctx.Expect(env.IsGitLab).To(specs.BeTrue())
		})
	})
}
