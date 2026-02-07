package theme

import (
	"testing"
	"time"
)

// TestTheme_Validate_Valid tests that a valid theme passes validation.
func TestTheme_Validate_Valid(t *testing.T) {
	th := Theme{
		Name:      "Leg Lasso Series",
		Program:   "adults",
		StartDate: time.Now(),
		EndDate:   time.Now().AddDate(0, 0, 28),
	}
	if err := th.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestTheme_Validate_EmptyName tests that an empty name fails validation.
func TestTheme_Validate_EmptyName(t *testing.T) {
	th := Theme{Program: "adults", StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 28)}
	if err := th.Validate(); err == nil {
		t.Error("expected error for empty name")
	}
}

// TestTheme_Validate_InvalidProgram tests that an invalid program fails validation.
func TestTheme_Validate_InvalidProgram(t *testing.T) {
	th := Theme{Name: "Test", Program: "seniors", StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 28)}
	if err := th.Validate(); err == nil {
		t.Error("expected error for invalid program")
	}
}

// TestTheme_Validate_StartAfterEnd tests that start after end fails validation.
func TestTheme_Validate_StartAfterEnd(t *testing.T) {
	th := Theme{Name: "Test", Program: "adults", StartDate: time.Now().AddDate(0, 0, 28), EndDate: time.Now()}
	if err := th.Validate(); err == nil {
		t.Error("expected error for start after end")
	}
}

// TestTheme_Status tests status calculation relative to now.
func TestTheme_Status(t *testing.T) {
	now := time.Date(2026, 2, 15, 12, 0, 0, 0, time.UTC)
	th := Theme{
		StartDate: time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 3, 9, 23, 59, 59, 0, time.UTC),
	}

	if s := th.Status(now); s != StatusActive {
		t.Errorf("got %q, want %q", s, StatusActive)
	}
	if s := th.Status(time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)); s != StatusUpcoming {
		t.Errorf("got %q, want %q", s, StatusUpcoming)
	}
	if s := th.Status(time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)); s != StatusCompleted {
		t.Errorf("got %q, want %q", s, StatusCompleted)
	}
}

// TestTheme_WeekNumber tests week calculation within a theme block.
func TestTheme_WeekNumber(t *testing.T) {
	th := Theme{
		StartDate: time.Date(2026, 2, 3, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 3, 2, 23, 59, 59, 0, time.UTC),
	}

	if w := th.WeekNumber(time.Date(2026, 2, 5, 12, 0, 0, 0, time.UTC)); w != 1 {
		t.Errorf("got week %d, want 1", w)
	}
	if w := th.WeekNumber(time.Date(2026, 2, 12, 12, 0, 0, 0, time.UTC)); w != 2 {
		t.Errorf("got week %d, want 2", w)
	}
	if w := th.WeekNumber(time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)); w != 0 {
		t.Errorf("got week %d, want 0 (before start)", w)
	}
}
