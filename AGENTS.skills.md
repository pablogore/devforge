# AGENTS.skills.md — DevForge

Contributors MUST understand these concepts to contribute safely.

---

## 1. Hexagonal Architecture

### Required Knowledge

- **Dependency Direction**: Infrastructure depends on interfaces, not the reverse
- **Inversion of Control**: Dependencies point inward toward domain
- **Port/Adapter Separation**: Ports define contracts, adapters implement them
- **Pure Domain Modeling**: Domain contains zero side effects

### Verification

- Can draw dependency graph
- Can identify which layer a file belongs to
- Can explain why domain cannot import ports

---

## 2. Semantic Versioning

### Required Knowledge

- **Major (X.0.0)**: Breaking changes
- **Minor (0.X.0)**: New features (backward compatible)
- **Patch (0.0.X)**: Backward compatible fixes
- **No Silent Breaking**: Breaking changes MUST bump major

### Conventional Commit Mapping

| Commit Type | Version Bump |
|-------------|--------------|
| feat | Minor |
| fix | Patch |
| perf | Patch |
| refactor | Patch (or Minor if API changes) |
| docs | None |
| test | None |
| chore | None |
| build | None |
| ci | None |
| revert | Depends on reverted commit |

---

## 3. Conventional Commits

### Required Knowledge

- **Format**: `<type>(<scope>)?<!>?:<description>`
- **Types**: feat, fix, refactor, perf, docs, test, chore, build, ci, revert
- **Breaking Changes**: Mark with `!` in type/scope or footer `BREAKING CHANGE:`
- **Scope**: Optional, describes affected component

### Examples

```
feat(auth): add OAuth2 provider
fix(api): correct response schema
docs: update README
feat(api)!: change response format
```

---

## 4. Deterministic CI

### Required Knowledge

- **Reproducible**: Same input → same output, every time
- **Idempotent**: Can run multiple times safely
- **No Hidden State**: All inputs explicit and validated
- **No Environment Drift**: Works in any CI environment

### Anti-Patterns (MUST AVOID)

- Using timestamps in version
- Reading uncommitted changes
- Assuming local git state
- Using environment variables that aren't explicit inputs

---

## 5. Governance Enforcement

### Required Knowledge

- **Architecture is ENFORCED**: Not a suggestion, not optional
- **Validation Failures are BLOCKING**: Not warnings, not suggestions
- **CI is GATEKEEPER**: Cannot be bypassed
- **Violations are POLICY BREACHES**: Not bugs to fix later

### Mindset

- "No" is the default answer to exceptions
- "It works on my machine" is not valid
- "We need to bypass for..." → NO
- "Can we skip..." → NO

---

## 6. Release Automation

### Required Knowledge

- **Tag = Release**: Tag version determines release
- **HEAD = Point**: Tag always at latest commit
- **Derivation = Truth**: Version comes from commits, not humans
- **Automation = Safety**: No manual steps possible

### Invariants

- [ ] Identical commits → identical version
- [ ] No manual tag creation possible
- [ ] No manual version bumping
- [ ] No override mechanism for version derivation

---

## 7. Testing Discipline

Contributors MUST understand:

- **Why domain tests must be pure**: Domain is the core of hexagonal architecture; impure tests would break the guarantee that domain contains zero side effects
- **Why application tests must mock ports**: Application orchestrates use cases; real infrastructure would introduce non-determinism and external dependencies
- **How to use `testify/mock` correctly**: Use mockery-generated mocks that implement port interfaces; mocks are injected, never created inside use cases
- **Why mocks must not contain business logic**: Mocks are contract implementations for testing; business logic belongs in domain
- **How to write deterministic tests**: No time-based assertions, no network calls, no file system dependencies, no random values
- **Why test isolation preserves architectural integrity**: Tests that respect layer boundaries verify that the architecture is enforceable

### Test Organization

- **Domain tests**: `*_test.go` in domain package, no external imports
- **Application tests**: `*_test.go` in application package, mocks in `mocks.go`
- **Adapter tests**: `*_test.go` in adapter package, may use real implementations for integration

### Anti-Patterns

- Domain test importing `testify/mock` → VIOLATION
- Application test using real git client → VIOLATION
- Test depending on current time → VIOLATION
- Test reading from file system → VIOLATION (unless explicitly testing adapter)

---

## 8. Cohesion and Functional Discipline

Contributors MUST understand:

### High Cohesion

- **Single Responsibility**: Each package has one job. If you cannot describe a package's purpose in one sentence, it likely has too many responsibilities
- **No Mixed Concerns**: Infrastructure code does not belong in domain. Domain logic does not belong in adapters
- **Package Naming**: Package name MUST reflect its single responsibility. `internal/utils` is a red flag

### Low Coupling

- **Dependency Direction**: Dependencies point inward. Infrastructure depends on domain, never the reverse
- **No Circular Imports**: A cannot import B if B imports A. This breaks governance enforcement
- **Ports Define Boundaries**: Adapters implement ports. Domain defines interfaces. Never skip the port

### Pure Functions

- **Why Pure Functions Increase Determinism**: Same input always produces same output. No hidden state means no unexpected behavior
- **Side Effects Are Contained**: File I/O, network calls, environment access—all isolated to adapters
- **Mutability is Minimized**: Prefer immutability. Mutable state creates non-determinism
- **Testability**: Pure functions are trivially testable. No mocking required

### Anti-Patterns

- Package named `internal/utils` → VIOLATION (no clear responsibility)
- Domain importing any adapter → VIOLATION (cross-layer import)
- Application performing os/exec → VIOLATION (side effect in orchestration)
- Global mutable state → VIOLATION (breaks determinism)
- Function that could be pure but isn't → VIOLATION (lazy impurity)

---

## 9. Test Generation Skill

When generating tests, the agent MUST:

1. Identify all branches in the function
2. Generate tests for each branch
3. Include negative scenarios
4. Simulate dependency failures
5. Verify deterministic behavior

Example expected test set:

- ✓ valid input
- ✓ empty input
- ✓ invalid input
- ✓ dependency failure
- ✓ boundary value
