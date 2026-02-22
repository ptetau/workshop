# QA Report: Calendar & Goals

Ref: [QA_AUDIT_PLAN.md](QA_AUDIT_PLAN.md) · Area 5

## Summary
Calendar is a full-featured SPA page loaded via JS. Not audited in detail here as it's a large JS-driven component in calendar.html. Training goals live on the training log page.

## Findings

| # | Sev | Finding | File:Line | Fix |
|---|-----|---------|-----------|-----|
| 1 | P3 | Calendar nav link text is "Calendar" everywhere — consistent across all roles | layout.html:123,159,176 | No action |
| 2 | P2 | Training goal on training log is hardcoded to "weekly" period only | member_training_log.html:334 | Future enhancement: allow monthly goals |
| 3 | P3 | Training goal display "3x weekly" is clear and concise | member_training_log.html:51,324 | No action |
| 4 | P2 | Calendar has no h1 visible — it renders its own header via JS, bypassing the template h1 convention | calendar.html (JS-rendered) | Acceptable for SPA-style calendar; the month/year header serves same purpose |

## Recommendations
No critical fixes needed. Calendar is the most interactive screen and works well as a standalone SPA component.
