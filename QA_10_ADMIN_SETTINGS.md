# QA Report: Admin Settings

Ref: [QA_AUDIT_PLAN.md](QA_AUDIT_PLAN.md) · Area 10

## Summary
Settings screens (accounts, features, terms, holidays, inactive) are clean CRUD forms. System Options page is the most well-built admin screen with excellent subtitle and layout.

## Findings

| # | Sev | Finding | File:Line | Fix |
|---|-----|---------|-----------|-----|
| 1 | P2 | Account management h1 "Account Management" — consistent | admin_accounts.html:3 | No action |
| 2 | P1 | Account creation shows password in `type="text"` — should be `type="password"` | admin_accounts.html:14 | Change to `type="password"` |
| 3 | P2 | Account role change uses `<select>` with `onchange` — no confirmation dialog for role changes | admin_accounts.html:57 | Add confirmation: "Change role to {role}?" |
| 4 | P2 | System Options h1 "System Options" with excellent subtitle — best contextual help in the app | admin_features.html:3-4 | No action |
| 5 | P3 | System Options "Save" button is at the top, not bottom — unconventional but reduces scroll for quick changes | admin_features.html:10-11 | Acceptable; table may be long |
| 6 | P2 | System Options beta cohort section has clear subtitle "Mark specific accounts as beta testers" — good | admin_features.html:35 | No action |
| 7 | P2 | Inactive members h1 "Inactive Members" — clear purpose | admin_inactive.html:3 | No action |
| 8 | P3 | Inactive members "Archive" button is red — correctly signals destructive action | admin_inactive.html:29 | No action |
| 9 | P2 | Inactive members Archive has `confirm()` dialog — good | admin_inactive.html:36 | No action |
| 10 | P1 | Term deletion has **no confirmation** — deletes immediately | admin_terms.html:67 | Add `if (!confirm('Delete this term?')) return;` |
| 11 | P1 | Holiday deletion has **no confirmation** — deletes immediately | admin_holidays.html:67 | Add `if (!confirm('Delete this holiday?')) return;` |
| 12 | P2 | Terms/holidays use text inputs for dates instead of native `type="date"` | admin_terms.html:13-14, admin_holidays.html:13-14 | Change to `type="date"` (also noted in QA_09) |

## Recommendations (priority order)
1. **Fix password field** on account creation — `type="password"` not `type="text"`
2. **Add delete confirmation** to term and holiday deletion
3. **Add role-change confirmation** on accounts page
4. **Use native date inputs** on terms and holidays forms
