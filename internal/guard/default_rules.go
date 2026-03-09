package guard

// DefaultRules returns the default set of architectural rules for validation.
// Order is deterministic; rules run in slice order.
func DefaultRules() []ArchitecturalRule {
	return []ArchitecturalRule{
		NewNoTimeNowInDomainRule(),
		NewNoFmtPrintOutsideCmdRule(),
		NewNoCircularImportsRule(),
		NewDomainMustNotImportAdaptersRule(),
		NewAdaptersMustNotImportDomainRule(),
		NewNoCrossLayerImportRule(),
	}
}
