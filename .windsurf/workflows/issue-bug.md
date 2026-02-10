---
description: Investigate and file a bug — analyse code and browser behaviour, write a detailed report, and extrapolate to find similar issues
---

# Bug Investigation Workflow

Investigate a reported or suspected bug by analysing code, running the app in the browser, writing a detailed bug report as a GitHub Issue, and extrapolating the root cause to find similar bugs elsewhere in the codebase.

## Context

- Bugs use the `bug` label and are **higher priority than stories** — `/project-sprint` and `/project-backlog` will always schedule bugs before regular issues within the same milestone.
- A good bug report includes: reproduction steps, root cause theory, affected code, and extrapolation to similar patterns.

## Steps

### Phase A: Investigate and Theorise

1. **Understand the symptom** — read the user's description and identify:
   - What is the **expected** behaviour?
   - What is the **actual** behaviour?
   - Which user role and page/flow are affected?
   - Is this a regression (did it work before)?

2. **Search the codebase for the affected area** — use code search to find the relevant handler, projection, orchestrator, template, and storage code:

   ```
   Use code_search to find the handler, projection, or template involved in the bug.
   ```

   Read the relevant files to understand the current implementation. Trace the data flow from the HTTP handler through to storage and back to the template.

3. **Form a root cause theory** — based on code analysis, hypothesise what's going wrong:
   - Is it a logic error (wrong condition, off-by-one, missing nil check)?
   - Is it a data flow error (wrong field mapped, missing dependency, stale data)?
   - Is it a race condition or timing issue (async fetch, navigation race)?
   - Is it a schema/migration issue (missing column, wrong default)?
   - Is it a UI issue (wrong element ID, missing JS handler, template rendering)?

4. **Verify in the browser** — if the app can be run locally, start it and reproduce the bug:

   ```powershell
   go run ./cmd/server
   ```

   Use the Playwright MCP browser tools to:
   - Navigate to the affected page
   - Take a snapshot to see the current state
   - Check console logs for errors
   - Check network requests for failed API calls
   - Try to reproduce the exact symptom

   Document what you observe — screenshots, console errors, API response bodies.

5. **Confirm the theory** — correlate browser observations with the code analysis:
   - Does the console error match the code path you identified?
   - Does the API return unexpected data?
   - Does the template reference a field that doesn't exist?
   - Add temporary logging if needed to narrow down the root cause.

### Phase B: Extrapolate — Find Similar Bugs

6. **Generalise the root cause** — abstract the bug into a pattern:
   - If the bug is "handler doesn't check for nil store dependency" → search all handlers for the same pattern
   - If the bug is "template references a field not in the API response" → check all templates against their API contracts
   - If the bug is "missing auth check on endpoint" → audit all similar endpoints
   - If the bug is "race condition on navigation" → check all pages that do select-triggered navigation

7. **Search for the same pattern elsewhere** — use grep/code search to find other instances:

   ```
   Use grep_search with a pattern that matches the generalised root cause.
   ```

   For each match, assess whether it has the same bug or is safe. Document findings.

8. **Compile the extrapolation report** — list all affected locations:
   > **Extrapolation: [pattern name]**
   >
   > The root cause pattern is: [description]
   >
   > | Location | Status | Notes |
   > |----------|--------|-------|
   > | `handlers.go:handleFoo` | **Affected** | Same missing nil check |
   > | `handlers.go:handleBar` | Safe | Already has the guard |
   > | `projections/get_baz.go` | **Affected** | Same pattern |
   >
   > **Total:** N locations affected, M already safe.

### Phase C: File the Bug Report

9. **Check for duplicates** — search existing issues:

// turbo
   ```powershell
   gh issue list --state open --label bug --limit 50 --json number,title --template "{{range .}}#{{.number}} | {{.title}}{{println}}{{end}}"
   ```

   Also search closed bugs for prior fixes that may have regressed:
// turbo
   ```powershell
   gh issue list --state closed --label bug --limit 50 --json number,title --template "{{range .}}#{{.number}} | {{.title}}{{println}}{{end}}"
   ```

10. **Draft the bug issue** — use this template:

    ```markdown
    ### Bug Summary
    [One-line description of the defect]

    ### Steps to Reproduce
    1. [Step 1]
    2. [Step 2]
    3. [Step 3]

    ### Expected Behaviour
    [What should happen]

    ### Actual Behaviour
    [What actually happens — include error messages, screenshots, console output]

    ### Root Cause Analysis
    **Theory:** [Explanation of why the bug occurs]

    **Affected code:**
    - `[file:line]` — [what's wrong]
    - `[file:line]` — [what's wrong]

    ### Extrapolation
    [Summary of other locations with the same pattern — reference the table from Phase B]

    **Other locations to fix:**
    - [ ] `[file:function]` — [same pattern]
    - [ ] `[file:function]` — [same pattern]

    ### Suggested Fix
    [Brief description of the fix approach — minimal upstream fix preferred over downstream workaround]

    ### Test Plan
    - [ ] Unit test: [regression test for the root cause]
    - [ ] Browser test: [reproduce the symptom and verify the fix]
    - [ ] Verify extrapolated locations are also fixed
    ```

11. **Present the draft to the user** — show the complete bug report before creating:
    > "Here's the bug report I'll file:
    >
    > **[Title]** (`bug`, `area:[x]`, milestone: [Phase N])
    > [full body draft]
    >
    > I also found [N] other locations with the same pattern (listed in Extrapolation).
    >
    > Does this look right? Any changes before I create the issue?"

12. **Create the issue** — once approved:

    ```powershell
    gh issue create --title "Bug: [short description]" --body "[body]" --label "bug" --label "area:[x]" --milestone "[Phase N: Name]"
    ```

    Assign to the **current active milestone** (bugs are urgent, don't defer to later phases).

13. **Offer next action:**
    > "Bug filed as #[N]. Since bugs are high priority, would you like to:
    > - Run `/issue-start [N]` to fix it now
    > - Run `/project-sprint` to see it prioritised in the sprint plan
    > - Continue with what you were doing — the bug is tracked"
