---
description: Review changed files for OWASP Top 10 security compliance
---

# OWASP Security Review

Focused review against [OWASP.md](../../OWASP.md) — the OWASP Top 10 (2025) checklist for Go.

## Steps

1. **Run vulnerability check:**

```powershell
govulncheck ./...
```

2. **A01 — Access Control** — if routes or middleware changed:
   - Permissions enforced in middleware on every request via `RequireRole()`
   - Roles extracted from authenticated session (`GetSessionFromContext()`), never from client
   - Admin-only endpoints use `RequireRole(account.RoleAdmin)`

   | Don't | Do |
   |-------|------|
   | `if r.Header.Get("X-Role") == "admin"` | `session, _ := GetSessionFromContext(r.Context())` |
   | Per-handler role checks | `RequireRole(account.RoleAdmin)` middleware on the route |
   | Unprotected admin endpoint | Every route has `RequireRole()` or is public (login, static) |

3. **A02 — Cryptography** — if auth or secrets touched:
   - Passwords hashed with bcrypt (cost ≥ 12)
   - HTTPS enforced; HSTS headers set (Caddy handles this)
   - Secrets (API keys, CSRF key) from env vars, never hardcoded
   - Cookies: `HttpOnly=true`, `Secure=true`, `SameSite=Strict`

   | Don't | Do |
   |-------|------|
   | `sha256.Sum256([]byte(password))` | `bcrypt.GenerateFromPassword([]byte(password), 12)` |
   | `csrfKey := "hardcoded-secret"` | `csrfKey := os.Getenv("WORKSHOP_CSRF_KEY")` |
   | `Secure: false` on cookies | `Secure: true` — Caddy terminates TLS |

4. **A03 — Injection** — if queries or templates changed:
   - SQL uses `?` parameterized placeholders (no string formatting)
   - No `sh -c` with user input — use `exec.Command` with args
   - Uses `html/template` (auto-escapes), not `text/template`

   | Don't | Do |
   |-------|------|
   | `fmt.Sprintf("WHERE id = '%s'", id)` | `db.QueryRowContext(ctx, "WHERE id = ?", id)` |
   | `exec.Command("sh", "-c", userInput)` | `exec.Command("convert", inputFile, outputFile)` |
   | `text/template` for HTML | `html/template` — auto-escapes all output |

5. **A04 — Insecure Design** — if auth or validation changed:
   - Per-IP rate limiter on all endpoints (10 req/sec via `middleware.NewRateLimiter`)
   - Validation at every layer (routes → orchestrators → concepts)
   - Account lockout after 5 failed attempts (15 min)

   | Don't | Do |
   |-------|------|
   | No rate limit on `/login` | Rate limiter middleware applied globally |
   | Only validate in handler | Validate in orchestrator (input) AND concept (invariants) |
   | Unlimited login attempts | `account.RecordFailedLogin()` → lock after 5 failures |

6. **A05 — Security Misconfiguration** — if error handling or headers changed:
   - Internal errors never exposed to clients (log + return generic message)
   - Security headers: CSP, X-Frame-Options, X-Content-Type-Options, Referrer-Policy
   - Debug mode disabled in production (`WORKSHOP_ENV=production`)
   - Seeded accounts use `PasswordChangeRequired=true`

   | Don't | Do |
   |-------|------|
   | `http.Error(w, err.Error(), 500)` | `slog.Error("...", "err", err); http.Error(w, "Internal error", 500)` |
   | Missing security headers | `middleware.SecurityHeaders` applied in middleware chain |
   | Default password left unchanged | `PasswordChangeRequired=true` forces change on first login |

7. **A06 — Vulnerable Components** — if dependencies changed:
   - `govulncheck ./...` passes
   - Dependencies reviewed before adding
   - GitHub Actions pinned by full commit SHA

   | Don't | Do |
   |-------|------|
   | `uses: actions/checkout@v4` | `uses: actions/checkout@<full-sha>` |
   | Adding dependency without review | Check maintenance status, license, vulnerability history |

8. **A07 — Authentication** — if session or login changed:
   - Password minimum 12 characters
   - Forced password change on seeded accounts
   - Session expires after 24 hours
   - Session tokens are 32-byte cryptographically random hex
   - Cookie name: `workshop_session`

   | Don't | Do |
   |-------|------|
   | `MaxAge: 3600` (1 hour) | `MaxAge: 86400` (24 hours, matches server-side expiry) |
   | `rand.Int()` for session token | `crypto/rand.Read(b)` → 32 bytes → hex encoded |
   | Allow seeded admin to skip password change | `PasswordChangeRequired: true` on seed |

9. **A08 — Data Integrity** — if deserialization or CI changed:
   - Strict JSON decoding with `DisallowUnknownFields()`
   - CI/CD: protected branches, reviewed PRs, status checks required
   - File uploads validated (type, size, content)

   | Don't | Do |
   |-------|------|
   | `json.Unmarshal(body, &input)` | `decoder.DisallowUnknownFields(); decoder.Decode(&input)` |
   | Allow direct push to main | PR required with CI status check |

10. **A09 — Logging** — if logging changed:
    - Structured logging with `log/slog`
    - No passwords, tokens, or medical details in logs
    - Auth events use `auth_event` key; data mutations use `audit_event` key

    | Don't | Do |
    |-------|------|
    | `log.Printf("login: %s pass: %s", email, password)` | `slog.Info("auth_event", "event", "login_success", "email", email)` |
    | `slog.Info("injury", "details", injuryNotes)` | `slog.Info("audit_event", "action", "injury.update", "resource_id", id)` |

11. **A10 — SSRF** — if external HTTP calls added:
    - URL allowlist for any external fetches
    - Block requests to private IP ranges (127.x, 10.x, 172.16.x, 192.168.x)

    | Don't | Do |
    |-------|------|
    | `http.Get(userProvidedURL)` | Parse URL, check against allowlist, block private IPs |

12. **Report** — summarise findings:
    - ✅ `govulncheck` passed / ❌ vulnerabilities found
    - List any OWASP violations with file path, line number, and category (A01–A10)
    - Recommendation: **Approve** / **Request Changes**
