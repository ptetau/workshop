# QA Report: Admin Email

Ref: [QA_AUDIT_PLAN.md](QA_AUDIT_PLAN.md) · Area 8

## Summary
Email management is the most complex admin feature. Compose page is powerful with multiple recipient filters, scheduling, test-send, and preview. Well-built overall.

## Findings

| # | Sev | Finding | File:Line | Fix |
|---|-----|---------|-----------|-----|
| 1 | P2 | Email list h1 "Email Management" — consistent with admin naming pattern | admin_emails.html:4 | No action |
| 2 | P3 | Compose page h1 dynamically changes: "Compose Email" / "Edit Draft" / "Scheduled Email" / "Retry Failed Email" — excellent context | admin_compose_email.html:3,486-500 | No action |
| 3 | P2 | Compose page "Save Draft" button is blue (#2980b9), "Send Now" is green, "Schedule" is purple (#8e44ad) — colour-coded by action severity. Good UX but departs from app's orange primary | admin_compose_email.html:74-77 | Intentional — email actions use severity colours to prevent accidental sends |
| 4 | P2 | Compose page "Cancel" is a plain text link, not a button — inconsistent with other forms | admin_compose_email.html:78 | Acceptable — reduces visual weight for a navigation action |
| 5 | P1 | Compose page has massive code duplication: `saveDraft()`, `sendEmail()`, `testSendEmail()`, and `scheduleEmail()` all duplicate the "save draft first, then action" pattern | admin_compose_email.html:363-452 | Refactor: extract `ensureDraftSaved(callback)` helper to DRY this up |
| 6 | P3 | Email detail modal close button `&times;` is clean and consistent with notice detail pattern | admin_emails.html:131 | No action |
| 7 | P2 | Filter buttons use active state (orange bg) vs inactive (grey border) — good visual affordance | admin_emails.html:12-16 | No action |
| 8 | P3 | Email template page subtitle could clarify that the template wraps all outgoing emails | admin_email_template.html | Consider adding subtitle |

## Recommendations
1. **Refactor compose JS** — extract `ensureDraftSaved(callback)` to eliminate 4x duplicated save-then-act pattern
2. Minor: consider adding subtitle to email template page
