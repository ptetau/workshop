package projections

import (
	"context"
	"time"

	memberStore "workshop/internal/adapters/storage/member"
	"workshop/internal/domain/attendance"
	"workshop/internal/domain/member"
)

// InactiveMemberStore defines the member store interface needed by the inactive radar.
type InactiveMemberStore interface {
	List(ctx context.Context, filter memberStore.ListFilter) ([]member.Member, error)
}

// InactiveAttendanceStore defines the attendance store interface needed by the inactive radar.
type InactiveAttendanceStore interface {
	ListByMemberID(ctx context.Context, memberID string) ([]attendance.Attendance, error)
}

// GetInactiveMembersQuery carries input for the inactive radar projection.
type GetInactiveMembersQuery struct {
	DaysSinceLastCheckIn int // members inactive for at least this many days
}

// GetInactiveMembersDeps holds dependencies for the inactive radar.
type GetInactiveMembersDeps struct {
	MemberStore     InactiveMemberStore
	AttendanceStore InactiveAttendanceStore
}

// InactiveMemberResult represents a single inactive member.
type InactiveMemberResult struct {
	MemberID     string
	Name         string
	Email        string
	Program      string
	Status       string
	LastCheckIn  string // YYYY-MM-DD or "never"
	DaysInactive int
}

// QueryGetInactiveMembers returns members who haven't checked in for the specified number of days.
func QueryGetInactiveMembers(ctx context.Context, query GetInactiveMembersQuery, deps GetInactiveMembersDeps) ([]InactiveMemberResult, error) {
	if query.DaysSinceLastCheckIn <= 0 {
		query.DaysSinceLastCheckIn = 30
	}

	cutoff := time.Now().AddDate(0, 0, -query.DaysSinceLastCheckIn)

	// Get all non-archived members
	members, err := deps.MemberStore.List(ctx, memberStore.ListFilter{Limit: 10000})
	if err != nil {
		return nil, err
	}

	var results []InactiveMemberResult

	for _, m := range members {
		if m.IsArchived() {
			continue
		}

		records, err := deps.AttendanceStore.ListByMemberID(ctx, m.ID)
		if err != nil {
			return nil, err
		}

		if len(records) == 0 {
			// Never checked in
			results = append(results, InactiveMemberResult{
				MemberID:     m.ID,
				Name:         m.Name,
				Email:        m.Email,
				Program:      m.Program,
				Status:       m.Status,
				LastCheckIn:  "never",
				DaysInactive: -1,
			})
			continue
		}

		// Records are ordered DESC by check-in time from ListByMemberID
		lastCheckIn := records[0].CheckInTime
		if lastCheckIn.Before(cutoff) {
			days := int(time.Since(lastCheckIn).Hours() / 24)
			results = append(results, InactiveMemberResult{
				MemberID:     m.ID,
				Name:         m.Name,
				Email:        m.Email,
				Program:      m.Program,
				Status:       m.Status,
				LastCheckIn:  lastCheckIn.Format("2006-01-02"),
				DaysInactive: days,
			})
		}
	}

	return results, nil
}
