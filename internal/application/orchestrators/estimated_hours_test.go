package orchestrators

import (
	"context"
	"testing"
	"time"

	attendanceDomain "workshop/internal/domain/attendance"
	domain "workshop/internal/domain/estimatedhours"
)

// mockEstimatedHoursStore implements EstimatedHoursStoreForBulkAdd for testing.
type mockEstimatedHoursStore struct {
	saved []domain.EstimatedHours
}

// Save implements EstimatedHoursStoreForBulkAdd.
// PRE: e is a valid EstimatedHours
// POST: entry is appended to saved slice
func (m *mockEstimatedHoursStore) Save(_ context.Context, e domain.EstimatedHours) error {
	m.saved = append(m.saved, e)
	return nil
}

// mockAttendanceStoreForOverlap implements AttendanceStoreForOverlap for testing.
type mockAttendanceStoreForOverlap struct {
	records []attendanceDomain.Attendance
	deleted int
}

// ListByMemberIDAndDateRange implements AttendanceStoreForOverlap.
// PRE: memberID, startDate, endDate are non-empty
// POST: Returns records within the date range
func (m *mockAttendanceStoreForOverlap) ListByMemberIDAndDateRange(_ context.Context, memberID, startDate, endDate string) ([]attendanceDomain.Attendance, error) {
	var result []attendanceDomain.Attendance
	for _, r := range m.records {
		d := r.CheckInTime.Format("2006-01-02")
		if r.MemberID == memberID && d >= startDate && d <= endDate {
			result = append(result, r)
		}
	}
	return result, nil
}

// DeleteByMemberIDAndDateRange implements AttendanceStoreForOverlap.
// PRE: memberID, startDate, endDate are non-empty
// POST: Returns count of deleted records
func (m *mockAttendanceStoreForOverlap) DeleteByMemberIDAndDateRange(_ context.Context, memberID, startDate, endDate string) (int, error) {
	count := 0
	var remaining []attendanceDomain.Attendance
	for _, r := range m.records {
		d := r.CheckInTime.Format("2006-01-02")
		if r.MemberID == memberID && d >= startDate && d <= endDate {
			count++
		} else {
			remaining = append(remaining, r)
		}
	}
	m.records = remaining
	m.deleted += count
	return count, nil
}

var fixedTimeEstHours = time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)

func testNowEstHours() time.Time { return fixedTimeEstHours }

// TestExecuteBulkAddEstimatedHours tests the happy path: 13 weeks × 3h = 39h.
func TestExecuteBulkAddEstimatedHours(t *testing.T) {
	store := &mockEstimatedHoursStore{}

	input := BulkAddEstimatedHoursInput{
		MemberID:    "m1",
		StartDate:   "2026-01-01",
		EndDate:     "2026-03-31",
		WeeklyHours: 3,
		Note:        "trained at another gym",
		CreatedBy:   "coach-1",
	}
	deps := BulkAddEstimatedHoursDeps{
		EstimatedHoursStore: store,
		GenerateID:          func() string { return "est-1" },
		Now:                 testNowEstHours,
	}

	result, err := ExecuteBulkAddEstimatedHours(context.Background(), input, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalHours != 39 {
		t.Errorf("total hours = %v, want 39", result.TotalHours)
	}
	if result.Source != domain.SourceEstimate {
		t.Errorf("source = %q, want %q", result.Source, domain.SourceEstimate)
	}
	if result.Status != domain.StatusApproved {
		t.Errorf("status = %q, want %q", result.Status, domain.StatusApproved)
	}
	if result.Note != "trained at another gym" {
		t.Errorf("note = %q, want %q", result.Note, "trained at another gym")
	}
	if len(store.saved) != 1 {
		t.Errorf("expected 1 saved entry, got %d", len(store.saved))
	}
}

// TestCheckEstimatedHoursOverlap_HasOverlap tests overlap detection with existing attendance.
func TestCheckEstimatedHoursOverlap_HasOverlap(t *testing.T) {
	attStore := &mockAttendanceStoreForOverlap{
		records: []attendanceDomain.Attendance{
			{ID: "a1", MemberID: "m1", CheckInTime: time.Date(2026, 2, 5, 18, 0, 0, 0, time.UTC), MatHours: 1.5},
			{ID: "a2", MemberID: "m1", CheckInTime: time.Date(2026, 2, 12, 18, 0, 0, 0, time.UTC), MatHours: 2.0},
		},
	}

	result, err := CheckEstimatedHoursOverlap(context.Background(), "m1", "2026-02-01", "2026-04-30", attStore)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasOverlap {
		t.Fatal("expected HasOverlap=true")
	}
	if result.OverlapCount != 2 {
		t.Errorf("OverlapCount = %d, want 2", result.OverlapCount)
	}
	if result.OverlapHours != 3.5 {
		t.Errorf("OverlapHours = %v, want 3.5", result.OverlapHours)
	}
}

// TestCheckEstimatedHoursOverlap_NoOverlap tests no overlap found.
func TestCheckEstimatedHoursOverlap_NoOverlap(t *testing.T) {
	attStore := &mockAttendanceStoreForOverlap{records: nil}
	result, err := CheckEstimatedHoursOverlap(context.Background(), "m1", "2026-02-01", "2026-04-30", attStore)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasOverlap {
		t.Fatal("expected HasOverlap=false")
	}
}

// TestExecuteBulkAddEstimatedHours_ReplaceMode tests that overlapping attendance is deleted.
func TestExecuteBulkAddEstimatedHours_ReplaceMode(t *testing.T) {
	estStore := &mockEstimatedHoursStore{}
	attStore := &mockAttendanceStoreForOverlap{
		records: []attendanceDomain.Attendance{
			{ID: "a1", MemberID: "m1", CheckInTime: time.Date(2026, 2, 5, 18, 0, 0, 0, time.UTC), MatHours: 1.5},
		},
	}

	input := BulkAddEstimatedHoursInput{
		MemberID:    "m1",
		StartDate:   "2026-02-01",
		EndDate:     "2026-04-30",
		WeeklyHours: 3,
		CreatedBy:   "coach-1",
		OverlapMode: OverlapModeReplace,
	}
	deps := BulkAddEstimatedHoursDeps{
		EstimatedHoursStore: estStore,
		AttendanceStore:     attStore,
		GenerateID:          func() string { return "est-r1" },
		Now:                 testNowEstHours,
	}

	result, err := ExecuteBulkAddEstimatedHours(context.Background(), input, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(estStore.saved) != 1 {
		t.Errorf("expected 1 saved entry, got %d", len(estStore.saved))
	}
	if attStore.deleted != 1 {
		t.Errorf("expected 1 deleted attendance record, got %d", attStore.deleted)
	}
	if len(attStore.records) != 0 {
		t.Errorf("expected 0 remaining attendance records, got %d", len(attStore.records))
	}
	if result.TotalHours == 0 {
		t.Error("expected non-zero TotalHours")
	}
}

// TestExecuteBulkAddEstimatedHours_AddMode tests that overlapping attendance is kept and estimate is saved on top.
func TestExecuteBulkAddEstimatedHours_AddMode(t *testing.T) {
	estStore := &mockEstimatedHoursStore{}
	attStore := &mockAttendanceStoreForOverlap{
		records: []attendanceDomain.Attendance{
			{ID: "a1", MemberID: "m1", CheckInTime: time.Date(2026, 2, 5, 18, 0, 0, 0, time.UTC), MatHours: 1.5},
			{ID: "a2", MemberID: "m1", CheckInTime: time.Date(2026, 2, 12, 18, 0, 0, 0, time.UTC), MatHours: 2.0},
		},
	}

	input := BulkAddEstimatedHoursInput{
		MemberID:    "m1",
		StartDate:   "2026-02-01",
		EndDate:     "2026-04-30",
		WeeklyHours: 3,
		Note:        "extra training at partner gym",
		CreatedBy:   "coach-1",
		OverlapMode: OverlapModeAdd,
	}
	deps := BulkAddEstimatedHoursDeps{
		EstimatedHoursStore: estStore,
		AttendanceStore:     attStore,
		GenerateID:          func() string { return "est-add1" },
		Now:                 testNowEstHours,
	}

	result, err := ExecuteBulkAddEstimatedHours(context.Background(), input, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(estStore.saved) != 1 {
		t.Errorf("expected 1 saved entry, got %d", len(estStore.saved))
	}
	if attStore.deleted != 0 {
		t.Errorf("expected 0 deleted attendance records (add mode), got %d", attStore.deleted)
	}
	if len(attStore.records) != 2 {
		t.Errorf("expected 2 remaining attendance records, got %d", len(attStore.records))
	}
	if result.TotalHours == 0 {
		t.Error("expected non-zero TotalHours")
	}
	if result.Note != "extra training at partner gym" {
		t.Errorf("note = %q, want %q", result.Note, "extra training at partner gym")
	}
}

// TestExecuteBulkAddEstimatedHours_InvalidInput tests validation failure.
func TestExecuteBulkAddEstimatedHours_InvalidInput(t *testing.T) {
	store := &mockEstimatedHoursStore{}

	input := BulkAddEstimatedHoursInput{
		MemberID:    "", // invalid
		StartDate:   "2026-01-01",
		EndDate:     "2026-03-31",
		WeeklyHours: 3,
		CreatedBy:   "coach-1",
	}
	deps := BulkAddEstimatedHoursDeps{
		EstimatedHoursStore: store,
		GenerateID:          func() string { return "est-2" },
		Now:                 testNowEstHours,
	}

	_, err := ExecuteBulkAddEstimatedHours(context.Background(), input, deps)
	if err == nil {
		t.Fatal("expected validation error for empty member ID")
	}
	if len(store.saved) != 0 {
		t.Errorf("expected 0 saved entries on validation failure, got %d", len(store.saved))
	}
}
