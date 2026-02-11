package perf

import (
	"strconv"
	"sync"
	"testing"
	"time"
)

// TestCollector_Record_And_Snapshot verifies basic record and snapshot functionality.
func TestCollector_Record_And_Snapshot(t *testing.T) {
	c := NewCollector(100)
	now := time.Now()

	c.Record(Entry{Kind: KindRequest, Path: "GET /foo", StatusCode: 200, DurationMs: 10, Timestamp: now})
	c.Record(Entry{Kind: KindRequest, Path: "GET /foo", StatusCode: 200, DurationMs: 30, Timestamp: now})
	c.Record(Entry{Kind: KindQuery, Path: "ExecContext", DurationMs: 5, Timestamp: now})

	snap := c.Snapshot(now.Add(-time.Minute), 10)
	if snap.TotalRequests != 3 {
		t.Errorf("TotalRequests = %d, want 3", snap.TotalRequests)
	}
	if len(snap.SlowestPaths) != 1 {
		t.Fatalf("SlowestPaths len = %d, want 1", len(snap.SlowestPaths))
	}
	if snap.SlowestPaths[0].AvgMs != 20 {
		t.Errorf("AvgMs = %v, want 20", snap.SlowestPaths[0].AvgMs)
	}
	if len(snap.SlowestQueries) != 1 {
		t.Fatalf("SlowestQueries len = %d, want 1", len(snap.SlowestQueries))
	}
}

// TestCollector_RingBuffer_Overwrites verifies oldest entries are overwritten when full.
func TestCollector_RingBuffer_Overwrites(t *testing.T) {
	c := NewCollector(3)
	now := time.Now()

	for i := 0; i < 5; i++ {
		c.Record(Entry{Kind: KindRequest, Path: "GET /x", DurationMs: float64(i), Timestamp: now})
	}

	if c.TotalRecorded() != 5 {
		t.Errorf("TotalRecorded = %d, want 5", c.TotalRecorded())
	}

	// Buffer of size 3 should only have entries 2,3,4 (overwrote 0,1)
	snap := c.Snapshot(now.Add(-time.Minute), 10)
	if len(snap.SlowestPaths) != 1 {
		t.Fatalf("SlowestPaths len = %d, want 1", len(snap.SlowestPaths))
	}
	if snap.SlowestPaths[0].Count != 3 {
		t.Errorf("Count = %d, want 3 (ring buffer kept last 3)", snap.SlowestPaths[0].Count)
	}
}

// TestCollector_Percentiles verifies P50/P95/P99 calculation.
func TestCollector_Percentiles(t *testing.T) {
	c := NewCollector(200)
	now := time.Now()

	// Insert 100 entries: durations 1..100
	for i := 1; i <= 100; i++ {
		c.Record(Entry{Kind: KindRequest, Path: "GET /p", DurationMs: float64(i), Timestamp: now})
	}

	snap := c.Snapshot(now.Add(-time.Minute), 10)
	if snap.RequestP50Ms < 49 || snap.RequestP50Ms > 51 {
		t.Errorf("P50 = %v, want ~50", snap.RequestP50Ms)
	}
	if snap.RequestP95Ms < 94 || snap.RequestP95Ms > 96 {
		t.Errorf("P95 = %v, want ~95", snap.RequestP95Ms)
	}
	if snap.RequestP99Ms < 98 || snap.RequestP99Ms > 100 {
		t.Errorf("P99 = %v, want ~99", snap.RequestP99Ms)
	}
}

// TestCollector_Snapshot_FiltersBySince verifies old entries are excluded.
func TestCollector_Snapshot_FiltersBySince(t *testing.T) {
	c := NewCollector(100)
	old := time.Now().Add(-2 * time.Hour)
	recent := time.Now()

	c.Record(Entry{Kind: KindRequest, Path: "GET /old", DurationMs: 100, Timestamp: old})
	c.Record(Entry{Kind: KindRequest, Path: "GET /new", DurationMs: 10, Timestamp: recent})

	snap := c.Snapshot(time.Now().Add(-1*time.Hour), 10)
	if len(snap.SlowestPaths) != 1 {
		t.Fatalf("SlowestPaths len = %d, want 1 (old entry filtered)", len(snap.SlowestPaths))
	}
	if snap.SlowestPaths[0].Path != "GET /new" {
		t.Errorf("Path = %q, want GET /new", snap.SlowestPaths[0].Path)
	}
}

// TestCollector_ConcurrentWrites verifies goroutine safety of Record.
func TestCollector_ConcurrentWrites(t *testing.T) {
	c := NewCollector(1000)
	now := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				c.Record(Entry{Kind: KindRequest, Path: "GET /c", DurationMs: float64(n), Timestamp: now})
			}
		}(i)
	}
	wg.Wait()
	if c.TotalRecorded() != 1000 {
		t.Errorf("TotalRecorded = %d, want 1000", c.TotalRecorded())
	}
}

// --- Correctness: Edge Cases ---

// TestCollector_EmptySnapshot verifies Snapshot returns zero values when no entries exist.
func TestCollector_EmptySnapshot(t *testing.T) {
	c := NewCollector(100)
	snap := c.Snapshot(time.Now().Add(-time.Hour), 10)

	if snap.TotalRequests != 0 {
		t.Errorf("TotalRequests = %d, want 0", snap.TotalRequests)
	}
	if snap.RequestP50Ms != 0 || snap.RequestP95Ms != 0 || snap.RequestP99Ms != 0 {
		t.Errorf("percentiles should be 0 for empty snapshot, got P50=%v P95=%v P99=%v",
			snap.RequestP50Ms, snap.RequestP95Ms, snap.RequestP99Ms)
	}
	if len(snap.SlowestPaths) != 0 {
		t.Errorf("SlowestPaths should be empty, got %d", len(snap.SlowestPaths))
	}
	if len(snap.SlowestQueries) != 0 {
		t.Errorf("SlowestQueries should be empty, got %d", len(snap.SlowestQueries))
	}
}

// TestCollector_SingleEntry verifies percentiles with exactly one entry.
func TestCollector_SingleEntry(t *testing.T) {
	c := NewCollector(100)
	now := time.Now()
	c.Record(Entry{Kind: KindRequest, Path: "GET /one", DurationMs: 42, Timestamp: now})

	snap := c.Snapshot(now.Add(-time.Minute), 10)
	if snap.RequestP50Ms != 42 {
		t.Errorf("P50 = %v, want 42", snap.RequestP50Ms)
	}
	if snap.RequestP99Ms != 42 {
		t.Errorf("P99 = %v, want 42", snap.RequestP99Ms)
	}
}

// TestCollector_AllSameDuration verifies percentiles when all entries have identical duration.
func TestCollector_AllSameDuration(t *testing.T) {
	c := NewCollector(100)
	now := time.Now()
	for i := 0; i < 50; i++ {
		c.Record(Entry{Kind: KindRequest, Path: "GET /same", DurationMs: 7.5, Timestamp: now})
	}

	snap := c.Snapshot(now.Add(-time.Minute), 10)
	if snap.RequestP50Ms != 7.5 || snap.RequestP95Ms != 7.5 || snap.RequestP99Ms != 7.5 {
		t.Errorf("all percentiles should be 7.5, got P50=%v P95=%v P99=%v",
			snap.RequestP50Ms, snap.RequestP95Ms, snap.RequestP99Ms)
	}
}

// TestCollector_TopN_MorePathsThanN verifies topN truncation.
func TestCollector_TopN_MorePathsThanN(t *testing.T) {
	c := NewCollector(100)
	now := time.Now()
	for i := 0; i < 20; i++ {
		path := "GET /path" + strconv.Itoa(i)
		c.Record(Entry{Kind: KindRequest, Path: path, DurationMs: float64(i), Timestamp: now})
	}

	snap := c.Snapshot(now.Add(-time.Minute), 5)
	if len(snap.SlowestPaths) != 5 {
		t.Errorf("SlowestPaths len = %d, want 5 (topN truncation)", len(snap.SlowestPaths))
	}
	// Verify sorted descending by AvgMs
	for i := 1; i < len(snap.SlowestPaths); i++ {
		if snap.SlowestPaths[i].AvgMs > snap.SlowestPaths[i-1].AvgMs {
			t.Errorf("SlowestPaths not sorted: [%d].AvgMs=%v > [%d].AvgMs=%v",
				i, snap.SlowestPaths[i].AvgMs, i-1, snap.SlowestPaths[i-1].AvgMs)
		}
	}
}

// TestCollector_TopN_FewerPathsThanN verifies topN when fewer paths than N.
func TestCollector_TopN_FewerPathsThanN(t *testing.T) {
	c := NewCollector(100)
	now := time.Now()
	c.Record(Entry{Kind: KindRequest, Path: "GET /only", DurationMs: 5, Timestamp: now})

	snap := c.Snapshot(now.Add(-time.Minute), 10)
	if len(snap.SlowestPaths) != 1 {
		t.Errorf("SlowestPaths len = %d, want 1", len(snap.SlowestPaths))
	}
}

// TestCollector_KindSeparation verifies requests and queries are bucketed independently.
func TestCollector_KindSeparation(t *testing.T) {
	c := NewCollector(100)
	now := time.Now()
	c.Record(Entry{Kind: KindRequest, Path: "GET /api", DurationMs: 10, Timestamp: now})
	c.Record(Entry{Kind: KindRequest, Path: "GET /api", DurationMs: 20, Timestamp: now})
	c.Record(Entry{Kind: KindQuery, Path: "ExecContext", DurationMs: 5, Timestamp: now})
	c.Record(Entry{Kind: KindQuery, Path: "QueryContext", DurationMs: 3, Timestamp: now})

	snap := c.Snapshot(now.Add(-time.Minute), 10)
	if len(snap.SlowestPaths) != 1 {
		t.Fatalf("SlowestPaths len = %d, want 1 (only request paths)", len(snap.SlowestPaths))
	}
	if snap.SlowestPaths[0].Path != "GET /api" {
		t.Errorf("SlowestPaths[0].Path = %q, want GET /api", snap.SlowestPaths[0].Path)
	}
	if len(snap.SlowestQueries) != 2 {
		t.Fatalf("SlowestQueries len = %d, want 2", len(snap.SlowestQueries))
	}
}

// TestCollector_NewCollector_InvalidSize verifies size <= 0 uses default.
func TestCollector_NewCollector_InvalidSize(t *testing.T) {
	c := NewCollector(0)
	if c.size != DefaultRingSize {
		t.Errorf("size = %d, want %d (default)", c.size, DefaultRingSize)
	}
	c2 := NewCollector(-5)
	if c2.size != DefaultRingSize {
		t.Errorf("size = %d, want %d (default)", c2.size, DefaultRingSize)
	}
}

// --- Resilience: Concurrent Read + Write ---

// TestCollector_SnapshotDuringConcurrentWrites verifies snapshot consistency under load.
func TestCollector_SnapshotDuringConcurrentWrites(t *testing.T) {
	c := NewCollector(1000)
	now := time.Now()

	// Start writer goroutines
	var wg sync.WaitGroup
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					c.Record(Entry{Kind: KindRequest, Path: "GET /c", DurationMs: 1, Timestamp: now})
				}
			}
		}()
	}

	// Take snapshots concurrently — must not panic or return invalid data
	for i := 0; i < 20; i++ {
		snap := c.Snapshot(now.Add(-time.Minute), 10)
		// Percentiles must be non-negative
		if snap.RequestP50Ms < 0 || snap.RequestP95Ms < 0 || snap.RequestP99Ms < 0 {
			t.Fatalf("negative percentile: P50=%v P95=%v P99=%v",
				snap.RequestP50Ms, snap.RequestP95Ms, snap.RequestP99Ms)
		}
	}

	close(done)
	wg.Wait()
}

// --- Resilience: Memory Bounds ---

// TestCollector_MemoryBounded verifies ring buffer never grows beyond initial allocation.
func TestCollector_MemoryBounded(t *testing.T) {
	const ringSize = 100
	c := NewCollector(ringSize)
	now := time.Now()

	// Write 10x the ring size
	for i := 0; i < ringSize*10; i++ {
		c.Record(Entry{Kind: KindRequest, Path: "GET /mem", DurationMs: 1, Timestamp: now})
	}

	// Internal slice length must never exceed ringSize
	if len(c.entries) != ringSize {
		t.Errorf("entries len = %d, want %d (memory bounded)", len(c.entries), ringSize)
	}
	if cap(c.entries) != ringSize {
		t.Errorf("entries cap = %d, want %d (no realloc)", cap(c.entries), ringSize)
	}
}

// BenchmarkCollectorRecord measures per-call cost of Record().
func BenchmarkCollectorRecord(b *testing.B) {
	c := NewCollector(DefaultRingSize)
	e := Entry{Kind: KindRequest, Path: "GET /bench", StatusCode: 200, DurationMs: 1.5, Timestamp: time.Now()}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Record(e)
	}
}

// BenchmarkCollectorRecord_Parallel confirms no lock contention under concurrent writes.
func BenchmarkCollectorRecord_Parallel(b *testing.B) {
	c := NewCollector(DefaultRingSize)
	e := Entry{Kind: KindRequest, Path: "GET /bench", StatusCode: 200, DurationMs: 1.5, Timestamp: time.Now()}
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Record(e)
		}
	})
}

// BenchmarkCollectorSnapshot measures cost of computing percentiles + top-N.
func BenchmarkCollectorSnapshot(b *testing.B) {
	c := NewCollector(DefaultRingSize)
	now := time.Now()
	for i := 0; i < DefaultRingSize; i++ {
		c.Record(Entry{Kind: KindRequest, Path: "GET /bench", StatusCode: 200, DurationMs: float64(i % 100), Timestamp: now})
	}
	since := now.Add(-time.Hour)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Snapshot(since, 10)
	}
}
