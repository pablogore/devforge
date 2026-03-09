// Package coverage resolves coverage package patterns from .syntegrity.yml policies
// and builds the -coverpkg flag for go test.
//
// Architectural Layer: application support (used by steps and use cases)
//
// Responsibility: pattern validation, glob resolution via go list ./..., and
// building the comma-separated coverpkg list. Excludes vendor, testdata,
// examples, and generated directories.
package coverage //nolint:revive // var-naming: package name describes coverage policy; stdlib conflict accepted
