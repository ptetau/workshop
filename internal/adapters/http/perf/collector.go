package perf

import (
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// DefaultRingSize is the default capacity of the ring buffer.
const DefaultRingSize = 10000

// EntryKind distinguishes request vs query entries.
type EntryKind uint8

const (
	KindRequest EntryKind = iota
	KindQuery
)

// Entry is a single timing record stored in the ring buffer.
type Entry struct {
	Kind       EntryKind
	Path       string // HTTP path or "store.Method"
	StatusCode int    // HTTP status (0 for queries)
	DurationMs float64
	Timestamp  time.Time
}

// Collector is a fixed-size ring buffer for timing entries.
// Writes are non-blocking; when full, oldest entries are overwritten.
// Aggregation happens only on read (Snapshot).
type Collector struct {
	mu      sync.Mutex
	entries []Entry
	size    int
	pos     int
	count   int64 // total entries ever written (atomic for stats)
}

// NewCollector creates a collector with the given ring buffer capacity.
// PRE: size > 0
// POST: Returns a ready-to-use collector with pre-allocated storage
func NewCollector(size int) *Collector {
	if size <= 0 {
		size = DefaultRingSize
	}
	return &Collector{
		entries: make([]Entry, size),
		size:    size,
	}
}

// Record appends an entry to the ring buffer.
// PRE: e is a valid Entry
// POST: Entry stored; if buffer full, oldest entry overwritten
// Lock hold time: single index increment + struct copy (~nanoseconds).
func (c *Collector) Record(e Entry) {
	c.mu.Lock()
	c.entries[c.pos] = e
	c.pos = (c.pos + 1) % c.size
	c.mu.Unlock()
	atomic.AddInt64(&c.count, 1)
}

// TotalRecorded returns the total number of entries ever recorded.
// PRE: none
// POST: returns count >= 0
func (c *Collector) TotalRecorded() int64 {
	return atomic.LoadInt64(&c.count)
}

// Snapshot holds aggregated performance data computed on read.
type Snapshot struct {
	TotalRequests  int64
	RequestP50Ms   float64
	RequestP95Ms   float64
	RequestP99Ms   float64
	SlowestPaths   []PathStat
	SlowestQueries []PathStat
}

// PathStat aggregates timing for a single path or store.method.
type PathStat struct {
	Path    string
	AvgMs   float64
	MaxMs   float64
	Count   int
	TotalMs float64
}

// Snapshot computes aggregated stats from the ring buffer.
// This is expensive (sorts) and should only be called on dashboard page load.
// PRE: none
// POST: Returns a Snapshot with percentiles and top-N lists
func (c *Collector) Snapshot(since time.Time, topN int) Snapshot {
	c.mu.Lock()
	// Copy entries under lock — minimal critical section
	buf := make([]Entry, c.size)
	copy(buf, c.entries)
	c.mu.Unlock()

	var requestDurations []float64
	requestStats := make(map[string]*PathStat)
	queryStats := make(map[string]*PathStat)

	for _, e := range buf {
		if e.Timestamp.IsZero() || e.Timestamp.Before(since) {
			continue
		}
		switch e.Kind {
		case KindRequest:
			requestDurations = append(requestDurations, e.DurationMs)
			s, ok := requestStats[e.Path]
			if !ok {
				s = &PathStat{Path: e.Path}
				requestStats[e.Path] = s
			}
			s.Count++
			s.TotalMs += e.DurationMs
			if e.DurationMs > s.MaxMs {
				s.MaxMs = e.DurationMs
			}
		case KindQuery:
			s, ok := queryStats[e.Path]
			if !ok {
				s = &PathStat{Path: e.Path}
				queryStats[e.Path] = s
			}
			s.Count++
			s.TotalMs += e.DurationMs
			if e.DurationMs > s.MaxMs {
				s.MaxMs = e.DurationMs
			}
		}
	}

	// Compute averages
	for _, s := range requestStats {
		s.AvgMs = s.TotalMs / float64(s.Count)
	}
	for _, s := range queryStats {
		s.AvgMs = s.TotalMs / float64(s.Count)
	}

	snap := Snapshot{
		TotalRequests:  c.TotalRecorded(),
		SlowestPaths:   topByAvg(requestStats, topN),
		SlowestQueries: topByAvg(queryStats, topN),
	}

	if len(requestDurations) > 0 {
		sort.Float64s(requestDurations)
		snap.RequestP50Ms = percentile(requestDurations, 50)
		snap.RequestP95Ms = percentile(requestDurations, 95)
		snap.RequestP99Ms = percentile(requestDurations, 99)
	}

	return snap
}

// percentile returns the p-th percentile from a sorted slice.
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := (p / 100) * float64(len(sorted)-1)
	lower := int(math.Floor(idx))
	upper := int(math.Ceil(idx))
	if lower == upper || upper >= len(sorted) {
		return sorted[lower]
	}
	frac := idx - float64(lower)
	return sorted[lower]*(1-frac) + sorted[upper]*frac
}

// topByAvg returns the top N paths sorted by average duration (descending).
func topByAvg(stats map[string]*PathStat, n int) []PathStat {
	list := make([]PathStat, 0, len(stats))
	for _, s := range stats {
		list = append(list, *s)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].AvgMs > list[j].AvgMs
	})
	if len(list) > n {
		list = list[:n]
	}
	return list
}
