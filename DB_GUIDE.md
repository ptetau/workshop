# Database Developer Guide (SQLite + WAL)

This guide covers the "Golden Path" for the application's persistence layer, ensuring robustness, high concurrency, and alignment with our [Architecture Guidelines](GUIDELINES.md).

## 1. Directory Structure

We separate database logic into **Adapters**, respecting the [Storage Isolation](GUIDELINES.md#storage-interface) rule.

```text
/internal/
  └── adapters/
      └── storage/
          ├── db.go                # Global Connection setup & configuration
          ├── migrations/          # SQL files for schema changes
          │   ├── 001_init.sql
          │   └── 002_add_views.sql
          ├── member/              # Concept-specific storage
          │   └── store.go         # Implementation of domain.MemberStore
          └── attendance/          # Concept-specific storage
              └── store.go         # Implementation of domain.AttendanceStore
```

---

## 2. Configuration (`internal/adapters/storage/db.go`)

This is the "engine room." We use the `mattn/go-sqlite3` driver with strict pragmas to ensure high concurrency and data safety.

**Key Settings Applied:**

*   **`WAL Mode`**: Allows simultaneous readers and one writer.
*   **`Busy Timeout`**: Prevents "database locked" errors during high traffic.
*   **`Foreign Keys`**: Enforces data integrity at the DB level.

```go
package storage

import (
	"database/sql"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

func Init(dbPath string) *sql.DB {
	// 1. Connection String (The "Secret Sauce")
	// - **Driver**: `modernc.org/sqlite` (Pure Go, no CGO required).
	// - **Mode**: WAL (`?_journal_mode=WAL`).
	// - **Foreign Keys**: ON (`?_foreign_keys=on`).
	// - **Busy Timeout**: 5000ms.
	// - **Synchronous**: NORMAL (balance of speed/safety for WAL).
	dsn := dbPath + "?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on&_synchronous=NORMAL"

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		log.Fatalf("Fatal: Could not open DB: %v", err)
	}

	// 2. Connection Pool Settings
	// Since we are in WAL mode, we can allow multiple open connections.
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(1 * time.Hour)

	// 3. Health Check
	if err := db.Ping(); err != nil {
		log.Fatalf("Fatal: Database unreachable: %v", err)
	}

	log.Println("Database initialized successfully in WAL mode.")
	return db
}
```

---

## 3. Schema & Migrations

We use **Migrations** to manage schema changes. Each Concept owns its tables.

**Example `001_init.sql`:**

```sql
-- 1. Concept Tables (Isolated)
CREATE TABLE members (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    status TEXT DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE attendance (
    id TEXT PRIMARY KEY,
    member_id TEXT NOT NULL,
    checked_in_at DATETIME,
    FOREIGN KEY(member_id) REFERENCES members(id)
);

-- 2. Performance Indexes (For Projections)
CREATE INDEX idx_attendance_member ON attendance(member_id);
```

---

## 4. Query Patterns

### A. Reading (Projections)

Projections are the **only** place where cross-concept data is combined. They can read from multiple tables (or Views) to build a result.

```go
// internal/application/projections/member_summary.go

func QueryMemberSummary(ctx context.Context, db *sql.DB, memberID string) (*MemberSummary, error) {
    // Projections can JOIN tables for efficiency
    query := `
        SELECT m.id, m.email, COUNT(a.id) as total_attendance
        FROM members m
        LEFT JOIN attendance a ON m.id = a.member_id
        WHERE m.id = ?
        GROUP BY m.id`
    
    var s MemberSummary
    err := db.QueryRowContext(ctx, query, memberID).Scan(&s.ID, &s.Email, &s.TotalAttendance)
    return &s, err
}
```

### B. Writing (Concept Stores)

**Rule:** Writes are isolated to a single Concept. Complex workflows across concepts are coordinated by **Orchestrators**, not by giant transactions.

```go
// internal/adapters/storage/member/store.go

func (s *SQLiteStore) Save(ctx context.Context, m *member.Member) error {
    // 1. Transactions are for ATOMIC updates within ONE concept
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // 2. Upsert Logic
    query := `
        INSERT INTO members (id, email, status) VALUES (?, ?, ?)
        ON CONFLICT(id) DO UPDATE SET email=excluded.email, status=excluded.status`
    
    _, err = tx.ExecContext(ctx, query, m.ID, m.Email, m.Status)
    if err != nil {
        return err // e.g., constraint violation
    }

    // 3. Commit
    return tx.Commit()
}
```

---

## 5. Operations & Backups

### Managing "The File"

Because SQLite is just a file (`app.db`), you must treat it differently than a server-based DB.

1.  **Never delete the `.db-wal` or `.db-shm` files** while the app is running. These contain unwritten data.
2.  **Backups:** You cannot just copy/paste the `.db` file while the app is running (you might get a corrupted half-written file).

### The Recommended Backup Tool: Litestream

For a Go app, use [Litestream](https://litestream.io/). It runs as a sidecar process and streams changes to S3 (or Google Cloud Storage) in real-time.

**Example `litestream.yml`:**

```yaml
dbs:
  - path: /data/app.db
    replicas:
      - url: s3://my-bucket/db-backup
```

*This gives you "Point-in-Time Recovery" (you can restore your DB to the state it was in at exactly 2:43 PM yesterday).*

---

## 6. Troubleshooting

| Symptom | Cause | Solution |
| --- | --- | --- |
| **"database is locked"** | A writer is taking too long or transactions aren't closing. | Ensure all `tx` have `defer tx.Rollback()`. Check your `_busy_timeout` config. |
| **"no such table"** | Migrations didn't run or path is wrong. | Check the file path to your DB. If using relative paths, standard working dir rules apply. |
| **Performance is slow** | Missing indexes. | Run `EXPLAIN QUERY PLAN SELECT...` in a SQL tool to see if you are scanning full tables. |
