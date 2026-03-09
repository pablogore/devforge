package domain

import (
	"errors"
	"testing"

	"github.com/pablogore/go-specs/specs"
)

func TestVersion(t *testing.T) {
	specs.Describe(t, "Version", func(s *specs.Spec) {
		s.It("String formats as vMajor.Minor.Patch", func(ctx *specs.Context) {
			cases := []struct {
				v    Version
				want string
			}{
				{Version{0, 1, 0}, "v0.1.0"},
				{Version{1, 0, 0}, "v1.0.0"},
				{Version{2, 3, 4}, "v2.3.4"},
			}
			for _, c := range cases {
				ctx.Expect(c.v.String()).ToEqual(c.want)
			}
		})
	})
}

func TestParseVersion(t *testing.T) {
	specs.Describe(t, "ParseVersion", func(s *specs.Spec) {
		s.It("covers valid and invalid paths", func(ctx *specs.Context) {
			cases := []struct {
				tag     string
				want    Version
				wantErr bool
			}{
				{"v1.2.3", Version{1, 2, 3}, false},
				{"1.2.3", Version{1, 2, 3}, false},
				{"v0.0.0", Version{0, 0, 0}, false},
				{"v1.2", Version{}, true},
				{"v1.2.3.4", Version{}, true},
				{"vx.2.3", Version{}, true},
				{"v1.a.3", Version{}, true},
				{"v1.2.x", Version{}, true},
				{"", Version{}, true},
			}
			for _, c := range cases {
				got, err := ParseVersion(c.tag)
				if c.wantErr {
					ctx.Expect(err != nil).To(specs.BeTrue())
					ctx.Expect(errors.Is(err, ErrInvalidVersionFormat)).To(specs.BeTrue())
					continue
				}
				ctx.Expect(err).To(specs.BeNil())
				ctx.Expect(got).ToEqual(c.want)
			}
		})
	})
}

func TestParseCommitType(t *testing.T) {
	specs.Describe(t, "ParseCommitType", func(s *specs.Spec) {
		s.It("covers commit type paths", func(ctx *specs.Context) {
			cases := []struct {
				msg   string
				want  CommitType
				found bool
			}{
				{"feat: add", CommitTypeFeature, true},
				{"fix: bug", CommitTypeFix, true},
				{"feat(api): add", CommitTypeFeature, true},
				{"feat!: break", CommitTypeFeature, true},
				{"not a type: msg", "", false},
				{"", "", false},
				{"refactor: x", CommitTypeRefactor, true},
				{"perf: x", CommitTypePerf, true},
				{"docs: x", CommitTypeDocs, true},
				{"test: x", CommitTypeTest, true},
				{"chore: x", CommitTypeChore, true},
				{"build: x", CommitTypeBuild, true},
				{"ci: x", CommitTypeCI, true},
				{"revert: x", CommitTypeRevert, true},
			}
			for _, c := range cases {
				ct, ok := ParseCommitType(c.msg)
				ctx.Expect(ok).ToEqual(c.found)
				if ok {
					ctx.Expect(ct).ToEqual(c.want)
				}
			}
		})
	})
}

func TestCommitType_HasBreakingChange(t *testing.T) {
	specs.Describe(t, "CommitType.HasBreakingChange", func(s *specs.Spec) {
		s.It("detects breaking change in footer or suffix", func(ctx *specs.Context) {
			ctx.Expect(CommitTypeFeature.HasBreakingChange("feat!: break")).To(specs.BeTrue())
			ctx.Expect(CommitTypeFeature.HasBreakingChange("feat: break\n\nBREAKING CHANGE: api")).To(specs.BeTrue())
			ctx.Expect(CommitTypeFeature.HasBreakingChange("feat: normal")).To(specs.BeFalse())
		})
	})
}

func TestCommitType_ToBumpType(t *testing.T) {
	specs.Describe(t, "CommitType.ToBumpType", func(s *specs.Spec) {
		s.It("covers bump type paths", func(ctx *specs.Context) {
			cases := []struct {
				ct   CommitType
				msg  string
				want BumpType
			}{
				{CommitTypeFeature, "feat: x", BumpTypeMinor},
				{CommitTypeFeature, "feat!: x", BumpTypeMajor},
				{CommitTypeFix, "fix: x", BumpTypePatch},
				{CommitTypePerf, "perf: x", BumpTypePatch},
				{CommitTypeRefactor, "refactor: x", BumpTypePatch},
				{CommitTypeDocs, "docs: x", BumpTypeNone},
				{CommitTypeTest, "test: x", BumpTypeNone},
				{CommitTypeChore, "chore: x", BumpTypeNone},
				{CommitTypeBuild, "build: x", BumpTypeNone},
				{CommitTypeCI, "ci: x", BumpTypeNone},
				{CommitTypeRevert, "revert: x", BumpTypeNone},
			}
			for _, c := range cases {
				got := c.ct.ToBumpType(c.msg)
				ctx.Expect(got).ToEqual(c.want)
			}
		})
	})
}

func TestDeriveNextVersion(t *testing.T) {
	specs.Describe(t, "DeriveNextVersion", func(s *specs.Spec) {
		s.It("covers derivation paths", func(ctx *specs.Context) {
			cases := []struct {
				commits []string
				lastTag string
				want    Version
				wantErr error
			}{
				{[]string{"feat: add"}, "", Version{0, 2, 0}, nil},
				{[]string{"fix: bug"}, "", Version{0, 1, 1}, nil},
				{[]string{"chore: deps"}, "", Version{}, ErrNoReleaseableChanges},
				{[]string{}, "", Version{}, ErrNoReleaseableChanges},
				{[]string{"fix: bug"}, "v1.0.0", Version{1, 0, 1}, nil},
				{[]string{"feat: add"}, "v1.0.0", Version{1, 1, 0}, nil},
				{[]string{"feat!: break"}, "v1.0.0", Version{2, 0, 0}, nil},
				{[]string{"feat: add"}, "invalid", Version{}, ErrInvalidLastTag},
				{[]string{"fix: a", "feat: b"}, "v1.0.0", Version{1, 1, 0}, nil},
				{[]string{"feat: a", "fix!: b"}, "v1.0.0", Version{2, 0, 0}, nil},
				{[]string{"merge branch", "feat: add"}, "", Version{0, 2, 0}, nil},
				{[]string{"feat: change\n\nBREAKING CHANGE: api"}, "v1.0.0", Version{2, 0, 0}, nil},
				{[]string{"fix: bug"}, "v2.1.0", Version{2, 1, 1}, nil},
				{[]string{"feat: new"}, "v0.1.0", Version{0, 2, 0}, nil},
			}
			for _, c := range cases {
				got, err := DeriveNextVersion(c.commits, c.lastTag)
				if c.wantErr != nil {
					ctx.Expect(err != nil).To(specs.BeTrue())
					ctx.Expect(errors.Is(err, c.wantErr)).To(specs.BeTrue())
					continue
				}
				ctx.Expect(err).To(specs.BeNil())
				ctx.Expect(got).ToEqual(c.want)
			}
		})
	})
}

func TestValidateVersionFormat(t *testing.T) {
	specs.Describe(t, "ValidateVersionFormat", func(s *specs.Spec) {
		s.It("validates format paths", func(ctx *specs.Context) {
			ctx.Expect(ValidateVersionFormat("v1.0.0")).To(specs.BeNil())
			ctx.Expect(ValidateVersionFormat("v0.1.2")).To(specs.BeNil())
			err := ValidateVersionFormat("1.0.0")
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, ErrInvalidVersionFormat)).To(specs.BeTrue())
			err = ValidateVersionFormat("v1.0")
			ctx.Expect(errors.Is(err, ErrInvalidVersionFormat)).To(specs.BeTrue())
			err = ValidateVersionFormat("x1.0.0")
			ctx.Expect(errors.Is(err, ErrInvalidVersionFormat)).To(specs.BeTrue())
			err = ValidateVersionFormat("v1.0.0.0")
			ctx.Expect(errors.Is(err, ErrInvalidVersionFormat)).To(specs.BeTrue())
		})
	})
}
