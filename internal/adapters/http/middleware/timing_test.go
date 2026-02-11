package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
