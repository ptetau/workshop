---
description: Dashboard view of project progress across epics, milestones, and recent activity
---

# Project Status Workflow

Generate a comprehensive status report of the project: milestone progress, epic completion, recent activity, and blockers.

## Steps

1. **Gather project data** — run the helper script to collect milestones, epics, in-progress work, and recent completions:

   ```powershell
   pwsh scripts/project-report.ps1
   ```

   To focus on a specific milestone:
   ```powershell
   pwsh scripts/project-report.ps1 -Milestone "Phase 2: Kiosk MVP"
   ```

2. **Generate status report** — compile the script output into a formatted summary:

   ```
   # Project Status — [date]

   ## Milestone Progress
   | Milestone | Done | Total | Progress |
   |-----------|------|-------|----------|
   | Phase 0: Foundation | 2 | 8 | 25% ████░░░░░░░░ |
   | Phase 1: Core Data | 0 | 6 | 0% ░░░░░░░░░░░░ |
   | ... | | | |

   ## Epics
   | # | Epic | Stories Done | Status |
   |---|------|-------------|--------|
   | 7 | S1 Foundation | 2/6 | 🟡 In Progress |
   | 8 | S2 Kiosk | 0/12 | ⬜ Not Started |
   | ... | | | |

   ## In Progress
   - Branch: issue-42-member-checkin (no PR)
   - PR #5: feat(kiosk): launch kiosk mode

   ## Recently Completed
   - #21 US-1.5.1: Paginate a list — closed 2026-02-07
   - #22 US-1.5.2: Change row count — closed 2026-02-07

   ## Blockers / Notes
   - [any blocked issues or open questions]
   ```

6. **Suggest next action** — based on the status, recommend what to do:
   - If there's in-progress work → "Finish the open PR/branch first"
   - If a milestone is nearly complete → "Close out Phase X by finishing these N issues"
   - If no work in progress → "Run `/project-sprint` to pick the next issues"
   - If all current milestone issues are done → "All Phase X stories complete — move to Phase Y"
