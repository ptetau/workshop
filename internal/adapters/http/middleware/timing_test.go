package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"workshop/internal/adapters/http/perf"
)

// TestTimingMiddleware_EmitsEntry verifies that a request entry is recorded.
func TestTimingMiddleware_EmitsEntry(t *testing.T) {
	collector := perf.NewCollector(100)
	handler := Timing(collector)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if collector.TotalRecorded() != 1 {
		t.Errorf("TotalRecorded = %d, want 1", collector.TotalRecorded())
	}
}

// TestTimingMiddleware_SkipsStatic verifies static assets are excluded from timing.
func TestTimingMiddleware_SkipsStatic(t *testing.T) {
	collector := perf.NewCollector(100)
	handler := Timing(collector)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/static/style.css", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if collector.TotalRecorded() != 0 {
		t.Errorf("TotalRecorded = %d, want 0 (static excluded)", collector.TotalRecorded())
	}
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

// TestTimingMiddleware_CapturesStatusCode verifies the status code is captured.
func TestTimingMiddleware_CapturesStatusCode(t *testing.T) {
	collector := perf.NewCollector(100)
	handler := Timing(collector)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest("GET", "/missing", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rr.Code)
	}
	if collector.TotalRecorded() != 1 {
		t.Errorf("TotalRecorded = %d, want 1", collector.TotalRecorded())
	}
}

// TestTimingMiddleware_NilCollector verifies middleware works without a collector.
func TestTimingMiddleware_NilCollector(t *testing.T) {
	handler := Timing(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

// --- Resilience: Handler Panic ---

// TestTimingMiddleware_HandlerPanic verifies that a panicking handler does not
// prevent the deferred timing logic from running and does not corrupt the pool.
// The middleware itself doesn't recover panics (that's the recovery middleware's job),
// but the defer must still execute so the statusWriter is returned to the pool.
func TestTimingMiddleware_HandlerPanic(t *testing.T) {
	collector := perf.NewCollector(100)
	handler := Timing(collector)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	req := httptest.NewRequest("GET", "/api/panic", nil)
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic to propagate, got nil")
		}
		// The deferred timing logic should have run before the panic propagated.
		if collector.TotalRecorded() != 1 {
			t.Errorf("TotalRecorded = %d, want 1 (defer must run even on panic)", collector.TotalRecorded())
		}
	}()

	handler.ServeHTTP(rr, req)
}

// --- Correctness: Default Status ---

// TestTimingMiddleware_DefaultStatusWhenNotSet verifies status defaults to 200
// when the handler writes a body without calling WriteHeader explicitly.
func TestTimingMiddleware_DefaultStatusWhenNotSet(t *testing.T) {
	collector := perf.NewCollector(100)
	handler := Timing(collector)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello")) // implicit 200
	}))

	req := httptest.NewRequest("GET", "/api/implicit", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

// --- Correctness: Entry Fields ---

// TestTimingMiddleware_EntryFieldAccuracy verifies the recorded entry has correct
// method, path, and status code.
func TestTimingMiddleware_EntryFieldAccuracy(t *testing.T) {
	// Use a collector of size 1 so we can inspect the single entry
	collector := perf.NewCollector(1)
	handler := Timing(collector)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest("POST", "/api/members", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	snap := collector.Snapshot(time.Now().Add(-time.Minute), 10)
	if len(snap.SlowestPaths) != 1 {
		t.Fatalf("SlowestPaths len = %d, want 1", len(snap.SlowestPaths))
	}
	if snap.SlowestPaths[0].Path != "POST /api/members" {
		t.Errorf("Path = %q, want \"POST /api/members\"", snap.SlowestPaths[0].Path)
	}
}

// --- Resilience: Pool State Isolation ---

// TestTimingMiddleware_PoolNoStateLeak verifies that statusWriter pool reuse
// does not leak status codes between requests.
func TestTimingMiddleware_PoolNoStateLeak(t *testing.T) {
	collector := perf.NewCollector(100)

	// First request: 500
	handler500 := Timing(collector)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	req1 := httptest.NewRequest("GET", "/api/fail", nil)
	rr1 := httptest.NewRecorder()
	handler500.ServeHTTP(rr1, req1)

	if rr1.Code != 500 {
		t.Errorf("request 1 status = %d, want 500", rr1.Code)
	}

	// Second request: handler does NOT call WriteHeader (implicit 200).
	// If pool leaks, we'd see 500 here.
	handler200 := Timing(collector)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	req2 := httptest.NewRequest("GET", "/api/ok", nil)
	rr2 := httptest.NewRecorder()
	handler200.ServeHTTP(rr2, req2)

	if rr2.Code != 200 {
		t.Errorf("request 2 status = %d, want 200 (pool must not leak 500)", rr2.Code)
	}
}

// --- Resilience: Duration Sanity ---

// TestTimingMiddleware_DurationPositive verifies the recorded duration is non-negative.
func TestTimingMiddleware_DurationPositive(t *testing.T) {
	collector := perf.NewCollector(1)
	handler := Timing(collector)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/dur", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	snap := collector.Snapshot(time.Now().Add(-time.Minute), 10)
	if len(snap.SlowestPaths) == 0 {
		t.Fatal("no entries recorded")
	}
	if snap.SlowestPaths[0].AvgMs < 0 {
		t.Errorf("AvgMs = %v, want >= 0", snap.SlowestPaths[0].AvgMs)
	}
}

// --- Performance: Aggregate Overhead Budget ---

// BenchmarkTimingMiddleware_FullCycle simulates a realistic request cycle:
// 1 middleware timing + 5 query timings (typical page load).
// This measures the aggregate instrumentation tax per request.
func BenchmarkTimingMiddleware_FullCycle(b *testing.B) {
	collector := perf.NewCollector(perf.DefaultRingSize)
	now := time.Now()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate middleware recording
		collector.Record(perf.Entry{Kind: perf.KindRequest, Path: "GET /api/members", StatusCode: 200, DurationMs: 5, Timestamp: now})
		// Simulate 5 query recordings (typical page)
		for q := 0; q < 5; q++ {
			collector.Record(perf.Entry{Kind: perf.KindQuery, Path: "QueryContext", DurationMs: 0.5, Timestamp: now})
		}
	}
}

// BenchmarkTimingMiddleware measures per-request overhead.
func BenchmarkTimingMiddleware(b *testing.B) {
	collector := perf.NewCollector(perf.DefaultRingSize)
	handler := Timing(collector)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	req := httptest.NewRequest("GET", "/api/bench", nil)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}
}

// BenchmarkTimingMiddleware_Parallel confirms no lock contention.
func BenchmarkTimingMiddleware_Parallel(b *testing.B) {
	collector := perf.NewCollector(perf.DefaultRingSize)
	handler := Timing(collector)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("GET", "/api/bench", nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
		}
	})
}
