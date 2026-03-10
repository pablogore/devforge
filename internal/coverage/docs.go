// Package coverage resolves coverage package patterns from .devforge.yml policies
// and builds the -coverpkg flag for go test.
//
// Architectural Layer: application support (used by steps and use cases)
//
// Responsibility: pattern validation, glob resolution via go list ./..., and
// building the comma-separated coverpkg list. Excludes vendor, testdata,
// examples, and generated directories. Applies default exclusions (testkit,
// fixtures, fake, spy) and optional policies.coverage.exclude patterns so
// DevForge and specs runner produce identical coverage numbers.
package coverage //nolint:revive // var-naming: package name describes coverage policy; stdlib conflict accepted
