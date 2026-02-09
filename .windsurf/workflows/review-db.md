---
description: Review changed files for database concerns (schema, queries, storage patterns, operations)
---

# Database Review

Focused review against [DB_GUIDE.md](../../DB_GUIDE.md) — SQLite + WAL patterns, schema, queries, and operations.

## Steps

1. **Run build and tests:**

// turbo
```powershell
go build ./...
```

// turbo
```powershell
go test -race -count=1 ./internal/adapters/storage/...
```

2. **Migration check** — if the PR touches `sqlite_store.go`, `model.go`, or `db.go` with new/changed columns:
   - A new migration function must exist in `db.go` (never modify an existing migration)
   - The migration must be registered in the `migrations` slice
   - A data survival test must exist in `db_test.go` for the new migration
   - `expectedTables` in `db_test.go` must be updated if new tables were added

   **If any of these are missing, flag it as a required change.**

3. **Schema** (`internal/adapters/storage/db.go`) — if tables or columns changed:
   - Schema changes go in new migration functions (never modify existing ones)
   - IDs use `TEXT PRIMARY KEY` with UUID strings
   - Timestamps use `TEXT NOT NULL` stored as RFC3339 (`2006-01-02T15:04:05Z07:00`)
   - Required fields have `NOT NULL` constraints
   - Foreign keys declared with `FOREIGN KEY (col) REFERENCES table(id)`
   - Performance indexes on frequently queried columns (foreign keys, dates)
   - Restricted data (injuries, observations) in separate tables

   | Don't | Do |
   |-------|------|
   | `id INTEGER PRIMARY KEY AUTOINCREMENT` | `id TEXT PRIMARY KEY` with UUID strings |
   | `created_at DATETIME DEFAULT CURRENT_TIMESTAMP` | `created_at TEXT NOT NULL` with RFC3339 strings |
   | Omit `NOT NULL` on required fields | Always add `NOT NULL` constraints |
   | Put injury columns in member table | Separate `injury` table with restricted access |
   | Skip indexes on foreign keys | `CREATE INDEX idx_attendance_member ON attendance(member_id)` |

4. **Store structure** (`internal/adapters/storage/*/`) — if stores changed:
   - Split into `store.go` (interface) and `sqlite_store.go` (implementation)
   - One store per concept — no cross-concept stores
   - Interface has: `GetByID`, `Save`, `Delete`, `List`, plus concept-specific methods

   | Don't | Do |
   |-------|------|
   | Interface and implementation in one file | `store.go` (interface) + `sqlite_store.go` (implementation) |
   | `memberStore` also has attendance methods | One store per concept |
   | Missing `Count()` or `List()` on store | Include full CRUD + List + any concept-specific queries |

5. **Write queries** — if INSERT/UPDATE/DELETE queries changed:
   - All writes wrapped in transactions (`BeginTx` + `defer tx.Rollback()` + `Commit()`)
   - Transactions scoped to single concept — orchestrators coordinate across concepts
   - Upserts use `INSERT ... ON CONFLICT(id) DO UPDATE SET` (not `INSERT OR REPLACE`)
   - All queries use `?` parameterized placeholders
   - Store methods accept value types, not pointers

   | Don't | Do |
   |-------|------|
   | `db.ExecContext(ctx, query, ...)` without transaction | `tx, _ := db.BeginTx(ctx, nil); defer tx.Rollback()` |
   | `INSERT OR REPLACE INTO member` | `INSERT INTO member (...) ON CONFLICT(id) DO UPDATE SET ...` |
   | `fmt.Sprintf("WHERE id = '%s'", id)` | `"WHERE id = ?"` with parameterized placeholder |
   | `Save(ctx, m *member.Member)` | `Save(ctx, entity domain.Member)` — value types |
   | Writing to member + attendance in one TX | One TX per concept — orchestrator coordinates |

6. **Read queries** — if SELECT queries or projections changed:
   - Use `QueryRowContext` / `QueryContext` (context-aware)
   - List columns explicitly (no `SELECT *`)
   - Projections may JOIN across concept tables for reads
   - Return value types, not pointers

   | Don't | Do |
   |-------|------|
   | `db.QueryRow("SELECT ...")` | `db.QueryRowContext(ctx, "SELECT ...")` |
   | `SELECT * FROM member` | `SELECT id, email, name, status FROM member` |
   | `return &MemberSummary{...}, nil` | `return MemberSummary{...}, nil` — value types |

7. **Scan helpers** — if scan functions changed:
   - Scan all columns in the same order as the SELECT
   - Handle nullable columns with `sql.NullString`, `sql.NullInt64`, etc.
   - Parse timestamp strings with `time.Parse(time.RFC3339Nano, str)`
   - Convert integer booleans: `entity.Flag = intVal != 0`

   | Don't | Do |
   |-------|------|
   | Scan columns in different order than SELECT | Match SELECT column order exactly |
   | `var lockedUntil string` for nullable column | `var lockedUntil sql.NullString` |
   | `time.Parse("2006-01-02", str)` | `time.Parse(time.RFC3339Nano, str)` or the full format |

8. **Configuration** — if DSN or connection setup changed:
   - Driver: `modernc.org/sqlite` (pure Go, no CGO)
   - DSN uses `_pragma()` syntax: `?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)&_pragma=synchronous(NORMAL)`
   - Connection pool: `SetMaxOpenConns(25)`, `SetMaxIdleConns(25)`

   | Don't | Do |
   |-------|------|
   | `import _ "github.com/mattn/go-sqlite3"` | `import _ "modernc.org/sqlite"` — pure Go, cross-compiles |
   | `?_journal_mode=WAL` query string | `?_pragma=journal_mode(WAL)` for modernc driver |
   | `SetMaxOpenConns(1)` | `SetMaxOpenConns(25)` — WAL supports concurrent readers |

9. **Operations** — if backup or DB file handling changed:
   - Never delete `.db-wal` or `.db-shm` while app is running
   - Backups use `VACUUM INTO` or Litestream — never raw file copy
   - DB stored in persistent path with proper permissions (`/opt/workshop/`)

   | Don't | Do |
   |-------|------|
   | `cp workshop.db backup.db` while app running | `VACUUM INTO 'backup.db'` or Litestream |
   | Delete `.db-wal` or `.db-shm` manually | Stop the app first, or let SQLite manage them |

10. **Report** — summarise findings:
   - ✅ Storage tests passed / ❌ failures
   - List any DB violations with file path and line number
   - Recommendation: **Approve** / **Request Changes**
