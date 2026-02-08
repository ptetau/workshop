---
description: Search open issues by topic or area, ranked by impact
---

# Issue Search Workflow

Find open issues matching a topic, keyword, or area — sorted from most impactful to least.

## Steps

1. **Parse the search intent** — extract what the user wants to find. Map their query to:
   - **Area label** (if it matches an `area:*` label): `area:communication`, `area:kiosk`, `area:attendance`, `area:grading`, `area:curriculum`, `area:library`, `area:voting`, `area:member-mgmt`, `area:calendar`, `area:study`, `area:business`, `area:competition`, `area:privacy`, `area:foundation`
   - **Cross-cutting label** (if applicable): `needs-storage`, `needs-ui`, `security-impact`, `privacy-impact`, `cross-cutting`
   - **Keyword** (free-text search against title/body): e.g. "email", "belt", "check-in"

   A single search may use multiple strategies (label + keyword) to maximise recall.

2. **Search by label** — if an area or cross-cutting label matches:

// turbo
   ```powershell
   gh issue list --state open --label "[LABEL]" --limit 100 --json number,title,labels,milestone,body --jq 'sort_by(.milestone.number // 999, .number) | .[] | "#\(.number) \(.title) [M: \(.milestone.title // "none")] [\(.labels | map(.name) | join(", "))]"'
   ```

3. **Search by keyword** — if a keyword was identified (always run this even if a label matched, to catch issues outside the label):

// turbo
   ```powershell
   gh search issues "[KEYWORD]" --repo ptetau/workshop --state open --limit 50 --json number,title,labels,url --jq '.[] | "#\(.number) \(.title) [\(.labels | map(.name) | join(", "))]"'
   ```

4. **Merge and deduplicate** — combine results from steps 2 and 3, removing duplicates by issue number.

5. **Rank by impact** — sort the merged list using this priority order:
   1. **Milestone phase** (lower = higher priority — Phase 0 before Phase 10)
   2. **Epic dependency** — issues that unblock other issues rank higher
   3. **Cross-cutting labels** — issues with `cross-cutting`, `security-impact`, or `privacy-impact` get a boost
   4. **Issue number** (lower = created earlier, likely more foundational)

6. **Present results** — display a ranked table:

   > **Search results for "[query]"**
   >
   > | # | Priority | Title | Milestone | Labels |
   > |---|----------|-------|-----------|--------|
   > | #42 | 1 | US-7.1: Send attendance summary email | Phase 3 | area:communication, needs-ui |
   > | #45 | 2 | US-7.3: Email notification preferences | Phase 4 | area:communication, needs-storage |
   > | ... | | | | |
   >
   > **Found N open issues matching "[query]".**

   If no results found, suggest:
   - Broadening the search (try a different keyword or check label spelling)
   - Creating a new issue with `/issue-create`

7. **Offer next actions** based on the results:
   - "Run `/issue-start #N` to begin work on the highest-impact issue"
   - "Run `/issue-search [different topic]` to explore another area"
   - "Run `/project-sprint` to see how these fit into the overall plan"
