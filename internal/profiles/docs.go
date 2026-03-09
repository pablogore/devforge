// Package profiles defines CI/CD workflow compositions for DevForge.
//
// Architectural Layer: profiles
//
// This package is the composition root of Hexagonal Architecture. It wires
// adapters to use cases and provides entry points for CLI commands.
//
// Responsibility Boundaries:
//   - go-lib profile: library release workflow
//   - go-service profile: service release workflow
//   - Dependency injection configuration
//   - Coverage threshold management per profile
//
// Invariants:
//   - MUST NOT contain business logic
//   - MUST only compose and wire use cases
//   - Profile-specific configuration only (thresholds, etc.)
//
// This layer is the entry point that connects the CLI to the internal
// application architecture.
package profiles
