# Lint fixes for kit-core (DevForge PR)

Apply these in the **kit-core** repo so `forge pr --profile go-lib` passes. Aligned with AGENTS.md (GoDoc for exported symbols, handle errors).

---

## G104 – Handle errors (gosec)

### 1. `pkg/resilience/circuit_breaker_test.go` line 159

**Before:**
```go
		cb.Execute(func() error {
			return errors.New("operation failed")
		})
```

**After:**
```go
		_ = cb.Execute(func() error {
			return errors.New("operation failed")
		})
```
(or use `require.NoError(t, cb.Execute(...))` if you have testify)

### 2. `tenant/context_test.go` line 139

**Before:**
```go
	s.SetDefaultTenantID("new-default")
```

**After:**
```go
	_ = s.SetDefaultTenantID("new-default")
```
(or `require.NoError(t, s.SetDefaultTenantID("new-default"))`)

---

## G117 – Secret-named fields (gosec)

### 3. `security/config/config.go` lines 25, 36, 50

Add a single-line nolint with justification so the struct can keep its JSON keys:

**Before (example line 25):**
```go
	Secret      string
```

**After:**
```go
	Secret      string //nolint:gosec // G117: config field name required for JSON unmarshaling
```

Do the same for the two `ClientSecret` fields (lines 36 and 50): add `//nolint:gosec // G117: ...` on the line above each field.

---

## revive – Package comments

Add a one-line package comment at the top of each file (before `package`).

### 4. `clock/clock.go`
```go
// Package clock provides time utilities.
package clock
```

### 5. `repository/cursor_pagination.go`
```go
// Package repository provides data access types.
package repository
```
(Adjust wording to match what the package actually does.)

### 6. `timestamp/interfaces.go`
```go
// Package timestamp provides time and timestamp interfaces.
package timestamp
```

---

## revive – Unused parameters

Rename the parameter to `_` (or `_param` if the linter requires a name).

### 7. `listeners/listeners.go`
- Line 46: `func(l interface{})` → `func(_ interface{})`
- Line 56: same
- Line 66: same
- Line 75: `policy RetryPolicy` → `_ RetryPolicy`

### 8. `listeners/listeners_test.go`
- Line 19: `ctx context.Context` → `_ context.Context`
- Line 32: `ctx context.Context` → `_ context.Context`
- Line 46: `ctx context.Context` → `_ context.Context`

### 9. `pkg/resilience/config_test.go` line 48
- `t *testing.T` → `_ *testing.T` (or keep `t` and use it, e.g. `t.Helper()`)

### 10. `pkg/resilience/fallback_test.go` lines 30, 48, 102
- `err error` → `_ error` in each fallback func

### 11. `pkg/resilience/mocks.go` line 105
- `m *MockCircuitBreaker` → `_ *MockCircuitBreaker`

### 12. `pkg/resilience/resilient_handlers_test.go` line 158
- `event interface{}` → `_ interface{}`

### 13. `tenant/context_test.go` line 243
- `t *testing.T` → `_ *testing.T` (or use `t` in the test)

### 14. `fflags/evaluator_test.go` line 39 (if still present)
- `ctx context.Context` → `_ context.Context`

---

## revive – Exported symbols need comments

Add a one-line GoDoc comment (starting with the symbol name) for each.

### 15. `repository/query_options.go` line 9
```go
// SortDirectionAsc is the ascending sort direction.
SortDirectionAsc  = "asc"
```

### 16. `tenant/context.go`
- Before `TenantContextKey`: `// TenantContextKey is the type for context keys.`
- Before `TenantIDKey`: `// TenantIDKey is the context key for tenant ID.`
- Before `TenantContext`: `// TenantContext holds tenant-scoped data.`
- Before `TenantContextService`: `// TenantContextService manages tenant context.`

### 17. `tenant/interfaces.go`
- Before `TenantContextServiceInterface`: `// TenantContextServiceInterface defines tenant context operations.`
- Before `TenantValidator`: `// TenantValidator validates tenant data.`

### 18. `timestamp/timestamp_service.go`
- Before the struct: `// TimestampService provides timestamp operations.` (or similar)

### 19. `validation/errors.go` line 7
```go
// ErrorTypeRequired indicates a required-field validation error.
ErrorTypeRequired      ErrorType = "required"
```

---

## revive – Naming (stutter / var-naming)

These are suggestions; fixing them may require API/import changes.

### 20. `security/types/types.go` – “meaningless package names”
Rename package `types` to something specific (e.g. `securitytypes` or `sectype`) and update imports, **or** add a file-level `//revive:disable:var-naming` at the top of the file.

### 21. Stutter (tenant.TenantContext, etc.)
Either rename types (e.g. `TenantContext` → `Context` in package tenant) and update all usages, or disable for that file:
```go
//revive:disable:exported
```
at the top of `tenant/context.go` and `tenant/interfaces.go` and `timestamp/timestamp_service.go` if you prefer to keep current names.

---

## SA4006 (staticcheck) – “value of ctx is never used”

Use the returned `ctx` or assign to `_` so the new value is not “never used”.

### 22. `tenant/context_test.go` lines 224, 231, 255

**Line 224 – Before:**
```go
	ctx, id = EnsureCorrelationID(ctx)
```
**After (if you only need id):**
```go
	_, id = EnsureCorrelationID(ctx)
```

**Line 231 – Before:**
```go
	ctx, causID := EnsureCausationID(ctx, "corr")
```
**After:**
```go
	_, causID := EnsureCausationID(ctx, "corr")
```

**Line 255 – Before:**
```go
	ctx = EnsureTraceIDs(context.Background(), nil, nil)
```
**After (if ctx is not used later):**
```go
	_ = EnsureTraceIDs(context.Background(), nil, nil)
```
(or keep `ctx` and use it in an assertion if the test intends to verify context propagation).

---

## Summary

| Category        | Count | Action |
|----------------|-------|--------|
| G104            | 2     | Handle return value of Execute / SetDefaultTenantID |
| G117            | 3     | Add //nolint for Secret / ClientSecret with short justification |
| package-comments| 3     | Add package comment in clock, repository, timestamp |
| unused-parameter | 12  | Rename param to `_` (or use it) |
| exported        | 6+    | Add GoDoc for exported const/type |
| var-naming/stutter | optional | Rename or revive:disable |
| SA4006          | 3     | Use `_` for unused `ctx` or use `ctx` in test |

After applying these in **kit-core**, run:
`forge pr --profile go-lib`
again to confirm all 35 issues are resolved.
