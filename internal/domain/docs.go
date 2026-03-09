// Package domain contains the core business logic of DevForge.
//
// Architectural Layer: domain
//
// This package is the innermost layer of Hexagonal Architecture. It contains
// pure business logic with zero side effects.
//
// Responsibility Boundaries:
//   - Conventional commit validation and parsing
//   - Semantic version derivation from commit history
//   - Coverage percentage validation
//   - Govulncheck JSON output parsing (severity filtering: HIGH/CRITICAL)
//   - Error definitions for domain-level failures
//
// Invariants:
//   - NO imports from application, ports, or adapters packages
//   - NO os/exec, file IO, environment access, or logging
//   - All functions are pure (same input = same output)
//   - Version derivation depends only on commit messages and last tag
//
// This package defines the rules that govern release automation without
// any knowledge of infrastructure, CLI, or external systems.
package domain
