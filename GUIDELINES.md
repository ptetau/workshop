# Application Guidelines

Build software by separating **what data is** (Concepts), **how it changes** (Orchestrators), and **how it's viewed** (Projections).

---

## Quick Start

```powershell
# 1. Design interactively
go run ./tools/interview --root ./myapp --module myapp

# 2. Generate code
go run ./tools/scaffold init --root ./myapp --module myapp \
  --concept Order --field Order:UserID:string --method Order:Cancel \
  --orchestrator CancelOrder --param CancelOrder:OrderID:string \
  --projection OrderSummary --query OrderSummary:OrderID:string \
  --route "GET:/orders/{id}:OrderSummary" --route "POST:/orders/{id}/cancel:CancelOrder"

# 3. Verify compliance
go run ./tools/lintguidelines --root ./myapp --strict
```

**Generated structure:**
```
internal/
├── domain/order/model.go           ← Concept: state + methods
├── application/
│   ├── orchestrators/cancel_order.go   ← Coordinates workflows
│   └── projections/order_summary.go    ← Read-only views
└── adapters/
    ├── http/routes.go              ← GET→projection, POST→orchestrator
    └── storage/order/store.go      ← CRUD + List interface
```

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│  Routes                                                         │
│  GET → Projections (read-only)                                  │
│  POST/PUT/DELETE → Orchestrators (state changes)                │
└─────────────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┴───────────────────┐
        ▼                                       ▼
┌───────────────────┐                 ┌───────────────────┐
│   Orchestrators   │                 │   Projections     │
│   (Commands)      │                 │   (Queries)       │
│                   │                 │                   │
│ Coordinate across │                 │ Combine data from │
│ multiple concepts │                 │ multiple concepts │
└───────────────────┘                 └───────────────────┘
        │                                       │
        └───────────────────┬───────────────────┘
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Concepts                                │
│  Independent state + methods. Each owns its storage.            │
└─────────────────────────────────────────────────────────────────┘
```

### Key Terms

| Term | What It Is | Example |
|------|------------|---------|
| **Concept** | Self-contained domain entity with state, methods, and storage. Never calls other concepts. | `Order`, `Inventory`, `User` |
| **Orchestrator** | Command function coordinating changes across concepts. The only place cross-concept logic lives. | `CancelOrder`, `PlaceOrder` |
| **Projection** | Read-only query combining data from concepts. Rebuildable from concept state. | `OrderSummary`, `UserDashboard` |
| **Route** | HTTP endpoint. GET → projection; POST/PUT/DELETE → orchestrator. | `GET /orders/{id}`, `POST /orders` |

### Data Flow

| Path | Example |
|------|---------|
| **Read** | `GET /orders/{id}` → `QueryOrderSummary()` → reads Order + User → returns view |
| **Write** | `POST /orders/{id}/cancel` → `ExecuteCancelOrder()` → Order.Cancel() + Inventory.Release() → persists |

---

## Building Blocks

### Concepts

Concepts are independent domain entities. Each owns its state, methods, and storage.

```go
// internal/domain/order/model.go
type Order struct {
    ID          string
    UserID      string    // Reference by ID only, never embed User
    Status      string
    CancelledAt time.Time
}

var ErrAlreadyCancelled = errors.New("order already cancelled")

func (o *Order) Cancel() error {
    if o.Status == "cancelled" {
        return ErrAlreadyCancelled
    }
    o.Status = "cancelled"
    o.CancelledAt = time.Now()
    return nil
}
```

| Don't | Do |
|-------|------|
| Call another concept from a concept method | Keep methods self-contained |
| Store full `User` object in `Order` | Store `UserID` string |
| Share tables between concepts | Each concept owns its storage |

**When to create a concept:** Must have both (1) independent lifecycle AND (2) be referenceable by ID.

| Entity | Lifecycle? | Referenced? | Decision |
|--------|------------|-------------|----------|
| `Order` | Yes | Yes | **Concept** |
| `Address` | No (part of User) | No | Embed in User |
| `LineItem` | No (part of Order) | No | Embed in Order |

---

### Orchestrators

Orchestrators coordinate workflows across concepts. They validate inputs and handle failures.

```go
// internal/application/orchestrators/cancel_order.go

type CancelOrderDeps struct {
    OrderStore     OrderStoreForCancel
    InventoryStore InventoryStoreForCancel
}

func ExecuteCancelOrder(ctx context.Context, input CancelOrderInput, deps CancelOrderDeps) error {
    // Validate input
    if input.OrderID == "" {
        return ErrMissingOrderID
    }
    
    // Coordinate concepts
    order, err := deps.OrderStore.GetByID(ctx, input.OrderID)
    if err != nil {
        return fmt.Errorf("order not found: %w", err)
    }
    
    if err := order.Cancel(); err != nil {
        if errors.Is(err, ErrAlreadyCancelled) {
            return &ConflictError{Message: "order already cancelled"}
        }
        return err
    }
    deps.OrderStore.Save(ctx, order)
    
    // Coordinate with other concepts
    inventory, _ := deps.InventoryStore.GetByID(ctx, order.InventoryID)
    inventory.Release(order.Quantity)
    deps.InventoryStore.Save(ctx, inventory)
    
    return nil
}
```

| Don't | Do |
|-------|------|
| Put workflow logic in concepts | Coordinate in orchestrators |
| Use cross-concept transactions | Design compensating actions |
| Ignore partial failures | Define compensation per step |
| Use global store variables | Use a deps struct with store interfaces (dependency injection) |

---

### Projections

Projections are read-only views combining data from concepts.

```go
// internal/application/projections/order_summary.go
func QueryOrderSummary(ctx context.Context, query OrderSummaryQuery) (OrderSummaryResult, error) {
    // Validate query params
    if query.OrderID == "" {
        return OrderSummaryResult{}, ErrMissingOrderID
    }
    
    order, _ := orderStore.GetByID(ctx, query.OrderID)
    user, _ := userStore.GetByID(ctx, order.UserID)
    
    return OrderSummaryResult{
        OrderID:  order.ID,
        UserName: user.Name,
        Status:   order.Status,
    }, nil
}
```

| Don't | Do |
|-------|------|
| Mutate state in projections | Read and combine only |
| Denormalize concepts into each other | Build separate projections |
| Treat projections as source of truth | Rebuild from concept state |

---

### Storage Interface

Each concept has a `Store` interface with CRUD + List.

```go
// internal/adapters/storage/order/store.go
type Store interface {
    GetByID(ctx context.Context, id string) (domain.Order, error)
    Save(ctx context.Context, order domain.Order) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filter ListFilter) ([]domain.Order, error)
}
```

| Operation | Used By |
|-----------|---------|
| `GetByID`, `Save`, `Delete` | Orchestrators |
| `List` | Projections |

| Don't | Do |
|-------|------|
| Share storage across concepts | One interface per concept |
| Put business logic in storage | Keep as pure data access |

---

### Routes

Routes enforce the read/write split.

| Method | Calls | Example |
|--------|-------|---------|
| GET | `projections.Query*()` | `GET /orders/{id}` → `QueryOrderSummary` |
| POST, PUT, DELETE | `orchestrators.Execute*()` | `POST /orders/{id}/cancel` → `ExecuteCancelOrder` |

---

## Feature Integration Checklist (Harmony)

When introducing a **new product area** (or gating an existing one), make the integration consistent across the app shell.

**1) Choose / define the feature flag key**
- Add/confirm the key in `internal/domain/featureflag/defaults.go`.
- Pick defaults that match current roles (admin/coach/member/trial) and whether a beta override makes sense.

**2) Enforce on server routes**
- Pages: use `requireFeaturePage(...)` and redirect to `/dashboard` when disabled.
- APIs: use `requireFeatureAPI(...)` and return `403 Forbidden` when disabled.
- Apply to *all* entrypoints: list/detail/create/update/delete routes, not just the main page.

**3) Nav / templates**
- Wrap nav links in `{{ if featureEnabled "<key>" }}` so users don’t see links they can’t access.
- If a page has cross-links/buttons into the feature area, gate those too.

**4) Tests**
- Add at least one unit test proving the feature gate blocks access (page redirect or API 403).
- Add at least one happy-path test proving access is allowed when enabled.

**5) Storage / migrations**
- If the feature requires new persistence, add a migration and update `storage/db_test.go` expectations.
- Consider backfills for existing rows (idempotent migrations).

**6) Operational / rollout**
- Verify the admin UI for feature flags can toggle the key and that toggles take effect immediately.
- If production scripts/calls exist, ensure they use the same endpoints and are still permitted.

| Don't | Do |
|-------|------|
| Mutate state in GET handlers | Reads only; mutations via POST/PUT/DELETE |

---

## Rules

### Naming
All symbols must be explicit. No abbreviations.

| Don't | Do |
|-------|------|
| `usr`, `amt`, `cfg` | `userID`, `paymentAmountCents`, `config` |

### Validation Layers

| Layer | Validates |
|-------|-----------|
| Orchestrators | Inputs (required fields, formats) |
| Concepts | Invariants (business rules) |
| Projections | Query params |

### Error Handling

| Layer | Error Type | HTTP |
|-------|------------|------|
| Concept | Domain (`ErrAlreadyCancelled`) | 409 |
| Orchestrator | Validation (`ErrMissingOrderID`) | 400 |
| Storage | Not found | 404 |
| Storage | Infrastructure | 500 |

### Consistency
- Cross-concept consistency is **eventual**
- Orchestrators use **compensating actions** for partial failures
- Schema is defined in `db.go` with `CREATE TABLE IF NOT EXISTS` (idempotent on startup)
- Projections can be rebuilt from concept state

---

## Testing

| Layer | Test Type | Approach |
|-------|-----------|----------|
| Concepts | Unit | Call method, assert state |
| Orchestrators | Unit | Mock storage, verify call sequence |
| Orchestrators | Integration | In-memory storage, full workflow |
| Projections | Unit | Mock storage, verify result structure |

```go
// Unit: concept invariant
func TestOrder_Cancel(t *testing.T) {
    order := Order{Status: "pending"}
    err := order.Cancel()
    assert.NoError(t, err)
    assert.Equal(t, "cancelled", order.Status)
}

// Integration: orchestrator workflow
func TestCancelOrder_ReleasesInventory(t *testing.T) {
    store := NewInMemoryStore()
    store.Save(ctx, Order{ID: "1", InventoryID: "inv1", Quantity: 5})
    store.Save(ctx, Inventory{ID: "inv1", Reserved: 5})
    
    ExecuteCancelOrder(ctx, CancelOrderInput{OrderID: "1"})
    
    inv, _ := store.GetInventory(ctx, "inv1")
    assert.Equal(t, 0, inv.Reserved)
}
```

| Don't | Do |
|-------|------|
| Test with real databases | Mock storage for unit tests |
| Skip multi-concept workflows | Integration test full paths |
| Test only happy paths | Include failure + compensation |

---

## Tooling Reference

### Commands

| Tool | Purpose |
|------|---------|
| `interview` | Interactive design session |
| `scaffold` | Generate code structure |
| `lintguidelines` | Verify compliance |

### Lint Rules

| Rule | Checks |
|------|--------|
| `naming` | No abbreviations |
| `concept-coupling` | Concepts don't import each other |
| `route-query` | GET → projections only |
| `route-command` | POST/PUT/DELETE → orchestrators only |
| `storage-isolation` | One storage per concept |

### Scaffold Flags

```
--concept Name              Create concept
--field Concept:Field:Type  Add field to concept
--method Concept:Method     Add method to concept
--orchestrator Name         Create orchestrator
--param Orch:Param:Type     Add param to orchestrator
--projection Name           Create projection
--query Proj:Param:Type     Add query param
--result Proj:Field:Type    Add result field
--route "METHOD:path:Target" Create route
```

---

## Privacy & Compliance

All code must comply with **GDPR**, **NZ Privacy Act 2020**, and **SOC 2** requirements. Full guidance in `PRIVACY.md`.

### Data Handling Rules

| Don't | Do |
|-------|------|
| Store raw credit card numbers | Store payment processor tokens only |
| Use a single "I agree" checkbox | Separate granular consent per purpose |
| Hard-delete everything on member deletion | Anonymise PII, hard-delete medical, retain financials 7 years |
| Log passwords, tokens, or medical details | Redact sensitive fields from audit logs |
| Use production data in dev/test | Use synthetic seeded data (already implemented) |
| Store medical history | Collect only current injuries relevant to training |

### Audit Logging

Every state-changing orchestrator and security-relevant event must emit an audit log entry.

```go
// In orchestrators that mutate member data:
slog.Info("audit_event",
    "actor_id", actorID,
    "actor_role", actorRole,
    "action", "member.profile.update",
    "resource_type", "member",
    "resource_id", memberID,
)
```

| Layer | Audit Responsibility |
|-------|---------------------|
| Routes | Log authentication events (login, logout, lockout) |
| Orchestrators | Log all data mutations with actor context |
| Projections | Log sensitive data access (profile views, exports) |

### Consent & Deletion

| Requirement | Architecture Layer |
|-------------|-------------------|
| Consent records (versioned, granular) | Concept: `ConsentRecord` |
| Waiver re-signing on version change | Orchestrator: check version at login/check-in |
| Data deletion (anonymise + hard-delete) | Orchestrator: `ExecuteDeleteMember` |
| Data export (JSON/CSV) | Projection: `QueryMemberDataExport` |
| Deletion grace period (30 days) | Orchestrator: scheduled execution |

### Data Classification

Storage adapters must enforce access controls matching data classification:

| Classification | Storage Rule |
|---------------|-------------|
| **Restricted** (injuries, observations) | Separate tables, coach/admin-only access methods |
| **Confidential** (PII, sizes) | Standard member table, anonymisable |
| **Financial** (payments) | Admin-only access, 7-year retention |
| **Public** (schedules, programs) | No access restriction |

---

## Corollaries

Quick reference for common decisions:

| Question | Answer |
|----------|--------|
| Can Concept A call Concept B? | No. Use an orchestrator. |
| Where does cross-concept logic go? | Orchestrators only. |
| Can I denormalize for performance? | Build a projection instead. |
| How do I reference another concept? | By ID only. Never embed. |
| Where do I validate input? | Orchestrators. |
| Where do I enforce business rules? | Concept methods. |
| Can GET routes change state? | No. POST/PUT/DELETE only. |
| What if an orchestrator fails mid-workflow? | Compensating actions. |
| Is LineItem a concept? | No. No independent lifecycle. Embed. |
| Is Inventory a concept? | Yes. Independent lifecycle + referenced by ID. |
| Where do I log audit events? | Orchestrators (mutations) and routes (auth events). |
| Can I store medical history? | No. Current injuries only, in separate restricted table. |
| How do I delete a member? | Anonymise PII, hard-delete medical, retain payments 7 years. |
| Can I use production data in tests? | No. Use synthetic seeded data only. |
