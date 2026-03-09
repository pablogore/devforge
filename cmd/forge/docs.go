// Package main is the command-line entry point for DevForge (forge).
//
// Architectural Layer: cmd (CLI adapter)
//
// This package is the presentation layer of Hexagonal Architecture. It
// provides the CLI interface for invoking profiles.
//
// Responsibility Boundaries:
//   - Command-line argument parsing
//   - Profile selection and invocation
//   - Error handling and exit codes
//
// Invariants:
//   - MUST NOT contain business logic
//   - MUST only instantiate composition root
//   - Pure presentation layer
//
// This package is the outermost layer - it translates user commands into
// profile executions without knowing implementation details.
package main
