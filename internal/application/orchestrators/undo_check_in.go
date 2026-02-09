package orchestrators

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"workshop/internal/domain/attendance"
)

// UndoCheckInStore defines the attendance store interface needed for undo.
type UndoCheckInStore interface {
	GetByID(ctx context.Context, id string) (attendance.Attendance, error)
	Delete(ctx context.Context, id string) error
}

// UndoCheckInInput carries input for the undo check-in orchestrator.
type UndoCheckInInput struct {
	AttendanceID string
}

// UndoCheckInDeps holds dependencies for UndoCheckIn.
type UndoCheckInDeps struct {
	AttendanceStore UndoCheckInStore
	Now             func() time.Time // injectable for testing
}

// ExecuteUndoCheckIn removes an attendance record (un-check-in).
// PRE: AttendanceID is non-empty and refers to an existing record
// POST: Attendance record is deleted
// INVARIANT: Only today's check-ins can be undone (#38)
func ExecuteUndoCheckIn(ctx context.Context, input UndoCheckInInput, deps UndoCheckInDeps) error {
	if input.AttendanceID == "" {
		return errors.New("attendance ID is required")
	}

	a, err := deps.AttendanceStore.GetByID(ctx, input.AttendanceID)
	if err != nil {
		return errors.New("attendance record not found")
	}

	// Enforce today-only constraint (US-2.5.2 #38)
	now := time.Now()
	if deps.Now != nil {
		now = deps.Now()
	}
	today := now.Format("2006-01-02")
	checkinDate := a.CheckInTime.Format("2006-01-02")
	if checkinDate != today {
		return errors.New("can only undo today's check-ins")
	}

	if err := deps.AttendanceStore.Delete(ctx, input.AttendanceID); err != nil {
		return err
	}

	slog.Info("checkin_event", "event", "member_unchecked_in", "attendance_id", input.AttendanceID, "member_id", a.MemberID)
	return nil
}
