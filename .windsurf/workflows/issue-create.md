---
description: Create a new GitHub Issue with domain alignment, deduplication, and user stories
---

# Create Issue Workflow

Add new features or requirements as GitHub Issues with domain term consistency, deduplication against existing issues and PRD.md, and complete user stories.

## Context

- **PRD source of truth:** [PRD.md](../../PRD.md) (reference document)
- **Issue tracker:** GitHub Issues (living backlog — epics, stories)
- **Labels:** `area:*` (14 areas), `type:epic`, `type:story`, `cross-cutting`, `needs-ui`, `needs-storage`, `privacy-impact`, `security-impact`
- **Milestones:** Phase 0–10 (implementation priority)

## Steps

1. **Understand the addition** — read the user's request and identify:
   - What feature area it belongs to (Foundation, Kiosk, Attendance, Grading, Curriculum, Voting, Library, Communication, Member Management, Calendar, Study, Competition, Business, Privacy)
   - The core capability being described
   - Which roles are affected (Admin, Coach, Member, Trial, Guest)
   - Whether it's a new epic, a new story under an existing epic, or an extension to an existing story

2. **Check for duplicates** — search existing GitHub Issues AND PRD.md:

// turbo
   ```powershell
   gh issue list --limit 200 --json number,title,body,labels --jq '.[] | "#\(.number) \(.title) [\(.labels | map(.name) | join(", "))]"'
   ```

   - Search issue titles and bodies for keywords from the request
   - Search PRD.md for existing coverage in the matching section
   - **If the feature already exists or partially overlaps**, stop and ask the user:
     > "This appears to already be covered by:
     > - **Issue #[N]:** [title] — [quote relevant part]
     > - **PRD §X.Y:** [section name] — [quote relevant part]
     >
     > Options:
     > 1. **Skip** — existing coverage is sufficient
     > 2. **Extend** — add acceptance criteria or sub-tasks to the existing issue
     > 3. **New story** — create a new story issue under the same epic
     > 4. **New epic** — this is a distinct feature area warranting its own epic
     >
     > Which would you prefer?"

3. **Verify domain terminology** — check the addition uses terms consistently with PRD.md §1.4 Terminology. Key domain terms to verify:

   | PRD Term | Common Alternatives (flag these) |
   |----------|----------------------------------|
   | **Member** | user, student, athlete, client, customer |
   | **Coach** | instructor, teacher, trainer, sensei |
   | **Admin** | owner, operator, manager, superuser |
   | **Program** | discipline, art, martial art, style |
   | **Class** | session type, training type, lesson |
   | **Session** | class occurrence, slot, booking |
   | **Schedule** | timetable, calendar, roster |
   | **Check-in** | sign-in, attendance mark, log-in (attendance context) |
   | **Belt** | rank, grade, level |
   | **Grading** | promotion, belt test, examination |
   | **Mat Hours** | training hours, floor time, session hours |
   | **Flight Time** | (member-facing name for mat hours) |
   | **Kiosk** | terminal, check-in station, self-service |
   | **Notice** | announcement, notification, bulletin |
   | **Term** | semester, period, season, block |
   | **Holiday** | closure, break, off-day |
   | **Waiver** | liability form, disclaimer, release form |
   | **Observation** | coach note, feedback, assessment |
   | **Rotor** | curriculum rotation, schedule cycle |
   | **Theme** | category, strand, positional group |
   | **Topic** | technique, drill, lesson |
   | **Clip** | video, resource, media, link |
   | **Concept** (architecture) | entity, model, aggregate, domain object |
   | **Orchestrator** (architecture) | service, use case, command handler |
   | **Projection** (architecture) | view, query, read model |

   **If the addition uses non-standard terms**, stop and ask clarifying questions:
   > "I have some questions about the terminology to make sure this aligns with the existing PRD:
   >
   > 1. You mentioned '[non-standard term]' — in the PRD this is called '[correct term]'. Should I:
   >    a) Use the existing PRD term '[correct term]'
   >    b) Introduce '[non-standard term]' as a new concept (explain why it's distinct)
   >
   > 2. You mentioned '[ambiguous term]' — this could mean:
   >    a) [interpretation A using existing domain terms]
   >    b) [interpretation B using existing domain terms]
   >    c) Something else — please clarify

4. **Determine issue structure** — decide what to create:

   **If new epic needed:**
   - Title: `Epic: S[N] [Feature Name]`
   - Labels: `type:epic`, `area:[relevant-area]`
   - Milestone: assign to the appropriate Phase (0–10)
   - Body: description + subsection checklist

   **If new story (most common):**
   - Title: `US-[area].[sub].[seq]: [Short title]`
   - Labels: `type:story`, `area:[relevant-area]`, plus any cross-cutting labels
   - Milestone: same as parent epic
   - Body: follows the template in step 5

   **Cross-cutting labels to consider:** `cross-cutting`, `needs-ui`, `needs-storage`, `privacy-impact`, `security-impact`

5. **Draft the issue body** — use this template for story issues:

   ```markdown
   **Epic:** [Epic name] | **PRD:** §[X.Y] [Section name]

   ### User Story
   As a [role], I want to [action] so that [benefit].

   ### Acceptance Criteria
   - *Given* [precondition]
   - *When* [action]
   - *Then* [expected result]
   - *And* [additional result]

   ### Invariants
   - [Business rule that must always hold]

   ### Pre-conditions
   - [What must be true before this action]

   ### Post-conditions
   - [What must be true after this action completes]

   ### Test Plan

   **Unit tests** (must cover all invariants, pre-conditions, and post-conditions):
   - [ ] [Test: invariant — describe what must always hold]
   - [ ] [Test: pre-condition — describe what must be true before]
   - [ ] [Test: post-condition — describe what must be true after]
   - [ ] [Test: verify core functionality end-to-end at the orchestrator/projection level]

   **Browser tests** (must cover the user story via Playwright):
   - [ ] [Test: happy path — Given/When/Then from acceptance criteria]
   - [ ] [Test: edge case or error path if applicable]

   ### Implementation Notes
   - [Architectural layer affected]
   - [Related issues or dependencies]
   - [Data model changes if any]
   ```

   Guidelines for writing stories:
   - One story per distinct user action or capability
   - Use the correct role (Admin, Coach, Member, Trial, Guest)
   - Acceptance criteria should be testable (Given/When/Then)
   - Invariants are rules that must ALWAYS hold
   - Reference related issues by number (e.g., "Depends on #7")
   - Reference PRD sections for full context (e.g., "per §4.2 belt progression")

6. **Present the draft to the user for review** — show the complete issue(s) before creating:
   > "Here's what I'll create:
   >
   > **[Title]** (`type:story`, `area:[x]`, milestone: Phase [N])
   > [full body draft]
   >
   > Does this look right? Any changes before I create the issue?"

7. **Create the issue(s)** — once the user approves:

   ```powershell
   gh issue create --title "[title]" --body "[body]" --label "type:story" --label "area:[x]" --milestone "Phase [N]: [Name]"
   ```

   - If creating multiple related stories, create them in sequence
   - After creation, report the issue numbers

8. **Update the parent epic** — add a checkbox for the new story in the epic's body:

   ```powershell
   gh issue view [epic-number] --json body --jq '.body'
   ```

   - Edit the epic body to include `- [ ] #[new-issue-number] [story title]` in the checklist

9. **Update PRD.md** — add the new section to PRD.md at the correct location so the reference document stays in sync. Follow existing PRD conventions (heading, description, role access, user stories with acceptance criteria).
