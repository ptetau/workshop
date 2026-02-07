package orchestrators

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"workshop/internal/domain/attendance"
	"workshop/internal/domain/member"
	"workshop/internal/domain/waiver"

	"github.com/google/uuid"
)

// GuestMemberStore defines the store interface needed by GuestCheckIn.
type GuestMemberStore interface {
	Save(ctx context.Context, m member.Member) error
}

// GuestWaiverStore defines the store interface needed by GuestCheckIn.
type GuestWaiverStore interface {
	Save(ctx context.Context, w waiver.Waiver) error
}

// GuestAttendanceStore defines the store interface needed by GuestCheckIn.
type GuestAttendanceStore interface {
	Save(ctx context.Context, a attendance.Attendance) error
}

// GuestCheckInInput carries input for the guest check-in flow.
type GuestCheckInInput struct {
	Name          string
	Email         string
	AcceptedTerms bool
	IPAddress     string
	ScheduleID    string
	ClassDate     string
}

// GuestCheckInDeps holds dependencies for GuestCheckIn.
type GuestCheckInDeps struct {
	MemberStore     GuestMemberStore
	WaiverStore     GuestWaiverStore
	AttendanceStore GuestAttendanceStore
}

// GuestCheckInResult holds the output of the guest check-in flow.
type GuestCheckInResult struct {
	MemberID     string
	WaiverID     string
	AttendanceID string
}

// ExecuteGuestCheckIn creates a guest member, signs a waiver, and checks them in.
// PRE: Name and Email must be non-empty; AcceptedTerms must be true
// POST: Guest member, waiver, and attendance records created
func ExecuteGuestCheckIn(ctx context.Context, input GuestCheckInInput, deps GuestCheckInDeps) (GuestCheckInResult, error) {
	if input.Name == "" {
		return GuestCheckInResult{}, errors.New("guest name is required")
	}
	if input.Email == "" {
		return GuestCheckInResult{}, errors.New("guest email is required")
	}
	if !input.AcceptedTerms {
		return GuestCheckInResult{}, errors.New("guest must accept waiver terms")
	}

	now := time.Now()
	memberID := uuid.New().String()
	waiverID := uuid.New().String()
	attendanceID := uuid.New().String()

	// Step 1: Create guest member
	guestMember := member.Member{
		ID:      memberID,
		Name:    input.Name,
		Email:   input.Email,
		Program: member.ProgramAdults,
		Status:  member.StatusActive,
	}
	if err := guestMember.Validate(); err != nil {
		return GuestCheckInResult{}, err
	}
	if err := deps.MemberStore.Save(ctx, guestMember); err != nil {
		return GuestCheckInResult{}, err
	}

	// Step 2: Sign waiver
	guestWaiver := waiver.Waiver{
		ID:            waiverID,
		MemberID:      memberID,
		AcceptedTerms: true,
		IPAddress:     input.IPAddress,
		SignedAt:       now,
	}
	if err := guestWaiver.Validate(); err != nil {
		return GuestCheckInResult{}, err
	}
	if err := deps.WaiverStore.Save(ctx, guestWaiver); err != nil {
		return GuestCheckInResult{}, err
	}

	// Step 3: Check in
	guestAttendance := attendance.Attendance{
		ID:          attendanceID,
		MemberID:    memberID,
		CheckInTime: now,
		ScheduleID:  input.ScheduleID,
		ClassDate:   input.ClassDate,
	}
	if err := guestAttendance.Validate(); err != nil {
		return GuestCheckInResult{}, err
	}
	if err := deps.AttendanceStore.Save(ctx, guestAttendance); err != nil {
		return GuestCheckInResult{}, err
	}

	slog.Info("guest_event", "event", "guest_checked_in", "member_id", memberID, "name", input.Name)

	return GuestCheckInResult{
		MemberID:     memberID,
		WaiverID:     waiverID,
		AttendanceID: attendanceID,
	}, nil
}
