package calendar

import (
	"errors"
	"time"
)

// CompetitionInterest tracks which members are interested in attending a competition.
type CompetitionInterest struct {
	ID        string
	EventID   string
	MemberID  string
	CreatedAt time.Time
}

// Validate checks the interest record's invariants.
// PRE: ci fields may be empty (validation will catch this).
// POST: Returns nil if valid, error with descriptive message otherwise.
// INVARIANT: EventID and MemberID must be non-empty.
func (ci *CompetitionInterest) Validate() error {
	if ci.EventID == "" {
		return errors.New("event_id is required")
	}
	if ci.MemberID == "" {
		return errors.New("member_id is required")
	}
	return nil
}
