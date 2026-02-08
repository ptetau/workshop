---
description: View open issues ordered by priority (milestone then issue number)
---

# Backlog Workflow

Display the open issue backlog ordered by implementation priority: milestone phase first (earliest = highest priority), then issue number within each milestone.

## Steps

1. **Generate the project report** — run the helper script to gather all data:

// turbo
   ```powershell
   pwsh scripts/project-report.ps1
   ```

2. **Show prioritised backlog** — list all open issues grouped by milestone, earliest milestone first:

// turbo
   ```powershell
   gh issue list --state open --limit 200 --json number,title,labels,milestone --jq 'sort_by(.milestone.number // 999, .number) | group_by(.milestone.title // "No milestone") | .[] | ["", "### \(.[0].milestone.title // "No milestone")", (.[] | "  #\(.number) \(.title) [\(.labels | map(.name) | join(", "))]")] | .[]'
   ```

   If the jq grouping is too complex, fall back to a simpler sorted list:
   ```powershell
   gh issue list --state open --limit 200 --json number,title,labels,milestone --jq 'sort_by(.milestone.number // 999, .number) | .[] | "#\(.number) \(.title) [\(.labels | map(.name) | join(", "))] (\(.milestone.title // "no milestone"))"'
   ```

3. **Summarise the backlog** — present a structured overview:

   > **Backlog Summary**
   >
   > | Milestone | Open | Next issue |
   > |-----------|------|------------|
   > | Phase 0: Foundation | 6 | #21 US-1.5.1: Paginate a list |
   > | Phase 1: Core Data | 4 | #... |
   > | ... | | |
   >
   > **Total open issues:** N
   >
   > The highest-priority unstarted issue is **#[N]: [title]** in **[milestone]**.

4. **Filter options** — if the user wants to narrow the view, offer filters:
   - By milestone: `gh issue list --state open --milestone "Phase 2: Kiosk MVP" ...`
   - By area label: `gh issue list --state open --label "area:kiosk" ...`
   - By cross-cutting label: `gh issue list --state open --label "needs-storage" ...`
   - Stories only: `gh issue list --state open --label "type:story" ...`
   - Epics only: `gh issue list --state open --label "type:epic" ...`

5. **Offer next action** — based on the backlog:
   - "Run `/issue-start [NUMBER]` to begin work on the highest-priority issue"
   - "Run `/project-sprint` to get a recommended implementation order with dependency analysis"
   - "Run `/issue-create` to add a new requirement to the backlog"
