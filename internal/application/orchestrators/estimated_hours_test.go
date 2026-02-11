package orchestrators

import (
	"context"
	"testing"
	"time"

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
