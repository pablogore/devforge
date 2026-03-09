package doctor

// PolicySuggestion is a suggested policy file and the rules it should contain.
type PolicySuggestion struct {
	File  string
	Rules []string
}

// GenerateSuggestions analyzes the repository at root and returns suggested policy files and rules.
// It does not modify the repository. Suggestions are derived from DetectDangerousImports,
// DetectTimeNowUsage, and DetectAdapterImports.
func GenerateSuggestions(root string) []PolicySuggestion {
	var out []PolicySuggestion

	dangerous := DetectDangerousImports(root)
	if len(dangerous) > 0 {
		rules := make([]string, 0, len(dangerous))
		for _, imp := range dangerous {
			rules = append(rules, "forbid_import: "+imp)
		}
		out = append(out, PolicySuggestion{File: "security.yaml", Rules: rules})
	}

	if DetectTimeNowUsage(root) {
		out = append(out, PolicySuggestion{
			File:  "domain.yaml",
			Rules: []string{"forbid_time_now: domain"},
		})
	}

	if DetectAdapterImports(root) {
		out = append(out, PolicySuggestion{
			File:  "architecture.yaml",
			Rules: []string{"forbid_import: internal/adapters"},
		})
	}

	return out
}
