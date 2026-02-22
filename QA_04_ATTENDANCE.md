# QA Report: Attendance & Training

Ref: [QA_AUDIT_PLAN.md](QA_AUDIT_PLAN.md) · Area 4

## Summary
Attendance and training log screens are feature-rich. Training log is the most complex member-facing screen. Kiosk is standalone and well-designed for touch.

## Findings

| # | Sev | Finding | File:Line | Fix |
|---|-----|---------|-----------|-----|
| 1 | P2 | Attendance h1 uses emoji "📊 Today's Attendance" / "📊 Attendance" — inconsistent with admin screens that have no emoji | get_attendance_today.html:24 | Remove emoji from h1 for consistency (see QA_03 #10) |
| 2 | P2 | Attendance empty-state "Check In Member" button uses blue (#2196f3) — should be orange | get_attendance_today.html:98 | Change to `var(--orange)` |
| 3 | P2 | Attendance "Back to Dashboard" and "View All Members" links use blue (#667eea) — inconsistent with orange brand | get_attendance_today.html:106-108 | Change to `var(--orange)` |
| 4 | P2 | Training log stat card "Flight Time" label differs from grading config which says "Flight Time (hrs)" — same concept, good alignment | member_training_log.html:21, admin_grading.html:37 | No action — "Flight Time" is the domain term |
| 5 | P3 | Training log "Week Streak" suffix "w" — could be clearer as "weeks" on first render | member_training_log.html:24-25 | Minor; "0w" is compact and clear enough |
| 6 | P2 | Training log "Training Goal" section uses hardcoded "weekly" period — no option for monthly or custom | member_training_log.html:74-79,334 | Acceptable for MVP; note for future enhancement |
| 7 | P2 | Training log "Submit Estimated Hours" subtitle is good educational text: "Trained elsewhere while travelling? Submit an estimate for review." | member_training_log.html:82 | No action — good contextual help |
| 8 | P3 | Training log uses `<details>` for "New Estimate" — consistent with member profile's "Add Estimated Hours" pattern | member_training_log.html:84 | Good consistency |
| 9 | P2 | Kiosk header uses red (#e94560) branding — different from app's orange (#F9B232) | kiosk.html:12 | Intentional — kiosk is dark-themed standalone; red is the kiosk accent |
| 10 | P1 | Kiosk `resetKiosk()` has a bug: `stepDone.classList.remove('hidden')` followed by `stepDone.classList.add('hidden')` — the remove is unnecessary | kiosk.html:279-280 | Remove the `remove('hidden')` line |
| 11 | P2 | Kiosk "Guest Check-In" redirects to waiver form `/forms/sign-waiver` — good flow for walk-ins | kiosk.html:269 | No action |
| 12 | P3 | Kiosk exit uses `prompt()` for password — functional but basic; could be a proper modal | kiosk.html:288-289 | Low priority; works fine for the use case |

## Recommendations
1. **Fix kiosk resetKiosk() bug** — remove redundant `classList.remove('hidden')`
2. **Standardise link colours** in attendance to orange
3. **Remove emoji** from attendance h1 (cross-cutting with QA_03)
