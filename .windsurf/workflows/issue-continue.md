---
description: Resume work on an in-progress GitHub issue — review progress, gap-check against criteria, plan remaining work
---

# Continue Issue Workflow

Resume work on a GitHub issue that already has a feature branch. Reviews what has been written against the issue's acceptance criteria, identifies gaps, and produces a plan to finish — or offers to abandon the branch and start fresh.

## When to use

This workflow is recommended by `/issue-start` when it detects an existing `issue-[N]-*` branch for the target issue. You can also invoke it directly when returning to unfinished work.

## Steps

1. **Detect the in-progress branch** — find the branch for the issue:

// turbo
   ```powershell
   git branch --list "issue-*" | ForEach-Object { $_.Trim() }
   ```

   If the user provided an issue number, match `issue-[NUMBER]-*`. If on a feature branch already, use it directly:

// turbo
   ```powershell
   git branch --show-current
   ```

2. **Load issue context** — read the full issue details:

   ```powershell
   gh issue view [NUMBER] --json number,title,body,labels,milestone
   ```

   Extract:
   - **User story** (As a... I want... so that...)
   - **Acceptance criteria** (Given/When/Then)
   - **Invariants, pre/post-conditions**
   - **Test plan** (unit tests + browser tests)

3. **Switch to the branch and review changes** — check out the WIP branch and examine what's been done:

   ```powershell
   git checkout issue-[NUMBER]-[desc]
   ```

// turbo
   ```powershell
   git log --oneline main..HEAD
   ```

// turbo
   ```powershell
   git diff --stat main..HEAD
   ```

   Read the changed files to understand the current state of implementation.

4. **Run automated checks** — see what currently passes:

   ```powershell
   pwsh scripts/check-all.ps1
   ```

   Note any failures — these indicate incomplete or broken work.

5. **Gap analysis** — compare the implementation against the issue's requirements. For each item, mark its status:

   **Acceptance criteria:**
   - [ ] or [x] — each Given/When/Then criterion

   **Invariants / Pre-conditions / Post-conditions:**
   - [ ] or [x] — each invariant, pre-condition, post-condition

   **Unit tests:**
   - [ ] or [x] — test exists and passes for each invariant/pre/post-condition
   - [ ] or [x] — core functionality test

   **Browser tests:**
   - [ ] or [x] — Playwright test for each acceptance criterion
   - [ ] or [x] — error/edge-case path covered

6. **Present options to the user** — based on the gap analysis:

   > "Issue #[N]: [title] — **[X of Y] criteria met**
   >
   > ### Progress
   > - [x] [completed item]
   > - [ ] [remaining item]
   >
   > ### Options
   > 1. **Continue** — finish the remaining work (plan below)
   > 2. **Abandon** — delete the branch and start fresh with `/issue-start`
   >
   > ### Plan to finish (if continuing)
   > [numbered steps for remaining work, in dependency order]
   >
   > Which would you prefer?"

7. **If continuing** — proceed with implementation of the remaining items from the plan.

8. **If abandoning** — delete the branch and return to main:

   ```powershell
   git checkout main
   git branch -D issue-[NUMBER]-[desc]
   ```

   If the branch was pushed, also delete the remote:
   ```powershell
   git push origin --delete issue-[NUMBER]-[desc]
   ```

   Then suggest running `/issue-start` to begin fresh.
