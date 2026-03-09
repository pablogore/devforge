// Package env provides environment variable access as defined by AGENTS.md.
//
// Architectural Layer: adapters (Hexagonal Architecture — AGENTS.md §2)
//
// This package is an infrastructure adapter. It implements the ports.EnvProvider
// interface so that environment access is isolated behind a port; application
// and domain layers never call os.Getenv directly.
//
// Responsibility Boundaries (AGENTS.md §7 — High Cohesion, Low Coupling):
//   - Single responsibility: read environment variables by key.
//   - All environment variable access in the policy engine flows through this adapter.
//   - Used only for explicit CI inputs; version derivation MUST NOT rely on env
//     except where AGENTS.md permits "explicit CI inputs" (§3 Deterministic Release).
//
// Invariants:
//   - MUST implement ports.EnvProvider and nothing else.
//   - MUST contain NO business logic; side effects (os.Getenv) are isolated here.
//   - Dependency direction: adapters depend on ports only; no imports from domain/application.
//
// See: AGENTS.md (Hexagonal Architecture, Code Standards docs.go).
package env
