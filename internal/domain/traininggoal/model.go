package traininggoal

import (
	"errors"
	"time"
)

// Domain errors
var (
	ErrEmptyMemberID = errors.New("member ID is required")
	ErrZeroTarget    = errors.New("target must be greater than zero")
	ErrInvalidPeriod = errors.New("period must be one of: weekly, monthly")
)

// Period constants
const (
	PeriodWeekly  = "weekly"
	PeriodMonthly = "monthly"
)

// TrainingGoal represents a member-set personal target (e.g., "3 sessions per week").
type TrainingGoal struct {
	ID        string
	MemberID  string
	Target    int    // e.g., 3
	Period    string // weekly, monthly
	CreatedAt time.Time
	Active    bool
}

// Validate checks if the TrainingGoal has valid data.
// PRE: TrainingGoal struct is populated
// POST: Returns nil if valid, error otherwise
func (g *TrainingGoal) Validate() error {
	if g.MemberID == "" {
		return ErrEmptyMemberID
	}
	if g.Target <= 0 {
		return ErrZeroTarget
	}
	if g.Period != PeriodWeekly && g.Period != PeriodMonthly {
		return ErrInvalidPeriod
	}
	return nil
}
