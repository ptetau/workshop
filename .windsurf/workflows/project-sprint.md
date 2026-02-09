---
description: Plan next work by reviewing milestones, open issues, and dependencies
---

# Sprint Plan Workflow

Review project status across milestones, identify what to work on next, and produce a prioritised short list of issues to tackle. Designed for solo-dev or small-team planning.

## Steps

1. **Gather project data** — run the helper script to see milestones, epics, open issues, and in-progress work:

   ```powershell
   pwsh scripts/project-report.ps1
   ```

   To focus on a specific milestone:
   ```powershell
   pwsh scripts/project-report.ps1 -Milestone "Phase 2: Kiosk MVP"
   ```

2. **Identify the active milestone** — the earliest milestone with open issues is the current focus. If the user specifies a milestone, use that instead.

3. **Check for in-progress work** — review the "In-Progress Work" section of the report. If there's in-progress work, flag it:
   > "You have work in progress:
   > - Branch: `issue-42-member-checkin` (no PR yet)
   > - PR #5: feat(kiosk): launch kiosk mode (#27)
   >
   > Consider finishing these before starting new work."

4. **Analyse dependencies** — for the open issues in the active milestone, identify a good implementation order:
   - **Foundation/infrastructure stories first** (storage, schema, domain models)
   - **Then orchestrators and business logic**
   - **Then UI/handler stories**
   - Group stories that share the same domain concept (e.g., all attendance stories together)
   - Flag stories that depend on unimplemented stories from earlier milestones

5. **Suggest next issues** — recommend 3–5 issues to work on next, with reasoning:

   > **Suggested next issues (in order):**
   >
   > 1. **#39 US-3.1.1: Attendance creates mat hours** — foundational for all attendance features, needs storage layer first
   > 2. **#40 US-3.1.2: Coach views today's attendance** — builds on #39, adds the read projection
   > 3. **#43 US-3.3.1: View training log** — member-facing projection, depends on attendance data from #39
   >
   > **Rationale:** These three form a vertical slice through the attendance domain (storage → write → read). Completing them together gives a working end-to-end feature.
   >
   > **Blocked:**
   > - #44 US-3.3.2: Milestone achievement — depends on #45 milestone configuration
   >
   > Pick an issue to start, or tell me a different focus area.

6. **User selects** — once the user picks an issue (or accepts the suggestion), offer to run `/issue-start` immediately:

   > "Ready to start #39? I can run `/issue-start` to set up the branch and implementation plan."

## Tips for Effective Planning

- **Vertical slices**: prefer implementing a thin end-to-end feature (domain → storage → orchestrator → handler → template) over completing all domain models first
- **One concept at a time**: finish all stories for a domain concept before moving to the next (e.g., all of `attendance` before starting `grading`)
- **Test as you go**: each story should have passing tests before moving to the next
- **Commit often**: one commit per story keeps the history clean and makes rollback easy
