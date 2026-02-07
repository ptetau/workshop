package orchestrators

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"workshop/internal/domain/attendance"
	"workshop/internal/domain/member"

	"github.com/google/uuid"
)

// AttendanceStore defines the interface for attendance persistence.
type AttendanceStore interface {
	Save(ctx context.Context, a attendance.Attendance) error
}

// CheckInSearchStore defines the member store interface needed for name search.
type CheckInSearchStore interface {
	GetByID(ctx context.Context, id string) (member.Member, error)
	SearchByName(ctx context.Context, query string, limit int) ([]member.Member, error)
}

// SearchMembersInput carries input for name-based member search.
type SearchMembersInput struct {
	Query string
	Limit int
}

// SearchMembersResult carries the shortlist of matching members.
type SearchMembersResult struct {
	Members []member.Member
}

// SearchMembersDeps holds dependencies for SearchMembers.
type SearchMembersDeps struct {
	MemberStore CheckInSearchStore
}

// ExecuteSearchMembers performs a fuzzy name search and returns a shortlist
// of matching Active members for the check-in autocomplete.
// PRE: Query must be non-empty
// POST: Returns up to Limit matching active members
func ExecuteSearchMembers(ctx context.Context, input SearchMembersInput, deps SearchMembersDeps) (SearchMembersResult, error) {
	if input.Query == "" {
		return SearchMembersResult{Members: []member.Member{}}, nil
	}
	if input.Limit <= 0 {
		input.Limit = 10
	}

	members, err := deps.MemberStore.SearchByName(ctx, input.Query, input.Limit)
	if err != nil {
		return SearchMembersResult{}, err
	}
	if members == nil {
		members = []member.Member{}
	}

	return SearchMembersResult{Members: members}, nil
}

// CheckInMemberInput carries input for the check-in orchestrator.
// MemberID is obtained by the caller after the user selects from the
// name-search shortlist — never typed directly.
type CheckInMemberInput struct {
	MemberID   string // selected from search shortlist
	ScheduleID string // optional: which class they're checking into
	ClassDate  string // optional: date of the class (YYYY-MM-DD)
}

// CheckInMemberDeps holds dependencies for CheckInMember.
type CheckInMemberDeps struct {
	MemberStore     CheckInSearchStore
	AttendanceStore AttendanceStore
}

// ExecuteCheckInMember coordinates member check-in.
// PRE: MemberID is a valid member selected from the name-search shortlist
// POST: Attendance record created with CheckInTime=now
// INVARIANT: Cannot check in twice without checking out (enforced by UI/business logic)
func ExecuteCheckInMember(ctx context.Context, input CheckInMemberInput, deps CheckInMemberDeps) error {
	if input.MemberID == "" {
		return errors.New("member must be selected from the search results")
	}

	// Verify member exists and is active
	m, err := deps.MemberStore.GetByID(ctx, input.MemberID)
	if err != nil {
		return errors.New("member not found")
	}
	if m.IsArchived() {
		return errors.New("archived members cannot check in")
	}

	// Create attendance record
	a := attendance.Attendance{
		ID:          uuid.New().String(),
		MemberID:    input.MemberID,
		CheckInTime: time.Now(),
		ScheduleID:  input.ScheduleID,
		ClassDate:   input.ClassDate,
	}

	if err := a.Validate(); err != nil {
		return err
	}

	if err := deps.AttendanceStore.Save(ctx, a); err != nil {
		return err
	}

	slog.Info("checkin_event", "event", "member_checked_in", "member_id", input.MemberID, "name", m.Name, "schedule_id", input.ScheduleID)
	return nil
}
