// Package exec provides external command execution.
//
// Architectural Layer: adapters (exec)
//
// This package implements the CommandRunner port interface using os/exec.
//
// Responsibility Boundaries:
//   - Run commands with output capture
//   - Run commands with combined output
//   - Support for working directory specification
//
// Invariants:
//   - MUST implement ports.CommandRunner interface
//   - Contains ALL os/exec usage in the system
//   - Contains NO business logic
//
// This adapter is infrastructure - it handles all external command execution
// required by the application layer through a defined port interface.
//
//nolint:revive // package name matches adapter purpose; stdlib conflict accepted
package exec
