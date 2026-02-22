# QA Report: Content & Library

Ref: [QA_AUDIT_PLAN.md](QA_AUDIT_PLAN.md) · Area 6

## Summary
Library, themes, and curriculum are well-built with good contextual subtitles. Notice management is feature-complete. One naming inconsistency between nav and page titles.

## Findings

| # | Sev | Finding | File:Line | Fix |
|---|-----|---------|-----------|-----|
| 1 | P2 | Nav says "Themes" but page h1 says "Theme Carousel" — mismatch | layout.html:139,167, themes.html:3 | Change h1 to "Themes" to match nav, keep subtitle for context |
| 2 | P3 | Themes subtitle "This week's curriculum across all classes, driven by the rotor system." is good educational text | themes.html:4 | No action — exactly the kind of helpful context we want |
| 3 | P3 | Library subtitle "Curated clips promoted by coaches. Isolate sequences and study on repeat." — excellent | library.html:4 | No action |
| 4 | P2 | Library "Add a Clip" section available to admin/coach/member — members can contribute clips; good UX | library.html:30 | No action |
| 5 | P2 | Curriculum h1 "Curriculum Management" — nav says "Curriculum". The "Management" suffix is admin-oriented but coaches also see this page | curriculum.html:3, layout.html:137,158 | Change to just "Curriculum" for all roles |
| 6 | P2 | Notice management h1 "Notice Management" — consistent with other admin pages pattern (Term Management, Holiday Management, etc.) | admin_notices.html:3 | Keep — admin-only screens can say "Management" |
| 7 | P3 | Notice form has markdown hint "(Markdown supported)" — good contextual help | admin_notices.html:22 | No action |
| 8 | P2 | Notice colour picker uses round buttons — visually distinct from the rest of the app's square/2px-radius design. Intentional for colour swatches | admin_notices.html:75 | No action — colour swatches should be round |
| 9 | P3 | Notices "Save as Draft" / "Save & Publish" — clear two-step workflow | admin_notices.html:50-51 | No action |

## Recommendations
1. **Fix Themes h1**: "Theme Carousel" → "Themes" (match nav link)
2. **Fix Curriculum h1**: "Curriculum Management" → "Curriculum" (coaches see this too)
