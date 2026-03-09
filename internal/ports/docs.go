// Package ports defines interfaces that abstract external dependencies.
//
// Architectural Layer: ports
//
// This package is the boundary layer of Hexagonal Architecture. It defines
// contracts (interfaces) that adapters must implement.
//
// Responsibility Boundaries:
//   - GitClient: Git repository operations
//   - CommandRunner: External command execution
//   - EnvProvider: Environment variable access
//
// Invariants:
//   - Contains NO implementation code
//   - Contains NO business logic
//   - Interfaces defined here must be implementable by adapters
//
// Ports are the "driver" side of hexagonal architecture - they define
// what the application layer needs to function.
package ports
