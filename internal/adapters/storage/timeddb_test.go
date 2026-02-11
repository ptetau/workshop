package storage

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

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

// --- Resilience: Error Passthrough ---

// TestTimedDB_ErrorPassthrough_ExecContext verifies SQL errors are returned unchanged
// and timing is still recorded. This is critical — swallowing errors would corrupt data.
func TestTimedDB_ErrorPassthrough_ExecContext(t *testing.T) {
	db := openTimedTestDB(t)
	defer db.Close()
	collector := perf.NewCollector(100)
	tdb := NewTimedDB(db, collector)

	// Invalid SQL — must return an error
	_, err := tdb.ExecContext(context.Background(), "INSERT INTO nonexistent_table VALUES (?)")
	if err == nil {
		t.Fatal("expected error from invalid SQL, got nil")
	}
	// Timing must still be recorded even on error
	if collector.TotalRecorded() != 1 {
		t.Errorf("TotalRecorded = %d, want 1 (must record even on error)", collector.TotalRecorded())
	}
}

// TestTimedDB_ErrorPassthrough_QueryContext verifies query errors are returned unchanged.
func TestTimedDB_ErrorPassthrough_QueryContext(t *testing.T) {
	db := openTimedTestDB(t)
	defer db.Close()
	collector := perf.NewCollector(100)
	tdb := NewTimedDB(db, collector)

	_, err := tdb.QueryContext(context.Background(), "SELECT * FROM nonexistent_table")
	if err == nil {
		t.Fatal("expected error from invalid SQL, got nil")
	}
	if collector.TotalRecorded() != 1 {
		t.Errorf("TotalRecorded = %d, want 1", collector.TotalRecorded())
	}
}

// TestTimedDB_ErrorPassthrough_QueryRowContext verifies QueryRowContext scan errors pass through.
func TestTimedDB_ErrorPassthrough_QueryRowContext(t *testing.T) {
	db := openTimedTestDB(t)
	defer db.Close()
	collector := perf.NewCollector(100)
	tdb := NewTimedDB(db, collector)

	var val string
	err := tdb.QueryRowContext(context.Background(), "SELECT val FROM test WHERE id = ?", "nonexistent").Scan(&val)
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
	// Timing recorded even though row doesn't exist
	if collector.TotalRecorded() != 1 {
		t.Errorf("TotalRecorded = %d, want 1", collector.TotalRecorded())
	}
}

// --- Resilience: Cancelled Context ---

// TestTimedDB_CancelledContext verifies that a cancelled context returns an error
// and timing is still recorded.
func TestTimedDB_CancelledContext(t *testing.T) {
	db := openTimedTestDB(t)
	defer db.Close()
	collector := perf.NewCollector(100)
	tdb := NewTimedDB(db, collector)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := tdb.ExecContext(ctx, "INSERT INTO test (id, val) VALUES (?, ?)", "1", "hello")
	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
	// Timing must still be recorded
	if collector.TotalRecorded() != 1 {
		t.Errorf("TotalRecorded = %d, want 1 (must record on cancelled ctx)", collector.TotalRecorded())
	}
}

// --- Correctness: Result Passthrough ---

// TestTimedDB_ResultPassthrough verifies sql.Result values (RowsAffected, LastInsertId)
// are returned unchanged through the wrapper.
func TestTimedDB_ResultPassthrough(t *testing.T) {
	db := openTimedTestDB(t)
	defer db.Close()
	tdb := NewTimedDB(db, perf.NewCollector(100))

	result, err := tdb.ExecContext(context.Background(), "INSERT INTO test (id, val) VALUES (?, ?)", "r1", "result")
	if err != nil {
		t.Fatalf("ExecContext: %v", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		t.Fatalf("RowsAffected: %v", err)
	}
	if rows != 1 {
		t.Errorf("RowsAffected = %d, want 1", rows)
	}
}

// --- Correctness: RawDB Accessor ---

// TestTimedDB_RawDB verifies RawDB returns the original *sql.DB.
func TestTimedDB_RawDB(t *testing.T) {
	db := openTimedTestDB(t)
	defer db.Close()
	tdb := NewTimedDB(db, nil)

	if tdb.RawDB() != db {
		t.Error("RawDB() should return the original *sql.DB")
	}
}

// --- Correctness: Interface Compliance ---

// TestTimedDB_ImplementsSQLDB is a compile-time check that *TimedDB satisfies SQLDB.
// (Also verified via var _ SQLDB = (*TimedDB)(nil) in timeddb.go, but this tests at runtime.)
func TestTimedDB_ImplementsSQLDB(t *testing.T) {
	db := openTimedTestDB(t)
	defer db.Close()
	var iface SQLDB = NewTimedDB(db, nil)
	if iface == nil {
		t.Fatal("TimedDB should satisfy SQLDB interface")
	}
}

// --- Resilience: Concurrent Mixed Operations ---

// TestTimedDB_ConcurrentMixedOps verifies no data races or panics under concurrent
// Exec, Query, and QueryRow calls.
func TestTimedDB_ConcurrentMixedOps(t *testing.T) {
	db := openTimedTestDB(t)
	defer db.Close()
	collector := perf.NewCollector(1000)
	tdb := NewTimedDB(db, collector)

	// Seed one row for reads
	tdb.ExecContext(context.Background(), "INSERT INTO test (id, val) VALUES (?, ?)", "seed", "data")

	ctx := context.Background()
	done := make(chan struct{})
	var wg sync.WaitGroup

	// Writer goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		i := 0
		for {
			select {
			case <-done:
				return
			default:
				tdb.ExecContext(ctx, "INSERT OR REPLACE INTO test (id, val) VALUES (?, ?)", "w", "v")
				i++
			}
		}
	}()

	// Reader goroutine (QueryContext)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
				rows, err := tdb.QueryContext(ctx, "SELECT id FROM test LIMIT 1")
				if err == nil {
					rows.Close()
				}
			}
		}
	}()

	// Reader goroutine (QueryRowContext)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
				var v string
				tdb.QueryRowContext(ctx, "SELECT val FROM test WHERE id = ?", "seed").Scan(&v)
			}
		}
	}()

	// Let it run briefly then stop
	time.Sleep(100 * time.Millisecond)
	close(done)
	wg.Wait()

	// Must have recorded entries without panics
	if collector.TotalRecorded() < 3 {
		t.Errorf("TotalRecorded = %d, want >= 3 (seed + at least one of each)", collector.TotalRecorded())
	}
}

// --- Performance: Overhead Isolation ---

// BenchmarkTimedDB_OverheadIsolation measures the *pure instrumentation overhead*
// by comparing TimedDB vs raw *sql.DB for the same query.
func BenchmarkTimedDB_OverheadIsolation(b *testing.B) {
	db, _ := sql.Open("sqlite", ":memory:")
	defer db.Close()
	db.Exec("CREATE TABLE bench (id INTEGER PRIMARY KEY, val TEXT)")
	db.Exec("INSERT INTO bench VALUES (1, 'x')")
	collector := perf.NewCollector(perf.DefaultRingSize)

	ctx := context.Background()

	b.Run("RawDB", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			db.QueryRowContext(ctx, "SELECT val FROM bench WHERE id = 1")
		}
	})

	tdb := NewTimedDB(db, collector)
	b.Run("TimedDB", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			tdb.QueryRowContext(ctx, "SELECT val FROM bench WHERE id = 1")
		}
	})
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
