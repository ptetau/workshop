package orchestrators

import (
	"context"
	"testing"
	"time"

	domain "workshop/internal/domain/estimatedhours"
)

// mockEstHoursStoreForSelfEstimate implements EstimatedHoursStoreForSelfEstimate for testing.
type mockEstHoursStoreForSelfEstimate struct {
	saved   []domain.EstimatedHours
	entries map[string]domain.EstimatedHours
}

// Save implements EstimatedHoursStoreForSelfEstimate.
// PRE: e is a valid EstimatedHours
// POST: entry is appended to saved slice and stored in entries map
func (m *mockEstHoursStoreForSelfEstimate) Save(_ context.Context, e domain.EstimatedHours) error {
	m.saved = append(m.saved, e)
	if m.entries == nil {
		m.entries = make(map[string]domain.EstimatedHours)
	}
	m.entries[e.ID] = e
	return nil
}

// GetByID implements EstimatedHoursStoreForSelfEstimate.
// PRE: id is non-empty
// POST: returns the entry or error if not found
func (m *mockEstHoursStoreForSelfEstimate) GetByID(_ context.Context, id string) (domain.EstimatedHours, error) {
	e, ok := m.entries[id]
	if !ok {
		return domain.EstimatedHours{}, context.Canceled // simulate not found
	}
	return e, nil
}

var fixedTimeSelfEst = time.Date(2026, 2, 10, 12, 0, 0, 0, time.UTC)

func testNowSelfEst() time.Time { return fixedTimeSelfEst }

// TestExecuteSubmitSelfEstimate_HappyPath tests a member submitting a self-estimate.
func TestExecuteSubmitSelfEstimate_HappyPath(t *testing.T) {
	store := &mockEstHoursStoreForSelfEstimate{entries: make(map[string]domain.EstimatedHours)}

	input := SubmitSelfEstimateInput{
		MemberID:    "m1",
		StartDate:   "2026-01-01",
		EndDate:     "2026-02-01",
		WeeklyHours: 3,
		Note:        "Trained at Checkmat SP while travelling",
	}
	deps := SubmitSelfEstimateDeps{
		EstimatedHoursStore: store,
		GenerateID:          func() string { return "se-1" },
		Now:                 testNowSelfEst,
	}

	result, err := ExecuteSubmitSelfEstimate(context.Background(), input, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Source != domain.SourceSelfEstimate {
		t.Errorf("source = %q, want %q", result.Source, domain.SourceSelfEstimate)
	}
	if result.Status != domain.StatusPending {
		t.Errorf("status = %q, want %q", result.Status, domain.StatusPending)
	}
	if result.CreatedBy != "m1" {
		t.Errorf("created_by = %q, want %q", result.CreatedBy, "m1")
	}
	if result.TotalHours == 0 {
		t.Error("expected non-zero TotalHours")
	}
	if len(store.saved) != 1 {
		t.Errorf("expected 1 saved entry, got %d", len(store.saved))
	}
}

// TestExecuteSubmitSelfEstimate_InvalidInput tests validation failure.
func TestExecuteSubmitSelfEstimate_InvalidInput(t *testing.T) {
	store := &mockEstHoursStoreForSelfEstimate{entries: make(map[string]domain.EstimatedHours)}

	input := SubmitSelfEstimateInput{
		MemberID:    "", // invalid
		StartDate:   "2026-01-01",
		EndDate:     "2026-02-01",
		WeeklyHours: 3,
	}
	deps := SubmitSelfEstimateDeps{
		EstimatedHoursStore: store,
		GenerateID:          func() string { return "se-2" },
		Now:                 testNowSelfEst,
	}

	_, err := ExecuteSubmitSelfEstimate(context.Background(), input, deps)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if len(store.saved) != 0 {
		t.Errorf("expected 0 saved entries, got %d", len(store.saved))
	}
}

// TestExecuteReviewSelfEstimate_Approve tests approving a pending self-estimate.
func TestExecuteReviewSelfEstimate_Approve(t *testing.T) {
	store := &mockEstHoursStoreForSelfEstimate{
		entries: map[string]domain.EstimatedHours{
			"se-1": {
				ID: "se-1", MemberID: "m1", StartDate: "2026-01-01", EndDate: "2026-02-01",
				WeeklyHours: 3, TotalHours: 15, Source: domain.SourceSelfEstimate,
				Status: domain.StatusPending, CreatedBy: "m1",
			},
		},
	}

	input := ReviewSelfEstimateInput{
		ID:         "se-1",
		Action:     "approve",
		ReviewerID: "admin-1",
	}
	deps := ReviewSelfEstimateDeps{
		EstimatedHoursStore: store,
		Now:                 testNowSelfEst,
	}

	result, err := ExecuteReviewSelfEstimate(context.Background(), input, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != domain.StatusApproved {
		t.Errorf("status = %q, want %q", result.Status, domain.StatusApproved)
	}
	if result.ReviewedBy != "admin-1" {
		t.Errorf("reviewed_by = %q, want %q", result.ReviewedBy, "admin-1")
	}
	if result.TotalHours != 15 {
		t.Errorf("total_hours = %v, want 15 (unadjusted)", result.TotalHours)
	}
}

// TestExecuteReviewSelfEstimate_ApproveWithAdjustment tests adjusting hours on approve.
func TestExecuteReviewSelfEstimate_ApproveWithAdjustment(t *testing.T) {
	store := &mockEstHoursStoreForSelfEstimate{
		entries: map[string]domain.EstimatedHours{
			"se-2": {
				ID: "se-2", MemberID: "m1", StartDate: "2026-01-01", EndDate: "2026-02-01",
				WeeklyHours: 3, TotalHours: 18, Source: domain.SourceSelfEstimate,
				Status: domain.StatusPending, CreatedBy: "m1",
			},
		},
	}

	input := ReviewSelfEstimateInput{
		ID:            "se-2",
		Action:        "approve",
		AdjustedHours: 12,
		ReviewNote:    "Reduced to 12h based on partner gym schedule",
		ReviewerID:    "admin-1",
	}
	deps := ReviewSelfEstimateDeps{
		EstimatedHoursStore: store,
		Now:                 testNowSelfEst,
	}

	result, err := ExecuteReviewSelfEstimate(context.Background(), input, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalHours != 12 {
		t.Errorf("total_hours = %v, want 12 (adjusted)", result.TotalHours)
	}
	if result.ReviewNote != "Reduced to 12h based on partner gym schedule" {
		t.Errorf("review_note = %q", result.ReviewNote)
	}
}

// TestExecuteReviewSelfEstimate_Reject tests rejecting a pending self-estimate.
func TestExecuteReviewSelfEstimate_Reject(t *testing.T) {
	store := &mockEstHoursStoreForSelfEstimate{
		entries: map[string]domain.EstimatedHours{
			"se-3": {
				ID: "se-3", MemberID: "m1", StartDate: "2026-01-01", EndDate: "2026-02-01",
				WeeklyHours: 3, TotalHours: 15, Source: domain.SourceSelfEstimate,
				Status: domain.StatusPending, CreatedBy: "m1",
			},
		},
	}

	input := ReviewSelfEstimateInput{
		ID:         "se-3",
		Action:     "reject",
		ReviewNote: "Unable to verify training at that gym",
		ReviewerID: "admin-1",
	}
	deps := ReviewSelfEstimateDeps{
		EstimatedHoursStore: store,
		Now:                 testNowSelfEst,
	}

	result, err := ExecuteReviewSelfEstimate(context.Background(), input, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != domain.StatusRejected {
		t.Errorf("status = %q, want %q", result.Status, domain.StatusRejected)
	}
	if result.ReviewNote != "Unable to verify training at that gym" {
		t.Errorf("review_note = %q", result.ReviewNote)
	}
}

// TestExecuteReviewSelfEstimate_NotPending tests reviewing a non-pending entry fails.
func TestExecuteReviewSelfEstimate_NotPending(t *testing.T) {
	store := &mockEstHoursStoreForSelfEstimate{
		entries: map[string]domain.EstimatedHours{
			"se-4": {
				ID: "se-4", MemberID: "m1", StartDate: "2026-01-01", EndDate: "2026-02-01",
				WeeklyHours: 3, TotalHours: 15, Source: domain.SourceSelfEstimate,
				Status: domain.StatusApproved, CreatedBy: "m1",
			},
		},
	}

	input := ReviewSelfEstimateInput{
		ID:         "se-4",
		Action:     "approve",
		ReviewerID: "admin-1",
	}
	deps := ReviewSelfEstimateDeps{
		EstimatedHoursStore: store,
		Now:                 testNowSelfEst,
	}

	_, err := ExecuteReviewSelfEstimate(context.Background(), input, deps)
	if err == nil {
		t.Fatal("expected error for reviewing non-pending entry")
	}
}
