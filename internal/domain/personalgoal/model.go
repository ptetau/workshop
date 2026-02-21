package personalgoal

import (
	"errors"
	"time"
)

// Domain errors
var (
	ErrEmptyTitle    = errors.New("title is required")
	ErrEmptyMemberID = errors.New("member ID is required")
	ErrZeroTarget    = errors.New("target must be greater than zero")
	ErrInvalidDates  = errors.New("end date must be after start date")
)

// PersonalGoal represents a member's personal training goal (e.g., "50 rear naked chokes in April").
type PersonalGoal struct {
	ID          string
	MemberID    string
	Title       string    // e.g., "50 rear naked strangles during April"
	Description string    // optional details
	Target      int       // e.g., 50
	Unit        string    // e.g., "submissions", "sessions", "hours", "techniques"
	StartDate   time.Time // period start
	EndDate     time.Time // period end
	Color       string    // hex color for calendar display (default: #F9B232)
	Progress    int       // current progress (0 initially)
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Validate checks if the PersonalGoal has valid data.
// PRE: PersonalGoal struct is populated
// POST: Returns nil if valid, error otherwise
func (g *PersonalGoal) Validate() error {
	if g.Title == "" {
		return ErrEmptyTitle
	}
	if g.MemberID == "" {
		return ErrEmptyMemberID
	}
	if g.Target <= 0 {
		return ErrZeroTarget
	}
	if !g.EndDate.After(g.StartDate) {
		return ErrInvalidDates
	}
	return nil
}

// SetDefaultColor assigns a default color if none is set.
// PRE: PersonalGoal struct is populated
// POST: Color field is set to default if empty
func (g *PersonalGoal) SetDefaultColor() {
	if g.Color == "" {
		g.Color = "#F9B232"
	}
}

// UpdateProgress updates the progress value.
// PRE: progress >= 0
// POST: Progress field is updated, UpdatedAt is set to now
func (g *PersonalGoal) UpdateProgress(progress int) {
	if progress < 0 {
		progress = 0
	}
	g.Progress = progress
	g.UpdatedAt = time.Now()
}

// IsActiveForDate checks if the goal is active on a given date.
// PRE: date is valid
// POST: returns true if date falls within StartDate and EndDate (inclusive)
func (g *PersonalGoal) IsActiveForDate(date time.Time) bool {
	dateStr := date.Format("2006-01-02")
	startStr := g.StartDate.Format("2006-01-02")
	endStr := g.EndDate.Format("2006-01-02")
	return dateStr >= startStr && dateStr <= endStr
}

// ProgressPercentage returns the completion percentage (0-100).
// PRE: Target > 0
// POST: returns percentage as integer (capped at 100)
func (g *PersonalGoal) ProgressPercentage() int {
	if g.Target <= 0 {
		return 0
	}
	pct := (g.Progress * 100) / g.Target
	if pct > 100 {
		return 100
	}
	return pct
}
