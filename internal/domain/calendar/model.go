package calendar

import (
	"errors"
	"time"
)

// Event type constants.
const (
	TypeEvent       = "event"       // club event (seminar, social, grading day, gym closure)
	TypeCompetition = "competition" // competition with optional registration URL
)

// Max length constants.
const (
	MaxTitleLength           = 200
	MaxDescriptionLength     = 2000
	MaxLocationLength        = 200
	MaxRegistrationURLLength = 2048
)

// Event represents a club calendar event or competition.
// PRE: Title is non-empty. StartDate is set. Type is "event" or "competition".
// INVARIANT: EndDate >= StartDate when EndDate is set.
type Event struct {
	ID              string
	Title           string
	Type            string // "event" or "competition"
	Description     string
	Location        string
	StartDate       time.Time
	EndDate         time.Time // zero value means single-day event
	RegistrationURL string    // only for competitions
	CreatedBy       string    // account ID
	CreatedAt       time.Time
}

// Validate checks the event's invariants.
// PRE: none
// POST: returns nil if valid, error describing the first violation otherwise
func (e *Event) Validate() error {
	if e.Title == "" {
		return errors.New("event title cannot be empty")
	}
	if len(e.Title) > MaxTitleLength {
		return errors.New("event title cannot exceed 200 characters")
	}
	if e.Type != TypeEvent && e.Type != TypeCompetition {
		return errors.New("event type must be 'event' or 'competition'")
	}
	if e.StartDate.IsZero() {
		return errors.New("event start date is required")
	}
	if !e.EndDate.IsZero() && e.EndDate.Before(e.StartDate) {
		return errors.New("event end date cannot be before start date")
	}
	if len(e.Description) > MaxDescriptionLength {
		return errors.New("event description cannot exceed 2000 characters")
	}
	if len(e.Location) > MaxLocationLength {
		return errors.New("event location cannot exceed 200 characters")
	}
	if len(e.RegistrationURL) > MaxRegistrationURLLength {
		return errors.New("registration URL cannot exceed 2048 characters")
	}
	return nil
}

// IsMultiDay returns true if the event spans more than one day.
// PRE: none
// POST: returns true if EndDate is set and on a different calendar day than StartDate
func (e *Event) IsMultiDay() bool {
	if e.EndDate.IsZero() {
		return false
	}
	return e.EndDate.After(e.StartDate) &&
		e.EndDate.Format("2006-01-02") != e.StartDate.Format("2006-01-02")
}
