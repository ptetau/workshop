# QA Report: Performance

Ref: [QA_AUDIT_PLAN.md](QA_AUDIT_PLAN.md) · Area 11

## Summary
Performance infrastructure is solid — `TimedDB` wrapper logs slow queries. Calendar index was recently added (migration 14). Storage layer uses standard SQLite patterns. Several areas need review.

## Findings

| # | Sev | Finding | File:Line | Fix |
|---|-----|---------|-----------|-----|
| 1 | P3 | TimedDB slow query threshold defaults to 200ms, configurable via `SLOW_QUERY_MS` env var — good | timeddb.go:17-26 | No action |
| 2 | P2 | TimedDB logs both `slog.Warn` and records to `query_timing` table — dual instrumentation is thorough | timeddb.go:52-80 | No action |
| 3 | P2 | Calendar `ListByDateRange` has composite index on `(start_date, end_date)` via migration 14 — good | calendar store, migration 14 | No action |
| 4 | P2 | Member list uses server-side pagination with `LIMIT/OFFSET` — correct for large datasets | handlers.go:548+ | No action |
| 5 | P2 | Grading readiness endpoint fetches all members then filters in Go — potential N+1 if member count grows large | grading store | Consider adding a dedicated readiness query with JOINs |
| 6 | P2 | Inactive members query `/api/members/inactive?days=N` likely scans full attendance table | inactive handler | Ensure index on `check_in.created_at` or `check_in.member_id` |
| 7 | P3 | Email recipients search uses `LIKE '%query%'` — will table-scan for large member sets | email handlers | Acceptable for current scale; FTS if membership exceeds 1000+ |
| 8 | P2 | Curriculum overview endpoint makes multiple sequential queries (class types → rotors → themes → topics → schedules) per request | curriculum handlers | Could be optimised with JOINs but acceptable at current scale |
| 9 | P3 | All admin list endpoints (terms, holidays, accounts, milestones) return full datasets — no pagination | Various admin templates | Fine for small datasets; add pagination if data grows |
| 10 | P2 | Training log loads multiple API calls in parallel from JS (stats, goals, badges, attendance, milestones) — good parallel loading pattern | member_training_log.html | No action — parallel is correct |

## Recommendations
1. **Verify indexes exist** on `check_in(member_id)`, `check_in(created_at)`, `attendance(class_date)`
2. **Monitor grading readiness** query performance as member count grows
3. No critical performance issues found at current scale
