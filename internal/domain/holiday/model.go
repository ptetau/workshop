package holiday

import (
	"errors"
	"strings"
	"time"
)

// Domain errors
var (
	ErrEmptyName      = errors.New("holiday name cannot be empty")
	ErrEmptyDate      = errors.New("holiday date cannot be zero")
	ErrInvalidDates   = errors.New("start date must be before or equal to end date")
	ErrEmptyStartDate = errors.New("start date cannot be zero")
	ErrEmptyEndDate   = errors.New("end date cannot be zero")
)

// Holiday represents a day (or range) when classes are not held.
type Holiday struct {
	ID        string
	Name      string
	StartDate time.Time
	EndDate   time.Time
}

// Validate checks if the Holiday has valid data.
// PRE: Holiday struct is populated
// POST: Returns nil if valid, error otherwise
func (h *Holiday) Validate() error {
	if strings.TrimSpace(h.Name) == "" {
		return ErrEmptyName
	}
	if h.StartDate.IsZero() {
		return ErrEmptyStartDate
	}
	if h.EndDate.IsZero() {
		return ErrEmptyEndDate
	}
	if h.StartDate.After(h.EndDate) {
		return ErrInvalidDates
	}
	return nil
}

// Contains returns true if the given date falls within this holiday.
// PRE: date is a valid time
// INVARIANT: Holiday fields are not mutated
func (h *Holiday) Contains(date time.Time) bool {
	d := date.Truncate(24 * time.Hour)
	start := h.StartDate.Truncate(24 * time.Hour)
	end := h.EndDate.Truncate(24 * time.Hour)
	return (d.Equal(start) || d.After(start)) && (d.Equal(end) || d.Before(end))
}
