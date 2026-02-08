# Data Privacy & Compliance Guide

Workshop Jiu Jitsu CRM privacy and compliance requirements. As both **Controller** (defining *why* data is collected) and **Processor** (defining *how* it is handled), we embed Privacy by Design from day one.

Complying with GDPR effectively "gold plates" compliance with the **NZ Privacy Act 2020**, as GDPR is generally stricter.

---

## 1. Database & Architecture (SOC 2 Security)

SOC 2's "Security" criteria requires protection against unauthorized access.

### 1.1 Encryption

| Layer | Requirement | Implementation |
|-------|-------------|----------------|
| **At Rest** | Database volume encrypted. No sensitive fields in plain text. | SQLite: use filesystem-level encryption (LUKS/BitLocker). Passwords: bcrypt cost ≥ 12 (already implemented). Medical notes: stored separately with access controls. |
| **In Transit** | TLS (HTTPS) for all connections. No unencrypted HTTP. | Reverse proxy (Caddy/nginx) terminates TLS. Internal: localhost only. Cookie: `Secure=true` in production. |

### 1.2 Role-Based Access Control (RBAC)

**Principle:** Least Privilege. Every API endpoint checks permissions server-side, never relying on the UI.

| Role | Can Access | Cannot Access |
|------|-----------|---------------|
| **Admin** | Everything | — |
| **Coach** | Attendance, member profiles (read), injuries (read), grading proposals, observations, curriculum | Payment details, member deletion, system config |
| **Member** | Own profile, own attendance, own training log, own messages | Other members' data, admin views, coach observations |
| **Trial** | Own profile, own attendance, schedule | Member features, grading, voting |
| **Guest** | Waiver flow only | Everything else |

**Implementation:**
- `RequireRole()` middleware on every route (already implemented)
- `requireAdmin()` helper for admin-only handlers (already implemented)
- No sensitive data in client-side storage (localStorage, sessionStorage)

### 1.3 Audit Logging

**Rule:** Know *who* did *what* and *when*.

**Auditable Events:**

| Category | Events |
|----------|--------|
| **Authentication** | Login success/failure, logout, account lockout, DevMode impersonation |
| **Data Access** | Member profile viewed, training log exported, grading readiness list viewed |
| **Data Mutation** | Member created/updated/archived/deleted, belt promotion, grading config changed, estimated hours added |
| **Consent** | Waiver signed, marketing consent granted/revoked, data export requested |
| **System** | Schedule/holiday/term changes, rotor version activated, notice published |

**Schema:**

```
AuditLog {
    id:          UUID
    timestamp:   datetime (UTC)
    actor_id:    string (account ID of the person performing the action)
    actor_role:  string (role at time of action, including impersonated role)
    action:      string (e.g., "member.profile.view", "grading.proposal.approve")
    resource_type: string (e.g., "member", "grading_record")
    resource_id: string (ID of the affected entity)
    metadata:    JSON (additional context — field changes, IP address, etc.)
}
```

**Rules:**
- Audit log is **append-only** — no updates or deletes
- Never log passwords, tokens, or full medical details in metadata
- Retain audit logs for **7 years** (aligned with NZ tax retention)

---

## 2. Data Rights & Handling (GDPR / NZ Privacy Act 2020)

### 2.1 Consent Management

**Granular consent — never a single "I agree to everything" checkbox.**

| Consent Type | Required? | Default | Notes |
|-------------|-----------|---------|-------|
| **Terms of Service** | Yes | N/A (must accept) | Cannot use system without accepting |
| **Liability Waiver** | Yes | N/A (must sign) | Versioned — re-prompt on update |
| **Marketing Emails** | No | **Unchecked** | Must be opt-in, never pre-checked |
| **Photo/Video Usage** | No | **Unchecked** | For social media, website |
| **Injury Data Collection** | Yes (for training) | N/A | Explain purpose: "to ensure coach awareness of current injuries" |

**Schema:**

```
ConsentRecord {
    id:            UUID
    member_id:     string
    consent_type:  string (terms_of_service | waiver | marketing | photo_video | injury_data)
    granted:       boolean
    granted_at:    datetime (UTC)
    revoked_at:    datetime (UTC, nullable)
    version:       string (version of the document consented to)
    ip_address:    string
}
```

**Rules:**
- Store the *version* of the document they consented to
- If a waiver or ToS is updated, prompt re-consent on next login
- "Revoke Consent" accessible from the member portal at any time
- Revoking marketing consent takes effect immediately
- Revoking waiver consent triggers Admin notification (may affect training eligibility)

### 2.2 Right to be Forgotten (Deletion / Anonymisation)

**When a member requests data deletion:**

| Data Category | Action | Reason |
|--------------|--------|--------|
| **PII** (name, email, phone, address) | **Anonymise** — replace with `DELETED_<hash>` | GDPR Article 17 |
| **Medical/injury data** | **Hard delete** | Special category data |
| **Sizing info** (belt, gi, rash top, t-shirt) | **Hard delete** | PII adjacent |
| **Attendance records** | **Anonymise** — keep dates/counts, remove member link | Business analytics |
| **Payment records** | **Retain** transaction ID + amount for **7 years** | NZ IRD tax requirements |
| **Grading records** | **Anonymise** — keep belt progression data, remove name | Historical integrity |
| **Audit logs** | **Retain** with anonymised actor reference | SOC 2 requirement |
| **Coach observations** | **Hard delete** | Contains subjective PII |
| **Messages** | **Hard delete** | Private communications |
| **Consent records** | **Retain** (proof of lawful processing) | Legal requirement |

**Implementation:**
- `ExecuteDeleteMember` orchestrator performs all steps atomically
- Generates a deletion receipt (JSON) stored in audit log
- Admin confirms deletion with reason
- 30-day grace period before hard deletion (configurable)

### 2.3 Data Portability (Export)

Members can export their data at any time.

**Export includes:**
- Personal profile (name, email, program, belt, sizes)
- Attendance history (dates, classes, mat hours)
- Training log (milestones, streaks, belt progression)
- Consent records
- Messages received

**Export excludes:**
- Coach observations (private to coaches)
- Other members' data
- System configuration

**Format:** JSON or CSV, downloadable from member portal.

---

## 3. Special Category Data (Health / Injuries)

Under GDPR, health data requires **higher protection**.

### 3.1 Segregation

- Injury/medical data stored in a **separate table** (`injuries`) with stricter access rules
- Coach observations that reference health stored in `coach_observations` with coach/admin-only access
- Never display medical details in member-facing views beyond "Active injury: [body part]"

### 3.2 Minimisation

- Only collect **current injuries relevant to training** — not full medical history
- Red Flag system: body part + active/inactive toggle. No diagnosis, no treatment details.
- PAR-Q forms: if collected, store responses only, not the underlying medical conditions

### 3.3 Waiver Versioning

| Field | Description |
|-------|-------------|
| `waiver_version` | Semantic version (e.g., "2.1") |
| `waiver_content_hash` | SHA-256 of the waiver text at time of signing |
| `signed_at` | UTC timestamp |
| `member_id` | Who signed |
| `ip_address` | Where they signed from |

**Rules:**
- When waiver text is updated, increment version
- System prompts re-signing on next login/check-in if member's signed version < current version
- Old waiver signatures preserved indefinitely (legal evidence)

---

## 4. NZ Context

### 4.1 Privacy Act 2020

- **Breach Notification:** If a data breach causes or is likely to cause "serious harm," report to the **NZ Privacy Commissioner** within 72 hours
- **Privacy Officer:** Designate a privacy officer (the Admin/owner in a small gym context)
- **Information Privacy Principles (IPPs):** 13 principles — GDPR compliance covers all of them

### 4.2 Data Sovereignty

- Host on cloud providers with **Sydney (ap-southeast-2)** or **Auckland** regions
- If using SQLite locally: data stays on-premises (strongest sovereignty)
- Document where data is stored in the privacy policy

### 4.3 IRD Tax Retention

- Payment/transaction records retained for **7 years** per IRD requirements
- Anonymised member data can still support tax records (transaction ID + amount sufficient)

---

## 5. Developer Checklist

| Feature | Compliance Check | Status |
|---------|-----------------|--------|
| **Passwords** | bcrypt cost ≥ 12 | ✅ Implemented |
| **Session cookies** | HttpOnly, SameSite=Strict | ✅ Implemented |
| **CSRF protection** | gorilla/csrf on all POST routes | ✅ Implemented |
| **RBAC** | RequireRole middleware on every route | ✅ Implemented |
| **Account lockout** | 5 failed attempts → 15 min lock | ✅ Implemented |
| **Audit logging** | Structured slog for auth events | ✅ Partial (auth only) |
| **Audit log table** | Immutable append-only audit_logs table | ⬜ TODO |
| **Consent management** | Granular consent records with versioning | ⬜ TODO |
| **Data export** | Member self-service JSON/CSV export | ⬜ TODO |
| **Data deletion** | Anonymise PII, hard-delete medical, retain payments | ⬜ TODO |
| **Waiver versioning** | Version + content hash + re-prompt | ⬜ TODO |
| **Medical segregation** | Injuries in separate table with strict access | ✅ Implemented |
| **TLS** | HTTPS in production (reverse proxy) | ⬜ TODO (infra) |
| **Encrypted backups** | Tested monthly | ⬜ TODO (infra) |
| **MFA** | For Admin accounts | ⬜ TODO |
| **Synthetic test data** | No production data in dev/test | ✅ Implemented |
| **Payment tokens** | Store Stripe/GoCardless token only, never raw card numbers | ⬜ TODO (when payment integrated) |
| **Breach response plan** | Document and test incident response | ⬜ TODO |

---

## 6. Data Classification

| Classification | Examples | Access | Retention |
|---------------|----------|--------|-----------|
| **Public** | Class schedule, program names | Anyone | Indefinite |
| **Internal** | Attendance counts, rotor themes | Coach, Admin | Indefinite |
| **Confidential** | Member name, email, belt, sizes | Member (own), Coach, Admin | Until deletion request + 30 days |
| **Restricted** | Injuries, medical notes, observations | Coach, Admin only | Until deletion request (hard delete) |
| **Financial** | Payment records, fee amounts | Admin only | 7 years (IRD) |

---

## References

- [GDPR Full Text](https://gdpr-info.eu/)
- [NZ Privacy Act 2020](https://www.legislation.govt.nz/act/public/2020/0031/latest/LMS23223.html)
- [SOC 2 Trust Services Criteria](https://www.aicpa.org/resources/landing/system-and-organization-controls-soc-suite-of-services)
- [OWASP Top 10 2025](https://owasp.org/Top10/)
- [NZ Privacy Commissioner — Breach Notification](https://www.privacy.org.nz/privacy-breaches/notify-us/)
