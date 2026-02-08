# OWASP Top 10 (2025) Implementation Checklist for Go

Concrete implementation steps for meeting the OWASP Top 10 2025 (Release Candidate) standards in a Go web application.

---

## A01: Broken Access Control

**Goal:** Enforce permissions server-side, never relying on the UI.

| Don't | Do |
|-------|------|
| Check permissions only in UI | Enforce in middleware on every request |
| Trust client-provided role claims | Validate roles from authenticated session |
| Use role checks scattered in handlers | Centralize in middleware |

- [ ] **Middleware Enforcement:** Check permissions on *every* request, not just at the handler level
- [ ] **Context-Based Checks:** Extract user roles from the request context

```go
// Centralized role check — use RequireRole() middleware, not per-handler checks
func RequireRole(roles ...string) func(http.Handler) http.Handler {
    roleSet := make(map[string]bool, len(roles))
    for _, r := range roles {
        roleSet[r] = true
    }
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            session, ok := GetSessionFromContext(r.Context())
            if !ok || !roleSet[session.Role] {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

---

## A02: Cryptographic Failures

**Goal:** Protect sensitive data in transit and at rest.

| Don't | Do |
|-------|------|
| Store passwords in plaintext | Use bcrypt with cost ≥ 12 |
| Use MD5/SHA1 for passwords | Use bcrypt, scrypt, or Argon2 |
| Transmit secrets over HTTP | Enforce HTTPS everywhere |

- [ ] **Password Hashing:** Use `golang.org/x/crypto/bcrypt`
- [ ] **TLS Enforcement:** Redirect HTTP to HTTPS; set HSTS headers

```go
import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
    return string(bytes), err
}

func CheckPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

---

## A03: Injection

**Goal:** Prevent untrusted data from being interpreted as code.

| Don't | Do |
|-------|------|
| Concatenate SQL strings | Use parameterized queries |
| Build shell commands from user input | Avoid shells; use exec with args |
| Embed user input in templates unsafely | Use html/template (auto-escapes) |

- [ ] **SQL Injection:** Always use `?` placeholders or named parameters
- [ ] **Command Injection:** Never use `sh -c`; pass args directly to exec
- [ ] **Template Injection:** Use `html/template`, not `text/template`

```go
// SQL: parameterized query
row := db.QueryRow("SELECT * FROM users WHERE id = ?", userID)

// Command: no shell, direct args
cmd := exec.Command("convert", inputFile, outputFile)

// Template: auto-escaped
tmpl := template.Must(template.ParseFiles("page.html"))
tmpl.Execute(w, data)
```

---

## A04: Insecure Design

**Goal:** Build security into the architecture, not as an afterthought.

| Don't | Do |
|-------|------|
| Add security checks after development | Design with threat modeling upfront |
| Allow unlimited retries | Implement rate limiting |
| Trust all internal services | Apply zero-trust between services |

- [ ] **Threat Modeling:** Document threats during design phase
- [ ] **Rate Limiting:** Per-IP rate limiter on all endpoints (10 req/sec)
- [ ] **Defense in Depth:** Validate at every layer (routes, orchestrators, concepts)

```go
// Per-IP rate limiter applied as middleware (middleware.NewRateLimiter)
limiter := middleware.NewRateLimiter(10, time.Second) // 10 req/sec per IP

return middleware.Chain(mux,
    middleware.SecurityHeaders,
    middleware.CSRF(csrfKey),
    middleware.Auth(sessions),
    middleware.RateLimit(limiter),
)
```

---

## A05: Security Misconfiguration

**Goal:** Secure defaults, minimal attack surface.

| Don't | Do |
|-------|------|
| Run with debug mode in production | Disable debug; use environment flags |
| Expose stack traces to users | Log internally; return generic errors |
| Leave default credentials unchanged | Seed with `PasswordChangeRequired=true` to force change on first login |

- [ ] **Error Handling:** Never expose internal errors to clients
- [ ] **Security Headers:** Set CSP, X-Frame-Options, X-Content-Type-Options
- [ ] **Minimal Exposure:** Disable directory listing; remove unused endpoints

```go
func SecureHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("Content-Security-Policy", "default-src 'self'")
        next.ServeHTTP(w, r)
    })
}
```

---

## A06: Vulnerable and Outdated Components

**Goal:** Keep dependencies updated and audited.

| Don't | Do |
|-------|------|
| Pin to old versions indefinitely | Run `go get -u` regularly |
| Ignore vulnerability reports | Use `govulncheck` in CI |
| Import abandoned packages | Prefer well-maintained dependencies |

- [ ] **Dependency Audit:** Run `govulncheck ./...` in CI pipeline
- [ ] **Update Cadence:** Review and update dependencies monthly
- [ ] **Minimal Dependencies:** Avoid unnecessary third-party packages

```powershell
# Install govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest

# Run vulnerability check
govulncheck ./...
```

---

## A07: Identification and Authentication Failures

**Goal:** Strong authentication, secure session management.

| Don't | Do |
|-------|------|
| Allow weak passwords | Enforce minimum length + complexity |
| Store session tokens in localStorage | Use HttpOnly, Secure cookies |
| Allow unlimited login attempts | Lock accounts after failures |

- [ ] **Session Security:** HttpOnly + Secure + SameSite cookies
- [ ] **Password Policy:** Minimum 12 characters, check against breached passwords
- [ ] **Account Lockout:** Temporary lockout after 5 failed attempts

```go
http.SetCookie(w, &http.Cookie{
    Name:     "workshop_session",
    Value:    token,            // 32-byte cryptographically random hex
    HttpOnly: true,
    Secure:   true,             // HTTPS only (Caddy handles TLS)
    SameSite: http.SameSiteStrictMode,
    Path:     "/",
    MaxAge:   86400,            // 24 hours — matches server-side session expiry
})
```

---

## A08: Software and Data Integrity Failures

**Goal:** Verify integrity of code and data.

| Don't | Do |
|-------|------|
| Trust unsigned updates | Verify checksums/signatures |
| Deserialize untrusted data blindly | Validate structure before processing |
| Allow arbitrary file uploads | Validate type, size, and content |

- [ ] **Checksum Verification:** Verify hashes for external downloads
- [ ] **CI/CD Security:** Sign commits; protected branches; reviewed PRs
- [ ] **Deserialization:** Use strict JSON decoding with known structs

```go
// Strict JSON decoding - fails on unknown fields
decoder := json.NewDecoder(r.Body)
decoder.DisallowUnknownFields()
if err := decoder.Decode(&input); err != nil {
    http.Error(w, "Invalid request", http.StatusBadRequest)
    return
}
```

---

## A09: Security Logging and Monitoring Failures

**Goal:** Detect and respond to attacks.

| Don't | Do |
|-------|------|
| Log to stdout only | Use structured logging to persistent storage |
| Log passwords or tokens | Redact sensitive fields |
| Ignore failed logins | Alert on suspicious patterns |

- [ ] **Structured Logging:** Use `log/slog` with JSON output
- [ ] **Audit Trail:** Log authentication events, permission denials, data changes
- [ ] **Alerting:** Set up alerts for anomalous patterns

```go
import "log/slog"

// Auth events use the "auth_event" key
slog.Info("auth_event", "event", "login_success", "email", input.Email, "role", acct.Role)
slog.Info("auth_event", "event", "login_failed", "email", input.Email, "reason", "wrong_password")

// Data mutations use the "audit_event" key
slog.Info("audit_event", "actor_id", actorID, "action", "member.update", "resource_id", memberID)
```

---

## A10: Server-Side Request Forgery (SSRF)

**Goal:** Prevent attackers from making requests on your behalf.

| Don't | Do |
|-------|------|
| Fetch arbitrary user-provided URLs | Allowlist permitted domains |
| Allow requests to internal networks | Block private IP ranges |
| Follow redirects blindly | Limit or disable redirects |

- [ ] **URL Validation:** Parse and validate before fetching
- [ ] **Allowlist:** Only permit known external domains
- [ ] **Network Isolation:** Block requests to 127.0.0.1, 10.x, 172.16.x, 192.168.x

```go
import "net/url"

func IsAllowedURL(rawURL string) bool {
    parsed, err := url.Parse(rawURL)
    if err != nil {
        return false
    }
    
    allowlist := []string{"api.example.com", "cdn.example.com"}
    for _, allowed := range allowlist {
        if parsed.Host == allowed {
            return true
        }
    }
    return false
}
```

---

## Integration with Guidelines

These security controls integrate with our architecture:

| Security Control | Architecture Layer |
|-----------------|-------------------|
| Access control middleware | Routes |
| Input validation | Orchestrators |
| Business rule enforcement | Concepts |
| Query param validation | Projections |
| Secure storage | Storage adapters |

Run security checks:
```powershell
# Lint for guideline compliance
go run ./tools/lintguidelines --root . --strict

# Check for vulnerabilities
govulncheck ./...

# Run tests including security scenarios
go test ./... -v
```
