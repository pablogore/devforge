# DevForge

A CI policy engine and deterministic pipeline runner for Go repositories. It enforces repository governance, release rules, and quality thresholds through profile-based pipelines.

---

## 1. Project Overview

**DevForge** is a **CI policy engine**: it defines and runs validation and release workflows in a deterministic way. It is not a generic CI platform—it encodes governance and release rules so that the same commit history and repository state always produce the same outcomes.

**What it does:**

- **Deterministic pipeline runner** — Pipelines are ordered sequences of steps. Execution is reproducible; version derivation and validation depend only on explicit inputs (e.g. commit history, working tree), not timestamps or environment drift.
- **Repository governance** — Enforces conventions (e.g. conventional commits), architectural boundaries, static analysis, and coverage thresholds before merge or release.

**Key capabilities:**

| Capability | Description |
|------------|-------------|
| **Conventional commit validation** | Validates PR titles and commit messages against a conventional-commit pattern. |
| **Semantic version derivation** | Derives the next version from commit history since the last tag; no manual version bumps. |
| **Architecture guard rules** | Configurable rules (e.g. no cross-layer imports, no `time.Now` in domain) validated during PR. |
| **Static analysis** | Runs golangci-lint and govulncheck as pipeline steps; repositories can add custom plugins via `.devforge.yml`. |
| **Profile-based pipelines** | Different profiles (e.g. go-lib, go-service) define different step sets and thresholds. |
| **Deterministic releases** | Release flow: preconditions → version derivation → tag creation → verification → goreleaser; idempotent and reproducible. |

---

## 2. Key Concepts

| Concept | Description |
|--------|-------------|
| **Profile** | A named CI “flavor” (e.g. `go-lib`, `go-service`) that provides entry points for PR, release, and doctor. Profiles register pipelines and wire adapters to use cases. |
| **Pipeline** | A named, ordered list of steps (e.g. PR validation, release, doctor). Use cases run a pipeline by executing each step in sequence via a StepRunner. |
| **Step** | A single check or operation (e.g. golangci-lint, govulncheck, version-derivation). Steps implement the `Step` interface: `Name() string`, `Run(ctx *Context) error`. |
| **StepRunner** | Executes one step and logs duration and outcome. Used by use cases to run each step in a pipeline; does not change step behavior. |
| **Pipeline Registry** | A map of pipeline name → pipeline. Profiles register pipelines at init; use cases (or helpers) retrieve them by name (e.g. `GetPipeline("pr")`, `GetPipeline("release")`). |
| **Step Registry** | A map of step name → constructor. Each step implementation registers itself in `init()`; the CLI and `RunSteps` resolve steps by name (e.g. for `forge run gofmt staticcheck`). |

**How they interact:** The CLI selects a profile and invokes the profile’s PR, release, or doctor entry point. The profile obtains a pipeline (from the pipeline registry or by building it) and passes it to an application use case. The use case runs the pipeline’s steps in order using a StepRunner. Steps receive a shared `Context` (workdir, command runner, git client, logger, etc.) and do not depend on concrete infrastructure—only on ports exposed via that context.

---

## 3. Basic Usage

**Validate a pull request (profile auto-detected or explicit):**

```bash
forge pr
forge pr --profile go-lib
forge pr --profile go-lib --mode full
forge pr --profile go-lib --workdir . --base-ref origin/main --title "feat: add feature"
```

**Run the release pipeline:**

```bash
forge release
forge release --profile go-lib
forge release --profile go-service --workdir .
```

**Run environment and repo checks (doctor):**

```bash
forge doctor
forge doctor --profile go-lib
```

**Run individual steps by name (no pipeline):**

```bash
forge run gofmt staticcheck
forge run --workdir /path/to/repo golangci-lint govulncheck
```

Use `forge <command> --help` for flags and `forge run --help` for the list of available step names.

**Configuration priority:** For `pr`, `release`, and `doctor`, profile and mode are resolved in order: **CLI flags** override **`.devforge.yml`**, then **auto-detection** (profile) or **default** (mode: full).

---

## Installation

Releases provide precompiled binaries for **linux** (amd64, arm64), **darwin** (amd64, arm64), and **windows** (amd64). Each release includes raw binaries and a `checksums.txt` file for verification.

### Quick install

```bash
curl -sSL https://raw.githubusercontent.com/devforge/devforge/main/scripts/install.sh | bash
```

This detects your OS and architecture, downloads the matching binary, and installs it to `/usr/local/bin/forge` (or `~/bin` on Windows).

### Manual install

**Linux (example: amd64):**

```bash
curl -L https://github.com/devforge/devforge/releases/latest/download/forge-linux-amd64 \
  -o forge
chmod +x forge
sudo mv forge /usr/local/bin/
```

For other platforms use the same pattern with the appropriate artifact: `forge-linux-arm64`, `forge-darwin-amd64`, `forge-darwin-arm64`, `forge-windows-amd64.exe`. Verify downloads using `checksums.txt` from the [release assets](https://github.com/devforge/devforge/releases).

---

## 4. Repository Configuration

Repositories can customize DevForge using a **`.devforge.yml`** file in the repository root. **If the file does not exist, DevForge falls back to default behavior** (profile auto-detection, mode `full`, no plugins). The file is optional.

| Field | Purpose |
|-------|---------|
| **profile** | Selects the pipeline profile (for example `go-lib` or `go-service`). |
| **mode** | Selects pipeline intensity for PR validation: `quick`, `full`, or `deep`. |
| **pipeline** | Optional. Control which steps run: `disable` (list of step names to remove) or `enable` (whitelist of step names). If both are set, `enable` takes precedence. See [Pipeline customization](#pipeline-customization) below. |
| **plugins** | Defines additional validation steps executed after the main pipeline (e.g. license checks, proto lint). Each plugin has a `name` (for logging) and a `run` command (executed via `bash -c`). |

**Example `.devforge.yml`:**

```yaml
profile: go-service
mode: full

plugins:
  - name: proto-lint
    run: buf lint

  - name: license-check
    run: go-licenses check ./...
```

### Pipeline customization

Repositories can control which steps run using the **`pipeline`** key in `.devforge.yml`. If no pipeline section is provided, the default pipeline runs unchanged.

**Disable specific steps:**

```yaml
pipeline:
  disable:
    - coverage
    - govulncheck
```

**Whitelist mode (only listed steps run):**

```yaml
pipeline:
  enable:
    - go-mod-tidy
    - conventional-commit
    - architectural-guard
    - static-analysis
    - govulncheck
    - test
```

Step names are the internal names used by DevForge (e.g. `go-mod-tidy`, `conventional-commit`, `architectural-guard`, `policy-pack`, `static-analysis`, `govulncheck`, `test`). If both `enable` and `disable` are set, `enable` takes precedence (whitelist).

CLI flags (`--profile`, `--mode`) take precedence over values in `.devforge.yml`.

---

## Pipeline modes

The pipeline can run in different modes for PR validation.

| Mode   | Description |
|--------|-------------|
| **quick** | Lighter checks; faster feedback. |
| **full**  | Default; full step set and thresholds. |
| **deep**  | Extended checks; more thorough. |

You can set the mode in `.devforge.yml`:

```yaml
profile: go-lib
mode: quick
```

CLI flags override configuration. To force a mode from the command line:

```bash
forge pr --mode deep
```

If `.devforge.yml` does not define `mode`, the default is `full`. Existing behavior is unchanged when mode is omitted.

---

## 5. Profiles

Profiles define which pipelines and thresholds apply to a repository. Profile can be set via `--profile`, `.devforge.yml`, or auto-detection.

| Profile | Represents | Typical use |
|---------|------------|-------------|
| **go-lib** | Go library or module without a `cmd/` entrypoint | Shared libraries, SDKs. PR: 94% coverage, complexity threshold 15, 2m static-analysis timeout. |
| **go-service** | Go application with a `cmd/` directory | Services, CLIs. PR: 80% coverage, complexity threshold 20, 3m static-analysis timeout. |

**Automatic detection:** If `--profile` is not set, DevForge infers the profile from the repository at `--workdir`:

- **go-service** — `go.mod` exists **and** a top-level `cmd/` directory exists.
- **go-lib** — Otherwise (e.g. `go.mod` only, or no `go.mod`; default is go-lib).

---

## 6. Pipelines

Pipelines are ordered sets of steps. The same logical pipeline (e.g. “PR”, “release”, “doctor”) can have different step lists or thresholds per profile.

| Pipeline | Purpose |
|----------|---------|
| **PR validation** | Runs before merge: go mod tidy, conventional commit, architectural guard, policy-pack, static analysis (golangci-lint), security (govulncheck), tests, coverage; then any plugins from `.devforge.yml`. Fails on first step error. |
| **Release** | Preconditions (branch, clean tree, full history) → version derivation → tag creation → tag verification → goreleaser. Deterministic; version is derived from commits since last tag. |
| **Doctor** | Environment and repo checks: git installed, goreleaser installed, full history, on main, working tree clean, tags accessible. Reports pass/fail per check. |

The **profile** determines which pipeline (and which step list) is executed for `pr`, `release`, and `doctor`. Pipelines are registered by profiles and can be retrieved by name from the pipeline registry.

---

## forge doctor

The `doctor` command analyzes the repository and suggests useful policy rules. It does **not** modify any files; it only prints recommendations.

**Example:**

```bash
forge doctor
```

**Example output:**

```
Doctor check results:
  [PASS] git installed: ...
  [PASS] goreleaser installed: ...
  ...

Suggested policies:

security.yaml
  forbid_import: net/http/pprof

domain.yaml
  forbid_time_now: domain

architecture.yaml
  forbid_import: internal/adapters
```

These suggestions can be added manually under `.devforge/policies/` (e.g. create `security.yaml`, `domain.yaml`, or `architecture.yaml` with the indicated rules). If no issues are detected, doctor prints: **No policy suggestions detected.**

### Automatic Policy Generation

You can generate policy packs automatically from the same analysis:

```bash
forge doctor --generate-policies
```

This command analyzes the repository and writes policy files under **`.devforge/policies/`** (e.g. `architecture.yaml`, `security.yaml`). If no issues are detected, it prints **No policies needed.** and does not create files. Generated policies are enforced by the DevForge PR pipeline (policy-pack step).

**Example workflow:**

```bash
forge doctor --generate-policies
git add .devforge/policies
git commit -m "chore: add generated policy packs"
```

---

## 7. Steps

Steps are individual checks or operations. Each step has a **name** (used in the step registry and in pipelines) and implements `Name()` and `Run(ctx *Context) error`.

**Static analysis and security**

DevForge uses **golangci-lint** for static analysis and **govulncheck** for vulnerability detection:

- **Static analysis:** `golangci-lint run` (with `--timeout=5m` in the pipeline). If the repository contains `.golangci.yml`, `.golangci.yaml`, or `.golangci.toml`, that configuration is used; otherwise golangci-lint runs with its default configuration. No config file is required.
- **Security analysis:** `govulncheck ./...` (with `-json` in the pipeline). Fails on HIGH or CRITICAL findings.

golangci-lint automatically includes many tools that are commonly run separately, such as: **gofmt**, **govet**, **staticcheck**, **gocyclo**, **errcheck**, **unused**, and **ineffassign**. The default PR pipeline runs a single golangci-lint step for static analysis and a separate govulncheck step for security.

**Examples:**

| Step name | Description |
|-----------|-------------|
| `gofmt` | Optional step: ensures code is formatted with `gofmt -s -l`. Available for `forge run gofmt`; the default PR pipeline uses golangci-lint, which includes formatting checks. |
| `golangci-lint` | Runs `golangci-lint run --timeout=5m`. Covers formatting, vet, staticcheck, gocyclo, errcheck, unused, ineffassign, and more. |
| `govulncheck` | Runs `govulncheck -json ./...`; fails on HIGH/CRITICAL. |
| `architectural-guard` | Validates repository against configurable architectural rules (e.g. no cross-layer imports). |
| `conventional-commit` | Validates PR title / commit message against conventional commit format. |
| **Plugins** | Custom steps from `.devforge.yml` (e.g. `proto-lint`, `license-check`); each runs via `bash -c "<run>"`. |
| `preconditions`, `version-derivation`, `create-tag`, `verify-tag`, `goreleaser` | Release pipeline steps. |
| `git-installed`, `goreleaser-installed`, `full-history`, `branch-main`, `working-tree-clean`, `tags-accessible` | Doctor checks. |
| `integration-tests` | Runs `go test -tags=integration`; only in **go-service** pipeline with **mode: deep**. See [Integration tests](#integration-tests) below. |

Steps are **registered automatically** in the step registry via `application.RegisterStep(name, constructor)` in each step package’s `init()`. The CLI command `forge run <step> [<step>...]` looks up steps by name and runs them in order.

---

### Integration tests

If a repository defines tests with `//go:build integration` (and optionally places them under `integrationtest/`), DevForge can execute them:

- **Automatic:** Use **mode: deep** for the go-service profile (e.g. `forge pr --profile go-service --mode deep`). The integration-tests step runs after unit tests and test-race.
- **Manual:** Run `forge run integration-tests`.

If no Go files in the repository contain `//go:build integration`, the step is skipped and succeeds without running tests. Libraries (go-lib profile) do not run integration tests in the pipeline.

---

## 8. Plugins

Plugins are custom validation steps defined in `.devforge.yml` under `plugins`. Each plugin has a **name** (for logging) and a **run** command (executed via `bash -c`). Plugins run **after** the standard pipeline (static analysis, tests, coverage). If a plugin command exits with a non-zero code, the pipeline fails.

**Writing a Hello World plugin**

1. Add a plugin entry to `.devforge.yml` in your repo root:

```yaml
plugins:
  - name: hello-world
    run: echo "Hello from DevForge plugin"
```

2. Run PR validation; the plugin runs after the built-in steps:

```bash
forge pr
```

3. To make the step fail the pipeline (e.g. for a real check), use a command that exits non-zero on failure:

```yaml
plugins:
  - name: license-check
    run: go-licenses check ./...
```

Plugins receive the same working directory as the rest of the pipeline; they run in sequence after the profile’s standard steps.

---

## Writing a Plugin (Hello World)

DevForge allows repositories to extend the validation pipeline using **plugins** defined in `.devforge.yml`. Plugins are **simple commands executed after the core pipeline completes**.

They are useful for:

- repository-specific validations
- additional linters
- custom scripts
- security checks
- language-specific tooling

### Minimal example

Minimal repository using a plugin:

```
repo/
├── .devforge.yml
├── go.mod
└── main.go
```

**Example `.devforge.yml`:**

```yaml
profile: go-lib

plugins:
  - name: hello-world
    run: echo "Hello from DevForge plugin"
```

This configuration:

- **Uses the go-lib pipeline profile** — Standard go-lib validation (static analysis via golangci-lint, security via govulncheck, tests, coverage) applies.
- **Runs the standard validation pipeline** — All core steps run first.
- **Executes the plugin command afterward** — The `hello-world` plugin runs once the core pipeline succeeds.

### What happens during CI

When you run:

```bash
forge pr
```

the following steps occur:

1. The repository profile is selected (from config, flags, or auto-detection).
2. The core validation pipeline runs (go-mod-tidy, conventional commit, architectural guard, policy-pack, static analysis via golangci-lint, govulncheck, tests, coverage).
3. Plugins defined in `.devforge.yml` execute in order.

**Example output:**

```
INFO running plugin plugin=hello-world
Hello from devforge plugin
INFO plugin completed plugin=hello-world duration_ms=...
```

### Running plugins locally

Developers can test plugins locally by running:

```bash
forge pr
```

This validates repository plugins before pushing changes, using the same pipeline and plugin order as in CI.

### Plugin capabilities

Plugin commands can run **any executable available** in the CI (or local) environment.

**Run a script:**

```yaml
plugins:
  - name: security-scan
    run: ./scripts/security-check.sh
```

**Run a linter:**

```yaml
plugins:
  - name: proto-lint
    run: buf lint
```

**Run a license check:**

```yaml
plugins:
  - name: license-check
    run: go-licenses check ./...
```

Plugins execute **sequentially** in the order defined in `.devforge.yml`. If any plugin exits with a non-zero code, the pipeline fails.

---

## Automatic Tool Installation

DevForge **automatically installs required analysis tools** when they are not present. You do not need to install them in CI or locally.

The following tools are installed on first use if missing:

- **golangci-lint** — via the official install script (pinned version)
- **govulncheck** — via `go install` (pinned version)

This allows CI workflows to run without manually installing tools.

**Example workflow:**

```yaml
steps:
  - uses: actions/checkout@v4
    with:
      fetch-depth: 0

  - uses: devforge/devforge@v1
```

No separate steps for installing golangci-lint or govulncheck are required; DevForge installs them when needed.

### Tool cache

DevForge automatically installs required tools and stores them in:

```
~/.devforge/tools
```

Tools are versioned and cached under this directory (e.g. `golangci-lint/v1.64.8/`, `govulncheck/v1.1.4/`). A `bin` directory exposes the active binaries and is prepended to `PATH`, so tools are installed once and reused across executions. This significantly speeds up CI pipelines.

**Installed tools:**

- **golangci-lint**
- **govulncheck**

**Example CI workflow:**

```yaml
steps:
  - uses: actions/checkout@v4
    with:
      fetch-depth: 0

  - uses: devforge/devforge@v1
```

No tool installation is required in the workflow.

---

## Plugins

DevForge supports **external plugins**: standalone binaries that extend the pipeline without modifying or recompiling DevForge.

Plugins are executables with the prefix `forge-plugin-<name>`, discovered via `PATH`. Examples:

- `forge-plugin-security`
- `forge-plugin-license`
- `forge-plugin-sbom`

They run as pipeline steps **after** core steps (e.g. golangci-lint, govulncheck, tests). No configuration is required; if the binary is on `PATH`, it is executed.

### Example plugin: forge-plugin-hello

```go
package main

import "fmt"

func main() {
	fmt.Println("hello from devforge plugin")
}
```

Build and place the binary on `PATH`:

```bash
go build -o forge-plugin-hello
# Ensure the directory containing forge-plugin-hello is on PATH
```

When you run `forge pr` (or `forge run plugin-hello`), DevForge will run the binary with the repository workdir as the current directory. Exit code 0 means success; non-zero fails the pipeline.

### Plugin configuration

Plugins can be configured using `.devforge.yml`. Use the **`plugins`** key as a map from plugin name to options (e.g. `enabled`, and any custom keys like `severity`). That configuration is passed to the plugin via the environment variable **`DEVFORGE_PLUGIN_CONFIG`** (JSON).

**Example `.devforge.yml`:**

```yaml
plugins:
  security:
    enabled: true
    severity: high

  license:
    enabled: true
```

- **`enabled`** — If `false`, that plugin step is skipped. Defaults to `true` when omitted.
- Other keys (e.g. `severity`) are included in the JSON passed to the plugin.

**Example plugin code reading config:**

```go
cfg := os.Getenv("DEVFORGE_PLUGIN_CONFIG")
fmt.Println("plugin config:", cfg)
// e.g. plugin config: {"severity":"high"}
```

### Plugin tools

Plugins must manage their own dependencies. DevForge does not install or manage tools for plugins.

A common convention is for a plugin to install its own tools under:

```
~/.devforge/plugins/<plugin>/tools
```

DevForge does not read or write this directory; it keeps the core runner simple and deterministic.

---

## 9. Policy Packs

Repositories can define **governance rules** using policy packs under `.devforge/policies/`. Policy packs are **optional**: if the directory does not exist or is empty, no policy validation runs and the pipeline behaves exactly as it does today.

Each policy file is YAML with a `name`, optional `type`, optional `severity`, and a `rules` map. Rules can be a single value (string) or a list of values.

### Example: `.devforge/policies/security.yaml`

```yaml
name: security
severity: error

rules:
  forbid_import:
    - net/http/pprof
    - internal/adapters
```

### Rules

- **`forbid_import`** — Fails if any package imports a path containing the given segment. Value can be a string or a list (e.g. `internal/adapters`, `net/http/pprof`).
- **`forbid_time_now`** — Fails if `time.Now()` appears in the given path. Value is a path segment (e.g. `domain` → `internal/domain`).

### Severity

- **`error`** (default) — A violation fails the PR pipeline.
- **`warning`** — A violation is logged but the pipeline continues; useful for gradual adoption.

### Behavior

Policy packs run after the architectural guard step during `forge pr`. If no `.devforge/policies` directory exists, the step is a no-op and the pipeline is unchanged.

---

## 10. Using DevForge in GitHub Actions

DevForge can be used as a **reusable GitHub Action** to validate repositories using deterministic CI pipelines.

### GitHub Action installation

The action **downloads the forge binary from GitHub Releases**, **caches it between workflow runs**, and runs the PR validation pipeline. **Go is not required** in consuming repositories.

- The binary is downloaded from the [DevForge releases](https://github.com/devforge/devforge/releases) for the runner OS and architecture.
- The binary is stored under `~/.devforge/bin` and cached by the action (cache key includes runner OS, arch, and DevForge version).
- If the cache is restored, the download step is skipped and the existing binary is reused.
- Downloads use `curl -f` so the workflow fails immediately if the release asset is missing.

### Action versioning

| Reference | Meaning |
|-----------|---------|
| **`@v1`** | Latest stable v1 release. Updated automatically after each release. Use for most workflows. |
| **`@v1.x.x`** | Immutable release (e.g. `@v1.0.0`, `@v1.2.3`). Never changes; use when you need a pinned version. |

After each release, the `v1` tag is moved to the new release commit so `devforge/devforge@v1` always runs the latest stable action and binary.

When the action runs, DevForge automatically:

- **Detects repository type (profile)** — Infers whether the repo is `go-lib` or `go-service` (or uses profile from `.devforge.yml` or workflow inputs).
- **Runs the appropriate validation pipeline** — Executes the PR validation steps (static analysis via golangci-lint, security via govulncheck, tests, coverage, etc.) for that profile.
- **Executes repository plugins** — Runs any additional validation steps defined in `.devforge.yml` (e.g. license-check, proto-lint) after the main pipeline.

### Minimal CI example

```yaml
name: CI

on:
  pull_request:

jobs:
  validate:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: devforge/devforge@v1
```

No additional setup is required.

The action downloads the forge binary from GitHub Releases, caches it between runs, and executes the PR validation pipeline.

**What happens when the workflow runs:** The action restores or downloads the forge binary, then DevForge analyzes the repository, selects the correct pipeline profile, runs the validation pipeline (and installs tools like golangci-lint and govulncheck under `~/.devforge/tools` when needed), and executes any optional plugins defined in `.devforge.yml`. The job fails if any step in the pipeline or a plugin fails.

---

## 11. Architecture

DevForge uses a layered (hexagonal) design. The CLI loads `.devforge.yml` (optional) and applies flag overrides before selecting profile and invoking pipelines. Dependencies point inward: CLI → profiles → application → ports; steps implement application’s Step interface; adapters implement ports; domain has no external dependencies.

```
CLI
 ↓
profiles
 ↓
application (pipelines / use cases)
 ↓
steps
 ↓
ports
 ↓
adapters
 ↓
domain
```

| Layer | Role |
|-------|------|
| **CLI** | Parses arguments, selects profile (or uses detection), invokes profile entry points. No business logic. |
| **profiles** | Composition root: wires adapters, registers pipelines, implements RunPRWithMode, RunRelease, RunDoctor, RunSteps. |
| **application** | Use cases (PR, release, doctor), Pipeline and Step abstractions, pipeline and step registries, StepRunner, Context. Depends only on ports and domain. |
| **steps** | Concrete Step implementations (PR, release, doctor, static analysis, guard). Depend on application (and guard where needed); receive ports via Context. |
| **ports** | Interfaces for command execution, git, env, logger, clock. No implementation. |
| **adapters** | Implementations of ports (exec, git, env, logger, clock). Used only by profiles when building use cases. |
| **domain** | Pure logic: version derivation, conventional commit validation, coverage parsing, govulncheck parsing. No I/O or infrastructure. |

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed diagrams and execution flows.

---

## 12. Extending DevForge

### Adding a new step

1. Add a new file under `internal/steps/` (e.g. `mystep.go`).
2. Implement the `application.Step` interface (`Name() string`, `Run(ctx *application.Context) error`). Use only `ctx` (workdir, Cmd, Git, Log, etc.); do not import adapters.
3. In `init()`, register the step: `application.RegisterStep("my-step", func() application.Step { return NewMyStep() })`.
4. The step is then available for pipelines and for `forge run my-step`.

### Adding a new pipeline

1. In a profile’s `init()` (or a new profile), build the step list (using existing step constructors or new ones from `internal/steps`).
2. Register it: `application.RegisterPipeline(application.Pipeline{Name: "my-pipeline", Steps: steps})`.
3. Use it by retrieving it by name (`application.GetPipeline("my-pipeline")`) when building use cases in that profile, or via a helper like `RunPipeline`.

### Adding a new profile

1. Add a new file under `internal/profiles/` (e.g. `go_app.go`).
2. Implement profile-specific step builders and thresholds if needed.
3. In `init()`, register pipelines with `application.RegisterPipeline(...)` for that profile’s PR, release, and doctor pipelines.
4. Register the profile: `Register(Profile{Name: "go-app", RunPRWithMode: ..., RunRelease: ..., RunDoctor: ...})`.
5. Optionally extend `internal/detection/repository.go` so that `DetectProfile(workdir)` can return `"go-app"` when appropriate.

---

## 13. Repository Structure

```
cmd/
  forge/                  CLI entrypoint; dispatches to profiles and application

internal/
  application/            Use cases, Pipeline, Step interface, Context, registries, StepRunner
  steps/                  All Step implementations (PR, release, doctor, static analysis, plugins, guard)
  profiles/               Composition root: profile and pipeline registration, RunSteps
  config/                 Repository config loading (.devforge.yml)
  ports/                  Interfaces: CommandRunner, GitClient, EnvProvider, Logger, Clock
  adapters/               Implementations: exec, git, env, logger, clock
  domain/                 Pure logic: version, conventional commit, coverage, govulncheck
  guard/                  Architectural rules (used by steps and validation)
  detection/              Repository type detection (go-lib vs go-service)
  mocks/                  Test doubles for ports
```

- **cmd** — Single binary; no business logic.
- **application** — Core abstractions and orchestration; no adapter or step imports.
- **steps** — All concrete steps; depend on application (and guard where needed).
- **profiles** — Wire adapters and pipelines; call application use cases.
- **ports** — Interfaces only.
- **adapters** — Infrastructure; implement ports.
- **domain** — Pure functions and types; no I/O.
- **detection** — Heuristic to choose profile when `--profile` is omitted.
- **config** — Loads `.devforge.yml`; used by CLI to resolve profile/mode/plugins when flags are not set. (Config file name preserved for backward compatibility.)

---

## 14. Design Goals

- **Deterministic CI** — Same inputs (commit history, working tree) produce the same version and validation results. No timestamps or randomness in version derivation or pipeline outcome.
- **Minimal configuration** — Profiles encode defaults (thresholds, timeouts, step sets). Repositories can optionally add `.devforge.yml` to set profile, mode, and plugins; CLI flags override config; auto-detection applies when neither is set.
- **Strong architectural boundaries** — Clear layers (CLI, profiles, application, steps, ports, adapters, domain) and dependency rules so that business logic stays testable and infrastructure stays swappable.
- **Composable pipelines** — Pipelines are ordered lists of steps; steps can be grouped (e.g. parallel or sequential groups) and reused across pipelines and profiles.
- **Simple extensibility** — New steps, pipelines, and profiles are added by implementing well-defined interfaces and registering in `init()`, without changing core use case or domain code.
