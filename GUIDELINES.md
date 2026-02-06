# Application Guidelines

1) All symbols must be explicit and readable; they communicate what the data is or what the function does.
   Don't use abbreviations like `usr` or `amt`; do use names like `userID` and `paymentAmountCents`.
2) All concepts are independent and have their own storage. Each concept type owns its own persistence schema/table/collection. They may reference each other by id in the weakest way possible.
   Don't share tables or embed another concept's schema; do store only foreign IDs.
3) If we wish to view multiple concepts together we can create a separate projection. We must not couple the core concepts. Projections may be computed on the fly or materialized for performance, but they remain separate from concept storage.
   Don't denormalize another concept into a core concept; do build a projection or materialized view.
4) Concepts can have methods that change their own state. Those methods cannot have strong coupling to any other concept. Strong coupling includes direct method calls or references to other concept types, shared persistence models/schemas, or tight synchronous dependencies where one concept must call another to function.
   Don't call another concept or share its schema in a concept method; do keep the method self-contained and use IDs only.
5) The system is altered by orchestrators (commands), which are functions that only ever marshall traffic to concepts. Orchestrators can plumb between multiple concepts, coordinating workflows and performing mapping/aggregation as needed.
   Don't embed workflow logic inside concepts; do coordinate steps in an orchestrator.
6) The default view (coupled to a route) is a projection that marshalls query params and reads state (either concepts, or projections on concepts) but must not change state (query only). All GET routes are projections; POST, PUT, and DELETE routes may call orchestrators to change state.
   Don't mutate state in GET handlers; do handle reads only and move state changes to POST/PUT/DELETE via orchestrators.

7) Cross-cutting concerns: orchestrators validate inputs, concepts enforce their own invariants, and projections validate query params.
   Don't put input validation in concept methods; do validate input in orchestrators and enforce invariants in concepts.
8) Consistency across multiple concepts is eventual. Orchestrators use compensating actions (sagas) to resolve partial failures.
   Don't use cross-concept transactions; do design compensating actions for each step.
9) Concepts own their migrations. Projections can be rebuilt from concept state when needed.
   Don't migrate projection storage by hand; do rebuild or regenerate projections from concept data.
10) Orchestrators handle partial failures with compensating actions.
   Don't ignore partial failures or leave half-complete workflows; do define compensation per step.

Corollaries
- A concept must never call another concept directly; only orchestrators coordinate cross-concept workflows.
  Don't call `Order.AdjustInventory()` from `Order`; do call an orchestrator that coordinates `Order` and `Inventory`.
- Projections are read-only and can be rebuilt or discarded without affecting concept correctness.
  Don't treat projections as sources of truth; do rebuild projections from concept state.
- Cross-concept identifiers are plain IDs only; no shared schema objects or embedded concept types.
  Don't store full `User` objects inside `Order`; do store `UserID`.
- Route handlers for GET perform queries only; state changes are confined to POST/PUT/DELETE through orchestrators.
  Don't update records in GET routes; do expose mutations via command routes.
- Validation is layered: orchestrators validate inputs, concepts enforce invariants, projections validate query params.
  Don't validate query params inside concepts; do validate them at the projection boundary.
