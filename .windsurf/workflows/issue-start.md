---
description: Pick up a GitHub issue, create a branch, and load full context for implementation
---

# Start Issue Workflow

Begin work on a GitHub issue. Creates a feature branch, loads the full requirement context from the issue and PRD, identifies affected architectural layers, and produces an implementation plan.

## Prerequisites

- Clean working tree (no uncommitted changes)
- On `main` branch with latest changes pulled

## Steps

1. **Identify the issue** — if the user provides an issue number, use it. Otherwise, show recent open issues:

// turbo
   ```powershell
   gh issue list --state open --limit 20 --json number,title,labels,milestone --jq '.[] | "#\(.number) \(.title) [\(.labels | map(.name) | join(", "))] (\(.milestone.title // "no milestone"))"'
   ```

   Ask the user which issue to work on if not specified.

2. **Check for existing branch** — before creating a new branch, check if work is already in progress:

// turbo
   ```powershell
   git branch --list "issue-[NUMBER]-*" | ForEach-Object { $_.Trim() }
   ```

   Also check for a remote branch:
// turbo
   ```powershell
   git branch -r --list "origin/issue-[NUMBER]-*" | ForEach-Object { $_.Trim() }
   ```

   **If a branch exists**, stop and recommend:
   > "Found existing branch `[branch-name]` for issue #[NUMBER].
   >
   > This issue already has work in progress. Options:
   > 1. **Continue** — run `/issue-continue` to review progress and finish the work
   > 2. **Start fresh** — delete the existing branch and create a new one
   >
   > Which would you prefer?"

   If the user chooses to continue, hand off to `/issue-continue`. If they choose to start fresh, delete the branch(es) and proceed with step 6.

3. **Load issue context** — read the full issue details:

   ```powershell
   gh issue view [NUMBER] --json number,title,body,labels,milestone,assignees
   ```

   Extract from the issue body:
   - **Epic reference** (parent epic)
   - **PRD section** (§X.Y reference)
   - **User story** (As a... I want... so that...)
   - **Acceptance criteria** (Given/When/Then)
   - **Invariants, pre/post-conditions**
   - **Implementation notes**

4. **Load PRD context** — read the referenced PRD section for full requirements context. Also read adjacent sections that may be affected (e.g., if implementing §3.1 Attendance, also skim §2.2 Check-In since they're connected).

5. **Load architectural context** — read the relevant project guides based on the issue's labels:
   - `needs-storage` → read [DB_GUIDE.md](../../DB_GUIDE.md) §3 (Schema & Migrations) and check existing schema in `internal/adapters/storage/db.go`. **If new columns or tables are needed, plan a migration function.**
   - `needs-ui` → check existing templates in `templates/` and handlers in `internal/adapters/http/`
   - `security-impact` → read [OWASP.md](../../OWASP.md)
   - `privacy-impact` → read [PRIVACY.md](../../PRIVACY.md)
   - Always skim [GUIDELINES.md](../../GUIDELINES.md) for the affected layers

6. **Check for dependencies** — search for related issues that this story depends on or that should be implemented first:

// turbo
   ```powershell
   gh issue list --state open --label "[same-area-label]" --json number,title,state --jq '.[] | "#\(.number) \(.title) [\(.state)]"'
   ```

   - Flag any open issues that are prerequisites
   - Note any closed issues that provide context or prior art

7. **Ensure clean state and create branch** — verify working tree is clean, pull latest, and create a feature branch:

// turbo
   ```powershell
   git status --short
   ```

   ```powershell
   git checkout main && git pull origin main
   ```

   ```powershell
   git checkout -b issue-[NUMBER]-[short-kebab-description]
   ```

   Branch naming convention: `issue-42-member-checkin-by-name`

8. **Identify affected layers** — based on the issue, list which files/packages will need changes:

   | Layer | Package | When needed |
   |-------|---------|-------------|
   | **Domain concept** | `internal/domain/[concept]/model.go` | New entity, validation, or domain logic |
   | **Orchestrator** | `internal/application/orchestrators/[action].go` | New use case or workflow |
   | **Projection** | `internal/application/projections/[query].go` | New read/query |
   | **Storage interface** | `internal/adapters/storage/[concept]/store.go` | New persistence operations |
   | **Storage impl** | `internal/adapters/storage/[concept]/sqlite_store.go` | SQLite queries |
   | **Schema** | `internal/adapters/storage/db.go` | New tables or columns |
   | **HTTP handler** | `internal/adapters/http/handlers.go` | New endpoint logic |
   | **Routes** | `internal/adapters/http/routes.go` | New URL mappings |
   | **Middleware** | `internal/adapters/http/middleware/` | Auth, rate limiting changes |
   | **Templates** | `templates/` | New or updated UI |

9. **Design tests first** — before writing implementation code, plan the required tests:

   **Unit tests** (one `_test.go` per source file touched):
   - For each **invariant** in the issue → a test that asserts the rule holds
   - For each **pre-condition** → a test that rejects invalid input / state
   - For each **post-condition** → a test that verifies the expected outcome
   - At least one test per orchestrator/projection verifying core functionality

   **Browser tests** (Playwright, one spec per user story):
   - For each **acceptance criterion** (Given/When/Then) → an automated browser test that walks through the UI flow
   - Cover happy path and at least one error/edge-case path
   - Tests should use the Playwright MCP tools or a `_test.go` that drives a test server

   List the test cases in the implementation plan so the user can review them before coding begins.

10. **Produce implementation plan** — write a concise, ordered plan:
   - List files to create or modify, in dependency order (domain → storage → orchestrator → handler → routes → templates)
   - Note any schema migrations needed
   - Note any new dependencies or packages
   - Identify which acceptance criteria map to which implementation steps
   - List the unit tests and browser tests to be written (from step 8)
   - Flag any open questions for the user

11. **Confirm with user** — present the plan and ask:
   > "Ready to implement issue #[N]: [title]
   >
   > **Branch:** `issue-[N]-[desc]`
   > **Affected layers:** [list]
   > **Plan:** [numbered steps]
   >
   > Shall I proceed?"
