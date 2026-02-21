package projections

import (
	"context"
	"fmt"
	"time"

	memberStore "workshop/internal/adapters/storage/member"
	domainAttendance "workshop/internal/domain/attendance"
)

// TrainingVolumeAttendanceStore defines the attendance store interface needed by the training volume projection.
type TrainingVolumeAttendanceStore interface {
	ListByMemberIDAndDateRange(ctx context.Context, memberID string, startDate string, endDate string) ([]domainAttendance.Attendance, error)
	ListByDateRange(ctx context.Context, startDate string, endDate string) ([]domainAttendance.Attendance, error)
}

// TrainingVolumeMemberStore defines the member store interface needed by the training volume projection.
type TrainingVolumeMemberStore interface {
	Count(ctx context.Context, filter memberStore.ListFilter) (int, error)
}

// GetTrainingVolumeQuery carries input for the training volume projection.
type GetTrainingVolumeQuery struct {
	MemberID string
	Range    string // "month" or "year"
	Compare  bool
	Now      time.Time // optional: if zero, time.Now() is used
}

// TrainingVolumeSeries represents a line on the graph.
type TrainingVolumeSeries struct {
	Name   string
	Values []float64
}

// GetTrainingVolumeResult carries the output of the training volume projection.
type GetTrainingVolumeResult struct {
	Range     string
	StartDate string // YYYY-MM-DD
	EndDate   string // YYYY-MM-DD
	Buckets   []string
	Series    []TrainingVolumeSeries
}

// GetTrainingVolumeDeps holds dependencies for the training volume projection.
type GetTrainingVolumeDeps struct {
	AttendanceStore TrainingVolumeAttendanceStore
	MemberStore     TrainingVolumeMemberStore
}

// QueryGetTrainingVolume aggregates attendance volume for charting.
// PRE: query.MemberID is non-empty, query.Range is "month" or "year"
// POST: Returns bucketed series for the member and an all-member average
func QueryGetTrainingVolume(ctx context.Context, query GetTrainingVolumeQuery, deps GetTrainingVolumeDeps) (GetTrainingVolumeResult, error) {
	now := query.Now
	if now.IsZero() {
		now = time.Now()
	}

	if query.MemberID == "" {
		return GetTrainingVolumeResult{}, fmt.Errorf("member_id is required")
	}

	if query.Range != "month" && query.Range != "year" {
		return GetTrainingVolumeResult{}, fmt.Errorf("range must be month or year")
	}

	var start, end time.Time
	if query.Range == "month" {
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
		end = start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	} else {
		start = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.Local)
		end = time.Date(now.Year(), 12, 31, 23, 59, 59, int(time.Second-time.Nanosecond), time.Local)
	}

	startDate := start.Format("2006-01-02")
	endDate := end.Format("2006-01-02")

	// Member attendance records for the range.
	memberRecords, err := deps.AttendanceStore.ListByMemberIDAndDateRange(ctx, query.MemberID, startDate, endDate)
	if err != nil {
		return GetTrainingVolumeResult{}, err
	}

	// All-member records for average line.
	allRecords, err := deps.AttendanceStore.ListByDateRange(ctx, startDate, endDate)
	if err != nil {
		return GetTrainingVolumeResult{}, err
	}

	activeCount := 0
	if deps.MemberStore != nil {
		active, err := deps.MemberStore.Count(ctx, memberStore.ListFilter{Status: "active"})
		if err == nil {
			activeCount += active
		}
		trial, err := deps.MemberStore.Count(ctx, memberStore.ListFilter{Status: "trial"})
		if err == nil {
			activeCount += trial
		}
	}
	if activeCount <= 0 {
		activeCount = 1
	}

	buckets := buildTrainingVolumeBuckets(query.Range, now)
	memberSeries := make([]float64, len(buckets))
	avgSeries := make([]float64, len(buckets))

	bucketIndex := make(map[string]int, len(buckets))
	for i, b := range buckets {
		bucketIndex[b] = i
	}

	addRecordsToSeries(query.Range, now, bucketIndex, memberSeries, memberRecords)

	allCounts := make([]float64, len(buckets))
	addRecordsToSeries(query.Range, now, bucketIndex, allCounts, allRecords)
	for i := range allCounts {
		avgSeries[i] = allCounts[i] / float64(activeCount)
	}

	series := []TrainingVolumeSeries{
		{Name: "You", Values: memberSeries},
		{Name: "All members avg", Values: avgSeries},
	}

	// Comparison mode is only supported for month range.
	if query.Range == "month" && query.Compare {
		lastMonthNow := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local).AddDate(0, -1, 0)
		lmStart := time.Date(lastMonthNow.Year(), lastMonthNow.Month(), 1, 0, 0, 0, 0, time.Local)
		lmEnd := lmStart.AddDate(0, 1, 0).Add(-time.Nanosecond)

		syLastYearNow := time.Date(now.Year()-1, now.Month(), 1, 0, 0, 0, 0, time.Local)
		syStart := time.Date(syLastYearNow.Year(), syLastYearNow.Month(), 1, 0, 0, 0, 0, time.Local)
		syEnd := syStart.AddDate(0, 1, 0).Add(-time.Nanosecond)

		lmRecords, err := deps.AttendanceStore.ListByMemberIDAndDateRange(ctx, query.MemberID, lmStart.Format("2006-01-02"), lmEnd.Format("2006-01-02"))
		if err == nil {
			lmSeries := make([]float64, len(buckets))
			addRecordsToSeries(query.Range, now, bucketIndex, lmSeries, lmRecords)
			series = append(series, TrainingVolumeSeries{Name: "Last month", Values: lmSeries})
		}

		syRecords, err := deps.AttendanceStore.ListByMemberIDAndDateRange(ctx, query.MemberID, syStart.Format("2006-01-02"), syEnd.Format("2006-01-02"))
		if err == nil {
			sySeries := make([]float64, len(buckets))
			addRecordsToSeries(query.Range, now, bucketIndex, sySeries, syRecords)
			series = append(series, TrainingVolumeSeries{Name: "Same month last year", Values: sySeries})
		}
	}

	return GetTrainingVolumeResult{
		Range:     query.Range,
		StartDate: startDate,
		EndDate:   endDate,
		Buckets:   buckets,
		Series:    series,
	}, nil
}

func buildTrainingVolumeBuckets(rng string, now time.Time) []string {
	if rng == "year" {
		return []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	}

	// Month: bucket by day-of-month ("1".."31") for the current month.
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	days := end.Day()
	b := make([]string, 0, days)
	for i := 1; i <= days; i++ {
		b = append(b, fmt.Sprintf("%d", i))
	}
	return b
}

func addRecordsToSeries(rng string, now time.Time, bucketIndex map[string]int, series []float64, records []domainAttendance.Attendance) {
	for _, r := range records {
		key := bucketKeyForAttendance(rng, now, r)
		idx, ok := bucketIndex[key]
		if !ok {
			continue
		}
		series[idx]++
	}
}

func bucketKeyForAttendance(rng string, now time.Time, a domainAttendance.Attendance) string {
	if rng == "year" {
		m := a.CheckInTime.Month()
		return [...]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}[int(m)-1]
	}

	// Month mode: key by day-of-month, but aligned to the *current month* day numbers.
	// When comparing against last month / last year, we still map to day numbers.
	return fmt.Sprintf("%d", a.CheckInTime.Day())
}
