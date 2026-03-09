# AGENTS.validation.md — DevForge

Validation is MANDATORY. All checks MUST pass before release.

---

## Structural Validation

Before any build, validate:

- [ ] Directory structure matches hexagonal model
- [ ] `domain/` contains NO imports from `application/`, `ports/`, or `adapters/`
- [ ] `application/` contains NO imports from `adapters/`
- [ ] NO cyclic dependencies exist
- [ ] All ports have corresponding adapter implementations
- [ ] Profiles only compose, never implement business logic
- [ ] CLI contains no business logic
- [ ] Every production package contains `docs.go`
- [ ] docs.go contains no executable code
- [ ] mocks.go (if present) contains only mock definitions
- [ ] No circular dependencies between packages
- [ ] No cross-layer import violations
- [ ] Domain contains only pure logic (no os/exec, no file IO, no env access)
- [ ] No shared mutable global state detected

---

## Determinism Validation

Every release MUST validate:

- [ ] Version derived from committed history only
- [ ] NO timestamps used in version calculation
- [ ] NO environment variables affect version (except explicit CI inputs)
- [ ] Version calculation is idempotent (re-runnable with same result)
- [ ] Identical commit history produces identical version
- [ ] CI environment MUST confirm full git history is available
- [ ] Tag resolution MUST be verified before version derivation
- [ ] If history is incomplete, RELEASE MUST FAIL

---

## Commit Validation

Every PR MUST validate:

- [ ] PR title matches Conventional Commit regex:
  ```
  ^(feat|fix|refactor|perf|docs|test|chore|build|ci|revert)(\(.+\))?!?: .+
  ```
- [ ] PR title is present (not empty)
- [ ] Commit type is valid (feat, fix, refactor, perf, docs, test, chore, build, ci, revert)
- [ ] Breaking change marker (`!`) used correctly (if applicable)

---

## Coverage Validation

Every test run MUST validate:

- [ ] Coverage output parsed successfully
- [ ] go-lib profile: coverage >= 90%
- [ ] go-service profile: coverage >= 80%
- [ ] Coverage file generated (coverage.out exists)
- [ ] Unit tests exist for domain logic
- [ ] Application layer isolates ports (mocks or fakes)
- [ ] No domain tests import testify/mock
- [ ] Coverage thresholds respected

---

## Test Coverage Validation

A pull request MUST fail validation if:

- [ ] New code paths lack tests
- [ ] Error returns are not tested
- [ ] Only happy path tests exist

Recommended metrics:

- Branch coverage >= 90%
- Every error path exercised

---

## Pre-Release Validation

Before tag creation, MUST validate:

- [ ] Working tree is clean (`git status --porcelain` empty)
- [ ] Target is main branch (`git branch --show-current` == "main")
- [ ] Tag does NOT already exist (`git tag -l <version>` empty)
- [ ] All quality validations passed (lint, vet, test, coverage)
- [ ] All governance validations passed (Conventional Commit)
- [ ] Version can be derived deterministically

---

## Release Validation

After artifact build, MUST validate:

- [ ] Tag created at HEAD (tagged commit == HEAD)
- [ ] HEAD commit hash MUST equal tag target commit hash
- [ ] Tag version matches derived version
- [ ] GoReleaser executed successfully
- [ ] Artifacts published to registry
- [ ] Tag was created by CI (not manually)

---

## Invalid States (MUST BLOCK)

| State | Action |
|-------|--------|
| Domain imports infra | BUILD FAIL |
| Application imports adapters | BUILD FAIL |
| Non-squash merge attempted | MERGE BLOCKED |
| Invalid PR title | MERGE BLOCKED |
| Coverage below threshold | RELEASE BLOCKED |
| Working tree dirty | RELEASE BLOCKED |
| Tag already exists | RELEASE BLOCKED |
| Not on main branch | RELEASE BLOCKED |
| govulncheck HIGH/CRITICAL | RELEASE BLOCKED |
| Version not derivable | RELEASE BLOCKED |
| Non-linear history detected | MERGE BLOCKED |
| Force push to main attempted | MERGE BLOCKED |
| Incomplete git history detected | RELEASE BLOCKED |
| New code paths without tests | MERGE BLOCKED |
| Error paths not tested | MERGE BLOCKED |
| Only happy path tests (no branch/error coverage) | MERGE BLOCKED |
