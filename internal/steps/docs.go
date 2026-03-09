// Package steps provides concrete implementations of the application Step interface.
//
// Architectural Layer: application (step implementations)
//
// This package contains all CI step types (golangci-lint, govulncheck, PR/release/doctor
// step builders, parallel/timeout groups, etc.). Each step implements
// application.Step and is registered with the application step registry in init().
//
// Responsibility Boundaries:
//   - Implement application.Step for each concrete step type
//   - Provide exported constructors (e.g. GolangCILintStep, ReleaseSteps)
//   - No use-case orchestration; steps are composed by profiles and application
//
// Invariants:
//   - MUST depend only on application (Step, Context, RegisterStep) and other
//     allowed internal packages (guard, mocks only in tests)
//   - Step logic and behavior MUST NOT depend on global state or environment
//     except as provided via application.Context
package steps
