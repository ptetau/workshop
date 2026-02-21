package projections

import (
	"context"
	"testing"
	"time"

	memberStore "workshop/internal/adapters/storage/member"
	domainAttendance "workshop/internal/domain/attendance"
)

type mockTrainingVolumeAttendanceStore struct {
	member []domainAttendance.Attendance
	all    []domainAttendance.Attendance
}

// ListByMemberIDAndDateRange returns seeded member attendance within range.
// PRE: memberID, startDate, endDate are non-empty
// POST: Returns matching records
func (m *mockTrainingVolumeAttendanceStore) ListByMemberIDAndDateRange(_ context.Context, memberID string, startDate string, endDate string) ([]domainAttendance.Attendance, error) {
	var out []domainAttendance.Attendance
	for _, a := range m.member {
		d := a.CheckInTime.Format("2006-01-02")
		if a.MemberID == memberID && d >= startDate && d <= endDate {
			out = append(out, a)
		}
	}
	return out, nil
}

// ListByDateRange returns seeded all-member attendance within range.
// PRE: startDate, endDate are non-empty
// POST: Returns matching records
func (m *mockTrainingVolumeAttendanceStore) ListByDateRange(_ context.Context, startDate string, endDate string) ([]domainAttendance.Attendance, error) {
	var out []domainAttendance.Attendance
	for _, a := range m.all {
		d := a.CheckInTime.Format("2006-01-02")
		if d >= startDate && d <= endDate {
			out = append(out, a)
		}
	}
	return out, nil
}

type mockTrainingVolumeMemberStore struct {
	active int
	trial  int
}

// Count returns counts for active/trial used to compute average line.
// PRE: filter has valid parameters
// POST: Returns the count
func (m *mockTrainingVolumeMemberStore) Count(_ context.Context, filter memberStore.ListFilter) (int, error) {
	switch filter.Status {
	case "active":
		return m.active, nil
	case "trial":
		return m.trial, nil
	default:
		return 0, nil
	}
}

// TestQueryGetTrainingVolume_MonthBucketsAndAverage verifies month daily bucketing and average line.
func TestQueryGetTrainingVolume_MonthBucketsAndAverage(t *testing.T) {
	loc := time.Local
	now := time.Date(2026, 2, 15, 12, 0, 0, 0, loc)
	memberID := "m1"

	att := &mockTrainingVolumeAttendanceStore{
		member: []domainAttendance.Attendance{
			{ID: "a1", MemberID: memberID, CheckInTime: time.Date(2026, 2, 1, 8, 0, 0, 0, loc)},
			{ID: "a2", MemberID: memberID, CheckInTime: time.Date(2026, 2, 1, 18, 0, 0, 0, loc)},
			{ID: "a3", MemberID: memberID, CheckInTime: time.Date(2026, 2, 2, 8, 0, 0, 0, loc)},
		},
		all: []domainAttendance.Attendance{
			// same as member + one other member on day 2
			{ID: "a1", MemberID: memberID, CheckInTime: time.Date(2026, 2, 1, 8, 0, 0, 0, loc)},
			{ID: "a2", MemberID: memberID, CheckInTime: time.Date(2026, 2, 1, 18, 0, 0, 0, loc)},
			{ID: "a3", MemberID: memberID, CheckInTime: time.Date(2026, 2, 2, 8, 0, 0, 0, loc)},
			{ID: "a4", MemberID: "m2", CheckInTime: time.Date(2026, 2, 2, 9, 0, 0, 0, loc)},
		},
	}

	deps := GetTrainingVolumeDeps{
		AttendanceStore: att,
		MemberStore:     &mockTrainingVolumeMemberStore{active: 3, trial: 1}, // denom=4
	}

	res, err := QueryGetTrainingVolume(context.Background(), GetTrainingVolumeQuery{MemberID: memberID, Range: "month", Now: now}, deps)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if res.Range != "month" {
		t.Fatalf("range=%q, want month", res.Range)
	}
	if len(res.Buckets) != 28 {
		t.Fatalf("buckets=%d, want 28", len(res.Buckets))
	}
	if len(res.Series) < 2 {
		t.Fatalf("series=%d, want >=2", len(res.Series))
	}

	you := res.Series[0].Values
	avg := res.Series[1].Values
	if you[0] != 2 {
		t.Fatalf("day1=%v, want 2", you[0])
	}
	if you[1] != 1 {
		t.Fatalf("day2=%v, want 1", you[1])
	}
	// all counts: day1=2, day2=2 => avg per member (4) = 0.5
	if avg[0] != 0.5 {
		t.Fatalf("avg day1=%v, want 0.5", avg[0])
	}
	if avg[1] != 0.5 {
		t.Fatalf("avg day2=%v, want 0.5", avg[1])
	}
}

// TestQueryGetTrainingVolume_CompareAddsSeries verifies comparison mode adds series for month range.
func TestQueryGetTrainingVolume_CompareAddsSeries(t *testing.T) {
	now := time.Date(2026, 2, 15, 12, 0, 0, 0, time.Local)
	memberID := "m1"

	deps := GetTrainingVolumeDeps{
		AttendanceStore: &mockTrainingVolumeAttendanceStore{},
		MemberStore:     &mockTrainingVolumeMemberStore{active: 1, trial: 0},
	}

	res, err := QueryGetTrainingVolume(context.Background(), GetTrainingVolumeQuery{MemberID: memberID, Range: "month", Compare: true, Now: now}, deps)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(res.Series) != 4 {
		t.Fatalf("series=%d, want 4 (you, avg, last month, same month last year)", len(res.Series))
	}
	if res.Series[2].Name != "Last month" {
		t.Fatalf("series[2].Name=%q, want Last month", res.Series[2].Name)
	}
	if res.Series[3].Name != "Same month last year" {
		t.Fatalf("series[3].Name=%q, want Same month last year", res.Series[3].Name)
	}
}

// TestQueryGetTrainingVolume_YearBuckets verifies year monthly bucketing.
func TestQueryGetTrainingVolume_YearBuckets(t *testing.T) {
	loc := time.Local
	now := time.Date(2026, 2, 15, 12, 0, 0, 0, loc)
	memberID := "m1"

	att := &mockTrainingVolumeAttendanceStore{
		member: []domainAttendance.Attendance{
			{ID: "a1", MemberID: memberID, CheckInTime: time.Date(2026, 1, 10, 8, 0, 0, 0, loc)},
			{ID: "a2", MemberID: memberID, CheckInTime: time.Date(2026, 2, 10, 8, 0, 0, 0, loc)},
			{ID: "a3", MemberID: memberID, CheckInTime: time.Date(2026, 2, 11, 8, 0, 0, 0, loc)},
			{ID: "a4", MemberID: memberID, CheckInTime: time.Date(2026, 12, 31, 8, 0, 0, 0, loc)},
		},
		all: []domainAttendance.Attendance{},
	}
	deps := GetTrainingVolumeDeps{AttendanceStore: att, MemberStore: &mockTrainingVolumeMemberStore{active: 1, trial: 0}}

	res, err := QueryGetTrainingVolume(context.Background(), GetTrainingVolumeQuery{MemberID: memberID, Range: "year", Now: now}, deps)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(res.Buckets) != 12 {
		t.Fatalf("buckets=%d, want 12", len(res.Buckets))
	}
	if res.Series[0].Values[0] != 1 {
		t.Fatalf("Jan=%v, want 1", res.Series[0].Values[0])
	}
	if res.Series[0].Values[1] != 2 {
		t.Fatalf("Feb=%v, want 2", res.Series[0].Values[1])
	}
	if res.Series[0].Values[11] != 1 {
		t.Fatalf("Dec=%v, want 1", res.Series[0].Values[11])
	}
}
