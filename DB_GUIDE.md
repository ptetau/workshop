# Database Developer Guide (SQLite + WAL)

This guide covers the "Golden Path" for the application's persistence layer, ensuring robustness, high concurrency, and alignment with our [Architecture Guidelines](GUIDELINES.md).

---

## 1. Directory Structure

We separate database logic into **Adapters**, respecting the [Storage Isolation](GUIDELINES.md#storage-interface) rule.

```text
/internal/
  └── adapters/
      └── storage/
          ├── db.go                  # Schema (CREATE TABLE IF NOT EXISTS) + connection setup
          ├── account/
          │   ├── store.go           # Store interface (GetByID, Save, Delete, List, Count)
          │   └── sqlite_store.go    # SQLite implementation of Store
          ├── member/
          │   ├── store.go
          │   └── sqlite_store.go
          └── attendance/
              ├── store.go
              └── sqlite_store.go
```

| Don't | Do |
|-------|------|
| Put schema in separate SQL migration files | Define schema inline in `db.go` with `CREATE TABLE IF NOT EXISTS` |
| Combine interface and implementation in one file | Split into `store.go` (interface) and `sqlite_store.go` (implementation) |
| Share a store across multiple concepts | One `store.go` + `sqlite_store.go` per concept |

---

## 2. Configuration (`internal/adapters/storage/db.go`)

This is the "engine room." We use the `modernc.org/sqlite` driver (pure Go, no CGO required) with strict pragmas to ensure high concurrency and data safety.

**Key Settings Applied:**

*   **`WAL Mode`**: Allows simultaneous readers and one writer.
*   **`Busy Timeout`**: 5000ms — prevents "database locked" errors during high traffic.
*   **`Foreign Keys`**: Enforces data integrity at the DB level.
*   **`Synchronous = NORMAL`**: Balance of speed and safety for WAL mode.

```go
package main

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

// DSN uses _pragma() syntax for modernc.org/sqlite
dsn := "workshop.db?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)&_pragma=synchronous(NORMAL)"

db, err := sql.Open("sqlite", dsn)
if err != nil {
	log.Fatalf("failed to open database: %v", err)
}

// Connection pool for WAL mode
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(25)

// Health check
if err := db.Ping(); err != nil {
	log.Fatalf("database unreachable: %v", err)
}
```

| Don't | Do |
|-------|------|
| Use `mattn/go-sqlite3` (requires CGO) | Use `modernc.org/sqlite` (pure Go, cross-compiles) |
| Use `?_journal_mode=WAL` query string syntax | Use `?_pragma=journal_mode(WAL)` syntax for modernc driver |
| Skip the busy timeout | Always set `_pragma=busy_timeout(5000)` to avoid "database is locked" errors |
| Set `SetMaxOpenConns(1)` | Set `SetMaxOpenConns(25)` — WAL mode supports concurrent readers |
| Forget to enable foreign keys | Always set `_pragma=foreign_keys(ON)` in the DSN |

---

## 3. Schema

All tables are defined in `db.go` using `CREATE TABLE IF NOT EXISTS`. The schema runs on every startup — idempotent by design.

```go
// internal/adapters/storage/db.go
schema := `
CREATE TABLE IF NOT EXISTS member (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    program TEXT NOT NULL,
    status TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS attendance (
    id TEXT PRIMARY KEY,
    member_id TEXT NOT NULL,
    check_in_time TEXT NOT NULL,
    schedule_id TEXT,
    class_date TEXT,
    FOREIGN KEY (member_id) REFERENCES member(id)
);
`
```

| Don't | Do |
|-------|------|
| Use `DATETIME` or `INTEGER` for timestamps | Use `TEXT NOT NULL` and store as RFC3339 strings (`2006-01-02T15:04:05Z07:00`) |
| Use `AUTOINCREMENT` integer IDs | Use `TEXT PRIMARY KEY` with UUID strings |
| Omit `NOT NULL` on required fields | Always add `NOT NULL` constraints |
| Forget foreign key declarations | Always declare `FOREIGN KEY (col) REFERENCES table(id)` |
| Put restricted data (injuries, observations) in the member table | Store in separate tables with stricter access methods |
| Skip performance indexes | Add `CREATE INDEX` for frequently queried columns (foreign keys, dates) |

---

## 4. Query Patterns

### A. Reading (Projections)

Projections are the **only** place where cross-concept data is combined. They can JOIN across tables.

```go
// internal/application/projections/member_summary.go

func QueryMemberSummary(ctx context.Context, db *sql.DB, memberID string) (MemberSummary, error) {
    query := `
        SELECT m.id, m.email, COUNT(a.id) as total_attendance
        FROM member m
        LEFT JOIN attendance a ON m.id = a.member_id
        WHERE m.id = ?
        GROUP BY m.id`
    
    var s MemberSummary
    err := db.QueryRowContext(ctx, query, memberID).Scan(&s.ID, &s.Email, &s.TotalAttendance)
    return s, err
}
```

| Don't | Do |
|-------|------|
| Mutate data in a projection | Projections are read-only — use orchestrators for writes |
| Use `db.QueryRow()` without context | Always use `db.QueryRowContext(ctx, ...)` |
| Return pointer types from projections | Return value types: `(MemberSummary, error)` not `(*MemberSummary, error)` |
| `SELECT *` | List columns explicitly — resilient to schema changes |

### B. Writing (Concept Stores)

**Rule:** Writes are isolated to a single concept. Complex workflows across concepts are coordinated by **Orchestrators**, not by giant transactions.

```go
// internal/adapters/storage/member/sqlite_store.go

func (s *SQLiteStore) Save(ctx context.Context, entity domain.Member) error {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    query := `
        INSERT INTO member (id, email, name, program, status) VALUES (?, ?, ?, ?, ?)
        ON CONFLICT(id) DO UPDATE SET email=excluded.email, name=excluded.name,
        program=excluded.program, status=excluded.status`
    
    _, err = tx.ExecContext(ctx, query, entity.ID, entity.Email, entity.Name, entity.Program, entity.Status)
    if err != nil {
        return err
    }

    return tx.Commit()
}
```

| Don't | Do |
|-------|------|
| Accept pointer types: `Save(ctx, m *member.Member)` | Accept value types: `Save(ctx, entity domain.Member)` |
| Write to multiple concept tables in one transaction | One transaction per concept — orchestrators coordinate across concepts |
| Use `db.Exec()` for writes | Always wrap writes in `BeginTx` + `defer tx.Rollback()` + `tx.Commit()` |
| Use `INSERT OR REPLACE` (deletes then re-inserts) | Use `INSERT ... ON CONFLICT(id) DO UPDATE SET` (true upsert) |
| Build SQL with `fmt.Sprintf` and user input | Always use `?` parameterized placeholders |
| Put business logic in the store | Stores are pure data access — business logic goes in concept methods or orchestrators |

---

## 5. Operations & Backups

### Managing "The File"

Because SQLite is just a file (`workshop.db`), you must treat it differently than a server-based DB.

1.  **Never delete the `.db-wal` or `.db-shm` files** while the app is running. These contain unwritten data.
2.  **Backups:** You cannot just `cp` the `.db` file while the app is running — you'll get a corrupted half-written file.

| Don't | Do |
|-------|------|
| `cp workshop.db backup.db` while app is running | Use `VACUUM INTO 'backup.db'` or Litestream for safe backups |
| Delete `.db-wal` or `.db-shm` while app is running | Stop the app first, or let SQLite manage these files |
| Store the DB in a temp directory | Store in a persistent path with proper permissions (`/opt/workshop/`) |

### The Recommended Backup Tool: Litestream

For continuous replication, use [Litestream](https://litestream.io/). It runs as a sidecar and streams WAL changes to S3 in real-time.

```yaml
# litestream.yml
dbs:
  - path: /opt/workshop/workshop.db
    replicas:
      - url: s3://my-bucket/workshop-backup
```

---

## 6. Troubleshooting

| Symptom | Cause | Solution |
| --- | --- | --- |
| **"database is locked"** | Writer taking too long or transactions not closing | Ensure all `tx` have `defer tx.Rollback()`. Verify `_pragma=busy_timeout(5000)` in DSN. |
| **"no such table"** | Schema didn't run or wrong DB path | Check that `storage.InitDB(db)` runs before any queries. Verify the DSN file path. |
| **Performance is slow** | Missing indexes | Run `EXPLAIN QUERY PLAN SELECT...` to check for full table scans. Add indexes on foreign keys and date columns. |
| **"foreign key mismatch"** | `_pragma=foreign_keys(ON)` not set | Ensure foreign keys are enabled in the DSN. Without it, FK constraints are silently ignored. |
| **Data loss after crash** | Deleted `.db-wal` while app was running | Never touch WAL/SHM files manually. Use `VACUUM INTO` for safe backups. |
