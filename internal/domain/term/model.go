package term

import (
	"errors"
	"strings"
	"time"
)

// Max length constants for user-editable fields.
const (
	MaxNameLength = 100
)

// Domain errors
var (
	ErrEmptyName      = errors.New("term name cannot be empty")
	ErrInvalidDates   = errors.New("start date must be before end date")
	ErrEmptyStartDate = errors.New("start date cannot be zero")
	ErrEmptyEndDate   = errors.New("end date cannot be zero")
)

// Term represents a school term (NZ school terms).
type Term struct {
	ID        string
	Name      string
	StartDate time.Time
	EndDate   time.Time
}

// Validate checks if the Term has valid data.
// PRE: Term struct is populated
// POST: Returns nil if valid, error otherwise
func (t *Term) Validate() error {
	if strings.TrimSpace(t.Name) == "" {
		return ErrEmptyName
	}
	if len(t.Name) > MaxNameLength {
		return errors.New("term name cannot exceed 100 characters")
	}
	if t.StartDate.IsZero() {
		return ErrEmptyStartDate
	}
	if t.EndDate.IsZero() {
		return ErrEmptyEndDate
	}
	if !t.StartDate.Before(t.EndDate) {
		return ErrInvalidDates
	}
	return nil
}

// Contains returns true if the given date falls within this term.
// PRE: date is a valid time
// INVARIANT: Term fields are not mutated
func (t *Term) Contains(date time.Time) bool {
	d := date.Truncate(24 * time.Hour)
	start := t.StartDate.Truncate(24 * time.Hour)
	end := t.EndDate.Truncate(24 * time.Hour)
	return (d.Equal(start) || d.After(start)) && (d.Equal(end) || d.Before(end))
}
