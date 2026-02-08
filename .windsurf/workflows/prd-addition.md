---
description: Add a new feature or requirement to the PRD with domain alignment, deduplication, and user stories
---

# PRD Addition Workflow

Safely add new features or requirements to [PRD.md](../../PRD.md) with domain term consistency, deduplication, and complete user stories.

## Steps

1. **Understand the addition** — read the user's request and identify:
   - What feature area it belongs to (Foundation, Kiosk, Attendance, Grading, Communication, etc.)
   - The core capability being described
   - Which roles are affected

2. **Check for duplicates** — search PRD.md for existing coverage:
   - Search for keywords from the request across all sections
   - Search for synonyms and related concepts
   - Check the user stories in the matching feature area
   - **If the feature already exists or partially overlaps**, stop and ask the user:
     > "This appears to already be covered in §X.Y — [section name]. Here's what's there:
     > [quote the existing content]
     >
     > Options:
     > 1. Skip — the existing coverage is sufficient
     > 2. Extend — add the new aspects to the existing section
     > 3. Replace — rewrite the section with the updated requirements
     > 4. Separate — this is different enough to warrant its own section
     >
     > Which would you prefer?"

3. **Verify domain terminology** — read the PRD's existing glossary and conventions, then check the addition uses terms consistently. Key domain terms to verify:

   | PRD Term | Common Alternatives (flag these) |
   |----------|----------------------------------|
   | **Member** | user, student, athlete, client, customer |
   | **Coach** | instructor, teacher, trainer, sensei |
   | **Admin** | owner, operator, manager, superuser |
   | **Program** | discipline, art, martial art, style |
   | **Class Type** | class, session type, training type |
   | **Schedule** | timetable, calendar, roster |
   | **Check-in** | sign-in, attendance mark, log-in (attendance context) |
   | **Belt** | rank, grade, level |
   | **Grading** | promotion, belt test, examination |
   | **Mat Hours** | training hours, floor time, session hours |
   | **Kiosk** | terminal, check-in station, self-service |
   | **Notice** | announcement, notification, bulletin |
   | **Term** | semester, period, season, block |
   | **Holiday** | closure, break, off-day |
   | **Waiver** | liability form, disclaimer, release form |
   | **Observation** | coach note, feedback, assessment |
   | **Theme** | topic, curriculum item, lesson plan |
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
   >
   > [continue for each terminology issue]"

4. **Determine placement** — identify where in the PRD the addition belongs:
   - Find the correct feature area section (§1–§14)
   - Determine if it's a new subsection or extends an existing one
   - If it's a cross-cutting concern, check if it belongs in §1 Foundation
   - Note the numbering convention (e.g., §3.4 would follow §3.3)

5. **Draft user stories** — write user stories in the PRD's existing format:
   ```
   **US-[area][number]: [Title]**
   As a [role], I want to [action] so that [benefit].

   *Acceptance Criteria:*
   - [ ] [Criterion 1]
   - [ ] [Criterion 2]

   *Invariants:*
   - [Business rule that must always hold]

   *Pre-conditions:*
   - [What must be true before this action]

   *Post-conditions:*
   - [What must be true after this action completes]
   ```

   Guidelines for writing stories:
   - One story per distinct user action or capability
   - Use the correct role (Admin, Coach, Member, Trial, Guest)
   - Acceptance criteria should be testable and specific
   - Invariants are rules that must ALWAYS hold (e.g., "A member can only belong to one program at a time")
   - Pre-conditions are what must be true BEFORE the action (e.g., "Member must have an active account")
   - Post-conditions are what must be true AFTER success (e.g., "Attendance record is created and mat hours are updated")
   - Reference existing concepts by their PRD section (e.g., "per §4.2 belt progression")

6. **Draft the PRD section** — write the addition following existing PRD conventions:
   - Section heading with correct numbering
   - Brief description of the feature
   - Role access table (which roles can do what)
   - User stories with acceptance criteria, invariants, pre/post-conditions
   - Data model additions if needed (new fields, tables, or relationships)
   - Reference related sections (e.g., "See also §3.1 Check-In Flow")

7. **Present the draft to the user for review** — show the complete addition before making any edits:
   > "Here's the proposed PRD addition for §X.Y:
   >
   > [full draft]
   >
   > Does this look right? Any changes before I add it?"

8. **Apply the edit** — once the user approves, add the section to PRD.md at the correct location. Do NOT overwrite or modify existing sections unless explicitly asked.
