---
description: Run the full code review checklist against changed files
---

# Code Review Workflow

Review code changes against all project guides: [Guidelines](../../GUIDELINES.md), [DB Guide](../../DB_GUIDE.md), [OWASP](../../OWASP.md), [Privacy](../../PRIVACY.md).

## Steps

1. **Run automated checks** — execute all verification commands and report results:

   ```powershell
   pwsh scripts/check-all.ps1
   ```

2. **Identify changed files** — determine which layers are affected by the changes:
   - `internal/domain/*/model.go` → Architecture §Concepts
   - `internal/application/orchestrators/*.go` → Architecture §Orchestrators
   - `internal/application/projections/*.go` → Architecture §Projections
   - `internal/adapters/http/routes.go` or `handlers.go` → Architecture §Routes
   - `internal/adapters/storage/**` → Architecture §Storage + Database checklist
   - `internal/adapters/storage/db.go` → Database §Schema
   - `internal/adapters/http/middleware/*.go` → Security checklist
   - Any file touching PII, injuries, observations, consent → Privacy checklist

3. **Architecture review** — for each changed layer, verify:
   - **Concepts**: methods only modify own state, no cross-domain imports, references by ID only, domain errors, PRE/POST/INVARIANT contracts
   - **Orchestrators**: validates input first, coordinates across concepts, compensating actions for failures, audit log emitted, dependency injection
   - **Projections**: read-only, validates query params, logs sensitive data access
   - **Routes**: GET→projection, POST/PUT/DELETE→orchestrator, no business logic, RequireRole middleware, auth events logged
   - **Storage**: one interface per concept, pure data access, parameterized queries, single-concept transactions, upserts with ON CONFLICT

4. **Database review** — if schema or queries changed, verify:
   - `CREATE TABLE IF NOT EXISTS` with `TEXT PRIMARY KEY`
   - Foreign keys, NOT NULL constraints, performance indexes
   - Restricted data (injuries, observations) in separate tables
   - All queries use `?` placeholders with context-aware methods
   - Writes wrapped in transactions, scoped to single concept
   - No cross-concept writes in a single transaction

5. **Security review (OWASP Top 10)** — for each changed area, verify:
   - **A01**: permissions enforced in middleware, roles from session
   - **A02**: bcrypt cost ≥ 12, HTTPS/HSTS, secrets from env vars, secure cookies
   - **A03**: parameterized SQL, no shell injection, html/template
   - **A04**: rate limiting, layered validation, account lockout
   - **A05**: no internal errors exposed, security headers, debug off in prod
   - **A06**: govulncheck passes, actions pinned by SHA
   - **A07**: password min 12 chars, forced change on seed, 24h session expiry
   - **A08**: strict JSON decoding, file upload validation
   - **A09**: structured slog, no secrets in logs, auth_event and audit_event keys
   - **A10**: URL allowlist, block private IPs

6. **Privacy review** — if PII, medical, or consent data is touched, verify:
   - PII accessed only through proper role checks
   - Medical data in separate table, coach/admin only
   - No production data in dev/test
   - Granular consent (never single checkbox), opt-in marketing
   - Deletion: anonymise PII, hard-delete medical, retain payments 7 years
   - Audit log: append-only, actor context, no secrets, 7-year retention
   - NZ context: 72h breach notification, data sovereignty, IRD 7-year retention

7. **Check Common Issues table** — scan for any of these anti-patterns:
   - Concept importing another concept → move to orchestrator
   - GET handler calling orchestrator → use projection
   - SQL string concatenation → use `?` placeholders
   - Password stored as SHA256 → use bcrypt cost ≥ 12
   - Stack traces returned to client → log internally, return generic message
   - Session in localStorage → use HttpOnly cookie
   - Missing rate limiting → add rate limiter middleware
   - Cross-concept write in one transaction → split into separate stores
   - Orchestrator using global stores → use deps struct with interfaces
   - `mattn/go-sqlite3` import → use `modernc.org/sqlite`
   - `?_journal_mode=WAL` DSN syntax → use `?_pragma=journal_mode(WAL)`
   - Pointer types in store methods (`*member.Member`) → use value types (`domain.Member`)
   - `INSERT OR REPLACE` → use `ON CONFLICT(id) DO UPDATE SET`
   - `DATETIME` columns → use `TEXT` with RFC3339 strings
   - `requireAdmin()` helper → use `RequireRole(account.RoleAdmin)`
   - Single "I agree" checkbox → separate granular consent per purpose
   - Hard-deleting all member data → anonymise PII, hard-delete medical, retain payments
   - Logging medical details → redact sensitive fields
   - Using production data in tests → use synthetic seeded data only

8. **Report findings** — summarise:
   - ✅ All automated checks passed / ❌ Failures found
   - List any checklist items that fail with file path and line number
   - List any Common Issues detected
   - Recommendation: **Approve** / **Request Changes** (with specific fixes needed)
