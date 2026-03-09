package guard

import (
	"testing"

	"github.com/pablogore/go-specs/specs"
)

func TestDefaultRules(t *testing.T) {
	specs.Describe(t, "DefaultRules", func(s *specs.Spec) {
		s.It("returns non-empty deterministic list", func(ctx *specs.Context) {
			rules := DefaultRules()
			ctx.Expect(len(rules) > 0).To(specs.BeTrue())
			names := make([]string, len(rules))
			for i, r := range rules {
				names[i] = r.Name()
			}
			ctx.Expect(len(names)).ToEqual(len(rules))
		})
		s.It("includes NoTimeNowInDomain and NoFmtPrint and layer rules", func(ctx *specs.Context) {
			rules := DefaultRules()
			nameSet := make(map[string]bool)
			for _, r := range rules {
				nameSet[r.Name()] = true
			}
			ctx.Expect(nameSet["NoTimeNowInDomain"]).To(specs.BeTrue())
			ctx.Expect(nameSet["NoFmtPrintOutsideCmd"]).To(specs.BeTrue())
			ctx.Expect(nameSet["NoCircularImports"]).To(specs.BeTrue())
			ctx.Expect(nameSet["DomainMustNotImportAdapters"]).To(specs.BeTrue())
			ctx.Expect(nameSet["AdaptersMustNotImportDomain"]).To(specs.BeTrue())
			ctx.Expect(nameSet["NoCrossLayerImports"]).To(specs.BeTrue())
		})
	})
}
