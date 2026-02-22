# QA Report: Authentication & Onboarding

Ref: [QA_AUDIT_PLAN.md](QA_AUDIT_PLAN.md) · Area 2

## Summary
Auth flows are clean and functional. Minor terminology and UX friction issues.

## Findings

| # | Sev | Finding | File:Line | Fix |
|---|-----|---------|-----------|-----|
| 1 | P2 | Login h1 says "Sign In" but button says "Log In" — inconsistent on the same page | login.html:5,22 | Change h1 to "Log In" to match button |
| 2 | P3 | Activate page uses JS-based form submission (fetch API) but change-password uses standard POST — inconsistent patterns | activate.html:14, change_password.html:15 | Not a user-facing issue but worth noting |
| 3 | P3 | Activate page password hint "Minimum 12 characters" appears below field; change-password has same pattern — consistent, good | activate.html:19, change_password.html:24 | No action needed |
| 4 | P2 | Admin dashboard h1 "Admin Dashboard" / coach "Coach Dashboard" / member "My Dashboard" or personalised greeting — role label is redundant when DevMode bar shows role | dashboard_admin.html:3, dashboard_coach.html:3, dashboard_member.html:21 | Keep as-is; the h1 is useful when DevMode bar is hidden in prod |
| 5 | P2 | Admin dashboard "Training" section has 8 shortcut buttons but some duplicate nav links (Members, Attendance, Grading, Schedules) | dashboard_admin.html:53-62 | Acceptable as quick-access hub; not confusing |
| 6 | P1 | Admin dashboard button sizes differ: admin uses `font-size:0.8rem`, coach uses `font-size:0.85rem` | dashboard_admin.html:54, dashboard_coach.html:101 | Standardise to 0.85rem across all dashboards |
| 7 | P2 | Coach dashboard "Actions" section has Check In, Launch Kiosk, Report Injury, Add Observation — good verb-first labels | dashboard_coach.html:99-105 | No action needed |
| 8 | P3 | Member dashboard trial banner uses gradient background that differs from the rest of the app's flat design | dashboard_member.html:6 | Intentional — trial onboarding should stand out |
| 9 | P2 | Member dashboard "Quick Links" section shows Themes link but nav bar doesn't show Themes for members | dashboard_member.html:104 | Either add Themes to member nav or remove from quick links |

## Recommendations
1. **Fix login h1**: "Sign In" → "Log In"
2. **Standardise dashboard button font-size** to 0.85rem across all roles
3. **Align member quick-links with nav**: Add Themes to member nav or remove from quick links
