package projections

import (
	"context"
	"time"

	"workshop/internal/adapters/storage/attendance"
	"workshop/internal/adapters/storage/injury"
	domainAttendance "workshop/internal/domain/attendance"
	domainClassType "workshop/internal/domain/classtype"
	domainInjury "workshop/internal/domain/injury"
	domainSchedule "workshop/internal/domain/schedule"
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
	Belt           string
	Stripe         int
	MatHours       float64
	ScheduleID     string
	ClassName      string
}

// GetAttendanceTodayResult carries the query result.
type GetAttendanceTodayResult struct {
	Attendees []AttendanceWithMember
}

// AttendanceTodayScheduleStore defines the schedule store interface for this projection.
type AttendanceTodayScheduleStore interface {
	GetByID(ctx context.Context, id string) (domainSchedule.Schedule, error)
}

// AttendanceTodayClassTypeStore defines the class type store interface for this projection.
type AttendanceTodayClassTypeStore interface {
	GetByID(ctx context.Context, id string) (domainClassType.ClassType, error)
}

// GetAttendanceTodayDeps holds dependencies for GetAttendanceToday.
type GetAttendanceTodayDeps struct {
	AttendanceStore    AttendanceStore
	MemberStore        MemberStore
	InjuryStore        InjuryStore
	GradingRecordStore GradingRecordStore            // optional: nil skips belt lookup
	ScheduleStore      AttendanceTodayScheduleStore  // optional: nil skips class name
	ClassTypeStore     AttendanceTodayClassTypeStore // optional: nil skips class name
}

// QueryGetAttendanceToday retrieves today's check-ins with injury indicators.
// PRE: Valid query parameters
// POST: Returns today's attendance with member details and injury flags
func QueryGetAttendanceToday(ctx context.Context, query GetAttendanceTodayQuery, deps GetAttendanceTodayDeps) (GetAttendanceTodayResult, error) {
	// Determine target date
	targetDate := time.Now().Truncate(24 * time.Hour)
	if query.Date != "" {
		if parsed, err := time.Parse("2006-01-02", query.Date); err == nil {
			targetDate = parsed.Truncate(24 * time.Hour)
		}
	}

	// Get all attendance records (filter by date in production)
	attendances, err := deps.AttendanceStore.List(ctx, attendance.ListFilter{
		Limit:  1000,
		Offset: 0,
	})
	if err != nil {
		return GetAttendanceTodayResult{}, err
	}

	// Filter to target date's records
	var todayAttendances []domainAttendance.Attendance
	for _, a := range attendances {
		if a.CheckInTime.Truncate(24 * time.Hour).Equal(targetDate) {
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
			MatHours:     a.MatHours,
			ScheduleID:   a.ScheduleID,
		}

		// Check for injury
		if inj, hasInjury := injuryMap[m.ID]; hasInjury {
			awm.HasInjury = true
			awm.InjuryBodyPart = inj.BodyPart
		}

		// Look up latest belt
		if deps.GradingRecordStore != nil {
			if records, err := deps.GradingRecordStore.ListByMemberID(ctx, m.ID); err == nil && len(records) > 0 {
				latest := records[0]
				for _, r := range records[1:] {
					if r.PromotedAt.After(latest.PromotedAt) {
						latest = r
					}
				}
				awm.Belt = latest.Belt
				awm.Stripe = latest.Stripe
			}
		}

		// Look up class name from schedule
		if a.ScheduleID != "" && deps.ScheduleStore != nil && deps.ClassTypeStore != nil {
			if sched, err := deps.ScheduleStore.GetByID(ctx, a.ScheduleID); err == nil {
				if ct, err := deps.ClassTypeStore.GetByID(ctx, sched.ClassTypeID); err == nil {
					awm.ClassName = ct.Name
				}
			}
		}

		result = append(result, awm)
	}

	return GetAttendanceTodayResult{Attendees: result}, nil
}
