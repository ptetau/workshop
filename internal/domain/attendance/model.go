package attendance

import (
	"errors"
	"time"
)

// Attendance holds state for the concept.
type Attendance struct {
	ID           string
	CheckInTime  time.Time
	CheckOutTime time.Time
	MemberID     string
	ScheduleID   string
	ClassDate    string  // YYYY-MM-DD format
	MatHours     float64 // hours credited from session duration
}

// Validate checks if the Attendance has valid data.
// PRE: Attendance struct is initialized
// POST: Returns error if validation fails, nil otherwise
// INVARIANT: MemberID must not be empty, CheckInTime must be set
func (a *Attendance) Validate() error {
	if a.MemberID == "" {
		return errors.New("attendance must be associated with a member")
	}
	if a.CheckInTime.IsZero() {
		return errors.New("check-in time must be set")
	}
	if !a.CheckOutTime.IsZero() && a.CheckOutTime.Before(a.CheckInTime) {
		return errors.New("check-out time cannot be before check-in time")
	}
	return nil
}

// IsCheckedOut returns true if the member has checked out.
// PRE: Attendance is initialized
// POST: Returns boolean indicating check-out status
func (a *Attendance) IsCheckedOut() bool {
	return !a.CheckOutTime.IsZero()
}

// Duration returns the duration of the attendance session.
// PRE: Attendance is initialized with CheckInTime
// POST: Returns duration, or time since check-in if not checked out
func (a *Attendance) Duration() time.Duration {
	if a.IsCheckedOut() {
		return a.CheckOutTime.Sub(a.CheckInTime)
	}
	// Still checked in, return duration so far
	return time.Since(a.CheckInTime)
}
