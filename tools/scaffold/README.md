# Scaffold CLI

Generate the initial shape of the application without adding business logic. The CLI creates concepts, orchestrators, projections, routes, and per-concept storage plus SQL migrations. All method bodies are empty except for TODO comments.

Run the CLI from the repository root.

## Quick Start

Create a concept, an orchestrator, a projection, and wire routes:

```powershell
go run ./tools/scaffold init ^
  --concept Order ^
  --field Order:Status:string ^
  --field Order:TotalCents:int ^
  --method Order:Approve ^
  --orchestrator CreateOrder ^
  --param CreateOrder:CustomerID:string ^
  --projection OrderSummary ^
  --query OrderSummary:OrderID:string ^
  --result OrderSummary:Status:string ^
  --route GET:/views/order-summary:OrderSummary ^
  --route POST:/orders:CreateOrder
```

This creates:
- `internal/domain/order/model.go`
- `internal/adapters/storage/order/store.go`
- `internal/application/orchestrators/create_order.go`
- `internal/application/projections/order_summary.go`
- `internal/adapters/http/routes.go`
- `internal/adapters/storage/migrations/*.sql`
- `.scaffold/state.json`

## Concepts

Create a concept with fields and methods:

```powershell
go run ./tools/scaffold init ^
  --concept InventoryItem ^
  --field InventoryItem:SKU:string ^
  --field InventoryItem:OnHand:int ^
  --method InventoryItem:Reserve
```

Notes:
- The tool always includes an `ID string` field.
- Concept storage stubs live under `internal/adapters/storage/<concept>`.
- A SQL create migration is generated for the concept.

## Orchestrators

Create an orchestrator with input parameters:

```powershell
go run ./tools/scaffold init ^
  --orchestrator ReserveInventory ^
  --param ReserveInventory:InventoryItemID:string ^
  --param ReserveInventory:Quantity:int
```

## Projections

Create a projection with query and result fields:

```powershell
go run ./tools/scaffold init ^
  --projection InventoryStatus ^
  --query InventoryStatus:InventoryItemID:string ^
  --result InventoryStatus:OnHand:int
```

## Routes

Add GET and POST routes:

```powershell
go run ./tools/scaffold init ^
  --route GET:/views/inventory-status:InventoryStatus ^
  --route POST:/inventory/reserve:ReserveInventory
```

Route rules:
- GET routes call projections only.
- POST/PUT/DELETE routes call orchestrators only.

## Updates and Migrations

Re-running the tool adds missing items and generates migrations for newly added fields.

```powershell
go run ./tools/scaffold init ^
  --concept Order ^
  --field Order:Notes:string
```

This adds an ALTER migration for the new field.

Use `--force` to overwrite existing files (including generated routes):

```powershell
go run ./tools/scaffold init ^
  --concept Order ^
  --force
```

The tool tracks state in `.scaffold/state.json`.
