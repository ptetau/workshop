package estimatedhours

import (
	"testing"
	"time"
)

// TestEstimatedHours_Validate_Valid tests that a valid entry passes validation.
func TestEstimatedHours_Validate_Valid(t *testing.T) {
	e := EstimatedHours{
		MemberID:    "m1",
		StartDate:   "2026-01-01",
		EndDate:     "2026-03-31",
		WeeklyHours: 3,
		Source:      SourceEstimate,
		Status:      StatusApproved,
		CreatedBy:   "coach-1",
		CreatedAt:   time.Now(),
	}
	if err := e.Validate(); err != nil {
		t.Errorf("expected valid, got error: %v", err)
	}
}

// TestEstimatedHours_Validate_EmptyMember tests that empty member ID fails.
func TestEstimatedHours_Validate_EmptyMember(t *testing.T) {
	e := EstimatedHours{
		MemberID:    "",
		StartDate:   "2026-01-01",
		EndDate:     "2026-03-31",
		WeeklyHours: 3,
		Source:      SourceEstimate,
		Status:      StatusApproved,
		CreatedBy:   "coach-1",
	}
	if err := e.Validate(); err != ErrEmptyMemberID {
		t.Errorf("expected ErrEmptyMemberID, got %v", err)
	}
}

// TestEstimatedHours_Validate_InvalidDateRange tests that start after end fails.
func TestEstimatedHours_Validate_InvalidDateRange(t *testing.T) {
	e := EstimatedHours{
		MemberID:    "m1",
		StartDate:   "2026-03-31",
		EndDate:     "2026-01-01",
		WeeklyHours: 3,
		Source:      SourceEstimate,
		Status:      StatusApproved,
		CreatedBy:   "coach-1",
	}
	if err := e.Validate(); err != ErrStartAfterEnd {
		t.Errorf("expected ErrStartAfterEnd, got %v", err)
	}
}

// TestEstimatedHours_Validate_ZeroHours tests that zero weekly hours fails.
func TestEstimatedHours_Validate_ZeroHours(t *testing.T) {
	e := EstimatedHours{
		MemberID:    "m1",
		StartDate:   "2026-01-01",
		EndDate:     "2026-03-31",
		WeeklyHours: 0,
		Source:      SourceEstimate,
		Status:      StatusApproved,
		CreatedBy:   "coach-1",
	}
	if err := e.Validate(); err != ErrInvalidWeekly {
		t.Errorf("expected ErrInvalidWeekly, got %v", err)
	}
}

// TestEstimatedHours_Validate_ExcessiveHours tests that >40 weekly hours fails.
func TestEstimatedHours_Validate_ExcessiveHours(t *testing.T) {
	e := EstimatedHours{
		MemberID:    "m1",
		StartDate:   "2026-01-01",
		EndDate:     "2026-03-31",
		WeeklyHours: 41,
		Source:      SourceEstimate,
		Status:      StatusApproved,
		CreatedBy:   "coach-1",
	}
	if err := e.Validate(); err != ErrWeeklyHoursTooHigh {
		t.Errorf("expected ErrWeeklyHoursTooHigh, got %v", err)
	}
}

// TestEstimatedHours_CalculateTotalHours tests the total hours calculation.
func TestEstimatedHours_CalculateTotalHours(t *testing.T) {
	e := EstimatedHours{
		StartDate:   "2026-01-01",
		EndDate:     "2026-03-31",
		WeeklyHours: 3,
	}
	if err := e.CalculateTotalHours(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Jan 1 to Mar 31 = 89 days = ceil(89/7) = 13 weeks → 13 × 3 = 39
	if e.TotalHours != 39 {
		t.Errorf("total hours = %v, want 39", e.TotalHours)
	}
}

// TestEstimatedHours_CalculateTotalHours_SingleDay tests minimum 1 week.
func TestEstimatedHours_CalculateTotalHours_SingleDay(t *testing.T) {
	e := EstimatedHours{
		StartDate:   "2026-01-01",
		EndDate:     "2026-01-01",
		WeeklyHours: 5,
	}
	if err := e.CalculateTotalHours(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Same day = 0 days, ceil(0/7) would be 0 but minimum is 1 week → 5
	if e.TotalHours != 5 {
		t.Errorf("total hours = %v, want 5", e.TotalHours)
	}
}

// TestEstimatedHours_Validate_InvalidSource tests that invalid source fails.
func TestEstimatedHours_Validate_InvalidSource(t *testing.T) {
	e := EstimatedHours{
		MemberID:    "m1",
		StartDate:   "2026-01-01",
		EndDate:     "2026-03-31",
		WeeklyHours: 3,
		Source:      "invalid",
		Status:      StatusApproved,
		CreatedBy:   "coach-1",
	}
	if err := e.Validate(); err != ErrInvalidSource {
		t.Errorf("expected ErrInvalidSource, got %v", err)
	}
}
