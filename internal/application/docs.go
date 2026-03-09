// Package application orchestrates use cases for DevForge.
//
// Architectural Layer: application
//
// This package coordinates domain logic through ports. It contains no business
// logic itself but orchestrates the flow of data and commands.
//
// Responsibility Boundaries:
//   - PR validation workflow orchestration (Step-based)
//   - Release workflow orchestration (Step-based)
//   - Coordination between domain and adapters via ports
//   - Precondition validation (branch, working tree, history)
//   - Step interface and Context for deterministic step execution
//
// Invariants:
//   - MUST import only domain and ports packages
//   - NO os/exec imports
//   - NO direct infrastructure usage
//   - All external dependencies injected via ports
//
// This layer implements use cases defined in the domain layer without
// knowledge of how ports are implemented.
package application
