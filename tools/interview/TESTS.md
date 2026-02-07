# Interview CLI - Simulation & Exploratory Tests

Interactive PRD-to-scaffold graph generator. Validates tool fitness for converting feature requests into compiled-ready applications.

## Quick Reference

```powershell
# Run all unit tests
go test ./tools/interview -v

# Run pairwise combination tests only
go test ./tools/interview -v -run Pairwise
```

---

## Current Test Coverage

| Test Suite | Coverage Focus |
|------------|----------------|
| `main_test.go` | `symbolify`, `buildScaffoldArgs` output |
| `pairwise_test.go` | All 256 input combinations, disambiguation flows |

---

## Exploratory Test Scenarios

### 1. Basic Interview Flow

**Purpose**: Verify minimal complete interview produces valid graph

```
Input Sequence:
- Concept: "Order" → done (fields) → done (methods) → done (concepts)
- Orchestrator: "Create Order" → done (params) → skip (route) → done
- Projection: "Order Summary" → done (query) → done (result) → skip (route) → done

Expected:
✓ Graph contains 1 concept named "Order"
✓ Graph contains 1 orchestrator named "CreateOrder"
✓ Graph contains 1 projection named "OrderSummary"
✓ ScaffoldArgs includes --concept, --orchestrator, --projection
```

### 2. Field Type Validation

**Purpose**: Confirm all supported types scaffold correctly

| Input Type | Expected Output |
|------------|-----------------|
| `string` | `string` |
| `int` | `int` |
| `bool` | `bool` |
| `time` | `time` |
| `custom` | passes through as-is |

```
Input:
- Concept: "Event" → Field: "Name" → Type: "string" 
- Field: "Timestamp" → Type: "time" → done (fields) → done

Expected: Fields array contains [{Name: "Name", Type: "string"}, {Name: "Timestamp", Type: "time"}]
```

### 3. Disambiguation Threshold Test

**Purpose**: Verify Levenshtein similarity triggers disambiguates at 0.8 threshold

| Candidate | Existing | Similarity | Triggers? |
|-----------|----------|------------|-----------|
| `Order` | `Orderr` | 0.83 | Yes |
| `Customer` | `Client` | 0.25 | No |
| `User` | `Users` | 0.80 | Yes |
| `Payment` | `Peyment` | 0.86 | Yes |

```
Input:
- Concept: "Order" → done → done → "Orderr" → "yes" → done
- Expected: Graph contains exactly 1 concept (merged)

- Concept: "Order" → done → done → "Orderr" → "no" → done → done → done
- Expected: Graph contains exactly 2 concepts
```

### 4. Route Method Validation

**Purpose**: Verify orchestrator routes accept POST/PUT/DELETE only

```
Input:
- Orchestrator: "CreateOrder" → done (params) → "POST" → "/orders"
  Expected: Route {Method: "POST", Path: "/orders", Target: "CreateOrder"}

- Orchestrator: "UpdateOrder" → done (params) → "PUT" → "/orders/{id}"
  Expected: Route {Method: "PUT", Path: "/orders/{id}", Target: "UpdateOrder"}

- Orchestrator: "DeleteOrder" → done (params) → "DELETE" → "/orders/{id}"
  Expected: Route {Method: "DELETE", Path: "/orders/{id}", Target: "DeleteOrder"}

- Orchestrator: "QueryOrder" → done (params) → "skip"
  Expected: No route generated
```

### 5. Projection GET Routes

**Purpose**: Verify projections only generate GET routes

```
Input:
- Projection: "OrderDetail" → done (query) → done (result) → "/views/order-detail"
  Expected: Route {Method: "GET", Path: "/views/order-detail", Target: "OrderDetail"}
```

### 6. Symbol Conversion

**Purpose**: Verify phrase-to-symbol conversion

| Input Phrase | Expected Symbol |
|--------------|-----------------|
| `order summary` | `OrderSummary` |
| `create-order` | `CreateOrder` |
| `user_profile` | `UserProfile` |
| `GET items` | `GetItems` |
| `multi word phrase here` | `MultiWordPhraseHere` |

### 7. Graph Output Files

**Purpose**: Verify graph.json and graph.dot are written correctly

```powershell
# After interview runs with --out .scaffold/test/graph
Test:
✓ .scaffold/test/graph.json exists and is valid JSON
✓ .scaffold/test/graph.dot exists and starts with "digraph Scaffold {"
✓ graph.json contains generatedAt timestamp
✓ graph.json contains source text (PRD content)
```

### 8. Empty Interview

**Purpose**: Verify tool handles immediate done gracefully

```
Input:
- done (concepts) → done (orchestrators) → done (projections)

Expected:
✓ Empty graph with no errors
✓ ScaffoldArgs is empty array
✓ graph.json still written with generatedAt
```

---

## Edge Cases to Explore

### Input Validation

| Scenario | Input | Expected Behavior |
|----------|-------|-------------------|
| Multi-word concept | `order item` | Prompt: "One word only." (retry) |
| Empty phrase | `` | Prompt: "Please provide a short phrase" |
| Done as name | `done` | Ends section |
| Yes/no typo | `ye` | Prompt: "Please answer 'yes' or 'no'." |

### Special Characters

| Input | Behavior |
|-------|----------|
| `Order#1` | Symbolify to `Order1` |
| `Create-Order` | Symbolify to `CreateOrder` |
| Path `/api/v1/orders` | Preserved in route |

---

## Integration Verification

After interview completes, verify scaffold executes:

```powershell
# Interview generates scaffold command
Get-Content PRD.md | go run ./tools/interview --root C:\temp\test --module testapp

# Verify scaffold was invoked with generated args
Test:
✓ internal/domain/<concept>/model.go exists
✓ internal/application/orchestrators/<orch>.go exists  
✓ internal/application/projections/<proj>.go exists
```

---

## Manual Smoke Test

```powershell
# Run interactive interview with sample PRD
echo "Build an e-commerce order system" | go run ./tools/interview --root C:\temp\manual --module ecommerce

# Interactive Prompts - Enter:
Concept: Order
Description: Order domain entity
Field: Status → string → Description: Order status → TotalCents → int → Description: Total amount → done
Method: Approve → Description: Approves order → Invariant: status=approved → Pre: status=new → Post: status=approved → Cancel → Description: Cancels order → Invariant: status=cancelled → Pre: status!=shipped → Post: status=cancelled → done
done (concepts)

Orchestrator: Create Order
Description: Creates a new order
Param: CustomerID → string → Description: The customer ID → done
Route: POST → /orders
done (orchestrators)

Projection: Order Summary
Description: Summary view of orders
Query: OrderID → string → Description: filter by ID → done
Result: Status → string → Description: status → Total → int → Description: total → done
Route: /views/order-summary
done (projections)

# Verify outputs exist and scaffold ran
dir C:\temp\manual\internal\domain\order
dir C:\temp\manual\internal\application
dir C:\temp\manual\.scaffold\interview
```
