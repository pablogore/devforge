package domain

import (
	"errors"
	"testing"

	"github.com/pablogore/go-specs/specs"
)

func TestValidateGovulncheckOutput(t *testing.T) {
	specs.Describe(t, "ValidateGovulncheckOutput", func(s *specs.Spec) {
		s.It("accepts empty or config-only stream", func(ctx *specs.Context) {
			ctx.Expect(ValidateGovulncheckOutput("")).To(specs.BeNil())
			ctx.Expect(ValidateGovulncheckOutput(`{"type":"config","version":"1.0"}`)).To(specs.BeNil())
		})
		s.It("accepts low severity only", func(ctx *specs.Context) {
			jsonOut := `{"type":"finding","severity":"LOW","osv":"GO-2024-1"}
{"type":"finding","Severity":"low","osv":"GO-2024-2"}`
			ctx.Expect(ValidateGovulncheckOutput(jsonOut)).To(specs.BeNil())
		})
		s.It("returns ErrGovulncheckHighOrCritical when HIGH present", func(ctx *specs.Context) {
			jsonOut := `{"type":"finding","severity":"LOW"}
{"type":"finding","severity":"HIGH","osv":"GO-2024-3"}`
			err := ValidateGovulncheckOutput(jsonOut)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, ErrGovulncheckHighOrCritical)).To(specs.BeTrue())
		})
		s.It("returns ErrGovulncheckHighOrCritical when CRITICAL present", func(ctx *specs.Context) {
			jsonOut := `{"type":"finding","Severity":"CRITICAL","osv":"GO-2024-4"}`
			err := ValidateGovulncheckOutput(jsonOut)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, ErrGovulncheckHighOrCritical)).To(specs.BeTrue())
		})
		s.It("detects Vuln nested severity", func(ctx *specs.Context) {
			jsonOut := `{"Vuln":{"database_specific":{"severity":"HIGH"},"id":"GO-2024-5"}}`
			err := ValidateGovulncheckOutput(jsonOut)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, ErrGovulncheckHighOrCritical)).To(specs.BeTrue())
		})
		s.It("ignores invalid lines", func(ctx *specs.Context) {
			jsonOut := `not json
{"severity":"MEDIUM"}`
			ctx.Expect(ValidateGovulncheckOutput(jsonOut)).To(specs.BeNil())
		})
		s.It("detects osv database_specific severity CRITICAL", func(ctx *specs.Context) {
			jsonOut := `{"osv":{"database_specific":{"severity":"CRITICAL"},"id":"GO-2024-6"}}`
			err := ValidateGovulncheckOutput(jsonOut)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, ErrGovulncheckHighOrCritical)).To(specs.BeTrue())
		})
		s.It("detects osv database_specific severity HIGH", func(ctx *specs.Context) {
			jsonOut := `{"osv":{"database_specific":{"severity":"HIGH"}}}`
			err := ValidateGovulncheckOutput(jsonOut)
			ctx.Expect(err != nil).To(specs.BeTrue())
			ctx.Expect(errors.Is(err, ErrGovulncheckHighOrCritical)).To(specs.BeTrue())
		})
		s.It("accepts Vuln nested without severity", func(ctx *specs.Context) {
			jsonOut := `{"Vuln":{"database_specific":{},"nested":{"id":"x"}}}`
			ctx.Expect(ValidateGovulncheckOutput(jsonOut)).To(specs.BeNil())
		})
		s.It("accepts Vuln database_specific empty severity", func(ctx *specs.Context) {
			jsonOut := `{"Vuln":{"database_specific":{"severity":""},"id":"GO-2024-7"}}`
			ctx.Expect(ValidateGovulncheckOutput(jsonOut)).To(specs.BeNil())
		})
	})
}
