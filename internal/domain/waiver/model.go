package waiver

import (
	"errors"
	"time"
)

// Waiver holds state for the concept.
type Waiver struct {
	ID            string
	AcceptedTerms bool
	IPAddress     string
	MemberID      string
	SignedAt      time.Time
}

// Validate checks if the Waiver has valid data.
// PRE: Waiver struct is initialized
// POST: Returns error if validation fails, nil otherwise
// INVARIANT: AcceptedTerms must be true, MemberID must not be empty
func (w *Waiver) Validate() error {
	if w.MemberID == "" {
		return errors.New("waiver must be associated with a member")
	}
	if !w.AcceptedTerms {
		return errors.New("terms must be accepted")
	}
	if w.SignedAt.IsZero() {
		return errors.New("signed date must be set")
	}
	return nil
}

// IsValid returns true if the waiver is still valid (not expired).
// PRE: Waiver is initialized
// POST: Returns boolean indicating validity
// INVARIANT: Waivers expire after 1 year
func (w *Waiver) IsValid() bool {
	return time.Since(w.SignedAt) < 365*24*time.Hour
}

// HasExpired returns true if the waiver has expired.
// PRE: Waiver is initialized
// POST: Returns boolean indicating expiry status
func (w *Waiver) HasExpired() bool {
	return !w.IsValid()
}
