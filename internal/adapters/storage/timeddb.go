package storage

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"workshop/internal/adapters/http/perf"
)

// SQLDB is the database interface used by all stores.
// Both *sql.DB and *TimedDB satisfy this interface.
type SQLDB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

// Compile-time check that *sql.DB satisfies SQLDB.
var _ SQLDB = (*sql.DB)(nil)

// DefaultSlowQueryMs is the default threshold for slow query warnings.
const DefaultSlowQueryMs = 50

var slowQueryMs int64
var slowQueryOnce sync.Once

// getSlowQueryThreshold returns the slow-query threshold in milliseconds.
func getSlowQueryThreshold() float64 {
	slowQueryOnce.Do(func() {
		ms := DefaultSlowQueryMs
		if v := os.Getenv("WORKSHOP_SLOW_QUERY_MS"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				ms = n
			}
		}
		atomic.StoreInt64(&slowQueryMs, int64(ms))
	})
	return float64(atomic.LoadInt64(&slowQueryMs))
}

// TimedDB wraps a *sql.DB to log slow queries and optionally record to a collector.
// Satisfies the SQLDB interface so it can be passed to any store constructor.
type TimedDB struct {
	db        *sql.DB
	collector *perf.Collector
	threshold float64
}

// Compile-time check that *TimedDB satisfies SQLDB.
var _ SQLDB = (*TimedDB)(nil)

// NewTimedDB wraps a *sql.DB with timing instrumentation.
// PRE: db is a valid database connection
// POST: Returns a TimedDB that logs slow queries and records to collector
func NewTimedDB(db *sql.DB, collector *perf.Collector) *TimedDB {
	return &TimedDB{
		db:        db,
		collector: collector,
		threshold: getSlowQueryThreshold(),
	}
}

// RawDB returns the underlying *sql.DB (needed for migrations and pool config).
// PRE: none
// POST: returns the unwrapped *sql.DB
func (t *TimedDB) RawDB() *sql.DB {
	return t.db
}

// logQuery logs and optionally records a query timing.
func (t *TimedDB) logQuery(op string, start time.Time) {
	durationMs := float64(time.Since(start).Microseconds()) / 1000.0

	if durationMs >= t.threshold {
		slog.Warn("slow_query",
			"op", op,
			"duration_ms", durationMs,
		)
	} else {
		slog.Debug("query",
			"op", op,
			"duration_ms", durationMs,
		)
	}

	if t.collector != nil {
		t.collector.Record(perf.Entry{
			Kind:       perf.KindQuery,
			Path:       op,
			DurationMs: durationMs,
			Timestamp:  start,
		})
	}
}

// ExecContext wraps sql.DB.ExecContext with timing.
// PRE: ctx is valid, query is non-empty
// POST: query executed, timing recorded to collector
func (t *TimedDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	start := time.Now()
	result, err := t.db.ExecContext(ctx, query, args...)
	t.logQuery("ExecContext", start)
	return result, err
}

// QueryContext wraps sql.DB.QueryContext with timing.
// PRE: ctx is valid, query is non-empty
// POST: query executed, timing recorded to collector
func (t *TimedDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	start := time.Now()
	rows, err := t.db.QueryContext(ctx, query, args...)
	t.logQuery("QueryContext", start)
	return rows, err
}

// QueryRowContext wraps sql.DB.QueryRowContext with timing.
// PRE: ctx is valid, query is non-empty
// POST: query executed, timing recorded to collector
func (t *TimedDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	start := time.Now()
	row := t.db.QueryRowContext(ctx, query, args...)
	t.logQuery("QueryRowContext", start)
	return row
}

// BeginTx wraps sql.DB.BeginTx with timing.
// PRE: ctx is valid
// POST: transaction started, timing recorded to collector
func (t *TimedDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	start := time.Now()
	tx, err := t.db.BeginTx(ctx, opts)
	t.logQuery("BeginTx", start)
	return tx, err
}

// Close closes the underlying database connection.
// PRE: none
// POST: database connection closed
func (t *TimedDB) Close() error {
	return t.db.Close()
}

// Ping verifies the database connection.
// PRE: none
// POST: returns nil if connection is alive
func (t *TimedDB) Ping() error {
	return t.db.Ping()
}

// SetMaxOpenConns sets the maximum number of open connections.
// PRE: n >= 0
// POST: pool limit updated
// INVARIANT: db is not nil
func (t *TimedDB) SetMaxOpenConns(n int) {
	t.db.SetMaxOpenConns(n)
}

// SetMaxIdleConns sets the maximum number of idle connections.
// PRE: n >= 0
// POST: idle pool limit updated
// INVARIANT: db is not nil
func (t *TimedDB) SetMaxIdleConns(n int) {
	t.db.SetMaxIdleConns(n)
}
