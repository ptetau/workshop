package projections

import (
	"context"
	"time"

	"workshop/internal/adapters/storage/attendance"
	"workshop/internal/adapters/storage/injury"
	domainAttendance "workshop/internal/domain/attendance"
	domainInjury "workshop/internal/domain/injury"
)

// GetAttendanceTodayQuery carries query parameters.
type GetAttendanceTodayQuery struct {
	Date string // Optional, defaults to today
}

// AttendanceWithMember represents attendance with member details.
type AttendanceWithMember struct {
	MemberID       string
	MemberName     string
	CheckInTime    time.Time
	CheckOutTime   time.Time
	HasInjury      bool
	InjuryBodyPart string
}

// GetAttendanceTodayResult carries the query result.
type GetAttendanceTodayResult struct {
	Attendees []AttendanceWithMember
}

// GetAttendanceTodayDeps holds dependencies for GetAttendanceToday.
type GetAttendanceTodayDeps struct {
	AttendanceStore AttendanceStore
	MemberStore     MemberStore
	InjuryStore     InjuryStore
}

// QueryGetAttendanceToday retrieves today's check-ins with injury indicators.
// PRE: Valid query parameters
// POST: Returns today's attendance with member details and injury flags
func QueryGetAttendanceToday(ctx context.Context, query GetAttendanceTodayQuery, deps GetAttendanceTodayDeps) (GetAttendanceTodayResult, error) {
	// Get all attendance records (filter by date in production)
	attendances, err := deps.AttendanceStore.List(ctx, attendance.ListFilter{
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		return GetAttendanceTodayResult{}, err
	}

	// Filter to today's records
	today := time.Now().Truncate(24 * time.Hour)
	var todayAttendances []domainAttendance.Attendance
	for _, a := range attendances {
		if a.CheckInTime.Truncate(24 * time.Hour).Equal(today) {
			todayAttendances = append(todayAttendances, a)
		}
	}

	// Get active injuries
	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)
	injuries, err := deps.InjuryStore.List(ctx, injury.ListFilter{
		Limit:  1000,
		Offset: 0,
	})
	if err != nil {
		return GetAttendanceTodayResult{}, err
	}

	injuryMap := make(map[string]domainInjury.Injury)
	for _, inj := range injuries {
		if inj.ReportedAt.After(sevenDaysAgo) {
			injuryMap[inj.MemberID] = inj
		}
	}

	// Build result with member details
	var result []AttendanceWithMember
	for _, a := range todayAttendances {
		// Get member details
		m, err := deps.MemberStore.GetByID(ctx, a.MemberID)
		if err != nil {
			continue // Skip if member not found
		}

		awm := AttendanceWithMember{
			MemberID:     m.ID,
			MemberName:   m.Name,
			CheckInTime:  a.CheckInTime,
			CheckOutTime: a.CheckOutTime,
		}

		// Check for injury
		if inj, hasInjury := injuryMap[m.ID]; hasInjury {
			awm.HasInjury = true
			awm.InjuryBodyPart = inj.BodyPart
		}

		result = append(result, awm)
	}

	return GetAttendanceTodayResult{Attendees: result}, nil
}
