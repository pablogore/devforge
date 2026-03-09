# AGENTS.rules.md — DevForge

These rules are MANDATORY and ENFORCED. Violations block CI.

---

## Architecture Rules

### Layer Boundaries (STRICT)

1. **domain layer**:
   - NO os/exec
   - NO file IO
   - NO environment access
   - NO logging
   - NO side effects
   - Pure functions only

2. **application layer**:
   - MUST depend only on domain + ports
   - NO direct infrastructure usage
   - NO os/exec imports
   - Orchestrates only

3. **adapters layer**:
   - MUST implement ports interfaces
   - Contains ALL system interaction
   - No business logic

4. **profiles**:
   - MUST NOT implement business logic
   - Only compose and wire use cases
   - Dependency injection only

5. **CLI**:
   - NO business logic
   - Only instantiates composition root
   - Presentation layer only

---

## Cohesion and Coupling Rules (STRICT)

- **Single Responsibility** — Each package MUST have one clearly defined responsibility
- **No Mixed Concerns** — A package MUST NOT mix domain logic and infrastructure
- **Cross-layer imports FORBIDDEN** — Domain cannot import application, application cannot import adapters
- **Pure by Default** — Functions that can be pure MUST be pure
- **Side Effects Isolated** — All side effects MUST be in adapters
- **No Shared Mutable State** — Global mutable state is FORBIDDEN
- **No Circular Dependencies** — Dependency graph must be acyclic
- **Dependencies Point Inward** — Infrastructure depends on interfaces, never the reverse

---

## Determinism Rules (STRICT)

- **NO timestamps in version** — Version is purely commit-derived
- **NO environment-dependent inputs** — Only explicit CI inputs allowed
- **NO uncommitted changes** — Version derived from committed history only
- **NO randomness** — Every computation is pure function
- **NO local git state assumptions** — Only parent chain and tags
- **Version calculation MUST be idempotent** — Running release process multiple times on same HEAD produces same result
- **CI MUST operate on full git history** — Shallow clones that prevent correct tag resolution are FORBIDDEN

---

## Git Discipline (STRICT)

- **Squash & merge ONLY** — No merge commits allowed
- **PR title defines version** — Conventional Commit format mandatory
- **Direct commits to main FORBIDDEN** — Blocked by branch protection
- **Linear history REQUIRED** — No merge commits
- **Force push to main FORBIDDEN** — Blocked by branch protection
- **PR title is source of truth** — Not commit messages

---

## Release Invariants (STRICT)

- **Tag MUST point to HEAD** — CI verifies tag commit == HEAD
- **Tag MUST NOT already exist** — Prevents duplicate releases
- **Version MUST match delta** — Derived from commits since last tag
- **Version MUST be derivable** — Release FAILS if undeterminable
- **No manual tagging** — Only CI creates tags
- **No force push to main** — Branch protection required
- **Working tree MUST be clean** — No uncommitted changes
- **Release ONLY on main branch** — No branch releases

---

## Quality Rules (STRICT)

Unit tests MUST cover:

- happy path
- error paths
- boundary conditions
- adversarial inputs
- dependency failures

Coverage thresholds are enforced by DevForge policies (according to the repository profile). AGENTS.rules.md does not specify numeric coverage percentages.

Additional requirements:
- **Lint MUST pass** — No warnings treated as errors
- **govulncheck MUST pass** — No HIGH or CRITICAL vulnerabilities
- **go mod tidy MUST pass** — No stale dependencies

---

## Documentation Rules (STRICT)

- **All exported symbols MUST have GoDoc comments** — Types, functions, methods, constants, and variables that are exported (capitalized) MUST include a comment starting with the symbol name.
- **Applies to all layers** — Required in domain, application, ports, adapters, profiles, and cmd. Internal packages and adapters are NOT exempt.
- **Package docs** — Every production package (non-test, non-mocks-only) MUST contain a `docs.go`; see AGENTS.md § Code Standards.

---

## Testing Rules (STRICT)

### General Rules

- Every exported behavior MUST have unit test coverage.
- Coverage thresholds remain enforced per profile.
- Tests MUST be deterministic.
- Tests MUST NOT depend on external state.

### Domain Layer Testing

- Domain tests MUST be pure.
- Domain tests MUST NOT use mocks.
- Domain tests MUST NOT perform IO.

### Application Layer Testing

- Application tests MUST isolate ports (via mocks or fakes).
- No real infrastructure allowed in application tests.
- Prefer in-memory fakes over testify/mock when behavior verification is not essential.

### Adapter Testing

- Adapters MAY use integration-style tests.
- External systems MUST be mocked unless explicitly testing integration.
- Adapter tests MUST remain deterministic.

### Testing Framework and Doubles

- Tests SHOULD use `github.com/pablogore/go-specs` (Describe / It / Context, ctx.Expect(...).ToEqual / To(BeNil()) where available).
- Prefer **fakes** (in-memory implementations of ports) over **mocks** (testify/mock, gomock) to reduce boilerplate and improve readability.
- Use mocks only when:
  - External systems are involved and interaction must be verified, or
  - Behavior verification (call count, arguments) is essential.
- If a package uses testify mocks, mocks MAY live in `internal/mocks` and implement ports.
- Domain layer MUST NOT use mocks; use pure tests only.

---

## Test-Driven Development (MANDATORY)

All development performed by agents MUST follow TDD.

Required order:

1. Write failing tests
2. Implement minimal code
3. Refactor while tests remain green

Agents MUST NOT implement production code before tests exist.

---

## Branch Coverage Requirements

Unit tests MUST cover all logical branches of the code.

For every function the following scenarios MUST be tested:

- happy path
- invalid input
- empty input
- boundary conditions
- dependency failure
- error propagation

Example conditions that MUST have tests:

- `if err != nil`
- `if value == ""`
- `if value > limit`
- `switch` cases
- `default` branches

---

## Error Path Coverage

Every code path returning an error MUST have a corresponding test.

Example:

```go
if err != nil {
    return fmt.Errorf(...)
}
```

A unit test MUST trigger this condition.

---

## Adversarial Testing

Tests MUST attempt to break the application.

Required destructive scenarios:

- invalid parameters
- empty values
- nil values
- corrupted input
- dependency failure
- boundary values

The system MUST fail safely.

---

## Layer-Specific Testing Rules

Testing rules aligned with architecture layers:

**DOMAIN**

- Pure unit tests only
- No mocks allowed
- Deterministic tests only

**APPLICATION**

- Ports MUST be mocked
- Dependency failures MUST be simulated

**ADAPTERS**

- Integration tests allowed
- Infrastructure failures MUST be tested

---

## Violation Handling

Any violation of above rules MUST block CI:

| Violation Type | Action |
|----------------|--------|
| Architectural boundary breach | BUILD FAIL |
| Git discipline violation | MERGE BLOCKED |
| Release invariant violation | TAG/RELEASE FAIL |
| Quality threshold violation | RELEASE BLOCKED |

CI failures are GOVERNANCE MECHANISMS, not bugs to work around.

---

## Terminology

| Term | Definition |
|------|------------|
| **Deterministic Release** | Same commit history always produces identical version |
| **Semantic Version Derivation** | Version bump computed from Conventional Commit types |
| **Idempotent Version Calculation** | Running release process multiple times on same HEAD produces same result |
| **Hexagonal Boundary** | Strict layer separation defined by ports |
