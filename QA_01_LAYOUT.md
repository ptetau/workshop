# QA Report: Layout & Navigation

Ref: [QA_AUDIT_PLAN.md](QA_AUDIT_PLAN.md) · Area 1

## Summary
Navigation is well-structured with role-based menus. Several cross-cutting inconsistencies found.

## Findings

| # | Sev | Finding | File:Line | Fix |
|---|-----|---------|-----------|-----|
| 1 | P2 | Login page h1 says "Sign In" but nav link says "Login" and logout button says "Logout" — mixed verb forms | login.html:5, layout.html:186,191 | Standardise: nav "Log In", button "Log Out", h1 "Log In" |
| 2 | P2 | `<title>` is always "Workshop App" — doesn't reflect current page | layout.html:6 | Pass page title from handlers or use `{{ template "title" }}` |
| 3 | P2 | Footer says "© 2026 Workshop Jiu Jitsu · Wellington" — hardcoded year | layout.html:199 | Minor but fine for now; could use JS `new Date().getFullYear()` |
| 4 | P2 | Coach nav has Kiosk inside More dropdown; admin nav has no Kiosk link at all | layout.html:169 | Add Kiosk to admin nav (More > Training group) |
| 5 | P3 | Coach "More" dropdown has no group labels unlike admin's grouped "More" | layout.html:160-171 | Add group labels (Training, Content) for coach More menu |
| 6 | P2 | Member nav shows "Training Log" and "Messages" but trial shows same — trial doesn't get Calendar unless feature-flagged | layout.html:172-183 | Correct; just noting trial nav is intentionally minimal |
| 7 | P3 | Mobile nav hamburger uses HTML entity ☰ — could use a proper icon for consistency | layout.html:118 | Low priority; works fine |
| 8 | P2 | Admin dashboard quick-links include "Class Types" but this isn't in the nav More menu | dashboard_admin.html:59, layout.html:126-153 | Add "Class Types" to admin More > Training group |

## Recommendations
1. **Standardise auth verbs**: "Log In" / "Log Out" everywhere (not "Sign In" / "Logout")
2. **Add Class Types and Kiosk** to admin nav More menu under Training group
3. **Add group labels** to coach More dropdown for parity with admin
