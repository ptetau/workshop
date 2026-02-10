package milestone

import (
	"errors"
	"time"
)

// Domain errors for member milestones
var (
	ErrEmptyMemberID    = errors.New("member ID cannot be empty")
	ErrEmptyMilestoneID = errors.New("milestone ID cannot be empty")
)

// MemberMilestone represents a milestone earned by a specific member.
type MemberMilestone struct {
	ID          string
	MemberID    string
	MilestoneID string
	EarnedAt    time.Time
	Notified    bool // whether the member has been shown the congratulatory notification
}

// Validate checks if the MemberMilestone has valid data.
// PRE: MemberMilestone struct is populated
// POST: Returns nil if valid, error otherwise
func (m *MemberMilestone) Validate() error {
	if m.MemberID == "" {
		return ErrEmptyMemberID
	}
	if m.MilestoneID == "" {
		return ErrEmptyMilestoneID
	}
	return nil
}
