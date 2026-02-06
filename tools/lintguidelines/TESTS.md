# Lintguidelines CLI - Simulation & Exploratory Tests

Static analysis linter for validating architecture guidelines. Enforces naming, coupling, routing, and storage isolation rules.

## Quick Reference

```powershell
# Run all linter tests
go test ./tools/lintguidelines -v

# Run integration tests only
go test ./tools/lintguidelines -v -run Integration

# Run pairwise rule combination tests
go test ./tools/lintguidelines -v -run Pairwise
```

---

## Current Test Coverage

| Test Suite | Coverage Focus |
|------------|----------------|
| `main_test.go` | All 4 rules triggered from synthetic files |
| `pairwise_test.go` | All 16 rule combinations |
| `integration_test.go` | Scaffolded apps pass lint in strict mode |

---

## Rule Reference

| Rule | Checks | Location |
|------|--------|----------|
| `naming` | Banned abbreviations in identifiers | `internal/**/*.go` |
| `concept-coupling` | No cross-concept imports | `internal/domain/<concept>/*.go` |
| `route-query` | GET → projections only | `internal/adapters/http/routes.go` |
| `route-command` | POST/PUT/DELETE → orchestrators only | `internal/adapters/http/routes.go` |
| `storage-isolation` | Storage matches concept, no cross-imports | `internal/adapters/storage/<concept>/` |

---

## Exploratory Test Scenarios

### 1. Naming Rule Tests

**Banned Abbreviations**: `usr`, `amt`, `cfg`, `tmp`, `svc`

| Identifier | Expected |
|------------|----------|
| `UserID` | ✓ Pass |
| `UsrID` | ✗ Violation: `naming` |
| `TotalAmount` | ✓ Pass |
| `TotalAmt` | ✗ Violation: `naming` |
| `Config` | ✓ Pass |
| `cfg` | ✗ Violation: `naming` |
| `tmpFile` | ✗ Violation: `naming` |
| `SvcClient` | ✗ Violation: `naming` |

**Allowed Abbreviations**: `ID`, `URL`, `HTTP`, `JSON`, `SQL`, `UUID`, `API`

| Identifier | Expected |
|------------|----------|
| `UserID` | ✓ Pass |
| `APIURL` | ✓ Pass |
| `HTTPClient` | ✓ Pass |
| `JSONData` | ✓ Pass |
| `SQLQuery` | ✓ Pass |
| `UUIDField` | ✓ Pass |

### 2. Concept Coupling Tests

**File Location**: `internal/domain/<concept>/*.go`

```go
// ✗ VIOLATION: imports another concept
package order
import "workshop/internal/domain/customer"

// ✗ VIOLATION: imports application layer
package order
import "workshop/internal/application/orchestrators"

// ✗ VIOLATION: imports adapters layer
package order
import "workshop/internal/adapters/http"

// ✓ PASS: imports own concept only
package order
import "workshop/internal/domain/order/types"

// ✓ PASS: imports stdlib
package order
import "context"
```

### 3. Route Discipline Tests

**GET handlers must**:
- Call `projections.Query*`
- NOT call `orchestrators.Execute*`

```go
// ✗ VIOLATION: GET calls orchestrator
func handleGetOrders(w http.ResponseWriter, r *http.Request) {
    if r.Method != "GET" { return }
    _ = orchestrators.ExecuteCreateOrder(r.Context(), input)
}

// ✗ VIOLATION: GET missing projection call
func handleGetOrders(w http.ResponseWriter, r *http.Request) {
    if r.Method != "GET" { return }
    w.Write([]byte("hello"))
}

// ✓ PASS: GET calls projection
func handleGetOrders(w http.ResponseWriter, r *http.Request) {
    if r.Method != "GET" { return }
    result, _ := projections.QueryOrderSummary(ctx, query)
    json.NewEncoder(w).Encode(result)
}
```

**POST/PUT/DELETE handlers must**:
- Call `orchestrators.Execute*`
- NOT call `projections.Query*`

```go
// ✗ VIOLATION: POST missing orchestrator
func handlePostOrders(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" { return }
    w.WriteHeader(204)
}

// ✗ VIOLATION: POST calls projection
func handlePostOrders(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" { return }
    _ = projections.QueryOrderSummary(ctx, query)
}

// ✓ PASS: POST calls orchestrator
func handlePostOrders(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" { return }
    _ = orchestrators.ExecuteCreateOrder(ctx, input)
}
```

### 4. Storage Isolation Tests

**Requirements**:
1. Storage package folder must match a concept directory
2. Package name must match folder name
3. Must not import other concepts
4. Must define `Store` interface

```go
// Location: internal/adapters/storage/order/store.go

// ✗ VIOLATION: package name mismatch
package storage  // should be: package order

// ✗ VIOLATION: imports another concept
package order
import "workshop/internal/domain/customer"

// ✗ VIOLATION: missing Store interface
package order
type OrderRepo struct{}

// ✓ PASS: correct setup
package order
import "workshop/internal/domain/order"
type Store interface {
    GetByID(ctx context.Context, id string) (order.Order, error)
    Save(ctx context.Context, value order.Order) error
}
```

**Orphan Storage Check**:
```
internal/adapters/storage/billing/store.go exists
internal/domain/billing/ does NOT exist
→ VIOLATION: storage-isolation "storage package has no matching concept"
```

---

## Output Format Tests

### Text Output (default)

```powershell
go run ./tools/lintguidelines --format text

# Expected format per violation:
# <file>:<line>:<column> [<rule>] <message>
internal/domain/order/model.go:5:2 [naming] avoid abbreviation "usr" in identifier "UsrID"
```

### JSON Output

```powershell
go run ./tools/lintguidelines --format json

# Expected: valid JSON array
[
  {
    "file": "internal/domain/order/model.go",
    "line": 5,
    "column": 2,
    "rule": "naming",
    "message": "avoid abbreviation \"usr\" in identifier \"UsrID\"",
    "severity": "warning"
  }
]
```

### Strict Mode

```powershell
go run ./tools/lintguidelines --strict
echo $LASTEXITCODE

# Expected:
# - Exit code 0: no violations
# - Exit code 2: violations found
```

---

## File Exclusion Tests

| Path | Linted? |
|------|---------|
| `tools/**/*.go` | ✗ Excluded |
| `vendor/**/*.go` | ✗ Excluded |
| `node_modules/**` | ✗ Excluded |
| `static/**` | ✗ Excluded |
| `.scaffold/**` | ✗ Excluded |
| `*.go` with `Code generated by scaffold` | ✗ Excluded |
| `internal/**/*.go` | ✓ Linted |

---

## Edge Cases

### Empty Repository

```powershell
# No internal/ directory
go run ./tools/lintguidelines --root C:\empty
# Expected: "ok" (no violations)
```

### Missing routes.go

```powershell
# internal/adapters/http/ exists but no routes.go
go run ./tools/lintguidelines --root C:\partial
# Expected: route rules skipped gracefully
```

### Malformed Go Files

```go
// internal/domain/broken/model.go
package broken
func Incomplete( {  // syntax error
```

```powershell
go run ./tools/lintguidelines
# Expected: error parsing file, exit 1
```

---

## Integration Verification

```powershell
# Scaffold a valid app and verify it passes lint
go run ./tools/scaffold init --root C:\temp\lint-test `
  --concept Order --field Order:Status:string

go run ./tools/lintguidelines --root C:\temp\lint-test --strict
echo "Exit code: $LASTEXITCODE"

# Expected: Exit code 0 (scaffolded apps are lint-clean)
```

---

## Manual Smoke Test

```powershell
# Create intentional violations
mkdir C:\temp\lint-manual\internal\domain\order -Force
@"
package order

import "workshop/internal/domain/customer"

type Order struct {
    UsrID string
    Amt   int
}
"@ | Out-File C:\temp\lint-manual\internal\domain\order\model.go

# Run linter
go run ./tools/lintguidelines --root C:\temp\lint-manual --format text

# Expected violations:
# - naming: "usr" in UsrID
# - naming: "amt" in Amt  
# - concept-coupling: imports customer
```
