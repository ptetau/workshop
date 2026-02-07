package kiosk

import (
	"errors"
	"time"
)

// Domain errors
var (
	ErrEmptyAccountID = errors.New("kiosk must be tied to a launching account")
	ErrNotActive      = errors.New("kiosk session is not active")
	ErrAlreadyActive  = errors.New("kiosk session is already active")
)

// Session represents an active kiosk mode session.
// Kiosk mode locks the tablet to check-in only; exiting requires the
// launching account's password or another coach/admin login.
type Session struct {
	ID        string
	AccountID string // The account that launched kiosk mode
	StartedAt time.Time
	EndedAt   time.Time
}

// Validate checks if the Session has valid data.
// PRE: Session struct is populated
// POST: Returns nil if valid, error otherwise
func (s *Session) Validate() error {
	if s.AccountID == "" {
		return ErrEmptyAccountID
	}
	if s.StartedAt.IsZero() {
		return errors.New("started_at cannot be zero")
	}
	return nil
}

// IsActive returns true if the kiosk session is currently active.
// INVARIANT: Session fields are not mutated
func (s *Session) IsActive() bool {
	return s.EndedAt.IsZero()
}

// End terminates the kiosk session.
// PRE: Session is currently active
// POST: EndedAt is set to current time
func (s *Session) End() error {
	if !s.IsActive() {
		return ErrNotActive
	}
	s.EndedAt = time.Now()
	return nil
}
