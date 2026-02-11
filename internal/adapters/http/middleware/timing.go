package middleware

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"workshop/internal/adapters/http/perf"
)

// DefaultSlowRequestMs is the default threshold for slow request warnings.
const DefaultSlowRequestMs = 200

// slowRequestMs is the cached threshold (read via atomic after first load).
var slowRequestMs int64

var slowRequestOnce sync.Once

// getSlowRequestThreshold returns the slow-request threshold in milliseconds.
func getSlowRequestThreshold() float64 {
	slowRequestOnce.Do(func() {
		ms := DefaultSlowRequestMs
		if v := os.Getenv("WORKSHOP_SLOW_REQUEST_MS"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				ms = n
			}
		}
		atomic.StoreInt64(&slowRequestMs, int64(ms))
	})
	return float64(atomic.LoadInt64(&slowRequestMs))
}

// requestIDCounter is an atomic counter for request IDs.
var requestIDCounter uint64

// statusWriter wraps http.ResponseWriter to capture the status code.
type statusWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader captures the status code and delegates to the underlying ResponseWriter.
// PRE: code is a valid HTTP status code
// POST: status stored, header written to underlying ResponseWriter
func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

// statusWriterPool reduces allocations on the hot path.
var statusWriterPool = sync.Pool{
	New: func() any {
		return &statusWriter{}
	},
}

// Timing returns middleware that logs request duration.
// Requests to /static/ are excluded.
// Normal requests log at DEBUG; slow requests (above threshold) log at WARN.
// If collector is non-nil, entries are recorded for the perf dashboard.
func Timing(collector *perf.Collector) func(http.Handler) http.Handler {
	threshold := getSlowRequestThreshold()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			// Skip static assets
			if strings.HasPrefix(path, "/static/") {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			reqID := atomic.AddUint64(&requestIDCounter, 1)

			sw := statusWriterPool.Get().(*statusWriter)
			sw.ResponseWriter = w
			sw.status = http.StatusOK
			defer func() {
				durationMs := float64(time.Since(start).Microseconds()) / 1000.0

				if durationMs >= threshold {
					slog.Warn("slow_request",
						"request_id", reqID,
						"method", r.Method,
						"path", path,
						"status", sw.status,
						"duration_ms", durationMs,
					)
				} else {
					slog.Debug("request",
						"request_id", reqID,
						"method", r.Method,
						"path", path,
						"status", sw.status,
						"duration_ms", durationMs,
					)
				}

				if collector != nil {
					collector.Record(perf.Entry{
						Kind:       perf.KindRequest,
						Path:       r.Method + " " + path,
						StatusCode: sw.status,
						DurationMs: durationMs,
						Timestamp:  start,
					})
				}

				sw.ResponseWriter = nil
				statusWriterPool.Put(sw)
			}()

			next.ServeHTTP(sw, r)
		})
	}
}
