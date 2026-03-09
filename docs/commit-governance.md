# Conventional Commit Governance

This repository enforces [Conventional Commits](https://www.conventionalcommits.org/) through **DevForge**. The rules below are derived from the validation and version-derivation logic in the codebase.

---

## 1. Commit Message Format

The commit subject (and thus the **PR title**, which becomes the squash-merge commit) must follow:

```
type(scope): description
```

- **type** — Required. One of the allowed types (see section 2).
- **(scope)** — Optional. A short identifier in parentheses (e.g. `api`, `auth`).
- **:** — Required. A colon followed by a space.
- **description** — Required. Short summary; must be non-empty after the colon.

### Breaking change formats

Breaking changes can be indicated in two ways:

1. **Exclamation after type/scope** (in the subject):
   ```
   type(scope)!: description
   ```
   Example: `feat(api)!: change response format`

2. **Footer in the body**:
   ```
   BREAKING CHANGE: explanation
   ```
   The validator only checks the **subject line**. Version derivation (in `internal/domain/version.go`) considers the full message and treats either `!` in the subject or the presence of `BREAKING CHANGE:` in the message as a breaking change for a **major** bump.

### Validator regex

The format is validated in `internal/domain/conventional.go` with this pattern:

```
^(feat|fix|refactor|perf|docs|test|chore|build|ci|revert)(\(.+\))?!?: .+$
```

In code (Go raw string):

```go
const ConventionalCommitPattern = `^(feat|fix|refactor|perf|docs|test|chore|build|ci|revert)(\(.+\))?!?: .+`
```

- Titles starting with `Merge ` (merge commits) are **accepted** without format validation.
- An empty title is rejected (PR title required).

---

## 2. Allowed Commit Types

The following types are allowed by the validator and recognized by version derivation:

| Type | Meaning |
|------|---------|
| **feat** | A new user-facing feature. Drives a **minor** version bump (unless breaking). |
| **fix** | A bug fix. Drives a **patch** version bump. |
| **refactor** | Code change that neither fixes a bug nor adds a feature (e.g. restructuring). **Patch** bump. |
| **perf** | Performance improvement. **Patch** bump. |
| **docs** | Documentation only (README, comments, etc.). No version change. |
| **test** | Adding or updating tests. No version change. |
| **chore** | Maintenance (dependencies, tooling, config). No version change. |
| **build** | Build system or external dependency changes. No version change. |
| **ci** | CI configuration or scripts. No version change. |
| **revert** | Reverts a previous commit. Treated as no version change by the current logic; the reverted commit’s effect is not automatically “undone” in version calculation. |

---

## 3. Version Impact (SemVer)

Version bumping is implemented in `internal/domain/version.go`. The next version is derived from **all commit messages** since the last tag; the **maximum** bump among those commits is applied once.

### Rules

| Condition | Bump | Effect |
|-----------|------|--------|
| Any commit has `!` in the subject or `BREAKING CHANGE:` in the message | **Major** | `major++`, minor and patch reset to 0 |
| At least one **feat** (no breaking) | **Minor** | `minor++`, patch reset to 0 |
| Only **fix**, **refactor**, or **perf** (no breaking) | **Patch** | `patch++` |
| Only **docs**, **test**, **chore**, **build**, **ci**, or **revert** | **None** | No releaseable change; release fails with `ErrNoReleaseableChanges` |

### Examples (from `lastTag`)

- Last tag: `v1.2.3`. Commits: `fix: correct bug` → next version **v1.2.4** (patch).
- Last tag: `v1.2.3`. Commits: `feat(auth): add OAuth` → next version **v1.3.0** (minor).
- Last tag: `v1.2.3`. Commits: `feat(api)!: change contract` → next version **v2.0.0** (major).
- Last tag: `v1.2.3`. Commits: `fix: bug`, `feat: new thing` → next version **v1.3.0** (minor wins over patch).
- Last tag: `v1.2.3`. Commits: `docs: readme`, `chore: deps` → **no new version** (release fails if attempted).
- No previous tag: commits `feat: first feature` → next version **v0.1.0** (initial version from code).

---

## 4. Repository Governance Rules

- **PR title is the source of truth** — The title of the pull request is what gets validated and what becomes the single commit on `main` after squash. It must follow Conventional Commit format.
- **Squash & merge is required** — One commit per PR; that commit’s message is the PR title. Non-linear history (e.g. merge commits that don’t match policy) is not allowed.
- **Linear history is enforced** — Branch protection and CI ensure a linear history; Conventional Commit applies to the resulting squash commit.
- **Conventional Commit validation is mandatory before release** — Governance validation (including the conventional-commit check) must pass before version derivation and release. CI blocks merge and release on failure.

---

## 5. Where Validation Happens

- **Implementation** — `internal/domain/conventional.go`:
  - `ValidateConventionalCommit(title string) error` checks the subject against `ConventionalCommitPattern`.
  - Empty title returns `ErrPRTitleRequired`; non-matching title returns `ErrInvalidConventionalCommit`.

- **When it runs**:
  - **Locally / in CI**: when you run **`forge pr`** (PR validation mode).
  - **CI step name**: **`conventional-commit`** — this step runs the validator; if it fails, the PR pipeline fails.

---

## 6. PR Flow

```
Developer opens PR
        ↓
PR title set (source of truth for the squash commit)
        ↓
forge pr runs (e.g. in CI or locally)
        ↓
Conventional Commit check (step: conventional-commit)
        ↓
Other CI validations (lint, tests, coverage, etc.)
        ↓
All pass → Squash merge to main
        ↓
Release pipeline: version derived from commit history (Conventional Commits since last tag)
```

The conventional-commit step validates the **PR title** only; after squash merge, that title is the commit message used for version derivation.

---

## 7. Quick Reference

- [ ] PR title matches: `type(scope)?!?: description`
- [ ] `type` is one of: feat, fix, refactor, perf, docs, test, chore, build, ci, revert
- [ ] Colon and space after type/scope (and optional `!`)
- [ ] Non-empty description
- [ ] Use `!` or `BREAKING CHANGE:` only when introducing a breaking change (major bump)
- [ ] Squash & merge; no direct commits to main

---

## 8. Examples

### Valid

```
feat(api): add user endpoint
fix(cache): prevent stale reads
refactor(parser): simplify logic
perf(query): add index
docs: update README
chore(deps): bump go to 1.21
feat(auth)!: remove legacy login
feat(api): change contract

BREAKING CHANGE: response schema updated
```

### Invalid

| Title | Why invalid |
|-------|-------------|
| `update code` | Missing type and colon. |
| `bug fix` | Missing type and colon. |
| `feat add endpoint` | Missing colon after type. |
| `feature: new api` | Type must be `feat`, not `feature`. |
| `feat:` | Description after colon is required (regex: `: .+`). |
| *(empty)* | PR title is required. |

The validator uses the regex in section 1; any subject that does not match (and is not a `Merge ...` title) is rejected.
