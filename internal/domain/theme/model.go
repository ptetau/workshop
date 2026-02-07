package theme

import (
	"errors"
	"time"
)

// Theme represents a 4-week technical block for the gym's curriculum.
// PRE: Name and Program are non-empty.
// INVARIANT: StartDate is before EndDate.
type Theme struct {
	ID          string
	Name        string    // e.g. "Leg Lasso Series"
	Description string    // brief summary of the technical focus
	Program     string    // "adults" or "kids"
	StartDate   time.Time // first day of the 4-week block
	EndDate     time.Time // last day of the 4-week block
	CreatedBy   string    // account ID of the creator
	CreatedAt   time.Time
}

// StatusActive indicates the theme is currently running.
const StatusActive = "active"

// StatusUpcoming indicates the theme has not started yet.
const StatusUpcoming = "upcoming"

// StatusCompleted indicates the theme block has ended.
const StatusCompleted = "completed"

// Validate checks the theme's invariants.
// PRE: none
// POST: returns nil if valid, error describing the first violation otherwise
func (t *Theme) Validate() error {
	if t.Name == "" {
		return errors.New("theme name cannot be empty")
	}
	if t.Program == "" {
		return errors.New("theme program cannot be empty")
	}
	if t.Program != "adults" && t.Program != "kids" {
		return errors.New("theme program must be 'adults' or 'kids'")
	}
	if t.StartDate.IsZero() {
		return errors.New("theme start date cannot be empty")
	}
	if t.EndDate.IsZero() {
		return errors.New("theme end date cannot be empty")
	}
	if t.StartDate.After(t.EndDate) {
		return errors.New("theme start date cannot be after end date")
	}
	return nil
}

// Status returns the theme's current status relative to the given time.
// PRE: now is a valid time
// POST: returns StatusActive, StatusUpcoming, or StatusCompleted
func (t *Theme) Status(now time.Time) string {
	if now.Before(t.StartDate) {
		return StatusUpcoming
	}
	if now.After(t.EndDate) {
		return StatusCompleted
	}
	return StatusActive
}

// WeekNumber returns which week of the block the given time falls in (1-4).
// PRE: now is within the theme's date range
// POST: returns 1-4, or 0 if outside the range
func (t *Theme) WeekNumber(now time.Time) int {
	if now.Before(t.StartDate) || now.After(t.EndDate) {
		return 0
	}
	days := int(now.Sub(t.StartDate).Hours() / 24)
	week := (days / 7) + 1
	if week > 4 {
		week = 4
	}
	return week
}
