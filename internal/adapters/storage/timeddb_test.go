package storage

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"

	"workshop/internal/adapters/http/perf"
)

func openTimedTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	db.Exec("CREATE TABLE test (id TEXT PRIMARY KEY, val TEXT)")
	return db
}

// TestTimedDB_ExecContext verifies ExecContext records timing.
func TestTimedDB_ExecContext(t *testing.T) {
	db := openTimedTestDB(t)
	defer db.Close()
	collector := perf.NewCollector(100)
	tdb := NewTimedDB(db, collector)

	_, err := tdb.ExecContext(context.Background(), "INSERT INTO test (id, val) VALUES (?, ?)", "1", "hello")
	if err != nil {
		t.Fatalf("ExecContext: %v", err)
	}
	if collector.TotalRecorded() != 1 {
		t.Errorf("TotalRecorded = %d, want 1", collector.TotalRecorded())
	}
}

// TestTimedDB_QueryContext verifies QueryContext records timing.
func TestTimedDB_QueryContext(t *testing.T) {
	db := openTimedTestDB(t)
	defer db.Close()
	collector := perf.NewCollector(100)
	tdb := NewTimedDB(db, collector)

	tdb.ExecContext(context.Background(), "INSERT INTO test (id, val) VALUES (?, ?)", "1", "hello")

	rows, err := tdb.QueryContext(context.Background(), "SELECT id, val FROM test")
	if err != nil {
		t.Fatalf("QueryContext: %v", err)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		count++
		var id, val string
		rows.Scan(&id, &val)
	}
	if count != 1 {
		t.Errorf("rows = %d, want 1", count)
	}
	// 1 exec + 1 query = 2 recorded
	if collector.TotalRecorded() != 2 {
		t.Errorf("TotalRecorded = %d, want 2", collector.TotalRecorded())
	}
}

// TestTimedDB_QueryRowContext verifies QueryRowContext records timing.
func TestTimedDB_QueryRowContext(t *testing.T) {
	db := openTimedTestDB(t)
	defer db.Close()
	collector := perf.NewCollector(100)
	tdb := NewTimedDB(db, collector)

	tdb.ExecContext(context.Background(), "INSERT INTO test (id, val) VALUES (?, ?)", "1", "hello")

	var val string
	err := tdb.QueryRowContext(context.Background(), "SELECT val FROM test WHERE id = ?", "1").Scan(&val)
	if err != nil {
		t.Fatalf("QueryRowContext: %v", err)
	}
	if val != "hello" {
		t.Errorf("val = %q, want hello", val)
	}
}

// TestTimedDB_BeginTx verifies BeginTx records timing.
func TestTimedDB_BeginTx(t *testing.T) {
	db := openTimedTestDB(t)
	defer db.Close()
	collector := perf.NewCollector(100)
	tdb := NewTimedDB(db, collector)

	tx, err := tdb.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	tx.Exec("INSERT INTO test (id, val) VALUES (?, ?)", "1", "hello")
	tx.Commit()

	if collector.TotalRecorded() < 1 {
		t.Errorf("TotalRecorded = %d, want >= 1", collector.TotalRecorded())
	}
}

// TestTimedDB_NilCollector verifies TimedDB works without a collector.
func TestTimedDB_NilCollector(t *testing.T) {
	db := openTimedTestDB(t)
	defer db.Close()
	tdb := NewTimedDB(db, nil)

	_, err := tdb.ExecContext(context.Background(), "INSERT INTO test (id, val) VALUES (?, ?)", "1", "hello")
	if err != nil {
		t.Fatalf("ExecContext with nil collector: %v", err)
	}
}

// BenchmarkTimedDB_ExecContext measures per-call overhead of the timing wrapper.
func BenchmarkTimedDB_ExecContext(b *testing.B) {
	db, _ := sql.Open("sqlite", ":memory:")
	defer db.Close()
	db.Exec("CREATE TABLE bench (id INTEGER PRIMARY KEY, val TEXT)")
	collector := perf.NewCollector(perf.DefaultRingSize)
	tdb := NewTimedDB(db, collector)

	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tdb.ExecContext(ctx, "INSERT OR REPLACE INTO bench (id, val) VALUES (?, ?)", 1, "x")
	}
}

// BenchmarkTimedDB_Parallel confirms no lock contention under concurrent calls.
func BenchmarkTimedDB_Parallel(b *testing.B) {
	db, _ := sql.Open("sqlite", ":memory:")
	defer db.Close()
	db.Exec("CREATE TABLE bench (id INTEGER PRIMARY KEY, val TEXT)")
	collector := perf.NewCollector(perf.DefaultRingSize)
	tdb := NewTimedDB(db, collector)

	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tdb.QueryRowContext(ctx, "SELECT 1")
		}
	})
}
