// Package guard defines the compile-time extensible Architectural Guard system.
//
// Architectural Layer: boundary (between application and rule implementations)
//
// This package provides contracts only: the ArchitecturalRule interface and
// Context for validation. It does not import adapters; only ports are used.
// Rule implementations may live in application or other packages and are
// composed at compile time.
//
// Responsibility Boundaries:
//   - Define ArchitecturalRule interface
//   - Define Context required for rule validation
//   - No concrete rule implementations in this package (use default_rules or inject)
//
// Invariants:
//   - MUST NOT import internal/adapters
//   - Context and rules use only ports for external dependencies
//   - Execution order of rules is deterministic (slice order)
package guard
