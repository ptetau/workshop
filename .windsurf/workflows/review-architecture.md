---
description: Review changed files for architecture guideline compliance (concepts, orchestrators, projections, routes, storage)
---

# Architecture Review

Focused review against [GUIDELINES.md](../../GUIDELINES.md) §Concepts, §Orchestrators, §Projections, §Routes, §Storage.

## Steps

1. **Run lint check:**

// turbo
```powershell
go run ./tools/lintguidelines --root . --strict
```

2. **Review Concepts** (`internal/domain/*/model.go`) — for each changed concept:
   - Methods only modify own state — no side effects on other concepts
   - No imports from other `domain/` packages
   - References other concepts by ID only (`UserID string`, never `User User`)
   - Returns domain errors (`ErrAlreadyCancelled`), not generic strings
   - Entity has both independent lifecycle AND is referenceable by ID — otherwise embed it
   - PRE/POST/INVARIANT contracts documented on methods

   | Don't | Do |
   |-------|------|
   | `import "workshop/internal/domain/member"` in attendance | Keep each concept isolated — coordinate in orchestrators |
   | `User User` field in Order struct | `UserID string` — reference by ID only |
   | `errors.New("bad state")` | `var ErrAlreadyCancelled = errors.New("order already cancelled")` |
   | Make `Address` a concept | Embed in `Member` — no independent lifecycle |

3. **Review Orchestrators** (`internal/application/orchestrators/*.go`) — for each changed orchestrator:
   - Validates input before any processing
   - Coordinates across concepts (only place cross-concept logic lives)
   - Handles partial failures with compensating actions
   - Emits audit log entry via `slog.Info("audit_event", ...)`
   - Uses dependency injection via deps struct with store interfaces

   | Don't | Do |
   |-------|------|
   | `order.ProcessPayment()` calling inventory internally | `ExecutePlaceOrder()` coordinates Order + Inventory |
   | `var store = accountStore.New()` global | `type Deps struct { AccountStore AccountStoreForX }` |
   | Ignoring error from second Save | Define compensating action if first succeeded but second fails |
   | Skipping audit log | `slog.Info("audit_event", "actor_id", id, "action", "member.update", ...)` |

4. **Review Projections** (`internal/application/projections/*.go`):
   - Read-only — no state mutations
   - Validates query params
   - Uses `List()` from storage or cross-concept JOINs
   - Logs access to sensitive data (profile views, exports)

   | Don't | Do |
   |-------|------|
   | `member.SetStatus("active")` in a projection | Projections are read-only — use orchestrators for writes |
   | Return without validating query params | `if query.MemberID == "" { return ..., ErrMissing }` |

5. **Review Routes** (`internal/adapters/http/routes.go`, `handlers.go`):
   - GET handlers call `projections.Query*()`
   - POST/PUT/DELETE handlers call `orchestrators.Execute*()`
   - No business logic in handlers
   - `RequireRole()` middleware applied to every protected route
   - Auth events logged (login, logout, lockout)

   | Don't | Do |
   |-------|------|
   | GET handler calls `orchestrators.ExecuteCancel()` | GET → projection, POST → orchestrator |
   | `if account.Role == "admin"` in handler body | `RequireRole(account.RoleAdmin)` middleware on the route |
   | Business logic computing totals in handler | Move to projection or orchestrator |

6. **Review Storage** (`internal/adapters/storage/*/sqlite_store.go`):
   - One `Store` interface per concept (in `store.go`)
   - Pure data access — no business logic
   - Uses `?` parameterized placeholders
   - Transactions scoped to single concept (`BeginTx` + `defer tx.Rollback()`)
   - Upserts use `ON CONFLICT(id) DO UPDATE SET`
   - Value types, not pointers: `Save(ctx, entity domain.Member)`

   | Don't | Do |
   |-------|------|
   | `Save(ctx, m *member.Member)` | `Save(ctx, entity domain.Member)` — value types |
   | Business logic in store (checking status) | Pure data access only — logic goes in concept methods |
   | `memberStore` also writing to attendance table | One store per concept — orchestrators coordinate |

7. **Report** — summarise findings:
   - ✅ `lintguidelines --strict` passed / ❌ failures
   - List any violations with file path and line number
   - Recommendation: **Approve** / **Request Changes**
