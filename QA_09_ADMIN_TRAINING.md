# QA Report: Admin Training Config

Ref: [QA_AUDIT_PLAN.md](QA_AUDIT_PLAN.md) · Area 9

## Summary
Admin training screens (grading, milestones, self-estimates, schedules, class types) are consistent CRUD forms. Class type editing uses `prompt()` dialogs which is a notable UX gap.

## Findings

| # | Sev | Finding | File:Line | Fix |
|---|-----|---------|-----------|-----|
| 1 | P1 | Class type edit uses chained `prompt()` dialogs (4 in sequence) — extremely poor UX | admin_class_types.html:131-141 | Replace with inline edit form or modal |
| 2 | P1 | Delete class type warning "Schedules using it may break" — destructive action with cascade risk, but only `confirm()` protection | admin_class_types.html:157 | Acceptable with confirm; consider soft-delete or usage check |
| 3 | P2 | Grading page shows truncated member IDs `p.MemberID.substring(0,8)+'...'` — should show member name | admin_grading.html:62 | Change proposal list to display member name, not UUID snippet |
| 4 | P2 | Grading "Propose" button uses `alert('Proposal created!')` — should use inline feedback like other admin pages | admin_grading.html:157 | Replace alert with inline msg pattern |
| 5 | P2 | Milestone page h1 "Milestone Configuration" — consistent with admin naming | admin_milestones.html:3 | No action |
| 6 | P3 | Milestone form "Badge Icon" field uses text input for emoji — works but could show a preview | admin_milestones.html:26 | Low priority |
| 7 | P2 | Self-estimates h1 "Self-Estimate Review Queue" with helpful subtitle — excellent | admin_self_estimates.html:3-4 | No action |
| 8 | P2 | Self-estimates review modal buttons: "Approve" (green), "Reject" (red), "Cancel" (grey) — good colour coding | admin_self_estimates.html:22-24 | No action |
| 9 | P2 | Schedule page h1 "Schedule Management" — consistent | admin_schedules.html:3 | No action |
| 10 | P2 | Schedule time inputs use `type="text"` with placeholder "06:00" — should use `type="time"` for native picker | admin_schedules.html:28-29 | Change to `type="time"` for better UX on mobile |
| 11 | P1 | Term and holiday date inputs use `type="text"` with "2026-01-27" placeholder — should use `type="date"` for native picker | admin_terms.html:13-14, admin_holidays.html:13-14 | Change to `type="date"` |
| 12 | P2 | Delete schedule has no confirmation dialog — deletes immediately on click | admin_schedules.html:112-114 | Add `if (!confirm('Delete this schedule?')) return;` |
| 13 | P3 | Grading readiness "→ Hours" / "→ Sessions" toggle buttons are small and subtle — could be missed | admin_grading.html:90,119 | Acceptable; power-user feature |

## Recommendations (priority order)
1. **Replace `prompt()` class type edit** with inline form or modal
2. **Change date/time inputs to native types** (`type="date"`, `type="time"`) on terms, holidays, schedules
3. **Show member name** instead of truncated UUID on grading proposals
4. **Add delete confirmation** to schedule deletion
5. **Replace `alert()` with inline feedback** on grading proposal creation
