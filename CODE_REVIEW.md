# Code Review Guide

Checklist for reviewing code changes against our [Guidelines](GUIDELINES.md), [Database Guide](DB_GUIDE.md), [OWASP Top 10](OWASP.md), and [Privacy & Compliance](PRIVACY.md).

---

## Quick Reference

| Area | Check | Reference |
|------|-------|-----------|
| Architecture | Concepts don't call other concepts | [Guidelines §Concepts](GUIDELINES.md#concepts) |
| Architecture | Orchestrators coordinate workflows | [Guidelines §Orchestrators](GUIDELINES.md#orchestrators) |
| Architecture | GET → projection, POST/PUT/DELETE → orchestrator | [Guidelines §Routes](GUIDELINES.md#routes) |
| Architecture | Validation at every layer (routes, orchestrators, concepts) | [Guidelines §Validation](GUIDELINES.md#validation-layers) |
| Naming | No abbreviations (`usr`, `amt`, `cfg`) | [Guidelines §Naming](GUIDELINES.md#naming) |
| Database | WAL mode, busy timeout, foreign keys | [DB Guide §2](DB_GUIDE.md#2-configuration-internaladaptersstoragedbgo) |
| Database | Writes isolated to single concept | [DB Guide §4B](DB_GUIDE.md#b-writing-concept-stores) |
| Database | Projections allowed cross-concept JOINs | [DB Guide §4A](DB_GUIDE.md#a-reading-projections) |
| Security | Permissions checked in middleware | [OWASP A01](OWASP.md#a01-broken-access-control) |
| Security | Passwords hashed with bcrypt (cost ≥ 12) | [OWASP A02](OWASP.md#a02-cryptographic-failures) |
| Security | SQL uses parameterized queries | [OWASP A03](OWASP.md#a03-injection) |
| Security | Rate limiting on auth endpoints | [OWASP A04](OWASP.md#a04-insecure-design) |
| Security | Security headers (CSP, HSTS, X-Frame-Options) | [OWASP A05](OWASP.md#a05-security-misconfiguration) |
| Security | Dependencies audited with `govulncheck` | [OWASP A06](OWASP.md#a06-vulnerable-and-outdated-components) |
| Security | HttpOnly + Secure + SameSite cookies | [OWASP A07](OWASP.md#a07-identification-and-authentication-failures) |
| Security | Strict JSON decoding, no blind deserialization | [OWASP A08](OWASP.md#a08-software-and-data-integrity-failures) |
| Security | Structured logging, no secrets in logs | [OWASP A09](OWASP.md#a09-security-logging-and-monitoring-failures) |
| Security | URL allowlist, no fetching user-provided URLs | [OWASP A10](OWASP.md#a10-server-side-request-forgery-ssrf) |
| Privacy | Granular consent, never single "I agree" | [Privacy §2.1](PRIVACY.md#21-consent-management) |
| Privacy | Anonymise PII on deletion, hard-delete medical | [Privacy §2.2](PRIVACY.md#22-right-to-be-forgotten-deletion--anonymisation) |
| Privacy | Injuries in separate table, coach/admin only | [Privacy §3.1](PRIVACY.md#31-segregation) |
| Privacy | Audit log all mutations with actor context | [Privacy §1.3](PRIVACY.md#13-audit-logging) |

---

## 1. Architecture Checklist

### Concepts (`internal/domain/*/model.go`)
- [ ] Methods only modify own state
- [ ] No imports from other `domain/` packages
- [ ] References other concepts by ID only (never embed full struct)
- [ ] Returns domain errors (`ErrAlreadyCancelled`, not generic strings)
- [ ] Has both independent lifecycle AND is referenceable by ID
- [ ] PRE/POST/INVARIANT contracts documented on methods

### Orchestrators (`internal/application/orchestrators/*.go`)
- [ ] Validates input before any processing
- [ ] Coordinates across concepts (only place cross-concept logic lives)
- [ ] Handles partial failures with compensating actions
- [ ] Emits audit log entries for all data mutations
- [ ] Uses dependency injection (store interfaces, not globals)

### Projections (`internal/application/projections/*.go`)
- [ ] Read-only — no state mutations
- [ ] Validates query params
- [ ] Uses `List()` from storage or cross-concept JOINs
- [ ] Logs access to sensitive data (profile views, exports)

### Routes (`internal/adapters/http/routes.go`)
- [ ] GET handlers call `projections.Query*()`
- [ ] POST/PUT/DELETE handlers call `orchestrators.Execute*()`
- [ ] No business logic in handlers
- [ ] `RequireRole()` middleware applied to every protected route
- [ ] Auth events logged (login, logout, lockout)

### Storage (`internal/adapters/storage/*/store.go`)
- [ ] One `Store` interface per concept
- [ ] Pure data access — no business logic
- [ ] Uses parameterized queries (`?` placeholders)
- [ ] Transactions scoped to single concept (`BeginTx` + `defer tx.Rollback()`)
- [ ] Upserts use `ON CONFLICT(id) DO UPDATE SET`

---

## 2. Database Checklist

### Schema (`internal/adapters/storage/db.go`)
- [ ] New tables use `CREATE TABLE IF NOT EXISTS`
- [ ] `TEXT PRIMARY KEY` for IDs (UUIDs)
- [ ] Foreign keys declared with `FOREIGN KEY ... REFERENCES`
- [ ] Performance indexes for frequently queried columns
- [ ] `NOT NULL` constraints on required fields
- [ ] Restricted data (injuries, observations) in separate tables

### Queries
- [ ] All queries use `?` parameterized placeholders — never string concatenation
- [ ] `QueryRowContext`/`QueryContext` used (context-aware)
- [ ] Writes wrapped in transactions
- [ ] No cross-concept writes in a single transaction
- [ ] Projections may JOIN across tables for reads

### Operations
- [ ] Never delete `.db-wal` or `.db-shm` while app is running
- [ ] Backups use Litestream or `VACUUM INTO` — never raw file copy

---

## 3. Security Checklist (OWASP Top 10)

### A01 — Access Control
- [ ] Permissions enforced in middleware on every request
- [ ] Roles extracted from authenticated session, never from client
- [ ] Admin-only endpoints use `RequireRole(account.RoleAdmin)`

### A02 — Cryptography
- [ ] Passwords hashed with bcrypt (cost ≥ 12)
- [ ] HTTPS enforced; HSTS headers set
- [ ] Secrets (API keys, CSRF key) from env vars, never hardcoded
- [ ] Cookies: `HttpOnly=true`, `Secure=true`, `SameSite=Strict`

### A03 — Injection
- [ ] SQL uses `?` placeholders (no string formatting)
- [ ] No `sh -c` with user input — use `exec.Command` with args
- [ ] Uses `html/template` (auto-escapes), not `text/template`

### A04 — Insecure Design
- [ ] Rate limiting on auth and sensitive endpoints
- [ ] Validation at every layer (routes → orchestrators → concepts)
- [ ] Account lockout after 5 failed attempts (15 min)

### A05 — Security Misconfiguration
- [ ] Internal errors never exposed to clients (log + return generic message)
- [ ] Security headers: CSP, X-Frame-Options, X-Content-Type-Options, Referrer-Policy
- [ ] Debug mode disabled in production (`WORKSHOP_ENV=production`)
- [ ] No directory listing on static file server

### A06 — Vulnerable Components
- [ ] `govulncheck ./...` passes in CI
- [ ] Dependencies reviewed before adding
- [ ] GitHub Actions pinned by full commit SHA

### A07 — Authentication Failures
- [ ] Password minimum 12 characters
- [ ] Forced password change on seeded accounts
- [ ] Session expires after 24 hours
- [ ] Session tokens are cryptographically random (32 bytes)

### A08 — Data Integrity
- [ ] Strict JSON decoding with `DisallowUnknownFields()`
- [ ] CI/CD: signed commits, protected branches, reviewed PRs
- [ ] File uploads validated (type, size, content)

### A09 — Logging Failures
- [ ] Structured logging with `log/slog`
- [ ] No passwords, tokens, or medical details in logs
- [ ] Auth events logged with account ID (`auth_event` key)
- [ ] All data mutations logged with actor context (`audit_event` key)

### A10 — SSRF
- [ ] URL allowlist for any external fetches
- [ ] Block requests to private IP ranges (127.x, 10.x, 172.16.x, 192.168.x)

---

## 4. Privacy & Compliance Checklist

### Data Handling
- [ ] PII accessed only through proper role checks (member → own, coach/admin → any)
- [ ] Medical/injury data in separate table with coach/admin-only access
- [ ] No production data in dev/test — use synthetic seed only
- [ ] Data classification respected (Public / Internal / Confidential / Restricted / Financial)

### Consent
- [ ] Consent is granular (separate checkboxes per purpose)
- [ ] Marketing consent is opt-in (never pre-checked)
- [ ] Waiver version tracked — re-prompt on update
- [ ] Consent records stored with version, timestamp, IP

### Deletion & Export
- [ ] Member deletion: anonymise PII, hard-delete medical, retain payments 7 years
- [ ] Deletion has 30-day grace period
- [ ] Data export available (JSON/CSV) — excludes coach observations
- [ ] Audit logs retained with anonymised actor reference

### Audit Logging
- [ ] Audit log is append-only (no updates or deletes)
- [ ] Logs: actor_id, actor_role, action, resource_type, resource_id, timestamp
- [ ] Never log passwords, tokens, or full medical details
- [ ] Retain audit logs for 7 years (NZ IRD alignment)

### NZ Context
- [ ] Breach notification plan: report to NZ Privacy Commissioner within 72 hours
- [ ] Data sovereignty: hosted in AU/NZ region or on-premises
- [ ] Payment records retained 7 years (IRD requirement)

---

## 5. Review Commands

```powershell
# Verify guideline compliance
go run ./tools/lintguidelines --root . --strict

# Check for vulnerabilities
govulncheck ./...

# Run tests with race detector
go test -race -count=1 ./...

# Build to catch compile errors
go build ./...

# Format check
gofmt -l .
```

---

## 6. Common Issues

| Issue | Fix | Reference |
|-------|-----|-----------|
| Concept imports another concept | Move coordination to orchestrator | [Guidelines §Concepts](GUIDELINES.md#concepts) |
| GET handler calls orchestrator | Use projection for reads | [Guidelines §Routes](GUIDELINES.md#routes) |
| Orchestrator uses global store variables | Use a deps struct with store interfaces | [Guidelines §Orchestrators](GUIDELINES.md#orchestrators) |
| SQL string concatenation | Use parameterized query (`?`) | [OWASP A03](OWASP.md#a03-injection) |
| Password stored as SHA256 | Use bcrypt (cost ≥ 12) | [OWASP A02](OWASP.md#a02-cryptographic-failures) |
| Error returned with stack trace | Log internally, return generic message | [OWASP A05](OWASP.md#a05-security-misconfiguration) |
| Session in localStorage | Use HttpOnly cookie | [OWASP A07](OWASP.md#a07-identification-and-authentication-failures) |
| No rate limiting on auth | Add rate limiter middleware | [OWASP A04](OWASP.md#a04-insecure-design) |
| `requireAdmin()` helper | Use `RequireRole(account.RoleAdmin)` — one pattern for all roles | [OWASP A01](OWASP.md#a01-broken-access-control) |
| Using `mattn/go-sqlite3` | Use `modernc.org/sqlite` (pure Go, no CGO) | [DB Guide §2](DB_GUIDE.md#2-configuration-internaladaptersstoragedbgo) |
| `?_journal_mode=WAL` DSN syntax | Use `?_pragma=journal_mode(WAL)` for modernc driver | [DB Guide §2](DB_GUIDE.md#2-configuration-internaladaptersstoragedbgo) |
| Pointer types in store methods | Use value types: `Save(ctx, entity domain.Member)` | [DB Guide §4B](DB_GUIDE.md#b-writing-concept-stores) |
| `INSERT OR REPLACE` (deletes then re-inserts) | Use `ON CONFLICT(id) DO UPDATE SET` (true upsert) | [DB Guide §4B](DB_GUIDE.md#b-writing-concept-stores) |
| `DATETIME` column type for timestamps | Use `TEXT NOT NULL` with RFC3339 strings | [DB Guide §3](DB_GUIDE.md#3-schema) |
| Cross-concept write in one TX | Split into separate stores, use compensating actions | [DB Guide §4B](DB_GUIDE.md#b-writing-concept-stores) |
| Missing performance index | Add index on frequently queried columns | [DB Guide §3](DB_GUIDE.md#3-schema) |
| Single "I agree" checkbox | Separate granular consent per purpose | [Privacy §2.1](PRIVACY.md#21-consent-management) |
| Hard-deleting all member data | Anonymise PII, hard-delete medical, retain payments | [Privacy §2.2](PRIVACY.md#22-right-to-be-forgotten-deletion--anonymisation) |
| Logging medical details | Redact sensitive fields from logs | [Privacy §1.3](PRIVACY.md#13-audit-logging) |
| Using production data in tests | Use synthetic seeded data only | [Privacy §5](PRIVACY.md#5-developer-checklist) |

---

## 7. Approval Criteria

**Approve if all applicable items pass:**

- [ ] `go build ./...` succeeds
- [ ] `gofmt -l .` returns no output
- [ ] `go run ./tools/lintguidelines --root . --strict` passes
- [ ] `govulncheck ./...` reports no actionable issues
- [ ] `go test -race -count=1 ./...` passes
- [ ] No items from Common Issues table
- [ ] Architecture checklist complete for changed layers
- [ ] Security checklist complete for changed areas
- [ ] Privacy checklist complete if PII or medical data is touched
- [ ] Database checklist complete if schema or queries changed
