---
description: Review changed files for privacy and compliance (GDPR, NZ Privacy Act, SOC 2)
---

# Privacy & Compliance Review

Focused review against [PRIVACY.md](../../PRIVACY.md) — GDPR, NZ Privacy Act 2020, and SOC 2 requirements.

## Steps

1. **Identify data touched** — classify the data in the changed files:

   | Classification | Examples | Access | Retention |
   |---------------|----------|--------|-----------|
   | **Public** | Schedules, programs | Anyone | Indefinite |
   | **Internal** | Attendance counts, themes | Coach, Admin | Indefinite |
   | **Confidential** | Name, email, belt, sizes | Member (own), Coach, Admin | Until deletion + 30 days |
   | **Restricted** | Injuries, observations | Coach, Admin only | Until deletion (hard delete) |
   | **Financial** | Payments, fees | Admin only | 7 years (NZ IRD) |

2. **RBAC** — if routes or access patterns changed:
   - Permissions enforced server-side via `RequireRole()` middleware
   - Members can only access their own data (scope queries to `session.AccountID`)
   - Coach/admin-only data (injuries, observations) uses separate store methods

   | Don't | Do |
   |-------|------|
   | Check role in the UI and trust it | `RequireRole()` middleware on every route |
   | Write a custom `requireAdmin()` helper | `RequireRole(account.RoleAdmin)` — one pattern for all roles |
   | Let members query all profiles | Scope to `session.AccountID` for member-facing routes |
   | Store role in localStorage | Extract from server-side session via `GetSessionFromContext()` |

3. **Medical / Restricted data** — if injuries, observations, or health data touched:
   - Injuries stored in separate `injury` table with coach/admin-only access
   - Never display medical details to members beyond "Active injury: [body part]"
   - Only collect current injuries relevant to training — not full medical history
   - Never log injury details in audit events — log only that it was reported/updated

   | Don't | Do |
   |-------|------|
   | Add injury columns to the member table | Separate `injury` table with restricted access methods |
   | Show diagnosis to members | Display only "Active injury: [body part]" |
   | Store medical history or treatment details | Body part + active/inactive toggle only |
   | `slog.Info("audit_event", "injury_notes", notes)` | `slog.Info("audit_event", "action", "injury.update", "resource_id", id)` |

4. **Consent** — if registration, waivers, or marketing changed:
   - Consent is granular (separate checkboxes per purpose)
   - Marketing consent is opt-in (never pre-checked)
   - Waiver version tracked — re-prompt on update
   - Consent records stored with version, timestamp, IP

   | Don't | Do |
   |-------|------|
   | Single "I agree to everything" checkbox | Separate consent per: ToS, waiver, marketing, photo, injury |
   | Pre-check the marketing checkbox | Default unchecked — opt-in only |
   | Silently update waiver text | Increment version, re-prompt on next login/check-in |
   | Hard-delete consent records | Retain as proof of lawful processing |

5. **Deletion / Anonymisation** — if member deletion or data cleanup changed:
   - PII: anonymise with `DELETED_<hash>`
   - Medical/injury data: hard delete
   - Observations, messages: hard delete
   - Attendance, grading: anonymise (keep stats, remove member link)
   - Payments: retain 7 years (NZ IRD)
   - Consent records: retain (legal proof)
   - 30-day grace period before hard deletion

   | Don't | Do |
   |-------|------|
   | Hard-delete everything | Anonymise PII, hard-delete medical, retain payments 7 years |
   | Delete immediately on request | 30-day grace period before hard deletion |
   | Skip deletion receipt | Store JSON receipt in audit log with all actions taken |
   | Delete consent records | Retain as proof of lawful processing |
   | Delete payment records | Retain transaction ID + amount for 7 years (NZ IRD) |

6. **Audit logging** — if orchestrators or auth handlers changed:
   - Audit log is append-only (no updates or deletes)
   - Every data mutation emits `slog.Info("audit_event", ...)` with actor_id, action, resource_type, resource_id
   - Auth events emit `slog.Info("auth_event", ...)`
   - Never log passwords, tokens, or medical details

   | Don't | Do |
   |-------|------|
   | Update or delete audit log entries | Append-only — immutable |
   | Skip audit log on data mutation | Every orchestrator that mutates data must emit `audit_event` |
   | `slog.Info("audit_event", "password", pw)` | Never log passwords, tokens, or medical content |
   | `fmt.Println("updated member")` | `slog.Info("audit_event", "actor_id", id, "action", "member.update")` |

7. **Data export** — if export or member portal changed:
   - Export includes: profile, attendance, training log, consent, messages
   - Export excludes: coach observations, other members' data, system config
   - Format: JSON or CSV

8. **NZ context** — if hosting, breach handling, or retention changed:
   - Breach notification: report to NZ Privacy Commissioner within 72 hours
   - Data sovereignty: EU or AU/NZ regions (current: OVH France, GDPR-compliant)
   - Payment records retained 7 years (IRD requirement)

9. **Report** — summarise findings:
   - Data classifications affected by the change
   - List any privacy violations with file path and line number
   - Recommendation: **Approve** / **Request Changes**
