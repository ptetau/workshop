---
description: Complete work on a GitHub issue — run checks, commit, create PR, and close the issue
---

# Finish Issue Workflow

Wrap up work on the current feature branch. Runs all automated checks, commits with a conventional message, creates a PR linked to the issue, waits for CI, and merges.

## Prerequisites

- On a feature branch (not `main`)
- Implementation work is complete

## Steps

1. **Identify the issue** — extract the issue number from the current branch name:

// turbo
   ```powershell
   git branch --show-current
   ```

   Branch format: `issue-[NUMBER]-[description]` → extract the number.
   If the branch doesn't follow this convention, ask the user for the issue number.

2. **Load the issue** — read the acceptance criteria so we can verify:

   ```powershell
   gh issue view [NUMBER] --json number,title,body,labels,milestone
   ```

3. **Run all automated checks** — execute the full verification suite:

   ```powershell
   pwsh scripts/check-all.ps1
   ```

   **If any check fails**, stop and report the failures. Do not proceed until all checks pass (govulncheck is warn-only).

4. **Verify unit tests** — confirm that tests exist and pass for all invariants, pre-conditions, and post-conditions:

   - For each **invariant** in the issue body → there must be a `_test.go` that asserts the rule holds
   - For each **pre-condition** → there must be a test that rejects invalid input / state
   - For each **post-condition** → there must be a test that verifies the expected outcome
   - At least one test per orchestrator/projection verifying core functionality

   **If any test is missing**, stop and write it before proceeding.

5. **Verify browser tests** — confirm that Playwright browser tests cover the user story:

   - For each **acceptance criterion** (Given/When/Then) → there must be an automated browser test that walks through the UI flow
   - Tests must cover the happy path and at least one error/edge-case path

   **If any browser test is missing**, stop and write it before proceeding.

6. **Review acceptance criteria** — walk through each criterion from the issue and verify it's been met:
   - For each *Given/When/Then*, confirm the implementation handles it
   - For each invariant, confirm it's enforced
   - Flag any criteria that aren't fully addressed

   If any criteria are unmet, report them and ask the user whether to:
   > 1. Fix them now before finishing
   > 2. Create a follow-up issue for the gaps
   > 3. Proceed anyway (with a note on the PR)

7. **Stage and commit** — create a conventional commit referencing the issue:

// turbo
   ```powershell
   git add -A
   ```

// turbo
   ```powershell
   git diff --cached --stat
   ```

   Show the user the diff summary. Commit message format:
   ```
   feat(area): short description (#NUMBER)

   - [bullet point per significant change]
   - Implements acceptance criteria for US-X.Y.Z

   Closes #NUMBER
   ```

   For bug fixes use `fix(area):`, for docs use `docs:`, for refactors use `refactor(area):`.

   ```powershell
   git commit -m "[message]"
   ```

8. **Push and create PR** — push the branch and create a pull request:

   ```powershell
   git push -u origin [branch-name]
   ```

   ```powershell
   gh pr create --title "[same as commit title]" --body "[PR body with issue link and change summary]" --milestone "[milestone name]"
   ```

   PR body template:
   ```markdown
   Closes #[NUMBER]

   ## Changes
   - [bullet per change]

   ## Acceptance Criteria
   - [x] [criterion 1 — met]
   - [x] [criterion 2 — met]
   - [ ] [criterion 3 — follow-up needed, see #NEW]

   ## Testing
   - All automated checks pass (build, vet, fmt, test -race, lintguidelines)
   - Unit tests cover all invariants, pre-conditions, and post-conditions
   - Browser tests cover all acceptance criteria (Playwright)
   ```

9. **Wait for CI** — check that GitHub Actions passes:

   ```powershell
   gh pr checks --watch
   ```

   If CI fails, diagnose and fix before merging.

10. **Merge** — squash-merge the PR and delete the branch:

   ```powershell
   gh pr merge --squash --delete-branch
   ```

11. **Verify closure** — confirm the issue was closed by the merge:

// turbo
   ```powershell
   gh issue view [NUMBER] --json state --jq '.state'
   ```

   If the issue is still open (e.g., `Closes` keyword was missed), close it manually:
   ```powershell
   gh issue close [NUMBER] --comment "Completed in PR #[PR-NUMBER]"
   ```

12. **Update epic** — if the parent epic has a checklist, check off the completed story:

    ```powershell
    gh issue view [EPIC-NUMBER] --json body --jq '.body'
    ```

    Edit the epic body to change `- [ ] #[NUMBER]` to `- [x] #[NUMBER]`.

13. **Return to main** — clean up local state:

// turbo
    ```powershell
    git checkout main && git pull origin main
    ```
