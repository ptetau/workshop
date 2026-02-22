# QA Audit Plan — Epic #281: Quality Assurance

**Goal:** Unreasonable expectations for quality, performance, and ease of use.

**Principles:**
- Don't make people do more work than necessary
- Keep terminology consistent and clear across all screens
- Offer useful subtitles to educate where needed; otherwise let the UI speak for itself
- Every screen should feel like it belongs to the same product

---

## Audit Checklist (per feature area)

### A. Terminology & Labels
- [ ] Page title (h1) uses consistent naming with nav link
- [ ] Button labels are verb-first, action-clear ("Save", "Check In", not "Submit")
- [ ] Form labels match the domain language used elsewhere
- [ ] Status values use same casing/wording across screens (e.g., "Active" vs "active")
- [ ] Date formats are consistent (DD Mon YYYY or similar, never raw ISO)
- [ ] No jargon without explanation

### B. Screen Structure & Hierarchy
- [ ] Every page has a clear h1 that matches what the nav promised
- [ ] Subtitle/description present where the screen's purpose isn't obvious
- [ ] Subtitle absent where it would be noise (simple, obvious screens)
- [ ] Card/section grouping is logical
- [ ] Action buttons are visually distinct from navigation

### C. Reduce User Effort
- [ ] Filters auto-submit or have clear "apply" affordance
- [ ] Forms pre-fill sensible defaults
- [ ] Destructive actions have confirmation
- [ ] Success/error feedback is immediate and clear
- [ ] Empty states have helpful messages (not just blank)
- [ ] Pagination/sorting state preserved across navigation
- [ ] No unnecessary clicks to reach common actions

### D. Consistency
- [ ] Color usage: primary (orange), secondary (dark), danger (red), muted (grey)
- [ ] Button styles: primary action = orange, secondary = dark/grey, danger = red
- [ ] Table styles: consistent headers, row hover, alignment
- [ ] Spacing and padding match across similar screens
- [ ] Icons/emoji usage is consistent (all screens use them or none do)

### E. Accessibility & Polish
- [ ] All form inputs have associated labels
- [ ] Interactive elements have hover states
- [ ] Mobile responsive (doesn't break on narrow viewport)
- [ ] Loading states for async operations
- [ ] No orphaned/dead links

### F. Performance (backend)
- [ ] Queries use indexes appropriately
- [ ] No N+1 query patterns in list views
- [ ] Pagination is server-side for large datasets
- [ ] TimedDB logging catches slow queries

---

## Feature Areas

| # | Area | Templates | Report File |
|---|------|-----------|-------------|
| 1 | Layout & Navigation | layout.html | QA_01_LAYOUT.md |
| 2 | Authentication & Onboarding | login, activate, change_password, dashboards | QA_02_AUTH.md |
| 3 | Member Management | get_member_list, get_member_profile, form_register_member, form_check_in_member, form_report_injury, form_sign_waiver | QA_03_MEMBERS.md |
| 4 | Attendance & Training | get_attendance_today, member_training_log, kiosk | QA_04_ATTENDANCE.md |
| 5 | Calendar & Goals | calendar | QA_05_CALENDAR.md |
| 6 | Content & Library | library, themes, curriculum | QA_06_CONTENT.md |
| 7 | Messaging | member_messages, member_inbox | QA_07_MESSAGING.md |
| 8 | Admin: Email | admin_emails, admin_compose_email, admin_email_template | QA_08_ADMIN_EMAIL.md |
| 9 | Admin: Training Config | admin_grading, admin_milestones, admin_self_estimates, admin_schedules, admin_class_types | QA_09_ADMIN_TRAINING.md |
| 10 | Admin: Settings | admin_accounts, admin_features, admin_terms, admin_holidays, admin_inactive, admin_notices, admin_perf | QA_10_ADMIN_SETTINGS.md |
| 11 | Performance | Storage layer, query patterns, indexes | QA_11_PERFORMANCE.md |

---

## Severity Levels

- **P0 — Broken:** Functionality doesn't work, data loss risk
- **P1 — Friction:** User has to do unnecessary work or gets confused
- **P2 — Inconsistency:** Terminology, style, or behavior differs between screens
- **P3 — Polish:** Minor visual or wording improvements

## Report Format

Each report follows this structure:

```
# QA Report: [Area Name]
## Summary
[One-line verdict]

## Findings
| # | Severity | Finding | File:Line | Fix |
|---|----------|---------|-----------|-----|

## Recommendations
[Ordered list of changes]
```
