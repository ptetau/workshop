---
description: View open issues ordered by priority (milestone then issue number)
---

# Backlog Workflow

Display the open issue backlog ordered by implementation priority: **bugs first** (across all milestones), then stories by milestone phase (earliest = highest priority), then issue number within each milestone.

## Steps

1. **Generate the project report** — run the helper script to gather all data:

// turbo
   ```powershell
   pwsh scripts/project-report.ps1
   ```

2. **Show prioritised backlog** — list all open issues grouped by milestone, earliest milestone first:

// turbo
   ```powershell
   gh issue list --state open --limit 200 --json number,title,labels,milestone --template "{{range .}}#{{.number}} | {{.title}} | {{if .milestone}}{{.milestone.title}}{{else}}No milestone{{end}} | {{range .labels}}{{.name}}, {{end}}{{println}}{{end}}"
   ```

3. **Summarise the backlog** — present a structured overview. **Always list bugs first**, then stories by milestone:

   > **Backlog Summary**
   >
   > **🐛 Open Bugs (fix first):**
   >
   > | # | Title | Milestone |
   > |---|-------|-----------|
   > | #52 | Bug: Progress bar crashes on null belt | Phase 3 |
   >
   > **Stories by Milestone:**
   >
   > | Milestone | Open | Next issue |
   > |-----------|------|------------|
   > | Phase 0: Foundation | 6 | #21 US-1.5.1: Paginate a list |
   > | Phase 1: Core Data | 4 | #... |
   > | ... | | |
   >
   > **Total open issues:** N (B bugs + S stories)
   >
   > The highest-priority issue is **#[N]: [title]** (bug/story) in **[milestone]**.

   If there are no open bugs, omit the bugs table and proceed with stories only.

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
