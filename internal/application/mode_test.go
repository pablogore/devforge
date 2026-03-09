package application

import (
	"strings"
	"testing"

	"github.com/pablogore/go-specs/specs"
)

func TestParseMode(t *testing.T) {
	specs.Describe(t, "ParseMode", func(s *specs.Spec) {
		s.It("covers valid and invalid mode paths", func(ctx *specs.Context) {
			cases := []struct {
				input   string
				want    RunMode
				wantErr bool
			}{
				{"quick", ModeQuick, false},
				{"full", ModeFull, false},
				{"deep", ModeDeep, false},
				{"", "", true},
				{"invalid", "", true},
				{"FULL", "", true},
			}
			for _, c := range cases {
				got, err := ParseMode(c.input)
				if c.wantErr {
					ctx.Expect(err != nil).To(specs.BeTrue())
					if c.input == "" {
						ctx.Expect(strings.Contains(err.Error(), "required")).To(specs.BeTrue())
					}
					if c.input == "invalid" {
						ctx.Expect(strings.Contains(err.Error(), "invalid mode")).To(specs.BeTrue())
						ctx.Expect(strings.Contains(err.Error(), "quick, full, deep")).To(specs.BeTrue())
					}
					continue
				}
				ctx.Expect(err).To(specs.BeNil())
				ctx.Expect(got).ToEqual(c.want)
			}
		})
	})
}
