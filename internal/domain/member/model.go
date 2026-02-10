package member

import (
	"errors"
	"strings"
)

// Max length constants for user-editable fields.
const (
	MaxNameLength = 100
)

// Business rule constants
const (
	StatusActive   = "active"
	StatusInactive = "inactive"
	StatusArchived = "archived"
	ProgramAdults  = "adults"
	ProgramKids    = "kids"
)

// Domain errors
var (
	ErrAlreadyArchived = errors.New("member is already archived")
	ErrNotArchived     = errors.New("member is not archived")
	ErrAlreadyActive   = errors.New("member is already active")
)

// Member holds state for the concept.
type Member struct {
	ID        string
	AccountID string
	Email     string
	Fee       int
	Frequency string
	Name      string
	Program   string
	Status    string
}

// Validate checks if the Member has valid data.
// PRE: Member struct is initialized
// POST: Returns error if validation fails, nil otherwise
// INVARIANT: Email must contain '@', Name must not be empty
func (m *Member) Validate() error {
	if strings.TrimSpace(m.Name) == "" {
		return errors.New("member name cannot be empty")
	}
	if len(m.Name) > MaxNameLength {
		return errors.New("member name cannot exceed 100 characters")
	}
	if !strings.Contains(m.Email, "@") {
		return errors.New("member email must be valid")
	}
	if m.Program != ProgramAdults && m.Program != ProgramKids {
		return errors.New("program must be 'adults' or 'kids'")
	}
	if m.Status != StatusActive && m.Status != StatusInactive && m.Status != StatusArchived {
		return errors.New("status must be 'active', 'inactive', or 'archived'")
	}
	return nil
}

// IsActive returns true if the member is currently active.
// INVARIANT: Status field is not mutated
func (m *Member) IsActive() bool {
	return m.Status == StatusActive
}

// IsArchived returns true if the member is archived.
// INVARIANT: Status field is not mutated
func (m *Member) IsArchived() bool {
	return m.Status == StatusArchived
}

// Archive sets the member status to archived.
// PRE: Member is not already archived
// POST: Status is set to archived
func (m *Member) Archive() error {
	if m.Status == StatusArchived {
		return ErrAlreadyArchived
	}
	m.Status = StatusArchived
	return nil
}

// Restore sets the member status back to active.
// PRE: Member is currently archived
// POST: Status is set to active
func (m *Member) Restore() error {
	if m.Status != StatusArchived {
		return ErrNotArchived
	}
	m.Status = StatusActive
	return nil
}
