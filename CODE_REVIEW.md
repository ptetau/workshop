# Code Review Guide

Checklist for reviewing code changes against our [Guidelines](GUIDELINES.md) and [OWASP Top 10](OWASP.md).

---

## Quick Reference

| Area | Check | Reference |
|------|-------|-----------|
| Architecture | Concepts don't call other concepts | [Guidelines: Building Blocks](GUIDELINES.md#concepts) |
| Architecture | Orchestrators coordinate workflows | [Guidelines: Orchestrators](GUIDELINES.md#orchestrators) |
| Architecture | GET → projection, POST/PUT/DELETE → orchestrator | [Guidelines: Routes](GUIDELINES.md#routes) |
| Naming | No abbreviations (usr, amt, cfg) | [Guidelines: Naming](GUIDELINES.md#naming) |
| Security | Permissions checked in middleware | [OWASP: A01](OWASP.md#a01-broken-access-control) |
| Security | Passwords hashed with bcrypt | [OWASP: A02](OWASP.md#a02-cryptographic-failures) |
| Security | SQL uses parameterized queries | [OWASP: A03](OWASP.md#a03-injection) |
| Security | Structured logging, no secrets | [OWASP: A09](OWASP.md#a09-security-logging-and-monitoring-failures) |

---

## Architecture Checklist

### Concepts
- [ ] Methods only modify own state
- [ ] No imports from other `domain/` packages
- [ ] References other concepts by ID only
- [ ] Returns domain errors (`ErrAlreadyCancelled`)

### Orchestrators
- [ ] Validates input before processing
- [ ] Coordinates multiple concepts
- [ ] Translates domain errors to HTTP errors
- [ ] Handles partial failures with compensation

### Projections
- [ ] Read-only (no state mutations)
- [ ] Validates query params
- [ ] Uses `List()` from storage

### Routes
- [ ] GET handlers call `projections.Query*()`
- [ ] POST/PUT/DELETE handlers call `orchestrators.Execute*()`
- [ ] No business logic in handlers

### Storage
- [ ] One `Store` interface per concept
- [ ] Pure data access (no business logic)
- [ ] Uses parameterized queries

---

## Security Checklist

### Authentication & Authorization
- [ ] Middleware checks permissions on every request
- [ ] Roles extracted from authenticated session
- [ ] HttpOnly + Secure + SameSite cookies
- [ ] Account lockout after failed attempts

### Data Protection
- [ ] Passwords hashed with bcrypt (cost ≥ 12)
- [ ] Sensitive data encrypted at rest
- [ ] HTTPS enforced; HSTS headers set

### Injection Prevention
- [ ] SQL uses `?` placeholders
- [ ] No `sh -c` with user input
- [ ] Uses `html/template` (auto-escapes)

### Input Validation
- [ ] Strict JSON decoding (`DisallowUnknownFields`)
- [ ] URL validation before fetching
- [ ] File upload validation (type, size, content)

### Logging & Monitoring
- [ ] Structured logging with `log/slog`
- [ ] No passwords/tokens in logs
- [ ] Auth events logged with user ID and IP

---

## Review Commands

```powershell
# Verify guideline compliance
go run ./tools/lintguidelines --root . --strict

# Check for vulnerabilities
govulncheck ./...

# Run tests
go test ./... -v

# Build to catch compile errors
go build ./...
```

---

## Common Issues

| Issue | Fix |
|-------|-----|
| Concept imports another concept | Move coordination to orchestrator |
| GET handler calls orchestrator | Use projection for reads |
| SQL string concatenation | Use parameterized query |
| Password stored as SHA256 | Use bcrypt |
| Error returned with stack trace | Log internally, return generic message |
| Session in localStorage | Use HttpOnly cookie |
| No rate limiting on auth endpoints | Add rate limiter middleware |

---

## Approval Criteria

**Approve if:**
- [ ] Passes `lintguidelines --strict`
- [ ] Passes `govulncheck`
- [ ] All tests pass
- [ ] No items from Common Issues
- [ ] Security checklist complete for changed areas
