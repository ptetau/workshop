# QA Report: Member Management

Ref: [QA_AUDIT_PLAN.md](QA_AUDIT_PLAN.md) · Area 3

## Summary
Member list is the most polished screen — server-side pagination, sort, search, filters, CSV import/export. Several legacy forms need style alignment.

## Findings

| # | Sev | Finding | File:Line | Fix |
|---|-----|---------|-----------|-----|
| 1 | P1 | Register form uses **blue** (#667eea) primary button, not orange — breaks brand consistency | form_register_member.html:60-74 | Change .btn-primary to `background: var(--orange)` |
| 2 | P1 | Register form Cancel links to `/index.html` — this is a dead route (should be `/members` or `/dashboard`) | form_register_member.html:27 | Fix href to `/members` |
| 3 | P1 | Injury form Cancel links to `/index.html` — same dead route | form_report_injury.html:42 | Fix href to `/dashboard` |
| 4 | P1 | Injury form requires **raw Member UUID** typed manually — terrible UX for coaches | form_report_injury.html:8-10 | Replace with name-search autocomplete (like check-in form) |
| 5 | P2 | Injury form uses **yellow** (#ffc107) primary button — should be orange | form_report_injury.html:77-91 | Change to `var(--orange)` |
| 6 | P1 | Waiver form uses **blue** (#667eea) primary button — breaks brand | form_sign_waiver.html:65-79 | Change to `var(--orange)` |
| 7 | P2 | Waiver form Cancel links to `/` — should be `/dashboard` for logged-in users | form_sign_waiver.html:33 | Change to `/dashboard` |
| 8 | P1 | Check-in form uses **blue** (#2196f3) primary button and focus color — breaks brand | form_check_in_member.html:44,71-75 | Change to orange theme |
| 9 | P2 | Check-in form bottom link to attendance uses blue (#667eea) not orange | form_check_in_member.html:33 | Use `var(--orange)` or `#F9B232` |
| 10 | P2 | Member profile h1 "👤 Member Profile" uses emoji prefix; member list h1 "📋 Member List" uses emoji — consistent with each other but inconsistent with admin screens which have no emoji | get_member_profile.html:20, get_member_list.html:21 | Either add emoji to all h1s or remove from all — recommend **remove** for cleaner look |
| 11 | P2 | Member profile "Back to Member List" link uses blue (#667eea) not orange | get_member_profile.html:150,152 | Change to `var(--orange)` |
| 12 | P2 | Member profile "Active" status uses orange checkmark but member list "Active" also uses orange — consistent, good | get_member_profile.html:53, get_member_list.html:119 | No action |
| 13 | P1 | Register, injury, waiver, check-in forms all override global styles with local `<style>` blocks using different colours and borders — fights layout.html | form_register_member.html:31-90, form_report_injury.html:46-107, etc. | Remove local style overrides; rely on layout.html globals |
| 14 | P3 | Member list belt CSS (`.belt-*`) is duplicated in 3 files: get_member_list.html, get_member_profile.html, get_attendance_today.html | Multiple files | Extract to layout.html or styles.css |
| 15 | P1 | Delete term/holiday has no confirmation dialog | admin_terms.html:67, admin_holidays.html:67 | Add `if (!confirm(...)) return;` |

## Recommendations (priority order)
1. **Fix dead Cancel links** → `/members` or `/dashboard` (register, injury forms)
2. **Fix all button colours** → orange brand across all forms
3. **Replace UUID input on injury form** with name-search autocomplete
4. **Remove local style blocks** from legacy forms — use global layout styles
5. **Standardise emoji usage** in h1 headings — recommend removing all emoji prefixes
6. **Extract belt CSS** to a shared location
7. **Add confirmation to delete term/holiday**
