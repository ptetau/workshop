package observation

import (
	"errors"
	"time"
)

// Max length constants for user-editable fields.
const (
	MaxContentLength = 1000
)

// Domain errors
var (
	ErrEmptyMemberID = errors.New("member ID is required")
	ErrEmptyAuthorID = errors.New("author ID is required")
	ErrEmptyContent  = errors.New("observation content cannot be empty")
)

// Observation represents a private per-student note written by a Coach or Admin.
// Not visible to the student. Used for technique feedback, grading observations,
// and behavioural notes.
type Observation struct {
	ID        string
	MemberID  string // The student being observed
	AuthorID  string // Coach or Admin AccountID
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate checks if the Observation has valid data.
// PRE: Observation struct is populated
// POST: Returns nil if valid, error otherwise
func (o *Observation) Validate() error {
	if o.MemberID == "" {
		return ErrEmptyMemberID
	}
	if o.AuthorID == "" {
		return ErrEmptyAuthorID
	}
	if o.Content == "" {
		return ErrEmptyContent
	}
	if len(o.Content) > MaxContentLength {
		return errors.New("observation content cannot exceed 1000 characters")
	}
	if o.CreatedAt.IsZero() {
		return errors.New("created_at must be set")
	}
	return nil
}
