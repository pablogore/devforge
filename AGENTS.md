# AGENTS.md — DevForge

DevForge is the operational CI standard and policy engine for open source Go projects.

It enforces:
- Deterministic Release Automation
- Semantic Version Derivation
- Conventional Commits
- Governance Validation
- Quality Thresholds
- Hexagonal Boundary Enforcement

This repository is STRICTER than any other repository in the ecosystem.

---

## 1. Policy Engine

DevForge is a policy engine, NOT a business service.

As a policy engine, it:
- Defines CI/CD governance rules
- Enforces architectural constraints
- Derives versions from commit semantics
- Blocks violations definitively

---

## 2. Hexagonal Architecture

DevForge strictly follows Hexagonal Architecture:

| Layer | Responsibility |
|-------|----------------|
| `domain/` | Pure logic, no IO, no side effects |
| `application/` | Use case orchestration via ports |
| `ports/` | Interfaces defining boundaries |
| `adapters/` | Infrastructure implementations |
| `profiles/` | Composition root for CI flows |
| `cmd/` | CLI adapter (presentation only) |

### Dependency Rules (MANDATORY)

- **Domain**: MUST NOT import application, ports, or adapters
- **Application**: MUST NOT import adapters
- **Adapters**: MUST implement ports interfaces
- **Profiles**: Wire adapters to use cases only
- **CLI**: Instantiates composition root only

---

## 3. Deterministic Release

Release automation is STRICTLY DETERMINISTIC and IDEMPOTENT.

### Determinism Guarantees (ENFORCED)

| Guarantee | Requirement |
|-----------|-------------|
| **Reproducible** | Identical commit history produces identical version |
| **Idempotent** | Version calculation can be run N times with same result |
| **No Hidden Assumptions** | All inputs are explicit and validated |
| **No Environment Drift** | Version derived from commit content only |
| **No Timestamps** | Version never includes time-based components |
| **No Randomness** | Version derivation is pure function of commit history |

### Inputs (EXPLICIT ONLY)

Version derivation MUST use only:
- Git commit history (parent chain)
- Commit messages (Conventional Commit format)
- Previous tag (if exists)

Version derivation MUST NOT use:
- Current time/date
- Environment variables (except explicit CI inputs)
- Uncommitted changes
- Local git state

### Git History Requirement

- CI MUST fetch full commit history required for Semantic Version Derivation.
- Shallow clones (fetch-depth: 1) are NOT permitted for release.
- Semantic Version Derivation requires access to previous tags.
- CI configuration MUST ensure complete tag history is available.
- Version derivation behavior MUST NOT depend on incomplete git history.

### Release Pipeline

On merge to main, the pipeline MUST execute in strict order:

    1. **Quality Validation** → lint, vet, test, coverage (enforced by DevForge according to the active repository profile)
2. **Governance Validation** → Conventional Commit, architecture
3. **Version Derivation** → Calculate next semantic version (PURE FUNCTION)
4. **Tag Creation** → Create git tag at HEAD
5. **Artifact Release** → Execute GoReleaser

Each step MUST pass before proceeding. No skipping allowed.

---

## 4. Release Invariants (STRICT)

These invariants are MANDATORY and ENFORCED:

| Invariant | Enforcement |
|-----------|-------------|
| Tag MUST point to HEAD | CI verifies tag commit == HEAD |
| Tag MUST NOT already exist | CI fails if tag exists |
| Version MUST match delta | Derived from Conventional Commits since last tag |
| Version MUST be derivable | Release FAILS if version cannot be computed |
| No manual tagging | Only CI creates tags (via goreleaser) |
| No force push to main | Branch protection enforced |
| Working tree MUST be clean | No uncommitted changes allowed |
| Release ONLY on main branch | No release from branches |

---

## 5. Git Discipline (STRICT)

| Rule | Enforcement |
|------|-------------|
| Squash & merge only | Non-linear history MUST fail validation; enforced via branch protection + CI |
| PR title defines version | Conventional Commit validated |
| Direct commits to main | Branch protection prevents |
| Linear history required | Non-linear history MUST fail validation |
| PR title is source of truth | Not commit messages |

---

## 6. Governance

This repository enforces strict governance:

| Rule | Enforcement |
|------|-------------|
| Squash & merge only | Branch protection + CI validation; non-linear history fails |
| Conventional Commit in PR title | Validated before release |
| Release only via merge to main | Branch protection enforced |
| No manual tagging | Tags created by CI only |
| No manual release commits | Semantic Version Derivation from commits |
| No force push to main | Branch protection enforced |

---

## 7. Architectural Quality Principles

DevForge enforces quality through explicit architectural constraints:

### High Cohesion

- Each package MUST have a single, clearly defined responsibility.
- Each file MUST belong clearly to one architectural layer.
- No mixed concerns within the same package.
- A package's name MUST reflect its single responsibility.

### Low Coupling

- Packages MUST depend only on required abstractions.
- No circular dependencies.
- No cross-layer shortcuts.
- No hidden dependencies.
- Dependencies MUST point inward toward domain.

### Pure Functions First

- Pure functions are first-class citizens.
- Business logic MUST prefer pure functions.
- Side effects MUST be isolated in adapters.
- Application orchestration MUST minimize mutation.
- Any function that CAN be pure MUST be pure.

---

## 8. Scope

DevForge governs:

- **go-lib** → Go libraries
- **go-service** → Go services

Future profiles MUST follow the same architectural rules.

---

## 8. Code Standards

### Documentation (docs.go)

Every production package (non-test, non-mocks-only package) MUST contain a `docs.go`.

**EXEMPT**:
- Packages containing only `_test.go` files
- Dedicated mock-only packages

`docs.go` MUST:
- Describe package purpose
- State architectural layer (domain, application, adapters, etc.)
- Define responsibility boundaries
- Declare invariants if applicable

`docs.go` MUST NOT contain executable code.

### GoDoc for Exported Symbols

All exported symbols (types, functions, methods, constants, variables) MUST include GoDoc comments.

| Rule | Scope |
|------|--------|
| **Exported = documented** | Every exported symbol MUST have a comment starting with its name |
| **Applies everywhere** | Required in domain, application, ports, adapters, profiles, and cmd |
| **No exceptions** | Internal packages and adapters are NOT exempt |

GoDoc comments MUST:
- Start with the symbol name (e.g. `// Foo does ...` for type or function `Foo`)
- Describe purpose and, when relevant, parameters and return values
- Be complete sentences where appropriate

### Testing Overview

Testing is part of governance enforcement and MUST respect hexagonal boundaries.

Coverage validation is enforced by DevForge based on the active repository profile. AGENTS documentation defines testing practices only and does not specify numeric coverage thresholds.

**If the coverage step fails**: Add tests per AGENTS.rules.md (Test-Driven Development, Branch Coverage Requirements, Error Path Coverage, Adversarial Testing). Every logical branch and every error return must have a test; happy-path-only tests are insufficient.

### Coverage Enforcement

Testing practices are defined in AGENTS.rules.md.

Coverage thresholds and enforcement policies are implemented by the DevForge tool according to the repository profile.

This ensures that governance documentation remains descriptive while enforcement remains deterministic in the CI tool.

---

## 9. Violation Handling (CRITICAL)

Any violation of architectural or release invariants MUST block CI.

CI failures are INTENTIONAL GOVERNANCE MECHANISMS, not errors to bypass.

| Violation Type | Action |
|----------------|--------|
| Architectural boundary breach | BUILD FAIL |
| Git discipline violation | MERGE BLOCKED |
| Release invariant violation | TAG/RELEASE FAIL |
| Quality threshold breach | RELEASE BLOCKED |

---

## 10. Terminology

| Term | Definition |
|------|------------|
| **Deterministic Release** | Same commit history always produces identical version |
| **Semantic Version Derivation** | Version bump computed from Conventional Commit types |
| **Hexagonal Boundary** | Strict layer separation defined by ports |
| **Policy Engine** | System that enforces rules, not implements business logic |
| **Governance Validation** | Checks that ensure compliance with project standards |
| **Pure Function** | Function with no side effects, same input = same output |
| **Idempotent** | Can be applied multiple times with same result |

---

## See Also

- [AGENTS.rules.md](./AGENTS.rules.md) — Strict mandatory rules
- [AGENTS.validation.md](./AGENTS.validation.md) — Validation requirements
- [AGENTS.skills.md](./AGENTS.skills.md) — Required contributor knowledge

These four documents together form the complete governance contract.
